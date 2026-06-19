# Evidence Baseline

Status: current release-evidence index, not a production-readiness claim.

This document is the module for pinned evidence baselines. It names the commit, toolchain, raw artifacts, summary docs, and freshness caveats that release-readiness docs cite. Updating evidence for a candidate starts here, then updates the summary docs named below.

## Terms

- **Pinned baseline**: exact commit, workflow run, or transcript bundle that supports an existing evidence claim.
- **Current candidate**: exact commit being considered for a new prerelease or production-readiness claim. It can be newer than the pinned baseline.
- **Fresh evidence**: evidence run against the current candidate, or an immutable workflow artifact tied to it.

## Current Release-Claim State

The current strongest release evidence remains exact-candidate prerelease evidence plus historical external-review evidence, not a production-readiness claim. The current pinned package-code evidence baseline for dependency review, SAST/gosec, Capslock, security/spec audit, and paired long fuzzing is `f7efa6a963a954952b1ecad3f46530f13799fe89`, captured under Go 1.26.4 in `docs/evidence/f7efa6a-20260619/`.

No package-code changes are included after this baseline in the current evidence index. Do not describe any newer commit as release-current on dependency, fuzz, Capslock, or security/spec evidence until those evidence lanes are refreshed at the exact candidate commit.

This baseline includes the accepted-ADR implementation sequence (ADR-0003, ADR-0001, ADR-0002, ADR-0009), issue #80's responder decoded-share reuse, PR #199's Go fix modernization, and PR #200's development-journal update. It also includes a fresh tag-ruleset capture, candidate GitHub status capture, fresh Scorecard run, and cross-toolchain vector-stability check in the same evidence bundle.

## Baseline Index

| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |
| --- | --- | --- | --- | --- |
| Dependency, vulnerability, and SAST/gosec review | `f7efa6a963a954952b1ecad3f46530f13799fe89`, Go 1.26.4 | `docs/evidence/f7efa6a-20260619/local-analysis.log`, `docs/evidence/f7efa6a-20260619/SHA256SUMS` | `docs/dependency-review.md` | Repeat when dependencies, Go toolchain, parser/framing, protocol, security-relevant code, or package-profile docs change before a stronger release claim. |
| Capslock capability analysis | `f7efa6a963a954952b1ecad3f46530f13799fe89`, Go 1.26.4 | `docs/evidence/f7efa6a-20260619/local-analysis.log`, `docs/evidence/f7efa6a-20260619/SHA256SUMS` | `docs/capslock-report.md` | Repeat when dependencies, imports, randomness handling, HKDF/HMAC usage, or Go toolchain change. Treat new broad capability classes as external-review findings. |
| Security/spec audit | `f7efa6a963a954952b1ecad3f46530f13799fe89`, Go 1.26.4 | `docs/evidence/f7efa6a-20260619/local-analysis.log`, `docs/evidence/f7efa6a-20260619/vector-stability.log`, `docs/evidence/f7efa6a-20260619/SHA256SUMS`, plus the audited docs named in `docs/security-spec-audit.md` | `docs/security-spec-audit.md`, `docs/security-assessment.md`, `docs/spec-matrix.md` | Repeat when protocol code, parser/framing code, package-profile docs, dependencies, toolchain, or targeted draft revision change. |
| Paired long fuzzing | `f7efa6a963a954952b1ecad3f46530f13799fe89`, Go 1.26.4 | `docs/evidence/f7efa6a-20260619/fuzz-m1mini.log`, `docs/evidence/f7efa6a-20260619/fuzz-imacpro.log`, final status captures, and `docs/evidence/f7efa6a-20260619/SHA256SUMS` | `docs/fuzz-evidence.md` | Repeat when parser, protocol, fuzz harness, dependency, or toolchain changes before a stronger release claim. |
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
