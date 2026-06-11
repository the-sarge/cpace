---
status: accepted
date: 2026-06-10
review-runs:
  - 20260609T224053-605bb721c00f68fec9eb9544 # ras consider — accept with revisions; revisions applied via ras fix --decisions, ras verify clean except the deferred [[0001]] link
---

# GC finalizer for `Session` to bound ISK lifetime on missed `Close`

## Status

**Accepted (2026-06-10).** This ADR captures a memory-handling decision surfaced by external code review (item H1.c). Gated per the project's ADR policy: the `ras consider` run above returned accept-with-revisions; the revisions were applied via a maintainer-decided resolution pass (`ras fix --decisions`) and re-gated, with `ras verify` returning clean except the deliberately deferred `[[0001-extract-cpace-core]]` cross-link — which resolved when ADR-0001 merged to main (PR #69, `e44436c`). Evidence trail: PR #66 comments and DEV-JOURNAL cpace.S15. The review concurred that the decision below is the right trade-off given Go's finalizer semantics and the package's documented best-effort memory-cleanup policy.

## Context

`Initiator.Finish` (`api.go:206-239`) returns `(messageC []byte, session *Session, err error)`. A caller that ignores the returned `*Session` (e.g. `msgC, _, err := initiator.Finish(...)` or a panic between `Finish` and the conventional `defer session.Close()`) leaves a `*Session` object that is unreachable to caller code but may remain allocated until GC discovers it and reclaims its memory. The `sessionState.isk` byte slice lives in the heap for that period.

`session.go:48-62` already implements the documented contract. The locking and zeroizing body is:

```go
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
```

`Close` is idempotent and zeroizes ISK best-effort. The nil-receiver `Close` contract is under separate consideration in [[0006-close-on-nil-convention]]; this finalizer question is independent because a cleanup only runs against a non-nil object that has become unreachable. The documented threat model in `docs/threat-model.md` and the `Session.Close` doc comment at `session.go:43-47` are honest about Go's lack of guaranteed memory zeroization. The remaining question is: when a caller forgets to call `Close`, should the runtime bound the ISK's lifetime by registering a cleanup that zeroizes key material after the session state becomes unreachable?

Today the answer is no. There are two separate lifetime gaps:

- In the unreachable-but-uncollected case, the caller has dropped the session and the runtime has not yet collected it. A cleanup could bound this window to eventual GC and cleanup execution, subject to Go's best-effort cleanup semantics.
- In reachable leak cases, such as a `*Session` held by a live connection-pool entry, closure capture, or leaked goroutine, the session is still reachable. A finalizer or cleanup never runs for those cases, so Option B does nothing for exactly the long-lived-reference bugs that motivate the review concern.

Go's mainline heap GC is precise; this ADR does not rely on conservative-GC behavior. Go's `runtime.SetFinalizer` (or the newer `runtime.AddCleanup` in Go 1.24+) provides a hook that runs after an object becomes unreachable, before its memory is reclaimed. This can be used to run a key-zeroization cleanup for garbage-collected session state. The cost is:

