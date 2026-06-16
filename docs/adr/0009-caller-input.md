---
status: accepted
date: 2026-06-16
review-runs:
  - 20260616T065527-779be2a67a01358b100aa80e # ras consider round 1 - required public-surface, test-seam, secret-lifetime, docs, diagnostics, and role-local precision fixes
  - 20260616T071012-53ce69dd6acfbef7baa79635 # ras consider round 2 - required thaw framing, CONTEXT/docs, evidence sequencing, empty local-AD, validation-order, and cap-test precision fixes
  - 20260616T072321-d625df82f8b30786dc5ac33d # ras consider round 3 - required context-including validation-order precision
  - 20260616T072321-d625df82f8b30786dc5ac33d-verification-1781595592834070000 # ras verify - unresolved: []
---

# Role-local caller input before v1.0.0

## Status

**Accepted (2026-06-16) - broad thaw for Caller input, narrow implementation authorization.** This ADR records a deliberate reopen of the release-readiness public-surface and package-profile freeze for **Caller input**. The accepted decision replaces the caller-input shape broadly within that module, but its authorization is narrowly limited to the follow-up `Input` implementation described below. It does not reopen CPace computation, **Message framing** wire bytes, peer-share rejection, **single-use state** lifecycle, release policy, evidence policy, or unrelated public API and package-profile choices.

The decision was gated per project policy. The first `ras consider` round required fixes to bind the final public `Input` surface, migrate deterministic test seams, make the secret-lifetime audit explicit, broaden documentation gates, choose public validation vocabulary, pin compatibility-flag behavior, and exercise distinct local associated data. The second round required removal of hypothetical thaw framing, CONTEXT.md and review/evidence sequencing gates, empty `LocalAssociatedData` preservation, validation-order precision, and cap-test migration precision. A fresh third round after those context changes required one remaining validation-order fix to include `Context`; `ras verify` on that third run returned `unresolved: []`.

## Context

The current public `Config` type is role-oriented around the protocol transcript: `InitiatorID` is always the party that calls `Start`, `ResponderID` is always the party that calls `Respond`, and `AssociatedData` is ADa for `Start` but ADb for `Respond`. Both parties must populate those fields in the same global orientation, so the responder must put the peer first and itself second. The docs warn about this because a common integration bug is for each side to put itself first, which produces different CI values and makes confirmation fail.

Internally, **Package-owned cap policy**, `acceptConfig`, and `normalizeConfig` already provide useful locality: they validate required fields, enforce caps, copy accepted slices, build CI, and prevent caller mutation from affecting **single-use state**. That internal implementation is reasonably deep. The shallow part is the public **Caller input** interface: callers must remember role orientation, fresh SessionID policy, the exceptional empty-session switch, package caps, copy semantics, and whether `AssociatedData` names local or protocol-role data.

The freeze normally blocks public-surface changes unless a review finding forces a reopen. For **Caller input**, the maintainer has elected a broad thaw before v1.0.0 so the package can settle the best v1.0.0 caller-input shape before the public contract is frozen again. That authorization is narrow: only the follow-up `Input` implementation may use this thaw, and all unrelated public API and package-profile choices remain frozen. Under that Caller input replacement, leaving `Config` as the primary interface carries caller complexity into v1.x. Adding a second long-term input path would reduce migration pain but leave two public modules describing the same facts.

There is also a memory-handling constraint. A deeper **Caller input** module must not introduce a long-lived validated object that stores password bytes beyond the `Start` or `Respond` call. The current flow copies caller input, builds **CPace core**, and eagerly clears accepted password bytes after generator derivation. A reusable validated-input object would lengthen password lifetime and make secret ownership less local.

## Decision

Replace the current public `Config` shape before v1.0.0 with one role-local **Caller input** type:

```go
type Input struct {
    Password            []byte
    SelfID              []byte
    PeerID              []byte
    Context             []byte
    SessionID           []byte
    LocalAssociatedData []byte
    AllowEmptySessionID bool
}
```

