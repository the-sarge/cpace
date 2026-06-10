---
status: accepted
date: 2026-05-22
review-runs:
  - 20260522T081906-6c67083f870be5ac1f971508 # phase 1 — this ADR
  - 20260522T150534-dc141248a1c30a4d025c1c1f # phase 2 — implementation plan
  - 20260522T152641-7d2ca5ac0cd36d5a1062254b # phase-2 re-run — revised plan
---

# Extract a deep CPace core

## Status

**Accepted (2026-05-22) — architecture.** The architecture itself — Candidate A, the seam placement, and the secret-ownership model — was endorsed by every review round below and is settled; future architecture reviews should not re-litigate it. That bar covers the architecture, not the surrounding document text: the plan converged at revision 3 *after* the last review round, so the binding wording (acceptance criteria, `clear()` contract, build sequence) postdates the reviews that are recorded here.

The decision was gated on independent multi-agent review (`ras consider`)
before enforcement:

- **Phase 1** — review of this ADR (run
  `20260522T081906-6c67083f870be5ac1f971508`): *proceed with changes*. The
  findings were folded into the amendments and acceptance criteria below.
- **Phase 2** — review of the implementation plan `docs/cpace-core-plan.md`
  (run `20260522T150534-dc141248a1c30a4d025c1c1f`): *proceed with changes*, ten
  findings, all plan-text precision; the plan was revised to address them.
- **Phase-2 re-run** — review of the revised plan (run
  `20260522T152641-7d2ca5ac0cd36d5a1062254b`): *proceed with changes*, six
  findings, all plan-text precision; the plan converged at revision 3.

No review round disputed the architecture — Candidate A, the seam placement,
and the secret-ownership model were validated each time. Implementation follows
the build sequence in `docs/cpace-core-plan.md`; the acceptance criteria below
remain binding on it.

**Revisions (2026-06-10).** A five-perspective review of this branch found record-integrity defects (tense, enumeration gaps, gating wording) and two undecided questions; this revision fixes the defects and records the decisions (see *Zero-value hardening* under Decision and *Sequencing against release blockers* below). The architecture is unchanged. The revised ADR and plan text must pass a confirming `ras consider` round before implementation begins; that run will be appended to `review-runs` when it completes.

## Context

The CPace cryptographic composition — generator derivation, scalar sampling,
Diffie-Hellman, transcript assembly, ISK derivation, confirmation tags, and
secret-clearing — has no module of its own. It is smeared across `crypto.go`
primitives, `strings.go` transcript builders, and four orchestration functions
in `api.go`. The security-critical invariant "every persistent secret is zeroed
on every path" is enforced by hand-placed deferred closures in the two `Finish`
methods — the constructors' `defer`s cover scratch secrets only, because the
persistent secrets deliberately survive construction — so verifying it requires
tracing each function and every early-return path.

## Decision

Extract a deep, unexported **CPace core** (see `CONTEXT.md`): stateful
`initiatorCore` and `responderCore` types held as a named `core` field by the
public `Initiator` and `Responder`, which become thin shells. (A named field,
not Go struct embedding — no core methods are promoted onto the public types.)

The core owns one role's cryptographic computation and the lifetime of its
**persistent** secrets — the initiator's role scalar, and the responder's ISK
(the responder's stored transcript is public wire data, but it is zeroed
alongside the ISK as hygiene) — exposing a single `clear()` per core. Scratch
secrets (the normalized password, the generator element, the Diffie-Hellman
point `k`, the responder's ephemeral scalar, the initiator's `finish`-local
ISK, and derivation buffers) are **not** core fields: they remain local
variables cleared eagerly at the narrowest scope inside core methods, matching
the current code with one deliberate improvement — the initiator's
`finish`-local ISK moves from two explicit per-path clears to a single `defer`,
which also covers panic paths. Storing scratch secrets on the core to be
cleared by `clear()` would extend a plaintext secret's lifetime across the
network round-trip — a regression, not the goal.

Resolved design points:

- **Secret ownership** — the core owns the lifetime of its persistent secrets;
  no persistent secret is held loose in `api.go`. Scratch secrets never become
  core fields and are cleared eagerly in place.
- **Seam content** — decoded cryptographic fields cross the seam; wire framing
  (`encode`/`decodeMessage*`) stays in front of it. The responder core MUST
  validate the peer share `Ya` as canonical and non-identity **before**
  generator derivation or scalar sampling, preserving today's
  validation-before-randomness error precedence.
- **Randomness** — the core constructor takes an `io.Reader`, a new
  deterministic seam for core-level tests. This is *added*, not a replacement:
  the unexported `startWithRandom` / `respondWithRandom` api-level seam is
  retained for deterministic full-pipeline tests and fuzzing.
- **Session** — the core constructs the `*Session`; the only secret that
  persists past `clear()` is the Session's own independent ISK clone.
- **Public interface** — unchanged signatures and types. With one recorded
  exception (next bullet), observable behavior is also unchanged, so this is
  internal-only and does not touch the frozen public API or profile policy.
