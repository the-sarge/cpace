# Project Plan

Status: living release-readiness plan after the policy/API decisions landed in
PRs #13-#17.

This document tracks current work. Historical review triage remains in
`docs/interview-results-triage.md`.

## Current Phase

The current phase is release readiness. Public API and package-profile policy
decisions are closed unless a new review finding reopens one. Do not describe
the package as production-ready until the release bar below is satisfied and
independent cryptographic review is complete.

## Release-Readiness PR Shape

Each release-readiness PR should include:

- the release-readiness gap being closed;
- the exact commit, command, workflow, or review artifact used as evidence;
- any residual risk or follow-up that remains after the PR;
- README, changelog, security, and spec documentation updates when release
  posture changes.
- no public API or package-profile changes; reopen the policy phase first if a
  new finding requires one.

## Closed Policy Decisions

All rows below are closed and preserved as the policy/API decision record.

| Area | Current behavior | Decision needed |
| --- | --- | --- |
| `Config.Rand` | Removed from the public API; scalar randomness always uses `crypto/rand.Reader`. | Done. Deterministic readers remain package-internal for tests and fuzzing only. |
| Empty `SessionID` | Rejected by default; `AllowEmptySessionID` preserves explicit draft-21 compatibility. | Done. Callers must opt into weaker empty-sid behavior deliberately. |
| Session lifecycle | `Session.Close` clears the session ISK best-effort and future `Export` calls fail with `ErrSessionClosed`. | Done. Non-secret metadata remains available after close. |
| Peer metadata | `PeerAssociatedData` and `PeerID` expose copied metadata bound into the confirmed exchange. | Done. Local AD/ID accessors are deferred until a concrete caller need appears. |
| Confirmation tag role separation | Draft-compatible tag input is unchanged. | Done. Keep draft-compatible tags; no package-added role labels. |
| Field size limits | Package-owned per-field caps: password and IDs 4 KiB, context and session ID 1 KiB, AD 64 KiB, public shares/tags exact-sized. | Done. Caps remain non-configurable and are not aggregate message limits. |
| Scalar sampling | Masked canonical 32-byte sampling with zero retry. | Done. Keep the draft-21 Ristretto255 recommendation; `SetUniformBytes` plus zero rejection/retry is an allowed alternative but would use 64-byte modulo reduction and define a different package profile. |

## Recommended PR Order

1. External review package.
   Prepare reviewer handoff notes for draft-compatible behavior, package-owned
   framing/profile choices, unsupported scope, and remaining release blockers.

## Completed Evidence

| Area | Evidence | Residual risk |
| --- | --- | --- |
| Dependency review | `docs/dependency-review.md` records `govulncheck -test -show verbose ./...` and advisory `gosec v2.26.1` results for commit `06f21c51645f54e2b7bde7c5b538479463be5d0e`. | Repeat on the exact release tag if dependencies, toolchain, or parser/security-relevant code changes. |
| Long fuzz evidence | `docs/fuzz-evidence.md` records all 14 registered targets on local smoke and long ARM/Intel runs for commit `06f21c51645f54e2b7bde7c5b538479463be5d0e`. | Repeat if parser, protocol, fuzz harness, dependency, or toolchain changes before release. |
| Security/spec audit | `docs/security-spec-audit.md` records review of `docs/security-assessment.md` and `docs/spec-matrix.md` against implementation commit `4a8f629e59f0cc5c8f9351abacfa511fe6e4f441`. | Repeat if protocol code, parser/framing code, package-profile docs, dependencies, toolchain, or the targeted draft revision changes. |
| Integration guidance | `docs/integration-guidance.md` documents outer PAKE/version negotiation, downgrade-protection, identity-orientation, and session-output guidance. | External reviewers should still evaluate whether this guidance is sufficient for real integrations. |

## Release Readiness

Before any production-readiness claim:

- run every fuzz target for more than five minutes on release hardware or in
  the long-fuzz workflow;
- repeat dependency review with `govulncheck -test -show verbose ./...`;
- review `docs/security-assessment.md` and `docs/spec-matrix.md` against the
  exact release commit;
- complete external review of package-owned framing and profile choices;
- obtain independent cryptographic review.

## Later Investigation

- OpenSSF Scorecard.
- CodeQL.
- Capslock.
- OSS-Fuzz.
- Longer continuous fuzzing campaigns.
- Offline Sage-derived extended vector dataset.
- Allocation measurements on hot paths before adding permanent allocation
  tests.