`Start(input Input)` treats `SelfID` as the initiator identity and `PeerID` as the responder identity. `Respond(input Input, messageA []byte)` treats `SelfID` as the responder identity and `PeerID` as the initiator identity. The implementation maps these role-local facts to the internal initiator-responder CI ordering before constructing **CPace core**.

`Password`, `Context`, and `SessionID` are shared session values both parties supply identically. `SelfID`, `PeerID`, and `LocalAssociatedData` are role-local values supplied from each caller's point of view and mapped by the package.

The final v1.0.0 public surface has one **Caller input** module. Public `Config` is removed before v1.0.0 and must not ship as a deprecated alias, wrapper, or peer input path. A branch-local compatibility adapter may exist only as temporary implementation scaffolding and must be removed before the implementation PR is ready for review.

The field name `LocalAssociatedData` is intentionally longer than `AssociatedData`: it names the caller's local associated data and removes the ADa/ADb role translation from caller memory. The returned `Session.PeerAssociatedData` remains the peer's local associated data from the confirmed exchange.

`LocalAssociatedData` remains optional. Nil and empty `Context` or `LocalAssociatedData` slices encode identically in transcript construction and validation; callers may omit local associated data without weakening the field's role-local naming. `Session.PeerAssociatedData()` keeps the existing **Session** behavior; callers should treat `len(session.PeerAssociatedData()) == 0` as the stable contract for omitted or empty peer local associated data.

The field name `AllowEmptySessionID` remains unchanged from the current `Config` spelling so draft-compatibility policy stays recognizable. The doc comment must say this flag exists only for draft-21 compatibility tests or profiles that deliberately accept weaker empty-session behavior.

No reusable validated-input object is introduced. **Caller input** remains a per-call value supplied to `Start` or `Respond`; accepted byte slices are copied, package caps are enforced before allocation-sensitive use, and accepted password bytes are cleared on the same lifetime rules as today.

`Input.Context` semantics do not expand in this ADR. Outer negotiation and downgrade protection remain caller-owned, and integrations still need to bind negotiated outer-protocol state into `Input.Context` or `Input.LocalAssociatedData`. This ADR deepens local role and input policy; it does not make `Session.TranscriptID` a complete channel binding.

The wire format, **Message framing** caps, CPace transcript ordering, **Peer-share rejection**, confirmation tags, **Session** behavior, **single-use state** lifecycle, and exported error identities remain unchanged except for any intentional public input type/name changes recorded by this ADR.

## Acceptance criteria

Multi-agent review concurrence on this ADR moves it `proposed -> accepted`. The criteria below are implementation-verification gates for the follow-up code PR.

