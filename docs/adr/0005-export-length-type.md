---
status: accepted
date: 2026-06-10
review-runs:
  - 20260609T225354-33836cfefacf9a5e5d403c7d # ras consider — accept with revisions; revisions applied via ras fix --decisions, ras verify clean
---

# Type of the `length` parameter on `Session.Export`

## Status

**Accepted (2026-06-10).** This ADR captures a v1.0.0 public-API decision surfaced by external code review (item M3). Gated per the project's ADR policy: the `ras consider` run above returned accept-with-revisions; the revisions were applied via a maintainer-decided resolution pass (`ras fix --decisions`) and re-gated, with `ras verify` returning clean (unresolved: []). Evidence trail: PR #66 comments and DEV-JOURNAL cpace.S15. The `Session.Export` signature is settled for the v1.0.0 freeze.

## Context

In the current snapshot, `session.go:64-87` declares:

```go
func (s *Session) Export(label, context []byte, length int) ([]byte, error) {
    // ...
    if length < 0 || length > maxHKDFOutput {
        return nil, fmt.Errorf("%w: invalid export length", ErrInvalidInput)
    }
    // ...
}
```

with `maxHKDFOutput = 255 * 64 = 16320` at `session.go:9`.

Two observations from the code review:

1. **`length int` accepts negative values that must be checked at runtime.** Go's `int` is signed; negative arguments are forbidden at the function semantics level but expressible at the type level. The current code handles it with a runtime check. That check is also a panic barrier: `crypto/hkdf.Key` returns a normal error when `length` exceeds the HKDF maximum, but panics with `makeslice: cap out of range` for negative `length`. The range check in `Session.Export` is therefore the only barrier between caller input and a panic in the security library, and must not be removed in favor of delegating validation to the standard library.

2. **`length == 0` is currently accepted silently.** HKDF with `length == 0` returns a result with length zero. Whether this is a meaningful API operation or a likely caller bug is undecided.

The public API is frozen for v1.0.0 unless this review reopens it. Any change to the `Export` signature (parameter type, parameter list, return shape) after the freeze is breaking.

Three independent decisions are intertwined here:

- **D1.** Parameter type: `int` vs `uint32` vs `int` with runtime check.
- **D2.** Whether `length == 0` is a valid request or rejected.
- **D3.** Upper bound: keep `maxHKDFOutput = 16320` (255 * 64, the HKDF-SHA512 maximum) or pick a smaller cap.

Go convention is split on D1. `crypto/hkdf.Key` takes `length int`. `crypto/rand.Read(p []byte)` takes a slice (sidestepping the question). Many Go APIs use `int` for length with a runtime check; some (especially network protocols) use `uint16` / `uint32` to express non-negativity at the type level. `int` is conventional for in-memory buffer sizes; `uint32` is conventional for wire-protocol field sizes.

On D2, both interpretations are defensible. Some HKDF callers legitimately want to ask for zero bytes (e.g., parameterized derivation in higher-level constructs). Others would call this a bug (why would you ever ask the KDF for zero bytes?). The cost of allowing it is a corner case in tests; the cost of rejecting it is one more `if` branch the caller might hit.

On D3, the upper bound is the HKDF construction's natural maximum. Restricting further (e.g., to 1 KiB) would be defensive but would also surprise callers who legitimately need more than 1 KiB. The current bound is the cleanest.

## Decision

Keep `length int` and reject negative values at runtime; do **not** switch to `uint32`. Accept `length == 0` as valid (return a result with length zero; nil vs empty slice shape is not part of the v1 contract). Keep `maxHKDFOutput = 16320`.

Rationale:

- **`int` matches `crypto/hkdf.Key`'s signature** and most other Go APIs that accept buffer sizes. Switching to `uint32` would force callers to type-cast at every call site (`uint32(len(buf))`), which is friction for a marginal type-safety win on a parameter that is already bounded.
- **`length == 0` is well-defined for HKDF** and the natural HKDF behaviour. Rejecting it would catch the uninitialized-length misuse case, but it would diverge from the underlying `crypto/hkdf.Key` semantics and special-case generic KDF wrappers; a wrong non-zero length is the same uncatchable bug class, so the expanded doc comment is the chosen mitigation.
- **The range check in `Session.Export` is both semantic validation and panic protection.** `crypto/hkdf.Key` returns a normal error on over-maximum length but panics on negative length, so this check is the sole barrier between caller input and a panic in the security library.
- **The current upper bound is the HKDF construction maximum.** Any tighter bound is arbitrary; any looser bound is impossible.

The range check in `Session.Export` is the right shape. Add one table-driven `TestExportLengthBoundaries` covering `-1`, `0`, `1`, `maxHKDFOutput - 1`, `maxHKDFOutput`, and `maxHKDFOutput + 1`.

## Acceptance criteria

