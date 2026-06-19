# Project Plan

Status: living release-readiness plan after the policy/API decisions landed in
PRs #13-#17 and the public `v0.1.2` external-review/evidence snapshot.

This document tracks current work. Historical review triage remains in
`docs/interview-results-triage.md`.

## Current Phase

The current phase is release readiness. Public API and package-profile policy decisions are closed unless a new review finding reopens one. ADR-0008 records the narrow public-lifecycle thaw for `Initiator.Close` and `Responder.Close`; ADR-0009 records a broad Caller input replacement whose authorization is narrowly limited to its follow-up `Input` implementation. Do not describe the package as production-ready until the release bar below is satisfied and independent cryptographic review is complete.

## Release-Readiness PR Shape

Each release-readiness PR should include:

- the release-readiness gap being closed;
- the exact commit, command, workflow, or review artifact used as evidence;
- any residual risk or follow-up that remains after the PR;
- README, changelog, security, and spec documentation updates when release
  posture changes.
- no public API or package-profile changes except the ADR-0009 caller-input follow-up implementation already authorized by that accepted ADR; reopen the policy phase first if a new finding requires any other public API or package-profile change.

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
   PR #48 covers phase 1: a reusable evidence-bundle policy and cross-toolchain vector-stability checklist. The `f7efa6a963a954952b1ecad3f46530f13799fe89` evidence bundle applies that policy with committed raw artifacts, `SHA256SUMS`, and vector-stability results. Keep the issue open only if additional release-packet acceptance criteria remain outside this exact-candidate bundle.
3. Exact-candidate evidence refresh.
   The current exact-candidate dependency review, long fuzzing, Capslock, security/spec audit support, tag-ruleset capture, GitHub status, Scorecard, and vector-stability evidence are indexed in `docs/evidence-baseline.md`. Repeat those lanes after any review-driven or security-relevant changes before making a stronger readiness claim.

## Completed Evidence

Current pinned evidence baselines and freshness caveats are indexed in `docs/evidence-baseline.md`; the table below keeps the release-readiness map and historical completed-evidence context.

| Area | Evidence | Residual risk |
| --- | --- | --- |
| Dependency review | `docs/evidence-baseline.md` indexes the current pinned dependency, vulnerability, and SAST/gosec baseline; `docs/dependency-review.md` carries the lane-specific summary and raw transcript link. | Repeat on the exact release tag if dependencies, toolchain, parser/framing, protocol, security-relevant code, or package-profile docs change. |
| Long fuzz evidence | `docs/evidence-baseline.md` indexes the current pinned paired long-fuzz baseline; `docs/fuzz-evidence.md` carries the lane-specific summary, raw log links, historical prerelease soak, and interim non-evidence gates. | Repeat if parser, protocol, fuzz harness, dependency, or toolchain changes before release. |
| Security/spec audit | `docs/evidence-baseline.md` indexes the current pinned security/spec audit baseline; `docs/security-spec-audit.md` records the audited implementation baseline. | Repeat if protocol code, parser/framing code, package-profile docs, dependencies, toolchain, or the targeted draft revision changes. |
| Integration guidance | `docs/integration-guidance.md` documents outer PAKE/version negotiation, downgrade-protection, role-local identity input, and session-output guidance. | External reviewers should still evaluate whether this guidance is sufficient for real integrations. |
| Release validation and CI hardening | `v0.1.2` is a signed annotated prerelease tag at commit `4e661bc1f925ebedf1f270668129d85bab73e468`. Tag-triggered Release Validation passed `Check`, `Race`, `Govulncheck`, and `Gosec` with SARIF upload in workflow run `25588835119`. Public background signal also includes CodeQL, OpenSSF Scorecard, Staticcheck Advisory, Actionlint, cross-platform smoke, scheduled vulnerability scanning, scheduled gosec, and scheduled fuzz regression. | CI evidence supports auditable prerelease hygiene, not production readiness. Keep release tags signed, watch scheduled lanes, and keep external and cryptographic review as release blockers. |
| External review handoff | `docs/external-review-handoff.md` summarizes supported scope, package-owned choices, evidence, review questions, and remaining release blockers for external reviewers. | The handoff is a review input, not a completed review. Findings still need to be tracked and resolved. |
| Threat model | `docs/threat-model.md` records assets, in-scope attackers, non-goals, security boundaries, and reviewer focus areas. | This is a self-authored review input, not an external assessment. Reviewers should check that the model matches real integration risks. |
| Release checklist | `docs/release-checklist.md` records exact-candidate validation, evidence refresh, signed-tag, release-validation, and GitHub-release steps. | The checklist must be executed against a future candidate before making stronger release-readiness claims. |
| Capslock capability analysis | `docs/evidence-baseline.md` indexes the current pinned Capslock capability-analysis baseline; `docs/capslock-report.md` carries the lane-specific summary and triage. | Capslock is experimental review signal, not a release gate. Repeat if dependencies, imports, randomness, HKDF/HMAC usage, or the Go toolchain change. |
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
- Caller input field-policy concentration: issue #136 was evaluated after the caller-input follow-up coverage landed. Keep the current small `input.go` validation/copy/normalization functions until future caller-input changes create drift; a private field-policy catalogue is not worth adding now without a behavior-preserving simplification.
