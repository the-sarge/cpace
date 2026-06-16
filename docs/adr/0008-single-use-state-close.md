---
status: accepted
date: 2026-06-16
review-runs:
  - 20260616T040116-7edbef4428d20850e0094ce1 # ras consider round 1 - required copy and lifecycle precision fixes
  - 20260616T041535-4f06c05a3b6dc7d3f4d7b388 # ras consider round 2 - required guard and loser-path precision fixes
  - 20260616T041535-4f06c05a3b6dc7d3f4d7b388-verification-1781583998570028000 # ras verify - unresolved: []
---

# Single-use state cleanup for abandoned `Initiator` and `Responder`

## Status

**Accepted (2026-06-16) - architecture and public-lifecycle narrow thaw.** This ADR records a narrow reopen of the release-readiness public-surface freeze to address the residual lifecycle risk already named by [[0001-extract-cpace-core]]. It does not reopen the broader validated-input, package-profile, suite-selection, or outer-negotiation design space.

The decision was gated per project policy. The first `ras consider` round required fixes for copied single-use values, failed-`Finish` deferred `Close`, synchronization-model wording, ADR-0006 linkage, diagnostic wording, and evidence-refresh linkage. The revised ADR changed the implementation direction to shared terminal state for constructed value copies. A fresh second `ras consider` round then required guard-invariant and loser-path verification precision; `ras verify` on that second run returned `unresolved: []`.

## Context

ADR-0001 extracted the deep **CPace core** and settled its secret-ownership model: `initiatorCore` owns the initiator role scalar until `Initiator.Finish`, and `responderCore` owns the responder working ISK until `Responder.Finish`. The public `Initiator` and `Responder` values are thin **single-use state** shells around those cores.

`Finish` already gives strong locality for state-consuming paths. It consumes the single-use state, defers `core.clear()` immediately after consumption, and clears persistent core secrets on success, parse failure, peer-share rejection, and confirmation failure. `Start` and `Respond` intentionally do not clear the core on success because the core must survive the network round trip until `Finish`.

The remaining gap is abandonment: a caller can successfully call `Start` or `Respond`, then cancel the exchange, close the transport, hit a timeout, or otherwise drop the returned single-use state without ever calling `Finish`. In that case the persistent core secret is left for ordinary garbage collection. ADR-0001 records this as a pre-existing residual public-lifecycle risk and states that it is resolvable only if the public freeze is reopened to add an abort or `Close`-style method on `Initiator` and `Responder`, distinct from `Session.Close`.

The maintainer has now elected to explore this narrow thaw before any broader validated-input or package-profile redesign. That sequencing matters: this ADR should define the lifecycle invariant for single-use state in a way that future input/profile work can preserve without reworking secret cleanup.

## Decision

Add `Close() error` methods to both public single-use state types:

```go
func (i *Initiator) Close() error
func (r *Responder) Close() error
```

`Close` is a local cleanup operation, not a protocol message. It does not send an alert, does not authenticate the peer, and never returns a `Session`. Its job is to let callers release the persistent secrets held by abandoned single-use state.

The lifecycle invariant becomes: `Finish` and `Close` are terminal operations on single-use state. Exactly one terminal operation wins. If `Finish` wins, it preserves today's semantics and may create a `Session`; a later `Close` is a no-op that returns `nil`. If `Finish` consumes the state but fails, a later deferred `Close` is also a no-op that returns `nil`. If `Close` wins, it clears the core-owned persistent secrets, marks the state spent, and a later `Finish` returns `ErrStateUsed`.

`Close` is idempotent and nil-safe. Calling `Close` on a nil `*Initiator` or nil `*Responder` returns `nil`, following the nil-receiver cleanup precedent from [[0006-close-on-nil-convention]]. Calling `Close` on a caller-fabricated zero-value `Initiator` or `Responder` returns `ErrInvalidInput` and does not consume the state, matching the zero-value non-consumption hardening rule from [[0001-extract-cpace-core]]. Zero-value `Close` reuses the existing same-type diagnostic wording from `Finish`: `uninitialized initiator` or `uninitialized responder`.

