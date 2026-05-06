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
| Confirmation tag role separation | Draft-compatible tag input is unchanged. | Done. Keep draft-compatible tags; no package-added role labels. |
| Field size limits | Package-owned per-field caps: password and IDs 4 KiB, context and session ID 1 KiB, AD 64 KiB, public shares/tags exact-sized. | Done. Caps remain non-configurable and are not aggregate message limits. |
| Scalar sampling | Masked canonical 32-byte sampling with zero retry. | Done. Keep the draft-21 Ristretto255 recommendation; `SetUniformBytes` plus zero rejection/retry is an allowed alternative but would use 64-byte modulo reduction and define a different package profile. |

## Recommended PR Order

All current policy rows are closed. Move to release-readiness work unless a new
review finding reopens a policy decision.

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
