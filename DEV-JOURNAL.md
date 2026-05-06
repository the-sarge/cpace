# Development Journal

**Append-only. New entries go at the END of this file.**

## cpace.S1 - 2026-05-05 20:34 EDT

**Main:** `f5ae7e2`
**Board:** CPace hardening workstreams merged; review triage documentation pending.
**Planner:** Josh

Recorded the post-review landing state for the CPace draft-21 hardening pass.
PRs #2 through #8 merged the documentation clarifications, draft-vector
coverage, fuzz expansion, security tooling, responder public-share
prevalidation, and best-effort memory hygiene workstreams.

This session adds the previously local interview-results triage document to the
repository and updates the changelog so the merged hardening work and triage
artifact are visible from the release notes.

---

## cpace.S2 - 2026-05-06 02:42 EDT

**Main:** `3f38b43`
**Board:** Policy/API decisions closed; release-readiness tracking begins.
**Planner:** Josh

Recorded the close of the CPace policy/API decision phase after PRs #13 through
#17 landed. Those PRs removed public randomness injection, required explicit
empty-session-ID compatibility opt-in, added session lifecycle and peer metadata
accessors, tightened framing/profile caps while keeping draft-compatible
confirmation tags, and documented the draft-21 scalar-sampling policy.

This session shifts the active project plan to release readiness. The next work
items are dependency review refresh, long fuzzing evidence, security/spec
documentation audit, and external review handoff before any production-readiness
claim.

---

## cpace.S3 - 2026-05-06 09:39 EDT

**Main:** `e1d0c6d`
**Board:** Dependency review and fuzz evidence for release readiness.
**Planner:** Josh

Combined the dependency-review refresh and long-fuzz evidence work into one
release-readiness branch. The advisory gosec lane flagged LEB128 integer
conversions, so this session records the parser cleanup and the clean rerun
alongside `govulncheck -test -show verbose ./...` evidence.

This session also starts recording fuzz evidence as a first-class artifact:
local smoke coverage plus long runs for all registered targets on ARM and Intel
hardware. The remaining release-readiness items are the security/spec audit and
external review handoff.

---

## cpace.S4 - 2026-05-06 12:20 EDT

**Main:** `4a8f629`
**Board:** Security/spec audit for release readiness.
**Planner:** Josh

Reviewed the security assessment and spec matrix against the merged
release-readiness implementation, the draft-21 source, and the recorded
dependency/fuzz evidence. No security/spec drift was found.

This session records the audit as `docs/security-spec-audit.md`, removes the
security/spec audit from the remaining release-readiness gaps, and leaves
external review handoff plus independent cryptographic review as the remaining
production-readiness blockers.

---

## cpace.S5 - 2026-05-06 14:49 EDT

**Main:** `5c97ae6`
**Board:** v0.1.0 draft-snapshot release preparation.
**Planner:** Josh

Prepared the first `v0.x` release snapshot after the project-owned
release-readiness gaps were reduced to integration guidance and release
hygiene. This session adds `docs/integration-guidance.md`, applies the
BSD-3-Clause license, keeps the release positioning explicit as unaudited and
not production-ready, and moves the changelog from `Unreleased` to `v0.1.0`.

The remaining production-readiness blockers are external review of
package-owned framing/CI/profile choices and independent cryptographic review.
