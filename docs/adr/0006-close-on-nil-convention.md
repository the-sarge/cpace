---
status: proposed
---

# `Session.Close` on nil pointer receivers - tolerant no-op or strict error

## Status

**Proposed - recorded, not yet enforced.** This ADR captures a v1.0.0 public-API contract decision surfaced by external code review (item M2). It stays `proposed` until an independent multi-agent review (`ras consider`) concurs. Once accepted, the contract for nil-receiver methods on `Session` is settled for the v1.0.0 freeze.

## Context

`session.go:48-62` implements:

```go
func (s *Session) Close() error {
    if s == nil || s.state == nil {
        return fmt.Errorf("%w: nil session", ErrInvalidInput)
    }
    // ...
}
```

while the three accessor methods (`TranscriptID`, `PeerAssociatedData`, `PeerID`) at `session.go:15-41` return `nil` on a nil receiver. `Export` returns `ErrInvalidInput` on a nil receiver, matching `Close`. The same guard also covers a non-nil receiver whose `state` is nil, such as `&Session{}` or `new(Session)`.

The nil-pointer receiver contract is asymmetric:

- `Close()` on a nil pointer -> wrapped `ErrInvalidInput`.
- `TranscriptID()` / `PeerAssociatedData()` / `PeerID()` on a nil pointer -> `nil`.
- `Export()` on a nil pointer -> wrapped `ErrInvalidInput`.

The current in-tree `api_test.go` `TestNilSessionClose` covers `Close`, `PeerAssociatedData`, and `PeerID` only. It does not assert the nil-receiver `TranscriptID` or `Export` behavior. The implementing commit must carry its own in-tree coverage for the full contract; unmerged review branches or pull requests are not load-bearing evidence for this ADR.

The surveyed Go standard-library behavior does not establish a nil-receiver `Close() -> nil` precedent:

- `*sync.Pool` - has no Close.
- `*net.TCPConn` - `Close()` on a typed nil receiver panics.
- `*os.File` - `(*os.File).Close()` on a nil receiver returns `os.ErrInvalid`; this is direct standard-library precedent for Option B's strict sentinel-error shape.
- `*sql.DB` - `(*sql.DB).Close()` on nil panics.
- `*bytes.Buffer` - doesn't implement Close.

The permanent record evaluates the four options in **Considered options** below, using labels A-D.

The current contract trades on safety: a caller running `defer session.Close()` on a `nil *Session` (e.g., after a failed `Finish` that returned `(nil, nil, err)`) gets a clean error rather than a panic. The cost is an ignored deferred error and the asymmetry with the accessors.

A common downstream pattern is `defer session.Close()` after a `Finish` that may have failed. Today's behaviour means `defer (*Session)(nil).Close()` returns an error that the caller often ignores because it is deferred, which is a no-op in practice. Switching only the nil-pointer case to `nil` makes that explicit: `defer session.Close()` after a failed `Finish` is a no-op. A failed `Finish` produces a nil pointer, not a caller-constructed zero-value `Session`.

## Decision

Switch `Session.Close()` on a nil `*Session` receiver to return `nil` (no-op). Keep `Close` strict for non-nil receivers whose `state` is nil, including `&Session{}` and `new(Session)`. Keep `Export` on nil and zero-value receivers returning `ErrInvalidInput`. Keep accessors returning `nil` on nil receivers.

The implementation must use the split guard shape:

```go
if s == nil {
    return nil
}
if s.state == nil {
    return fmt.Errorf("%w: nil session", ErrInvalidInput)
}
```

Rationale:

- **`defer session.Close()` after a failed `Finish` should be a no-op, not an error-emitter.** This common caller pattern only produces nil pointers, and the deferred error return is usually ignored.
- **The package still chooses sentinel errors over panics for invalid input.** Option B has direct `os.File` precedent and would be a defensible strict-error contract, but the failed-`Finish` defer path is more relevant to this API than preserving a returned error that deferred cleanup commonly discards.
- **`Export` on nil should remain an error** because callers MUST receive a return value, and `nil, nil` would be a silent "no key, no error" signal that masks a programmer bug. `Export` is not idempotent or terminal in the way `Close` is.
- **Accessors returning `nil`** is correct: they are documented as returning a copy of stored metadata; on nil receiver there is no metadata, so returning `nil` is unambiguous.
- **Zero-value `Session{}` stays strict.** A caller-constructed `&Session{}` is always a construction bug, not a failed-`Finish` cleanup case. Returning `ErrInvalidInput` matches `Export` on the same value, keeps zero-value behavior byte-identical to today at the v1.0.0 freeze, and preserves a reachable error path so `Close`'s error return is not vestigial.

## Acceptance criteria