- **Zero-value hardening (narrow policy reopen, decided 2026-06-10)** —
  `Initiator` / `Responder` are exported structs, so a caller can fabricate a
  zero value; today `Finish` on one panics inside the crypto (initiator) or
  consumes the state and returns `ErrMessage`/`ErrConfirmationFailed`
  (responder). The shells gain a core-presence guard: `Finish` on a fabricated
  zero value returns `ErrInvalidInput` **without** consuming the state. This is
  an observable behavior change for fabricated zero values only; it is recorded
  here as a deliberate, narrow reopen of the behavior freeze, and ships with a
  changelog entry and a pinning test (see Acceptance criteria).

## Acceptance criteria

Acceptance of this ADR records review concurrence with the decision; it does
not assert these criteria are met. The criteria below are implementation
gates — binding on the implementation, which is complete only when every one
of them is satisfied:

- **`clear()` contract** — `clear()` is idempotent and nil-safe. Each
  state-consuming shell method (`Finish`) defers core cleanup immediately after
  the single-use state is consumed, on **all** paths including parse and
  confirmation failure. (`Start` / `Respond` intentionally defer no core
  cleanup — their core must survive until `Finish`.) Core constructors and
  methods clear any partially-initialized secret before returning an error.
- **Scratch-secret cleanup** — `initiatorCore` / `responderCore` hold no
  `password`, generator, `k`, responder-scalar, or `finish`-local ISK field;
  core methods clear every scratch secret eagerly at the narrowest scope — by
  local `defer`, or inline immediately after last use, exactly as the sketches
  in `docs/cpace-core-plan.md` show.
- **Constant-time comparisons** — every comparison over secret-derived values
  (confirmation tags, identity checks) remains `hmac.Equal`; no
  `bytes.Equal` / `reflect.DeepEqual` is introduced on such values. Enforced by
  the plan's mandatory manual audit.
- **Zero-value guard** — a test pins the hardened behavior: `Finish` on a
  caller-fabricated zero-value `Initiator` / `Responder` returns
  `ErrInvalidInput` without consuming the single-use state, and the change is
  noted in the changelog.
- **Validation order** — the responder core decodes and validates `Ya` before
  deriving the generator or sampling the scalar;
  `TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness` is preserved
  or migrated to the new seam.
- **ISK isolation** — a regression test confirms `Session` construction
  deep-clones the core ISK and that `core.clear()` never mutates Session-owned
  key material.
- **Deterministic test seam** — `startWithRandom` / `respondWithRandom` (or
  equivalent unexported wrappers) still exist, so `fuzz_test.go`, `api_test.go`,
  and `bench_test.go` compile and run with deterministic entropy.

## Considered options

- **Stateful core objects (chosen)** — the only option that gives
  persistent-secret-clearing locality: afterwards the audit question reduces to
  reading two `clear()` methods.
- **Functional core, intermediates self-cleared** — relocates the math and the
  scratch-secret clearing, but the persistent secrets (scalar, ISK) still cross
  the seam and keep their `defer`s in `api.go`. Half the headline benefit.
- **Functional core, secrets returned** — relocates the math only; all clearing
  stays scattered. Rejected.

## Sequencing against release blockers

Decided 2026-06-10: implementation is **hard-gated on the external reviews**.
The reviewer packet (`docs/external-review-handoff.md`) pins reviewers to a
fixed baseline, and issues #29–#31 ask them to evaluate exactly the
orchestration and secret-lifetime code this refactor relocates.

- Implementation of this ADR must not begin until the external review
  (#29–#31) and the independent cryptographic review (#32) conclude, or a later
  explicit maintainer decision accepts the review churn.
- Before implementation begins, the revised ADR and plan text must pass the
  confirming `ras consider` round noted under Status.
- After implementation, the exact-candidate evidence refresh (#33) applies in
  full against the post-refactor commit — long-bar fuzz campaign, dependency
  review/SAST, security/spec audit, and Capslock — regardless of
  `go.mod`/`go.sum` being byte-identical. See the plan's *Evidence &
  release-readiness* section.

## Consequences

- The **persistent-secret** audit becomes a one-time read of two `clear()`
  methods. Scratch secrets remain auditable at their point of use, cleared
  eagerly as today.
- The core's interface becomes a direct test surface — draft test vectors can
  drive it without framing or single-use plumbing. Primitive-level vector tests
  are retained as internal-seam tests.
- Zeroization is guaranteed on internal-error paths and on state-consuming
  `Finish` paths (success or failure). A single-use state **abandoned** without
  `Finish` leaves its secret to GC — a pre-existing residual API-lifecycle risk
  that this refactor neither introduces nor worsens, and that is resolvable only
  if the public API freeze is reopened to add an abort/`Close`-style method on
  `Initiator` / `Responder` (distinct from the existing `Session.Close`).
- The refactor is invasive and security-relevant. The behavioral regression net
  (`api_test.go`) proves behavior is preserved but **cannot** prove zeroization
  is preserved — the secret-clearing relocation needs a manual audit pass.
- Per the project's evidence-discipline rule, the pinned evidence — fuzz,
  security/spec audit, dependency review/SAST, and Capslock — must be refreshed
  against the post-refactor commit before this work supports any stronger
  release claim. The plan's *Evidence & release-readiness* section records the
  full disposition; the unchanged-`go.mod` carve-out does **not** exempt the
  dependency-review/SAST re-run, which is sensitive to source relocation.
