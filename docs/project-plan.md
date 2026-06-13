# Project Plan

Status: living release-readiness plan after the policy/API decisions landed in
PRs #13-#17 and the public `v0.1.2` external-review/evidence snapshot.

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
| Field size limits | Package-owned per-field caps: password and IDs 4 KiB, context and session ID 1 KiB, AD 64 KiB, public shares/tags exact-sized; malformed framed inputs also have a 128 KiB aggregate decoder backstop. | Done. Valid message shapes remain governed by non-configurable per-field caps; the aggregate cap is an invalid-message throttle. |
| Scalar sampling | Masked canonical 32-byte sampling with zero retry. | Done. Keep the draft-21 Ristretto255 recommendation; `SetUniformBytes` plus zero rejection/retry is an allowed alternative but would use 64-byte modulo reduction and define a different package profile. |

## Recommended PR Order

1. External review cycle.
   Use `docs/external-review-handoff.md` to brief reviewers on
   draft-compatible behavior, package-owned framing/profile choices, unsupported
   scope, current evidence, and remaining release blockers. Track findings in
   focused follow-up PRs.
2. Evidence-process hardening for issue #44.
   PR #48 covers phase 1: a reusable evidence-bundle policy and
   cross-toolchain vector-stability checklist. Phase 2 remains applying that
   policy to the next exact-candidate packet. Keep the issue open until the
   packet includes committed raw artifacts with `SHA256SUMS`, or immutable
   workflow links following `docs/evidence/README.md`, plus the
   vector-stability result or an explicit unavailable-toolchain rationale.
3. Exact-candidate evidence refresh.
   After any review-driven changes, repeat dependency review, long fuzzing, and
   security/spec audit against the exact release-candidate commit before making
   any stronger readiness claim.

## Completed Evidence

| Area | Evidence | Residual risk |
| --- | --- | --- |
| Dependency review | `docs/dependency-review.md` records `govulncheck -test -show verbose ./...` and pinned `gosec@v2.26.1` results under Go 1.26.3 for v0.1.2 package-code candidate `2e09774f171dde8c62763d6e35a258b0fef88801`; raw transcript is in `docs/evidence/v012-candidate-20260508/`. | Repeat on the exact release tag if dependencies, toolchain, or parser/security-relevant code changes. |
| Long fuzz evidence | `docs/fuzz-evidence.md` records all 14 registered targets for `FUZZTIME=1h` on paired local ARM/Intel Go 1.26.3 runs for v0.1.2 package-code candidate `2e09774f171dde8c62763d6e35a258b0fef88801`; raw task logs are in `docs/evidence/v012-candidate-20260508/`. It also records supplemental `FUZZTIME=4h` soak evidence against signed tag `v0.1.2` commit `4e661bc1f925ebedf1f270668129d85bab73e468`: ARM passed all 14 targets, Intel had an all-target `FuzzProtocolConsistency` deadline failure and then a clean same-host targeted 4-hour rerun. Raw soak logs are in `docs/evidence/v012-soak-20260509/`. Earlier Go 1.26.3, Go 1.26.2, and older ARM/Intel runs remain historical evidence. | Repeat if parser, protocol, fuzz harness, dependency, or toolchain changes before release. |
| Security/spec audit | `docs/security-spec-audit.md` records review of `docs/security-assessment.md` and `docs/spec-matrix.md` against v0.1.2 package-code candidate `2e09774f171dde8c62763d6e35a258b0fef88801` under Go 1.26.3. | Repeat if protocol code, parser/framing code, package-profile docs, dependencies, toolchain, or the targeted draft revision changes. |
| Integration guidance | `docs/integration-guidance.md` documents outer PAKE/version negotiation, downgrade-protection, identity-orientation, and session-output guidance. | External reviewers should still evaluate whether this guidance is sufficient for real integrations. |
| Release validation and CI hardening | `v0.1.2` is a signed annotated prerelease tag at commit `4e661bc1f925ebedf1f270668129d85bab73e468`. Tag-triggered Release Validation passed `Check`, `Race`, `Govulncheck`, and `Gosec` with SARIF upload in workflow run `25588835119`. Public background signal also includes CodeQL, OpenSSF Scorecard, Staticcheck Advisory, Actionlint, cross-platform smoke, scheduled vulnerability scanning, scheduled gosec, and scheduled fuzz regression. | CI evidence supports auditable prerelease hygiene, not production readiness. Keep release tags signed, watch scheduled lanes, and keep external and cryptographic review as release blockers. |
| External review handoff | `docs/external-review-handoff.md` summarizes supported scope, package-owned choices, evidence, review questions, and remaining release blockers for external reviewers. | The handoff is a review input, not a completed review. Findings still need to be tracked and resolved. |
| Threat model | `docs/threat-model.md` records assets, in-scope attackers, non-goals, security boundaries, and reviewer focus areas. | This is a self-authored review input, not an external assessment. Reviewers should check that the model matches real integration risks. |
| Release checklist | `docs/release-checklist.md` records exact-candidate validation, evidence refresh, signed-tag, release-validation, and GitHub-release steps. | The checklist must be executed against a future candidate before making stronger release-readiness claims. |
| Capslock capability analysis | `docs/capslock-report.md` records Capslock `v0.3.2` results under Go 1.26.3 for v0.1.2 package-code candidate `2e09774f171dde8c62763d6e35a258b0fef88801`; raw transcript is in `docs/evidence/v012-candidate-20260508/`. | Capslock is experimental review signal, not a release gate. Repeat if dependencies, imports, randomness, HKDF/HMAC usage, or the Go toolchain change. |
| Performance benchmarks | `bench_test.go` and `task bench` cover full round trips, protocol phases, exporters, and message encoding/decoding with `-benchmem`. | Benchmark results are local comparison evidence, not release gates. Record host, Go version, exact command, and commit when sharing numbers. |
| OSS-Fuzz integration | `ossfuzz/` stages upstream project files for all 14 native Go fuzz targets. Local `build_fuzzers` and `check_build` validation passed with the repository mounted into a temporary `google/oss-fuzz` checkout on 2026-05-07. Upstream PR `google/oss-fuzz#15480` is open; CLA, header-check, and the upstream PR helper build passed on 2026-05-08. | Upstream onboarding still requires upstream review, merge, and follow-up monitoring after OSS-Fuzz starts running the project. |

## Release Readiness

Before any production-readiness claim:

- run every fuzz target for more than five minutes on release hardware or in
  the long-fuzz workflow;
- repeat dependency review with `govulncheck -test -show verbose ./...`;
- review `docs/security-assessment.md` and `docs/spec-matrix.md` against the
  exact release commit;
- execute the exact-candidate process in `docs/release-checklist.md`;
- complete external review of package-owned framing and profile choices;
- obtain independent cryptographic review.

## Later Investigation

- Longer continuous fuzzing campaigns.
- Offline Sage-derived extended vector dataset.
