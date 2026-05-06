# Project Plan

Status: living plan after the decision-free hardening workstreams landed in
PRs #2-#8.

This document tracks current work. Historical review triage remains in
`docs/interview-results-triage.md`.

## Current Phase

The next phase is policy and API decisions. Do not implement these as incidental
cleanup: each item changes package semantics, compatibility, or caller
responsibility.

## Policy PR Shape

Each policy PR should include:

- the decision being made and why the rejected alternatives are not being used;
- the compatibility effect for existing callers and draft-21 vectors;
- any migration note needed by downstream integrations;
- focused tests for the changed behavior and unchanged compatibility paths;
- README, changelog, and security/spec documentation updates when the public
  contract changes.

## Policy Decisions

| Area | Current behavior | Decision needed |
| --- | --- | --- |
| `Config.Rand` | Removed from the public API; scalar randomness always uses `crypto/rand.Reader`. | Done. Deterministic readers remain package-internal for tests and fuzzing only. |
| Empty `SessionID` | Rejected by default; `AllowEmptySessionID` preserves explicit draft-21 compatibility. | Done. Callers must opt into weaker empty-sid behavior deliberately. |
| Session lifecycle | `Session.Close` clears the session ISK best-effort and future `Export` calls fail with `ErrSessionClosed`. | Done. Non-secret metadata remains available after close. |
| Peer metadata | `PeerAssociatedData` and `PeerID` expose copied metadata bound into the confirmed exchange. | Done. Local AD/ID accessors are deferred until a concrete caller need appears. |
| Confirmation tag role separation | Draft-compatible tag input is unchanged. | Keep draft-compatible tags, or add role labels as a package-profile hardening break. |
| `maxFieldLength` | Parser cap is 1 MiB. | Keep, lower, or make configurable. |
| Scalar sampling | Masked canonical 32-byte sampling with zero retry. | Keep for draft conformance, or investigate a `SetUniformBytes`-based approach and prove compatibility/distribution properties. |

## Recommended PR Order

1. Framing and confirmation profile choices.
   Decide `maxFieldLength` and confirmation tag role separation after the
   compatibility posture is explicit.

2. Scalar sampling investigation.
   Treat any change as a protocol-conformance project, not a mechanical
   refactor. Keep this separate from API policy changes.

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
- Offline Sage-derived extended vector dataset.
- Allocation measurements on hot paths before adding permanent allocation
  tests.