- Finalizers run on a dedicated finalizer goroutine, not the goroutine that dropped the reference. They are best-effort and not guaranteed to run before process exit.
- Finalizers can resurrect objects (the object becomes unreachable, finalizer runs, finalizer's closure makes the object reachable again). Care is needed.
- Finalizers and cleanups add GC overhead; objects with cleanups require additional GC cycles to free.
- Cleanup closures can accidentally retain the target. With `runtime.AddCleanup`, a closure or cleanup argument that keeps the target reachable prevents the cleanup from ever running.
- `runtime.AddCleanup` (Go 1.24+) addresses several of these issues (no resurrection, multiple cleanups per object, cleanup runs after the last reference is gone). `go.mod` requires Go 1.26, so `AddCleanup` is available.

The package's documented memory policy is **best-effort**. The threat model explicitly scopes out local memory disclosure adversaries. A finalizer does not strengthen the threat model; it bounds the *expected* ISK lifetime under caller bugs.

The Go cryptography ecosystem largely avoids finalizers for key-material cleanup. `crypto/tls` does not register finalizers for session secrets. `crypto/aes` and related primitives do not. The age-encryption.org/age library does not. `awnumar/memguard` is a verifiable pro-finalizer outlier: it registers finalizers on secret containers. `os.File` uses cleanup machinery for file-descriptor release, but that is resource release rather than secret zeroization. The Go-crypto consensus remains caller responsibility, GC unpredictability, and the fact that Go's runtime makes guaranteed zeroization impossible anyway.

## Decision

Do **not** add a finalizer or `runtime.AddCleanup` cleanup to `Session` in v1.0.0. The package's contract — caller must `Close` when done — is documented and conventional Go. Finalizer-based zeroization adds complexity (cleanup ordering, reachable-object limitations, cleanup-target selection, and AddCleanup-vs-SetFinalizer API choice) without strengthening the documented threat model, and goes against the prevailing Go-cryptography convention.

The package may revisit this in a future minor release because adding a finalizer or cleanup is observation-only for correctly written callers and therefore non-breaking. Re-opening this decision requires one of two concrete triggers: a credible report or CVE demonstrating recovery of key material from GC-uncollected memory in a deployed Go service, or the Go standard library itself adopting `runtime.AddCleanup` for key-material zeroization.

This ADR rejects the option of adding a finalizer now on the grounds that:

1. The threat model already disclaims guaranteed zeroization, so a finalizer does not change the security posture, only the expected ISK heap-residency under buggy callers.
2. Go cryptography consensus is to leave finalizers off.
3. A cleanup does not help reachable leaks such as live pool entries, closure captures, or leaked goroutines; those sessions are never finalized.
4. A wrapper-attached cleanup is unsafe: Go may collect a receiver mid-method after its last use, so a cleanup attached to `*Session` can fire concurrently with an in-flight `Export`. An unlocked `clearBytes(state.isk)` would race with `Export`, while a lock-taking `Close` can spuriously produce `ErrSessionClosed`. Avoiding that requires attaching to `sessionState` or adding `runtime.KeepAlive` to all `Session` methods, which is materially more than a small cleanup hook.
5. Adding `runtime.AddCleanup` calls creates non-trivial test-flakiness risk for tests that assert ISK-clearing behaviour (the cleanup is not deterministic).
6. Removing a finalizer later is a behaviour change that downstream consumers may have come to rely on; not adding one preserves more freedom.

## Acceptance criteria

The implementation outcome (no finalizer in v1.0.0) requires:

- **Doc clarification** by extending the `Session.Close` paragraph in `doc.go` to state that callers MUST call `Close` when done with a `*Session`, that the package performs no zeroization on garbage collection, and that the paragraph cross-references this ADR.
- **Integration guidance** by adding a session-lifecycle subsection to `docs/integration-guidance.md` that tells callers to `defer session.Close()` immediately after each successful session creation, explains that copied `Session` values share close state, recommends conventional closer-leak static analysis or review checklists, and repeats that the package performs no zeroization on GC.
- **README quickstart correction** requiring the quickstart example to close both sessions with `defer initSession.Close()` and `defer respSession.Close()`.
- **Option A implementation outline**: keep the explicit `Close` path as the only zeroization path; do not add `runtime.SetFinalizer` or `runtime.AddCleanup` anywhere in package code. The acceptance PR records code-search evidence for the absence of those APIs rather than embedding shell invocations in this ADR.
- **Verification artifact**: `doc.go` cross-references this ADR and `task docs:check` passes.

## Considered options

- **A — No finalizer (recommended).** Documents caller responsibility, matches Go-cryptography consensus, preserves freedom to add later, and keeps closer-leak detection available to downstreams through conventional linting and code review. Implementation outline: keep `newSession` and `Session.Close` free of `runtime.SetFinalizer` and `runtime.AddCleanup`; record package-code search evidence in the acceptance PR. Cost: documentation updates.

- **B — `runtime.AddCleanup` zeroizing ISK.** A sound design would register a Go 1.24+ cleanup on the inner `sessionState`, not the outer `*Session` wrapper, because `Session` values are copyable and copies share the same state. The cleanup callback would receive the `isk` slice as its argument and call a package-private helper to zeroize that slice; the closure must not capture the cleanup target, because a target kept reachable by the cleanup closure or argument never reaches cleanup. This design needs regression coverage proving `copied := *s` remains exportable after `s` is dropped and GC is forced. It also needs deterministic tests of the package-private cleanup helper invoked directly; cleanup timing itself must not be asserted in CI beyond API-lifetime invariants, with at most a non-CI-gating eventually-runs check. Cost: helper, registration, regression tests, and lifetime review; substantially more than a small hook.

- **C — `runtime.SetFinalizer`.** The older finalizer API. Has resurrection semantics and only one finalizer per object. `AddCleanup` is strictly better when available. Considered only for completeness; reject in favor of B if a finalizer is desired.

- **D — Closer-leak static analysis and review guidance.** Do not add runtime cleanup; instead, document that `Session.Close() error` is `io.Closer`-shaped and should be covered by conventional closer-leak linting, code review checklists, and integration tests. This is the real non-finalizer mitigation for missed `Close` calls and belongs in `docs/integration-guidance.md`.

## Consequences

- **Option A (recommended):**
  - Zero implementation cost beyond documentation.
  - Caller bugs continue to leave ISK readable in heap until GC reclaims memory unzeroized; the package performs no zeroization on GC.
  - Reachable leaks remain reachable and would not be helped by a finalizer or cleanup.
  - The package contract is conventional Go; reviewers and downstream consumers should not be surprised.
  - Lint-based missed-`Close` detection remains available to downstreams because `Close() error` follows the `io.Closer` shape.
  - Adding a cleanup later is non-breaking for correctly written callers; removing a cleanup later would be an observable behavior change, so leaving it out now preserves v1.x flexibility.

- **Option B:**
  - Bounds expected ISK heap-residency only for unreachable-but-uncollected sessions; it does not help reachable leaks.
  - Adds non-deterministic ordering to ISK zeroization that may complicate future memory-handling tests.
  - Adds GC overhead (objects with cleanups require additional GC cycles to free).
  - Locks the runtime API choice (`AddCleanup`) for the lifetime of v1.x; pins the v1.x minimum Go version at >=1.24 (already satisfied by the current go 1.26 requirement).
  - Removing the cleanup later is a behaviour change.
  - Sets a precedent that other secret-holding types (none currently exist beyond `Session`) would need similar treatment.
  - Requires the `sessionState` attachment, no target-capturing cleanup closure, and copy-lifetime regression coverage. Wrapper-attached designs require `runtime.KeepAlive` across `Session` methods or risk races and spurious close errors, which strengthens the case for Option A.

- **Option C:**
  - Strictly inferior to B given `AddCleanup` is available.

- **Option D:**
  - Improves detection of missed `Close` calls without runtime cleanup semantics.
  - Does not bound ISK heap residency by itself and depends on downstream linting or review discipline.

## Notes for the reviewer

The decision in this ADR is *not* to ship a finalizer. The structure mirrors ADR-0001 — propose with full alternatives so the multi-agent review can dissent if there is a real case for option B. If the review concurs, the implementation work is documentation plus acceptance-PR evidence that package code contains no finalizer or cleanup registration. If the review dissents and prefers option B, the implementation work is the sound `runtime.AddCleanup` design described above, deterministic direct tests for the cleanup helper, and copy-lifetime regression coverage; CI must not assert cleanup timing beyond API-lifetime invariants.