- **Public input surface**: `Input` exists with the role-local fields named in the Decision, `Start` accepts `Input`, `Respond` accepts `Input`, and public `Config` is absent from the final v1.0.0 surface. Verification must include `rg "type Config|func Start|func Respond"` and `go doc` inspection showing that `Input` is the sole public caller-input module.
- **Role-local mapping**: `Start(Input{SelfID: A, PeerID: B, LocalAssociatedData: ADa})` and `Respond(Input{SelfID: B, PeerID: A, LocalAssociatedData: ADb}, messageA)` complete successfully for matching password, context, and SessionID, with `ADa != ADb`; reversing one side's role-local identities causes confirmation failure rather than silent success.
- **Peer metadata**: on a successful exchange with distinct local associated data, the initiator `Session.PeerID()` returns the responder's `SelfID`, the responder `Session.PeerID()` returns the initiator's `SelfID`, the initiator `Session.PeerAssociatedData()` returns `ADb`, and the responder `Session.PeerAssociatedData()` returns `ADa`.
- **Validation preservation and diagnostics**: empty password, empty `SelfID`, empty `PeerID`, empty SessionID without `AllowEmptySessionID`, and oversized fields are rejected with the same exported error identities as today; cap values and error categories remain unchanged; public validation messages use `self id`, `peer id`, and `local associated data` rather than leaked initiator/responder or old `AssociatedData` vocabulary. Empty `LocalAssociatedData` remains allowed, including nil and empty slices on either or both roles. Validation keeps today's two-phase structure but adopts uniform role-local field order for both `Start` and `Respond`: required-field checks run first in password, self id, peer id, session id order; cap checks then run in password, self id, peer id, context, session id, local associated data order. This intentionally normalizes responder empty-ID precedence to self-before-peer while preserving exported error identities, cap values, and error categories. Tests pin both `errors.Is` identities and message text for empty password, empty `SelfID`, empty `PeerID`, empty SessionID without `AllowEmptySessionID`, oversized password, oversized `SelfID`, oversized `PeerID`, oversized `Context`, oversized SessionID, and oversized `LocalAssociatedData`, including context-involving precedence cases.
- **Compatibility flag**: empty SessionID works only when both roles set `AllowEmptySessionID`, and asymmetric empty-session compatibility remains rejected in the same observable way as today. Non-empty matching `SessionID` exchanges succeed and preserve transcript/export behavior regardless of asymmetric `AllowEmptySessionID` values.
- **Copy ownership**: caller mutation of every accepted `Input` byte slice after `Start` or `Respond` does not affect **single-use state**, **CPace core**, `Session.PeerID`, `Session.PeerAssociatedData`, or exported key material.
- **Secret lifetime audit**: the implementation PR includes a named manual secret-lifetime audit artifact that inspects `normalizedConfig.wipe` or its successor, `newInitiatorCore`, `newResponderCore`, and the new `Input` acceptance code. The audit must confirm that no reusable validated-input object is introduced, no new persistent password field is introduced, and accepted password bytes are cleared on success, error, and panic paths.
- **Deterministic test seam migration**: `startWithRandom` and `respondWithRandom`, or explicitly named successors, accept `Input`; fuzz, benchmark, vector, API, example, and deterministic-random tests migrate away from public `Config` while preserving identical wire bytes and fuzz-seed outcomes. Verification must include `rg -n "startWithRandom|respondWithRandom|cpace\\.Config|\\bConfig\\{" --type go`, with `go test ./...` as the authoritative completeness check after public `Config` removal.
- **Protocol stability**: draft vectors, full exchanges, mismatch tests, peer-share rejection tests, **Message framing** tests, non-empty asymmetric compatibility tests, and fuzz seed invariants still pass after migration; no wire-format bytes change.
- **Documentation and public comments**: README, CONTEXT.md, examples, `doc.go`, exported comments, `docs/integration-guidance.md`, threat model, security assessment, spec matrix, external-review handoff, security/spec audit material, and evidence-facing docs are swept for `Config`, `InitiatorID`, `ResponderID`, `AssociatedData`, and old role-orientation guidance. Remaining matches must be intentionally historical ADR context, intentionally private implementation plumbing such as `acceptConfig`, `acceptedConfig`, `normalizeConfig`, or `normalizedConfig`, or updated to describe role-local input using `Input`, `SelfID`, `PeerID`, and `LocalAssociatedData` while keeping outer-negotiation downgrade guidance. The CONTEXT.md sweep must update **Caller input**, **Package-owned cap policy**, and any accepted/normalized-input successor vocabulary. Verification must include `rg -n "Config|InitiatorID|ResponderID|AssociatedData|role orientation" README.md CONTEXT.md *.go docs`.
- **Changelog**: `CHANGELOG.md` records the pre-v1 public input change, names the removed `Config` fields, and gives the migration rule: initiator uses `SelfID=initiator, PeerID=responder`; responder uses `SelfID=responder, PeerID=initiator`.
- **Review and evidence sequencing**: because external-review outreach is currently deferred, this ADR and its implementation may land before outreach starts; if they do, the reviewer packet, external-review handoff, and reviewer-facing anchors must be re-pinned after the implementation lands and before outreach resumes. Because the implementation changes public input and security-relevant validation code, existing pinned dependency-review, fuzz, Capslock, and security/spec evidence cannot support a stronger release claim for the post-change commit until the exact-candidate evidence refresh is completed. If this implementation lands before the consolidated post-ADR-0001 evidence refresh and reviewer-packet re-pin, it must be included in that pass; otherwise it requires its own exact-candidate refresh before a stronger release claim. The implementation PR does not perform that refresh unless it is explicitly scoped as a release-evidence PR.

