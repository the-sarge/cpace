---
status: proposed
---

# GC finalizer for `Session` to bound ISK lifetime on missed `Close`

## Status

**Proposed — recorded, not yet enforced.** This ADR captures a memory-handling decision surfaced by external code review (item H1.c). It stays `proposed` until an independent multi-agent review (`ras consider`) concurs. Acceptance is contingent on the review agreeing that adding a GC finalizer is the right trade-off given Go's finalizer semantics and the package's documented best-effort memory-cleanup policy.

## Context

`Initiator.Finish` (`api.go:206-239`) returns `(messageC []byte, session *Session, err error)`. A caller that ignores the returned `*Session` (e.g. `msgC, _, err := initiator.Finish(...)` or a panic between `Finish` and the conventional `defer session.Close()`) leaves a `*Session` that is reachable only through the runtime's reference graph until GC reclaims it. The `sessionState.isk` byte slice lives in the heap for that period.

`session.go:48-62` already implements the documented contract:

```go
func (s *Session) Close() error {
    if s == nil || s.state == nil {
        return fmt.Errorf("%w: nil session", ErrInvalidInput)
    }
    st := s.state
    st.mu.Lock()
    defer st.mu.Unlock()
    if st.closed {
        return nil
    }
    clearBytes(st.isk)
    st.isk = nil
    st.closed = true
    return nil
}
```

`Close` is idempotent and zeroizes ISK best-effort. The documented threat model in `docs/threat-model.md` and the `Session.Close` doc comment at `session.go:43-47` are honest about Go's lack of guaranteed memory zeroization. The remaining question is: when a caller forgets to call `Close`, should the runtime bound the ISK's lifetime by registering a finalizer that calls `Close()` on GC?

Today the answer is no. The ISK lives until the `*Session` is unreachable and GC reclaims its memory (no explicit zeroization). On runtimes with conservative GC or in caller bugs where the `*Session` is held by a long-lived structure (a connection-pool entry, a closure capture, a goroutine leak), the ISK may persist for the lifetime of the process.

Go's `runtime.SetFinalizer` (or the newer `runtime.AddCleanup` in Go 1.24+) provides a hook that runs when an object becomes unreachable, before its memory is reclaimed. This can be used to call `Close()` on garbage-collected sessions. The cost is:

- Finalizers run on a dedicated finalizer goroutine, not the goroutine that dropped the reference. They are best-effort and not guaranteed to run before process exit.
- Finalizers can resurrect objects (the object becomes unreachable, finalizer runs, finalizer's closure makes the object reachable again). Care is needed.
- Finalizers add GC pause-time cost proportional to the number of finalized objects.
- Finalizers create surprising lifetime behaviour: a `*Session` that captures a goroutine via a closure may keep the goroutine alive longer than expected.
- `runtime.AddCleanup` (Go 1.24+) addresses several of these issues (no resurrection, multiple cleanups per object, cleanup runs after the last reference is gone). `go.mod` requires Go 1.26, so `AddCleanup` is available.

The package's documented memory policy is **best-effort**. The threat model explicitly scopes out local memory disclosure adversaries. A finalizer does not strengthen the threat model; it bounds the *expected* ISK lifetime under caller bugs.

The Go cryptography ecosystem is split. `crypto/tls` does not register finalizers. `crypto/aes` and friends do not. Some third-party libraries (e.g., older versions of `golang.org/x/crypto/openpgp`) used finalizers; the consensus has trended away from them. The age-encryption.org/age library does not. The cited reasons include: caller responsibility, GC unpredictability, surprising goroutine-lifetime interactions, and the fact that Go's runtime makes guaranteed zeroization impossible anyway.

## Decision

Do **not** add a finalizer to `Session` in v1.0.0. The package's contract — caller must `Close` when done — is documented and conventional Go. Finalizer-based zeroization adds complexity (cleanup ordering, goroutine-lifetime surprises, AddCleanup-vs-SetFinalizer API choice) without strengthening the documented threat model, and goes against the prevailing Go-cryptography convention.

The package may revisit this in a future minor release (non-breaking — adding a finalizer/cleanup is observation-only for correctly-written callers) if real evidence emerges that downstream consumers are losing ISK bytes to caller bugs in deployed code.

This ADR rejects the option of adding a finalizer now on the grounds that:

1. The threat model already disclaims guaranteed zeroization, so a finalizer does not change the security posture, only the expected ISK heap-residency under buggy callers.
2. Go cryptography consensus is to leave finalizers off.
3. Adding `runtime.AddCleanup` calls creates non-trivial test-flakiness risk for tests that assert ISK-clearing behaviour (the cleanup is not deterministic).
4. Removing a finalizer later is a behaviour change that downstream consumers may have come to rely on; not adding one preserves more freedom.

## Acceptance criteria

The implementation outcome (no finalizer in v1.0.0) requires:

- **Doc clarification** in `doc.go` and `docs/integration-guidance.md` stating that callers MUST call `Close` (or `defer session.Close()`) when done with a `*Session`, and that the package does not automatically zeroize ISK on garbage collection.
- **Code search** of the test suite confirming no existing test relies on finalizer-based clearing (none do today).
- **Decision rationale recorded** so future contributors who notice the missed-Close lifetime gap do not re-litigate without new evidence.

## Considered options

- **A — No finalizer (recommended).** Documents caller responsibility, matches Go-cryptography consensus, preserves freedom to add later. Cost: a doc sentence.

- **B — `runtime.AddCleanup` zeroizing ISK.** Registers a Go 1.24+ cleanup at `Session` construction (in `newSession`, `api.go:344`) that calls `clearBytes(state.isk)` if the cleanup runs. Cost: ~5 LoC, plus careful test design to avoid relying on cleanup determinism.

- **C — `runtime.SetFinalizer`.** The older finalizer API. Has resurrection semantics and only one finalizer per object. `AddCleanup` is strictly better when available. Considered only for completeness; reject in favor of B if a finalizer is desired.

- **D — Document caller responsibility but ship a `MustClose` helper.** Provide a helper `MustClose(s *Session)` that panics if `s == nil`. Marginal value; encourages a different anti-pattern (panic in production).

## Consequences

- **Option A (recommended):**
  - Zero implementation cost.
  - Caller bugs continue to leave ISK readable in heap until GC reclaims memory unzeroized.
  - The package contract is conventional Go; reviewers and downstream consumers should not be surprised.

- **Option B:**
  - Bounds expected ISK heap-residency to GC pause cycles.
  - Adds non-deterministic ordering to ISK zeroization that may complicate future memory-handling tests.
  - Locks the runtime API choice (`AddCleanup`) for the lifetime of v1.x; `go.mod` would need to commit to Go 1.24+ permanently.
  - Removing the cleanup later is a behaviour change.
  - Sets a precedent that other secret-holding types (none currently exist beyond `Session`) would need similar treatment.

- **Option C:**
  - Strictly inferior to B given `AddCleanup` is available.

- **Option D:**
  - Encourages caller panics. Generally a Go anti-pattern.

## Notes for the reviewer

The decision in this ADR is *not* to ship a finalizer. The structure mirrors ADR-0001 — propose with full alternatives so the multi-agent review can dissent if there is a real case for option B. If the review concurs, the implementation work is purely doc: one paragraph in `doc.go`, one in `docs/integration-guidance.md`. If the review dissents and prefers option B, the implementation work is the `runtime.AddCleanup` call plus a new test asserting that cleanup eventually runs after dropping the last reference (using `runtime.GC()` and a settle window).