Multi-agent review concurrence on this ADR moves it proposed -> accepted (the decision is ratified at review time). The acceptance criteria below are implementation-verification gates: they bind the implementing change and must all be satisfied before v1.0.0 is tagged - not before this ADR is accepted.

- **`Session.Export` signature unchanged** from the current `func (s *Session) Export(label, context []byte, length int) ([]byte, error)`.
- **`length == 0` is documented as valid** in the doc comment on `Session.Export` in `session.go`, with text similar to: *"`length` must be in the range `[0, 16320]` (255 * 64, the HKDF-SHA512 maximum). A `length` of zero returns a result with length 0; callers must not distinguish nil from empty output for this case. Negative values and values exceeding the maximum are rejected with a wrapped `ErrInvalidInput`."*
- **New tests** added to `api_test.go`:
  - `TestExportLengthBoundaries` — table over `-1`, `0`, `1`, `maxHKDFOutput - 1`, `maxHKDFOutput`, `maxHKDFOutput + 1`.
  - Accepted rows, including `0` and `maxHKDFOutput`, must assert `err == nil` and `len(out) == length`. The zero row must not assert `out == nil` or `out != nil`; the v1 contract guarantees only the length, and callers must not distinguish nil from empty output for this case.
  - Rejected rows must assert an error and `errors.Is(err, ErrInvalidInput)`. The `-1` row also pins panic-freedom; any panic fails the test.
- **`CHANGELOG.md` updated** under Unreleased with: *"Pin Export length contract: documented as `[0, 16320]`, zero-length returns length 0."*

## Considered options

- **A — Keep `int`, accept `length == 0` (recommended).** Matches `crypto/hkdf.Key`, removes ambiguity via doc + tests, no API breakage.

- **B — Change to `uint32`.** Expresses non-negativity at the type level. Breaking change relative to any code already calling `Export`. Adds friction at every call site (`uint32(...)` casts). Limited type-safety win because negative literals are compile errors, but `uint32(someInt)` still type-checks when `someInt` holds a negative runtime value.

- **C — Keep `int`, reject `length == 0`.** Documents zero-length as a caller bug. Diverges from `crypto/hkdf.Key` semantics. Forces callers writing generic KDF wrappers to special-case the zero path.

- **D — Switch to `uint16`.** Maps to the maximum HKDF output (`uint16` can represent 65535, so 16320 fits). Strongest type-level expression but most surprising to callers used to `int` for buffer sizes.

- **E — Take a `[]byte` output buffer instead of `length`.** API shape `Export(label, context []byte, out []byte) error`. Avoids the integer-type question. Forces caller to pre-allocate; matches `io.Reader`-style. Larger refactor; not justified by review evidence.

## Consequences

- **Option A (recommended):**
  - Zero code change.
  - One table-driven `TestExportLengthBoundaries` pins the boundary contract.
  - Doc comment expanded with explicit range and zero-length semantics.
  - Accepting `length == 0` becomes part of the frozen v1 contract; rejecting it later is a breaking semantic change.
  - Residual risk remains: an accidental zero length, notably Go's zero value for an uninitialized `int`, silently yields an empty secret that HMAC and other variable-key-length primitives accept without error. This is accepted because a wrong non-zero length, such as `16` instead of `32`, is the same uncatchable bug class, and the expanded doc comment is the chosen mitigation.
  - Zero-length output contains no key material and must not be used where an application key is required.
  - This differs from the empty-`SessionID` rejection precedent: that default protects protocol security properties, while zero-length `Export` output is a local output-shape question.

- **Option B:**
  - Every existing call site must adopt `uint32` casts.
  - Type system catches negative literals (`Export(label, ctx, -1)`) at compile time. Does *not* catch `uint32(someInt)` where `someInt < 0` — the negative becomes a huge unsigned value caught only by the runtime range check.
  - Marginal real-world safety improvement; significant ergonomic cost.

- **Option C:**
  - Diverges from `crypto/hkdf.Key`.
  - Forces caller logic for zero-length cases.
  - Rejecting zero would catch the uninitialized-length misuse case described in Option A's residual-risk bullet, but the ADR judges the divergence cost higher because a wrong non-zero length is the same uncatchable bug class.

- **Option D:**
  - Surprising parameter type for an in-memory buffer size.
  - `uint16` has enough capacity: its maximum 65535 covers the current 16320 limit and even a hypothetical 256-byte hash limit of 65280. The rejection is ergonomic, not mathematical.
  - Forces casts at call sites and departs from Go convention for in-memory buffer sizes.

- **Option E:**
  - Different API shape entirely. Out of scope for a small fix.

## Implementation outline (Option A)

1. Expand the doc comment on `Session.Export` to state the accepted range and zero-length semantics.
2. Add `TestExportLengthBoundaries` to `api_test.go`.
3. No code change to `Session.Export` itself (the runtime check already handles all cases correctly).
4. Add a `CHANGELOG.md` Unreleased entry: "Pin Export length contract: documented as `[0, 16320]`, zero-length returns length 0."
