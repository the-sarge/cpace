---
status: proposed
---

# Type of the `length` parameter on `Session.Export`

## Status

**Proposed — recorded, not yet enforced.** This ADR captures a v1.0.0 public-API decision surfaced by external code review (item M3). It stays `proposed` until an independent multi-agent review (`ras consider`) concurs. Once accepted, the `Session.Export` signature is settled for the v1.0.0 freeze.

## Context

`session.go:64-87` declares:

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

1. **`length int` accepts negative values that must be checked at runtime.** Go's `int` is signed; negative arguments are forbidden at the function semantics level but expressible at the type level. The current code handles it with a runtime check.

2. **`length == 0` is currently accepted silently.** HKDF with `length == 0` returns a zero-byte slice. Whether this is a meaningful API operation or a likely caller bug is undecided.

The public API is frozen for v1.0.0 unless this review reopens it. Any change to the `Export` signature (parameter type, parameter list, return shape) after the freeze is breaking.

Three independent decisions are intertwined here:

- **D1.** Parameter type: `int` vs `uint32` vs `int` with runtime check.
- **D2.** Whether `length == 0` is a valid request or rejected.
- **D3.** Upper bound: keep `maxHKDFOutput = 16320` (255 × 64, the HKDF-SHA512 maximum) or pick a smaller cap.

Go convention is split on D1. `crypto/hkdf.Key` takes `length int`. `crypto/rand.Read(p []byte)` takes a slice (sidestepping the question). Many Go APIs use `int` for length with a runtime check; some (especially network protocols) use `uint16` / `uint32` to express non-negativity at the type level. `int` is conventional for in-memory buffer sizes; `uint32` is conventional for wire-protocol field sizes.

On D2, both interpretations are defensible. Some HKDF callers legitimately want to ask for zero bytes (e.g., parameterized derivation in higher-level constructs). Others would call this a bug (why would you ever ask the KDF for zero bytes?). The cost of allowing it is a corner case in tests; the cost of rejecting it is one more `if` branch the caller might hit.

On D3, the upper bound is the HKDF construction's natural maximum. Restricting further (e.g., to 1 KiB) would be defensive but would also surprise callers who legitimately need more than 1 KiB. The current bound is the cleanest.

## Decision

Keep `length int` and reject negative values at runtime; do **not** switch to `uint32`. Accept `length == 0` as valid (return a zero-byte slice). Keep `maxHKDFOutput = 16320`.

Rationale:

- **`int` matches `crypto/hkdf.Key`'s signature** and most other Go APIs that accept buffer sizes. Switching to `uint32` would force callers to type-cast at every call site (`uint32(len(buf))`), which is friction for a marginal type-safety win on a parameter that is already bounded.
- **`length == 0` is well-defined for HKDF** and the natural HKDF behaviour. Rejecting it would diverge from the underlying `crypto/hkdf.Key` semantics for no operational benefit.
- **The current upper bound is the HKDF construction maximum.** Any tighter bound is arbitrary; any looser bound is impossible.

The runtime check at `session.go:78-80` is the right shape. Add a test for the `length == 0` case to document the contract and one for the `length == maxHKDFOutput` upper-boundary case.

## Acceptance criteria

The implementation must satisfy these before this ADR moves `proposed → accepted` *and* before v1.0.0 is tagged:

- **`Session.Export` signature unchanged** from the current `func (s *Session) Export(label, context []byte, length int) ([]byte, error)`.
- **`length == 0` is documented as valid** in the `Export` doc comment at `session.go:64-67`, with text similar to: *"`length` must be in the range `[0, 255 × 64]`. A `length` of zero returns a zero-byte slice. Negative values and values exceeding the maximum are rejected with a wrapped `ErrInvalidInput`."*
- **New tests** added to `api_test.go`:
  - `TestExportLengthBoundaries` — table over `-1`, `0`, `1`, `maxHKDFOutput - 1`, `maxHKDFOutput`, `maxHKDFOutput + 1` asserting the correct accept/reject outcome.
  - The negative-length and over-maximum cases must assert both the error and `errors.Is(err, ErrInvalidInput)`.

## Considered options

- **A — Keep `int`, accept `length == 0` (recommended).** Matches `crypto/hkdf.Key`, removes ambiguity via doc + tests, no API breakage.

- **B — Change to `uint32`.** Expresses non-negativity at the type level. Breaking change relative to any code already calling `Export`. Adds friction at every call site (`uint32(...)` casts). Limited type-safety win because `uint32(-1)` still type-checks if the caller is careless.

- **C — Keep `int`, reject `length == 0`.** Documents zero-length as a caller bug. Diverges from `crypto/hkdf.Key` semantics. Forces callers writing generic KDF wrappers to special-case the zero path.

- **D — Switch to `uint16`.** Maps to the maximum HKDF output (16320 fits in `uint16`). Strongest type-level expression but most surprising to callers used to `int` for buffer sizes.

- **E — Take a `[]byte` output buffer instead of `length`.** API shape `Export(label, context []byte, out []byte) error`. Avoids the integer-type question. Forces caller to pre-allocate; matches `io.Reader`-style. Larger refactor; not justified by review evidence.

## Consequences

- **Option A (recommended):**
  - Zero code change.
  - Two new tests pin the contract.
  - Doc comment expanded with explicit range and zero-length semantics.

- **Option B:**
  - Every existing call site must adopt `uint32` casts.
  - Type system catches negative literals (`Export(label, ctx, -1)`) at compile time. Does *not* catch `uint32(someInt)` where `someInt < 0` — the negative becomes a huge unsigned value caught only by the runtime range check.
  - Marginal real-world safety improvement; significant ergonomic cost.

- **Option C:**
  - Diverges from `crypto/hkdf.Key`.
  - Forces caller logic for zero-length cases.
  - No operational benefit identified.

- **Option D:**
  - Surprising parameter type for an in-memory buffer size.
  - Pinning to `uint16` couples the API to the SHA-512 HKDF max; switching to a larger hash later would require a breaking change.

- **Option E:**
  - Different API shape entirely. Out of scope for a small fix.

## Implementation outline (Option A)

1. Expand the doc comment on `Session.Export` to state the accepted range and zero-length semantics.
2. Add `TestExportLengthBoundaries` to `api_test.go`.
3. No code change to `Session.Export` itself (the runtime check already handles all cases correctly).
4. Add a `CHANGELOG.md` Unreleased entry: "Pin Export length contract: documented as `[0, 255×64]`, zero-length returns an empty slice."
