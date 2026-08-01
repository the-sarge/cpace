# Evidence Baseline

Status: current release-evidence index, not a production-readiness claim.

This document is the module for pinned evidence baselines. It names the commit, toolchain, raw artifacts, summary docs, and freshness caveats that release-readiness docs cite. Updating evidence for a candidate starts here, then updates the summary docs named below.

## Terms

- **Pinned baseline**: exact commit, workflow run, or transcript bundle that supports an existing evidence claim.
- **Current candidate**: exact commit being considered for a new prerelease or production-readiness claim. It can be newer than the pinned baseline.
- **Fresh evidence**: evidence run against the current candidate, or an immutable workflow artifact tied to it.

## Current Release-Claim State

The frozen `v0.1.3` candidate source baseline is `1f19e278112fa037890848ed6c086addeffdca4e`, selected from a clean worktree under Go 1.26.5 after the PR #227 disposition and toolchain update. The dependency, vulnerability, SAST/gosec, Capslock capability, security/spec audit, Go 1.26.4 to Go 1.26.5 vector-stability, and paired long-fuzz lanes are fresh for that exact candidate; naming the SHA alone does not make the remaining lanes fresh.

The current dependency, vulnerability, and SAST/gosec evidence is the clean-worktree Go 1.26.5 capture in `docs/evidence/1f19e27-20260731/`. Module integrity passed, `govulncheck@v1.3.0` reported no vulnerabilities, and the test-inclusive pinned `gosec@v2.26.1` scan reported zero issues. The current protocol bundle is `docs/evidence/1f19e27-20260731-protocol/`: Capslock classes/counts are unchanged, the focused protocol audit passed, and the normalized 112-line vector event streams are bit-identical under Go 1.26.4 and Go 1.26.5 with shared SHA-256 `05e966e9b4cc8ea6883d5f8a750cc2d91098a3e59aee365ed852e717067efbd9`. These completed lanes are prerelease evidence, not a production-readiness claim.

The current paired long-fuzz evidence is the Go 1.26.5 maintainer-machine bundle in `docs/evidence/1f19e27-20260731-fuzz/`. The ARM all-target campaign passed all 14 registered targets; the Intel all-target campaign ran all 14 registered targets, ended nonzero on a `FuzzMessageARoundTrip` one-hour deadline miss with no new fuzz artifacts, and then passed a same-host one-hour targeted rerun of `FuzzMessageARoundTrip` with no new fuzz artifacts. Treat the lane as current exact-candidate fuzz evidence with that recorded limitation, not as two clean all-target passes.

The historical exact-candidate packet at `f7efa6a963a954952b1ecad3f46530f13799fe89`, captured under Go 1.26.4 in `docs/evidence/f7efa6a-20260619/`, remains the strongest completed bundle for tag-ruleset state, GitHub status, and Scorecard. Do not treat those lanes as current for `v0.1.3`; its Capslock, security/spec, vector, and paired long-fuzz artifacts are superseded by the exact-candidate protocol and fuzz bundles for those specific lanes.

The historical `f7efa6a963a954952b1ecad3f46530f13799fe89` packet covers the accepted-ADR implementation sequence (ADR-0003, ADR-0001, ADR-0002, ADR-0009), issue #80's responder decoded-share reuse, PR #199's Go fix modernization, and PR #200's development-journal update. It also contains the tag-ruleset, candidate GitHub status, Scorecard, and cross-toolchain vector-stability captures made for that packet; all of those artifacts are historical for frozen `v0.1.3` candidate source `1f19e278112fa037890848ed6c086addeffdca4e`.

Before the `v0.1.3` candidate gate, refresh tag-ruleset state, GitHub alert state, Scorecard, and release-validation evidence against the same exact candidate commit. Keep the completed dependency/SAST, protocol/capability/vector, and paired long-fuzz lanes pinned to their raw checked bundles unless a refresh trigger selects a new candidate.

## Baseline Index

| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |
| --- | --- | --- | --- | --- |
| Dependency, vulnerability, and SAST/gosec review | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 | `docs/evidence/1f19e27-20260731/local-analysis.log`, `docs/evidence/1f19e27-20260731/capture.sh`, `docs/evidence/1f19e27-20260731/SHA256SUMS` | `docs/dependency-review.md` | Repeat when dependencies, Go toolchain, parser/framing, protocol, security-relevant code, or package-profile docs change before a stronger release claim. |
| Capslock capability analysis | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 | `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`, `docs/evidence/1f19e27-20260731-protocol/capture.sh`, `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS` | `docs/capslock-report.md` | Repeat when dependencies, imports, randomness handling, HKDF/HMAC usage, or Go toolchain change. Treat new broad capability classes as external-review findings. |
| Security/spec audit and vector stability | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5 with Go 1.26.4 comparison | `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`, `docs/evidence/1f19e27-20260731-protocol/capture.sh`, `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS`, plus the audited docs named in `docs/security-spec-audit.md` | `docs/security-spec-audit.md`, `docs/security-assessment.md`, `docs/spec-matrix.md` | Repeat when protocol code, parser/framing code, package-profile docs, dependencies, toolchain, evidence-sensitive tooling, or targeted draft revision change. |
| Paired long fuzzing | `1f19e278112fa037890848ed6c086addeffdca4e`, Go 1.26.5; ARM all-target pass and Intel all-target deadline miss with same-host targeted recovery | `docs/evidence/1f19e27-20260731-fuzz/m1mini-campaign.log`, `docs/evidence/1f19e27-20260731-fuzz/imacpro-campaign.log`, `docs/evidence/1f19e27-20260731-fuzz/imacpro-FuzzMessageARoundTrip-targeted-rerun.log`, final status captures, and `docs/evidence/1f19e27-20260731-fuzz/SHA256SUMS` | `docs/fuzz-evidence.md` | Repeat when parser, protocol, fuzz harness, dependency, or toolchain changes before a stronger release claim. |
| Historical `v0.1.2` prerelease validation and soak | Signed tag `v0.1.2` at `4e661bc1f925ebedf1f270668129d85bab73e468` | `docs/evidence/v012-candidate-20260508/`, `docs/evidence/v012-soak-20260509/`, Release Validation run `25588835119` | `docs/project-plan.md`, `docs/external-review-handoff.md`, `docs/fuzz-evidence.md` | Historical prerelease evidence only. Do not use it as current exact-candidate evidence for newer commits. |
| Tag-authority ruleset capture | 2026-06-19 GitHub ruleset state | `docs/evidence/f7efa6a-20260619/rulesets-list.json`, `docs/evidence/f7efa6a-20260619/ruleset-16048307.json`, `docs/evidence/f7efa6a-20260619/ruleset-16048307-verify.json`, `docs/evidence/f7efa6a-20260619/SHA256SUMS` | `docs/release-checklist.md`, `docs/ci-policy.md` | Recapture before each release because GitHub repository ruleset state is admin-mutable. |

## Refresh Procedure

When refreshing evidence for a candidate:

1. Identify the exact candidate commit and update this module first.
2. Preserve raw logs or immutable workflow links according to `docs/evidence/README.md`.
3. Update the lane-specific summary docs named in the Baseline Index.
4. Regenerate `docs/evidence-baseline-summary-docs.txt` with `(cd tools/evidencebaseline && go run . --repo-root ../.. --write-summary-docs)`; the CI change classifier reads this generated adapter before Go is set up, while this module remains the source of truth.
5. Update `docs/project-plan.md` and `docs/external-review-handoff.md` only when the release-readiness posture or external-review packet changes.
6. Keep superseded baselines visible until the new evidence has its own raw artifacts, checksums or immutable workflow links, and residual-risk wording.