The implementation must satisfy these before this ADR moves `proposed -> accepted` *and* before v1.0.0 is tagged:

- **`Session.Close()` on a nil `*Session` receiver returns `nil`.**
- **`Session.Close()` on `&Session{}` or `new(Session)` returns an error matching `ErrInvalidInput`.**
- **`Export` on nil and zero-value receivers returns an error matching `ErrInvalidInput`.**
- **`TranscriptID`, `PeerAssociatedData`, and `PeerID` on a nil receiver continue to return `nil`.**
- **One in-tree test at the implementing commit** (for example, `TestNilReceiverMethods`) asserts the full receiver matrix: `(*Session)(nil).Close() == nil`; nil-receiver `Export` returns an error matching `ErrInvalidInput`; `TranscriptID`, `PeerAssociatedData`, and `PeerID` on nil return `nil`; `(&Session{}).Close()` returns an error matching `ErrInvalidInput`; and `(&Session{}).Export(...)` returns an error matching `ErrInvalidInput`.
- **Updated doc** on `Session.Close`: *"Close is idempotent and nil-safe; calling Close on a nil `*Session` returns nil."*
- **`CHANGELOG.md` under `## Unreleased`** records a bullet explicitly labeled as a pre-v1 contract/behavior change. Note: this is technically a breaking change for any caller that did `if err := s.Close(); errors.Is(err, ErrInvalidInput) { ... }` to detect a nil receiver. Such code is unusual but possible; the changelog should call this out explicitly.

## Considered options

- **A - Nil-pointer-tolerant Close, zero-value-strict Close, strict Export, nil accessors (recommended).** Optimizes failed-`Finish` deferred cleanup while preserving sentinel errors for constructed invalid sessions and `Export`.

- **B - Status quo.** `Close` returns `ErrInvalidInput` on nil pointers and zero-value receivers. Consistent within the package's existing strict-error policy and directly precedented by `(*os.File)(nil).Close()` returning `os.ErrInvalid`.

- **C - All methods panic on nil.** Matches `*sql.DB` and typed-nil `*net.TCPConn` behavior for `Close`. Catches programmer bugs loudly. Diverges from the package's no-panic-on-input policy.

- **D - All methods nil-tolerant.** `Export` on nil returns `nil, nil`. Aligns with accessors, but `nil, nil` is the canonical silent-failure shape, rejected for the same reason `scalarMultVFY` returning `clone(identityEncoding)` on failure is being changed in [[0003-peer-share-error-semantics]].

## Consequences

- **Option A (recommended):**
  - `defer session.Close()` after a failed `Finish` becomes a true no-op.
  - The nil-pointer `Close` contract changes from wrapped `ErrInvalidInput` to `nil`, which is a breaking change for callers detecting nil receivers with `errors.Is(err, ErrInvalidInput)`.
  - `Export` continues to surface programmer bugs, including nil and zero-value receivers.
  - Zero-value and nil-state `Session` receivers remain strict for `Close`, so `Close` retains a reachable non-nil error path and the `error` return is not vestigial.
  - Security impact is neutral: a nil receiver holds no ISK, so nil tolerance cannot mask a failed zeroization.
  - One CHANGELOG entry flagging the minor contract change.
  - Updated doc + test.

- **Option B (status quo):**
  - No behavior change. Keeps direct `os.File`-style strict sentinel-error precedent and continues to require the doc + test to remain accurate.

- **Option C:**
  - All nil-receiver method calls panic, including in `defer` blocks. `defer session.Close()` after a failed `Finish` panics during cleanup - likely *during* another error path - which is exactly the situation panics are worst at handling.
  - Diverges from the package's no-panic-on-input policy.

- **Option D:**
  - Reintroduces silent-failure shape for `Export`. Rejected.

## Implementation outline (Option A)

1. Change `session.go:48-51` from the combined `s == nil || s.state == nil` guard to the exact split guard: `if s == nil { return nil }` followed by `if s.state == nil { return fmt.Errorf("%w: nil session", ErrInvalidInput) }`.
2. Update the `Session.Close` doc comment at `session.go:43-47` to state nil-tolerance explicitly.
3. Update the existing `TestNilSessionClose` test or add a new `TestNilReceiverMethods` test to assert the full receiver matrix listed in the acceptance criteria, with coverage present in-tree at the implementing commit.
4. Add a bullet under `## Unreleased`, explicitly labeled as a pre-v1 contract/behavior change.
5. Audit `examples_test.go` and ADR-0004's defer-`Close` reference for consistency with the final contract. Under the split guard, error-checking `Close` in examples remains valid; the audit confirms there is no dead error handling.
6. Confirm `Export` and accessors are unchanged.
