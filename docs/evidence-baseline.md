# Evidence Baseline

Status: current release-evidence index, not a production-readiness claim.

This document is the module for pinned evidence baselines. It names the commit, toolchain, raw artifacts, summary docs, and freshness caveats that release-readiness docs cite. Updating evidence for a candidate starts here, then updates the summary docs named below.

## Terms

- **Pinned baseline**: exact commit, workflow run, or transcript bundle that supports an existing evidence claim.
- **Current candidate**: exact commit being considered for a new prerelease or production-readiness claim. It can be newer than the pinned baseline.
- **Fresh evidence**: evidence run against the current candidate, or an immutable workflow artifact tied to it.

## Current Release-Claim State

The frozen `v0.1.3` candidate source baseline is `1f19e278112fa037890848ed6c086addeffdca4e`, selected from a clean worktree under Go 1.26.5 after the PR #227 disposition and toolchain update. The dependency, vulnerability, SAST/gosec, Capslock capability, security/spec audit, Go 1.26.4 to Go 1.26.5 vector-stability, paired long-fuzz, and pre-publication GitHub release-control lanes are current for that candidate contract; naming the SHA alone does not make the future signed-tag Release Validation lane fresh.

The current dependency, vulnerability, and SAST/gosec evidence is the clean-worktree Go 1.26.5 capture in `docs/evidence/1f19e27-20260731/`. Module integrity passed, `govulncheck@v1.3.0` reported no vulnerabilities, and the test-inclusive pinned `gosec@v2.26.1` scan reported zero issues. The current protocol bundle is `docs/evidence/1f19e27-20260731-protocol/`: Capslock classes/counts are unchanged, the focused protocol audit passed, and the normalized 112-line vector event streams are bit-identical under Go 1.26.4 and Go 1.26.5 with shared SHA-256 `05e966e9b4cc8ea6883d5f8a750cc2d91098a3e59aee365ed852e717067efbd9`. These completed lanes are prerelease evidence, not a production-readiness claim.

The current paired long-fuzz evidence is the Go 1.26.5 maintainer-machine bundle in `docs/evidence/1f19e27-20260731-fuzz/`. The ARM all-target campaign passed all 14 registered targets; the Intel all-target campaign ran all 14 registered targets, ended nonzero on a `FuzzMessageARoundTrip` one-hour deadline miss with no new fuzz artifacts, and then passed a same-host one-hour targeted rerun of `FuzzMessageARoundTrip` with no new fuzz artifacts. Treat the lane as current exact-candidate fuzz evidence with that recorded limitation, not as two clean all-target passes.

The current GitHub release-control capture is `docs/evidence/1f19e27-20260801-github/`, observed from 2026-08-01 02:54:40 through 02:55:00 UTC at protected `main` `fa4ea296daae82e920ae6bb410e9bcb7c0061641`. Tag ruleset `16048307` was active for creation, update, and deletion of `refs/tags/v*` with no bypass actors and `current_user_can_bypass: never`; strict branch protection required exactly `Check`, `DCO`, `Dependency Gate`, and `SAST Gate`; open Code Scanning, Dependabot, and secret-scanning alert responses were empty; and the current Scorecard result was 7.4. The bundle also records recent vulnerability, SAST, cross-platform, Nightly Fuzz, and Autoscaled Fuzz status, including a Nightly Fuzz deadline failure recovered by a successful failed-job rerun and recurring autoscaled arm64 queue availability limits.

The frozen source SHA is an intermediate commit in PR #239 and has no direct GitHub check suites, runs, or statuses. PR #239's checked head `ff73e9ebcbdabfc949e8905caba526cec5057f38` contains that source commit plus policy/documentation files only, and all four required contexts passed on that checked descendant. Preserve that limitation rather than claiming the intermediate SHA itself carried required checks.

The historical `f7efa6a963a954952b1ecad3f46530f13799fe89` packet remains historical signal for its older Go 1.26.4 candidate. Its tag-ruleset, GitHub status, Scorecard, and cross-toolchain captures are superseded for `v0.1.3` by the current source-specific and GitHub release-control bundles above.