Copies of a constructed `Initiator` or `Responder` share the same terminal state. This mirrors `Session` copies sharing close state and avoids making the lifecycle invariant depend on callers never copying a Go value. The implementation therefore moves the terminal guard and core pointer behind an unexported state pointer owned by the shell; copying the shell copies the pointer, not the terminal guard. A terminal operation on either the original or a copy spends the shared state for all copies.

The post-refactor zero-value sentinel is `receiver.state == nil`. `Finish` and `Close` must split nil-receiver handling from zero-value-state handling before dereferencing the shared state. For constructed values, `state` and `state.core` are written exactly once at construction and are never reassigned or nilled; cleanup clears only core-owned fields through `clear()`. The existing pointer-stability comment in `api.go` must migrate from the shell's direct `core` field to the shared-state design.

`Close` should be safe to defer immediately after successful construction:

```go
initiator, messageA, err := cpace.Start(cfg)
if err != nil {
    return err
}
defer initiator.Close()
```

That pattern remains correct whether the exchange later succeeds, fails, or is abandoned. On success, `Finish` clears the core and the deferred `Close` returns `nil`; callers must still close any returned `Session` separately. The deferred `Close` error is intentionally safe to ignore in this pattern because the nil and zero-value error paths are unreachable after successful construction.

The implementation keeps the current CPace core seam. `Close` lives on the public shells, uses the same shared single-use terminal guard as `Finish`, and calls the existing core `clear()` methods. The terminal guard protects terminal ownership, not core execution itself: core operations run outside the lock, only the winning terminal operation may use or clear the core, losing `Finish` returns `ErrStateUsed`, losing `Close` returns `nil`, and `clear()` idempotence is only a backstop. The implementation must not move Config validation, package cap policy, package-profile binding, or outer-negotiation rules into the lifecycle seam.

## Acceptance criteria

Multi-agent review concurrence on this ADR moves it `proposed -> accepted`. The criteria below are implementation-verification gates for the follow-up code PR.

- **Public lifecycle methods**: `Initiator.Close() error` and `Responder.Close() error` exist, are documented, and are the only new public surface in the implementation PR.
- **Nil and zero-value behavior**: `(*Initiator)(nil).Close()` and `(*Responder)(nil).Close()` return `nil`; `(&Initiator{}).Close()` and `(&Responder{}).Close()` return errors matching `ErrInvalidInput`, use the `uninitialized initiator` / `uninitialized responder` wording, and do not consume the state.
- **Abandoned-state cleanup**: after successful `Start`, calling `Initiator.Close` clears the initiator core scalar; after successful `Respond`, calling `Responder.Close` clears the responder core ISK and transcript hygiene field.
- **Terminal ordering**: `Finish` after `Close` returns `ErrStateUsed`; `Close` after successful `Finish` returns `nil` and does not affect the returned `Session`; `Close` after a consuming failed `Finish` returns `nil`.
- **Idempotence, copying, and concurrency**: repeated `Close` calls return `nil` after the first successful close; value copies of constructed `Initiator` and `Responder` share terminal state; concurrent `Finish`/`Close` calls, including original-vs-copy cases, are race-safe with exactly one terminal operation winning.
- **Losing paths do not touch core**: implementation tests or an explicit implementation audit prove failed-consume paths return before any core read, finish call, or clear call. Prefer a deterministic internal counter or equivalent white-box check showing exactly one core clear across Finish-wins/Close-loses and Close-wins/Finish-loses cases for both roles, including original/copy pairs.
- **Existing Finish semantics preserved**: `Finish` still consumes state on parse failure, peer-share rejection, and confirmation failure, still clears core state on all those paths, and existing state-reuse tests continue to pass.
- **Documentation**: README, `docs/integration-guidance.md`, and relevant doc comments tell callers to `defer Close` on single-use state after successful `Start`/`Respond` when an exchange might be abandoned, and separately to `Close` returned sessions when done.
- **Changelog**: `CHANGELOG.md` records the narrow public-lifecycle addition and states that it closes the abandoned single-use state cleanup gap named by ADR-0001.
- **Evidence discipline**: because the implementation touches security-relevant lifecycle code, existing pinned dependency-review, fuzz, Capslock, and security/spec evidence cannot support a stronger release claim for the post-change commit until the exact-candidate evidence refresh is completed. The implementation PR does not perform that refresh. If this implementation lands before the already-planned consolidated post-ADR-0001 evidence refresh and reviewer-packet re-pin, it must be included in that pass; otherwise it requires its own exact-candidate refresh before a stronger release claim.