## Considered options

- **A - Replace `Config` with role-local `Input` before v1.0.0 (chosen).** Gives callers the most leverage: they provide self, peer, and local associated data, while the package owns role translation, validation, cap policy, and copy ownership. Cost: breaking pre-v1 public-surface change and documentation migration.
- **B - Add additive `StartInput` / `RespondInput` while keeping `Config`.** Reduces migration pain but leaves two public modules for the same facts and keeps the shallow `Config` interface alive into v1.x unless a later breaking removal occurs.
- **C - Keep `Config` and deepen only internal `acceptConfig` / `normalizeConfig`.** Improves maintainer locality but does not reduce caller burden; the public interface remains shallow.
- **D - Add a reusable validated input object.** Concentrates validation but lengthens password lifetime and creates a second secret owner. Rejected on secret-locality grounds.
- **E - Expand package profile and outer-binding support at the same time.** Could reduce downgrade-protection footguns, but it is a larger strategic package-scope change. This ADR keeps outer negotiation caller-owned and only deepens local input role semantics.

## Consequences

- Callers no longer need to remember the global initiator-responder identity ordering. Each side supplies `SelfID` and `PeerID`, and the package maps those facts to the transcript order.
- The public input interface becomes deeper: more behavior sits behind one smaller caller-facing rule, while validation, cap policy, copy ownership, and role translation concentrate in one implementation.
- The change is breaking relative to current `Config` callers. Because this is pre-v1.0.0, the migration is acceptable only if accepted before the v1 public-surface freeze is re-established.
- Internal tests should become less coupled to `acceptConfig` and `normalizeConfig` details. Most new coverage should exercise public `Start` / `Respond` behavior; direct internal tests remain appropriate only for package-owned cap policy and secret-lifetime audits that cannot be observed publicly.
- Future package-profile or outer-binding work should preserve the role-local **Caller input** invariant rather than reintroducing initiator/responder field orientation at the public seam.

## Implementation outline

Use TDD vertical slices. Do not write all tests first.

1. Add one public test proving a role-local `Input` initiator and responder complete an exchange and export matching key material; implement only enough public surface and mapping to pass.
2. Add one public test proving `PeerID` and `PeerAssociatedData` report the peer's role-local facts; implement the necessary mapping.
3. Add one public test proving reversed role-local identities fail confirmation; preserve current confirmation-failure identity.
4. Add public validation tests for empty password, `SelfID`, `PeerID`, SessionID, oversized renamed fields, two-phase validation precedence including `Context`, and the empty-session compatibility flag; add success coverage proving nil and empty `LocalAssociatedData` remain accepted. Implement validation by adapting **Package-owned cap policy** rather than duplicating it, and pin diagnostics to `self id`, `peer id`, and `local associated data`.
5. Add one mutation-after-construction test covering password, IDs, context, SessionID, and local associated data; preserve copy ownership.
6. Add cap-boundary tests only where the public behavior changes names; update `TestPackageOwnedCapPolicyPinsShippedValues` so caller-input cap name strings rotate to `self id`, `peer id`, and `local associated data`, while numeric cap values and **Message framing** cap names remain pinned.
7. Refactor after green: concentrate `Input` acceptance and role mapping behind the **Caller input** module, migrate `startWithRandom` / `respondWithRandom` or their explicit successors to `Input`, and remove public `Config` from the review-ready branch.
8. Update README, CONTEXT.md, examples, `doc.go`, integration guidance, public comments, security/spec/reviewer/evidence-facing docs that name old caller input fields, and `CHANGELOG.md`.
9. Add the required manual secret-lifetime audit artifact for accepted password ownership and clearing, then run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `task check`. Record exact-candidate evidence refresh as pending unless the PR is explicitly an evidence refresh.
