# Project Plan

Status: living release-readiness plan after publication of the single-suite `v0.1.3` prerelease. The next maintainer decision is whether to resume ADR-0010 consideration for a possible `v0.2.0`; external and cryptographic review issues #29-#32 remain the separate production-readiness path.

This document tracks current work. Historical review triage remains in
`docs/interview-results-triage.md`.

## Current Phase

The single-suite `v0.1.3` prerelease was published from signed annotated tag commit `b4174c6bf4bae78f4081c3d6d6baebff5f1cbbf1`, descended from frozen clean candidate source `1f19e278112fa037890848ed6c086addeffdca4e`. [Release Validation run 30686073480](https://github.com/the-sarge/cpace/actions/runs/30686073480) and consumer-style tag, release, SBOM, checksum, attestation, and post-SARIF checks passed. Public API and package-profile policy decisions remain closed unless a new review finding reopens one. ADR-0008 records the narrow public-lifecycle thaw for `Initiator.Close` and `Responder.Close`; ADR-0009 records a broad Caller input replacement whose authorization is narrowly limited to its completed follow-up `Input` implementation. ADR-0010 consideration and monorepo proposal work remain deferred pending a separate maintainer decision; any later accepted migration would target `v0.2.0`. `v0.1.3` remains explicitly non-production-ready; external review issues #29-#31 and independent cryptographic review issue #32 remain blockers for a future production-readiness claim.

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
| Scalar sampling | Masked canonical 32-byte sampling with an all-zero-sample retry. | Done. Keep the draft-21 Ristretto255 recommendation; `SetUniformBytes` plus zero rejection/retry is an allowed alternative but would use 64-byte modulo reduction and define a different package profile. |

## Recommended PR Order

1. Keep the frozen `v0.1.3` candidate source at `1f19e278112fa037890848ed6c086addeffdca4e`. PR #227 is included through merge commit `6be725fe617b2ad47fd260f39382d000c438e292`, and issue #231 pins the local and CI toolchain to the Go 1.26.5 security release without changing the Go 1.26 language version. Admit no additional public API, observable behavior, dependency, protocol, parser/framing, or package-profile work without an explicit policy reopen; any security-relevant change selects a new candidate and restarts affected evidence lanes.
2. Keep release-facing documentation led by the breaking pre-v1 `Config` → `Input` migration, exported `Suite` removal, and nil-safe `Session.Close` change, followed by the additive lifecycle and peer-share error APIs. State explicitly that the wire format and single draft-21 suite are unchanged.
3. Keep the assembled issue #237 packet at `docs/evidence/1f19e27-20260801-candidate/` aligned with issue #33. The dependency/govulncheck, test-inclusive gosec, Capslock, security/spec audit, Go 1.26.4 to Go 1.26.5 vector-stability, paired long-fuzz, pre-publication GitHub release-control, final pre-tag gate, signed-tag Release Validation, and post-SARIF verification lanes are current for the published candidate contract.
4. Preserve signed tag `v0.1.3`, [the published prerelease](https://github.com/the-sarge/cpace/releases/tag/v0.1.3), and [Release Validation run 30686073480](https://github.com/the-sarge/cpace/actions/runs/30686073480) as the immutable publication pointers; retain issues #29-#32 for the future production-readiness path.
5. Keep ADR-0010 consideration and monorepo work outside this branch until the maintainer makes the separate post-`v0.1.3` decision.

## Completed Evidence

Current pinned evidence baselines and freshness caveats are indexed in `docs/evidence-baseline.md`; the table below keeps the release-readiness map and historical completed-evidence context.

| Area | Evidence | Residual risk |
| --- | --- | --- |
| Dependency review | `docs/evidence-baseline.md` indexes the exact-candidate Go 1.26.5 dependency, vulnerability, and SAST/gosec baseline; `docs/dependency-review.md` carries the results, disposition, residual limitations, and raw transcript link. | Current for frozen candidate `1f19e278112fa037890848ed6c086addeffdca4e`; repeat if a dependency, toolchain, parser/framing, protocol, security-relevant code, or package-profile change selects a new candidate. |
| Long fuzz evidence | `docs/evidence-baseline.md` indexes the exact-candidate Go 1.26.5 paired long-fuzz baseline; `docs/fuzz-evidence.md` carries the lane-specific summary, raw log links, historical prerelease soak, and interim non-evidence gates. | Current for frozen candidate `1f19e278112fa037890848ed6c086addeffdca4e` with a recorded Intel all-target deadline non-pass recovered by a same-host targeted rerun; repeat if a parser, protocol, fuzz harness, dependency, or toolchain change selects a new candidate. |
| Security/spec audit | `docs/evidence-baseline.md` indexes the exact-candidate Go 1.26.5 audit and Go 1.26.4 comparison; `docs/security-spec-audit.md` records the current API, lifecycle, peer-share, Message framing, and package-profile results, and `docs/evidence/1f19e27-20260731-protocol/` preserves the raw transcript and checksums. | Current for frozen candidate `1f19e278112fa037890848ed6c086addeffdca4e`; this remains project-side self-review, not independent cryptographic review. Repeat after any listed audit refresh trigger. |
| GitHub release controls | `docs/evidence/1f19e27-20260801-github/` preserves checksummed tag-ruleset, branch-protection, required-check, alert, Scorecard, and scheduled-workflow state captured at protected `main` `fa4ea296daae82e920ae6bb410e9bcb7c0061641`, including Nightly Fuzz failure/recovery and autoscaled queue history. The issue #238 publication record adds the authorized break-glass interval and verified restoration of ruleset `16048307`. | Current point-in-time prerelease signal for the published candidate contract, not exact-candidate execution for scheduled runs on older SHAs. The source candidate is an intermediate PR #239 commit without direct check suites; all required contexts passed on its checked documentation-only descendant. Recapture mutable GitHub state before a future release claim. |
| Assembled candidate packet | `docs/evidence/1f19e27-20260801-candidate/` hash-links every current `v0.1.3` evidence lane, preserves the extracted release notes, and records the full pre-tag gate against a clean documentation-only descendant of the frozen source; issue #238 records the post-tag validation and consumer verification. | The packet and publication record are prerelease self-evidence, not independent review. The Intel long-fuzz deadline miss/targeted recovery and mutable GitHub-state limitations remain explicit. |
| Integration guidance | `docs/integration-guidance.md` documents outer PAKE/version negotiation, downgrade-protection, role-local identity input, and session-output guidance. | External reviewers should still evaluate whether this guidance is sufficient for real integrations. |
| Release validation and CI hardening | `v0.1.3` is a signed annotated prerelease tag at commit `b4174c6bf4bae78f4081c3d6d6baebff5f1cbbf1`. Tag-triggered Release Validation passed signed-tag verification, `Check`, `Race`, `Govulncheck`, `Gosec` with SARIF upload, SBOM generation, SBOM attestation, and release publication in run `30686073480`; consumer verification confirmed the prerelease/not-latest posture, release notes, SBOM checksum, Sigstore bundle, attestation, fresh remote tag signature, and empty post-SARIF alert set. | CI and consumer evidence support auditable prerelease hygiene, not production readiness. Keep release tags signed, watch scheduled lanes, and keep external and cryptographic review as production-readiness blockers. |
| External review handoff | `docs/external-review-handoff.md` summarizes supported scope, package-owned choices, evidence, review questions, and remaining release blockers for external reviewers. | The handoff is a review input, not a completed review. Findings still need to be tracked and resolved. |
| Threat model | `docs/threat-model.md` records assets, in-scope attackers, non-goals, security boundaries, and reviewer focus areas. | This is a self-authored review input, not an external assessment. Reviewers should check that the model matches real integration risks. |
| Release checklist | `docs/release-checklist.md` records exact-candidate validation, evidence refresh, signed-tag, release-validation, and GitHub-release steps. | The checklist must be executed against a future candidate before making stronger release-readiness claims. |
| Capslock capability analysis | `docs/evidence-baseline.md` indexes the exact-candidate Go 1.26.5 Capslock baseline; `docs/capslock-report.md` carries the unchanged-class/count result and triage, and `docs/evidence/1f19e27-20260731-protocol/` preserves raw normal/verbose output and checksums. | Current for frozen candidate `1f19e278112fa037890848ed6c086addeffdca4e`. Capslock remains experimental review signal, not a release gate; repeat after a capability refresh trigger. |
| Performance benchmarks | `bench_test.go` and `task bench` cover full round trips, protocol phases, exporters, and message encoding/decoding with `-benchmem`. | Benchmark results are local comparison evidence, not release gates. Record host, Go version, exact command, and commit when sharing numbers. |
| OSS-Fuzz-compatible staging | `ossfuzz/` keeps OSS-Fuzz-compatible project files for all 14 native Go fuzz targets. Local `build_fuzzers` and `check_build` validation passed with the repository mounted into a temporary `google/oss-fuzz` checkout on 2026-05-07; upstream PR `google/oss-fuzz#15480` was closed on 2026-05-11 after OSS-Fuzz maintainers declined the project for current project-size/user-base reasons and suggested ClusterFuzzLite instead. | Treat upstream OSS-Fuzz as unavailable for current release-readiness claims. Keep maintainer-controlled long fuzzing, scheduled hosted fuzzing, and autoscaled trusted-runner fuzzing as the active continuous-fuzz signal; consider ClusterFuzzLite separately if the project wants an OSS-Fuzz-adjacent CI service. |

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