## Considered options

- **A - Add `Close` on `Initiator` and `Responder` (chosen).** Best matches Go cleanup conventions and `Session.Close`, supports immediate `defer`, and names the operation as local resource cleanup rather than a peer-visible protocol action.
- **B - Add `Abort` on `Initiator` and `Responder`.** Communicates cancellation, but risks implying a protocol abort message or remote signal. It also diverges from the existing `Session.Close` cleanup vocabulary.
- **C - Add runtime cleanup or finalizers for single-use state.** Would not help reachable abandoned state, would add nondeterministic runtime behavior, and would repeat the trade-offs rejected for `Session` in ADR-0004. Caller-owned cleanup is the deeper module seam.
- **D - Keep the status quo.** Leaves the ADR-0001 residual lifecycle risk in place and keeps safe cancellation spread across caller discipline and garbage collection timing.
- **E - Wait for the broader validated-input/profile redesign.** Couples a concrete lifecycle cleanup gap to a larger public-surface redesign. This ADR intentionally keeps the lifecycle seam independent so future input/profile work can happen later without reworking cleanup.

## Consequences

- Callers gain a simple safe-default pattern: call `Close` with `defer` immediately after successful `Start` or `Respond`, then close any returned `Session` separately after successful `Finish`.
- Copies of constructed single-use state become safer because they share one terminal guard, matching the public lifecycle invariant instead of relying on caller discipline or `go vet` copylock warnings.
- The persistent-secret audit remains local: core constructors create persistent secrets, `Finish` and `Close` are the terminal cleanup paths, and `Session` owns only its independent ISK clone.
- The public surface grows before v1.0.0. This is intentional and narrowly scoped to the ADR-0001 residual risk.
- Future validated-input or package-profile work should preserve the lifecycle invariant: any public shape that creates single-use state must provide the same terminal cleanup semantics.
- The implementation is security-relevant. It improves abandoned-state cleanup, but it also requires exact-candidate evidence refresh before making stronger release-readiness claims. That refresh is deliberately deferred to the existing release-evidence workflow rather than bundled with this ADR or its implementation PR.

## Implementation outline

1. Move the terminal guard and core pointer into unexported state structs shared by value copies of `Initiator` and `Responder`; keep the public shell types as thin single-use state values with unexported fields only. The zero-value sentinel is a nil shared-state pointer, and constructed shared state holds a non-nil core pointer assigned once.
2. Add nil-safe, idempotent `Close` methods to `Initiator` and `Responder` in `api.go`. Each method checks for nil receiver, rejects caller-fabricated zero-value state with `ErrInvalidInput`, consumes the shared single-use state under the terminal guard, and clears the core only if this call won the terminal operation.
3. Keep `Finish` behavior unchanged except for using the shared terminal-state helper. Guard nil receiver before nil shared-state before any core dereference. Do not nil the shared-state pointer or core pointer after construction; clear core-owned fields through `clear()` so losing paths cannot observe a different construction state.
4. Add tests covering nil receiver, zero-value strictness, abandoned-state cleanup, repeated `Close`, `Finish` after `Close`, `Close` after successful `Finish`, `Close` after consuming failed `Finish`, Session independence, value-copy terminal-state sharing, loser-never-touches-core behavior, and concurrent `Finish`/`Close` under the race detector.
5. Update README, integration guidance, doc comments, and `CHANGELOG.md`.
6. Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `task check`. Record that exact-candidate evidence refresh remains pending rather than refreshing it in this change.