Before publication, complete the signed-tag Release Validation and post-release verification lane under issue #237. Recapture GitHub release-control state if an administrator changes repository policy or alerts, a new finding appears, or the delay before tagging makes the 2026-08-01 point-in-time state too stale for the release claim. Keep the completed dependency/SAST, protocol/capability/vector, paired long-fuzz, and release-control bundles pinned unless their stated refresh triggers select a new candidate or invalidate a mutable-state capture.

## Baseline Index

| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |
| --- | --- | --- | --- | --- |
| Dependency, vulnerability, and SAST/gosec review | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 | `docs/evidence/1f19e27-20260731/local-analysis.log`, `docs/evidence/1f19e27-20260731/capture.sh`, `docs/evidence/1f19e27-20260731/SHA256SUMS` | `docs/dependency-review.md` | Repeat when dependencies, Go toolchain, parser/framing, protocol, security-relevant code, or package-profile docs change before a stronger release claim. |
| Capslock capability analysis | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 | `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`, `docs/evidence/1f19e27-20260731-protocol/capture.sh`, `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS` | `docs/capslock-report.md` | Repeat when dependencies, imports, randomness handling, HKDF/HMAC usage, or Go toolchain change. Treat new broad capability classes as external-review findings. |
| Security/spec audit and vector stability | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 with Go 1.26.4 comparison | `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`, `docs/evidence/1f19e27-20260731-protocol/capture.sh`, `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS`, plus the audited docs named in `docs/security-spec-audit.md` | `docs/security-spec-audit.md`, `docs/security-assessment.md`, `docs/spec-matrix.md` | Repeat when protocol code, parser/framing code, package-profile docs, dependencies, toolchain, evidence-sensitive tooling, or targeted draft revision change. |
| Paired long fuzzing | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5; ARM all-target pass and Intel all-target deadline miss with same-host targeted recovery | `docs/evidence/1f19e27-20260731-fuzz/m1mini-campaign.log`, `docs/evidence/1f19e27-20260731-fuzz/imacpro-campaign.log`, `docs/evidence/1f19e27-20260731-fuzz/imacpro-FuzzMessageARoundTrip-targeted-rerun.log`, final status captures, and `docs/evidence/1f19e27-20260731-fuzz/SHA256SUMS` | `docs/fuzz-evidence.md` | Repeat when parser, protocol, fuzz harness, dependency, or toolchain changes before a stronger release claim. |
| Historical `v0.1.2` prerelease validation and soak | Signed tag `v0.1.2` at `4e661bc1f925ebedf1f270668129d85bab73e468` | `docs/evidence/v012-candidate-20260508/`, `docs/evidence/v012-soak-20260509/`, Release Validation run `25588835119` | `docs/project-plan.md`, `docs/external-review-handoff.md`, `docs/fuzz-evidence.md` | Historical prerelease evidence only. Do not use it as current exact-candidate evidence for newer commits. |
| GitHub release controls and scheduled signal | 2026-08-01 02:54:40–02:55:00 UTC; live `main` `fa4ea296daae82e920ae6bb410e9bcb7c0061641`, frozen source candidate `1f19e278112fa037890848ed6c086addeffdca4e` | `docs/evidence/1f19e27-20260801-github/`, including raw ruleset, branch-protection, check, alert, Scorecard, workflow/run/job responses, failure/recovery logs, and `SHA256SUMS` | `docs/release-checklist.md`, `docs/ci-policy.md`, `docs/v0.1.3-release-plan.md`, `docs/project-plan.md`, `docs/external-review-handoff.md` | Recapture before release if mutable GitHub policy, alert, or workflow state changes or the point-in-time capture becomes stale; do not treat scheduled runs on older SHAs as exact-candidate evidence. |

## Refresh Procedure

When refreshing evidence for a candidate:

1. Identify the exact candidate commit and update this module first.
2. Preserve raw logs or immutable workflow links according to `docs/evidence/README.md`.
3. Update the lane-specific summary docs named in the Baseline Index.
4. Regenerate `docs/evidence-baseline-summary-docs.txt` with `(cd tools/evidencebaseline && go run . --repo-root ../.. --write-summary-docs)`; the CI change classifier reads this generated adapter before Go is set up, while this module remains the source of truth.
5. Update `docs/project-plan.md` and `docs/external-review-handoff.md` only when the release-readiness posture or external-review packet changes.
6. Keep superseded baselines visible until the new evidence has its own raw artifacts, checksums or immutable workflow links, and residual-risk wording.
