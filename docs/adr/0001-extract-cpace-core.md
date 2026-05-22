---
status: accepted
---

# Extract a deep CPace core

## Status

**Accepted (2026-05-22).** Future architecture reviews should treat the
CPace-core shape recorded here as settled and not re-litigate it.

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

## Context

The CPace cryptographic composition — generator derivation, scalar sampling,
Diffie-Hellman, transcript assembly, ISK derivation, confirmation tags, and
secret-clearing — has no module of its own. It is smeared across `crypto.go`
primitives, `strings.go` transcript builders, and four orchestration functions
in `api.go`. The security-critical invariant "every persistent secret is zeroed
on every path" is enforced by hand-placed `defer`s spread across those four
functions, so verifying it requires tracing every function and every
early-return path.

## Decision

Extract a deep, unexported **CPace core** (see `CONTEXT.md`): stateful
`initiatorCore` and `responderCore` types embedded in the public `Initiator`
and `Responder`, which become thin shells.

The core owns one role's cryptographic computation and the lifetime of its
**persistent** secrets — the initiator's role scalar, and the responder's ISK —
exposing a single `clear()` per core. Scratch secrets (the normalized password,
the generator element, the Diffie-Hellman point `k`, the initiator's
`finish`-local ISK, and derivation buffers) are **not** core fields: they remain local variables cleared eagerly at the
narrowest scope inside core methods, exactly as the current code does. Storing
them on the core to be cleared by `clear()` would extend a plaintext secret's
lifetime across the network round-trip — a regression, not the goal.

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
- **Public interface** — unchanged. `Initiator` / `Responder` are fully opaque,
  so this is internal-only and does not touch the frozen public API or profile
  policy.

## Acceptance criteria

The implementation must satisfy these before this ADR moves
`proposed → accepted`:

- **`clear()` contract** — `clear()` is idempotent and nil-safe. Each public
  shell method defers core cleanup immediately after a core exists or the
  single-use state is consumed, on **all** paths including parse and
  confirmation failure. Core constructors and methods clear any
  partially-initialized secret before returning an error.
- **Scratch-secret cleanup** — `initiatorCore` / `responderCore` hold no
  `password`, generator, `k`, or `finish`-local ISK field; core methods clear
  every scratch secret eagerly via local `defer` at the narrowest scope.
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
  if the public API freeze is reopened to add a `Close`.
- The refactor is invasive and security-relevant. The behavioral regression net
  (`api_test.go`) proves behavior is preserved but **cannot** prove zeroization
  is preserved — the secret-clearing relocation needs a manual audit pass.
- Per the project's evidence-discipline rule, fuzz and audit evidence
  (`docs/fuzz-evidence.md`, `docs/security-spec-audit.md`) must be refreshed
  before this work supports any stronger release claim.
