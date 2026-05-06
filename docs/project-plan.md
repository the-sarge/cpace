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
| Empty `SessionID` | Accepted for draft-21 compatibility and documented as weaker. | Keep draft-permissive behavior, or reject empty values by default with an explicit compatibility escape hatch. |
| `Session.Discard()` | No public session-destruction method; internal consumed state is cleared best-effort. | Add a public lifecycle method that clears `Session.isk` and makes future `Export` fail, or keep the session API minimal. |
| Peer associated data | Peer AD is bound into transcripts but not exposed through accessors. | Add accessors to reduce application-layer mistakes, or keep messages opaque. |
| Confirmation tag role separation | Draft-compatible tag input is unchanged. | Keep draft-compatible tags, or add role labels as a package-profile hardening break. |
| `maxFieldLength` | Parser cap is 1 MiB. | Keep, lower, or make configurable. |
| Scalar sampling | Masked canonical 32-byte sampling with zero retry. | Keep for draft conformance, or investigate a `SetUniformBytes`-based approach and prove compatibility/distribution properties. |

## Recommended PR Order

1. Empty `SessionID` policy.
   Decide whether draft compatibility remains the default or moves behind an
   explicit compatibility option.

2. Session lifecycle and peer-data API.
   Consider `Session.Discard()` and peer associated-data accessors together
   because both are public API surface changes.

3. Framing and confirmation profile choices.
   Decide `maxFieldLength` and confirmation tag role separation after the
   compatibility posture is explicit.

4. Scalar sampling investigation.
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
