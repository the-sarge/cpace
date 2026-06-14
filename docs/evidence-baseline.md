# Evidence Baseline

Status: current release-evidence index, not a production-readiness claim.

This document is the module for pinned evidence baselines. It names the commit, toolchain, raw artifacts, summary docs, and freshness caveats that release-readiness docs cite. Updating evidence for a candidate starts here, then updates the summary docs named below.

## Terms

- **Pinned baseline**: exact commit, workflow run, or transcript bundle that supports an existing evidence claim.
- **Current candidate**: exact commit being considered for a new prerelease or production-readiness claim. It can be newer than the pinned baseline.
- **Fresh evidence**: evidence run against the current candidate, or an immutable workflow artifact tied to it.

## Current Release-Claim State

The current strongest release evidence remains historical external-review evidence, not a production-readiness claim. The current pinned package-code evidence baseline for dependency review, SAST/gosec, Capslock, security/spec audit, and paired long fuzzing is `933ece246e6170b11e838395bf36f852cba0cd02`, captured under Go 1.26.4 in `docs/evidence/go1264-20260611/`.

Security-relevant package-code changes landed after that baseline, including ADR-0003 and the later accepted-ADR/architecture deepening sequence through PR #104. Do not describe any newer commit as release-current on dependency, fuzz, Capslock, or security/spec evidence until those evidence lanes are refreshed at the exact candidate commit.

The latest merged code candidate when this module was introduced is PR #104 merge `aa3b30fe6f895655d2d2259e9e1e62c3ad34dc97`. Treat that commit as an evidence-refresh target, not a refreshed evidence baseline.

## Baseline Index

| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |
| --- | --- | --- | --- | --- |
| Dependency, vulnerability, and SAST/gosec review | `933ece246e6170b11e838395bf36f852cba0cd02`, Go 1.26.4 | `docs/evidence/go1264-20260611/local-analysis.log` | `docs/dependency-review.md` | Repeat when dependencies, Go toolchain, parser/framing, protocol, security-relevant code, or package-profile docs change before a stronger release claim. |
| Capslock capability analysis | `933ece246e6170b11e838395bf36f852cba0cd02`, Go 1.26.4 | `docs/evidence/go1264-20260611/local-analysis.log` | `docs/capslock-report.md` | Repeat when dependencies, imports, randomness handling, HKDF/HMAC usage, or Go toolchain change. Treat new broad capability classes as external-review findings. |
| Security/spec audit | `933ece246e6170b11e838395bf36f852cba0cd02`, Go 1.26.4 | `docs/evidence/go1264-20260611/local-analysis.log` plus the audited docs named in `docs/security-spec-audit.md` | `docs/security-spec-audit.md`, `docs/security-assessment.md`, `docs/spec-matrix.md` | Repeat when protocol code, parser/framing code, package-profile docs, dependencies, toolchain, or targeted draft revision change. |
| Paired long fuzzing | `933ece246e6170b11e838395bf36f852cba0cd02`, Go 1.26.4 | `docs/evidence/go1264-20260611/fuzz-mbp128.log`, `docs/evidence/go1264-20260611/fuzz-imacpro.log`, status captures, and `SHA256SUMS` | `docs/fuzz-evidence.md` | Repeat when parser, protocol, fuzz harness, dependency, or toolchain changes before a stronger release claim. |
| Historical `v0.1.2` prerelease validation and soak | Signed tag `v0.1.2` at `4e661bc1f925ebedf1f270668129d85bab73e468` | `docs/evidence/v012-candidate-20260508/`, `docs/evidence/v012-soak-20260509/`, Release Validation run `25588835119` | `docs/project-plan.md`, `docs/external-review-handoff.md`, `docs/fuzz-evidence.md` | Historical prerelease evidence only. Do not use it as current exact-candidate evidence for newer commits. |
| Tag-authority ruleset capture | 2026-06-10 GitHub ruleset state | `docs/evidence/tagruleset-20260610/` | `docs/release-checklist.md`, `docs/ci-policy.md` | Recapture before each release because GitHub repository ruleset state is admin-mutable. |

## Refresh Procedure

When refreshing evidence for a candidate:

1. Identify the exact candidate commit and update this module first.
2. Preserve raw logs or immutable workflow links according to `docs/evidence/README.md`.
3. Update the lane-specific summary docs named in the Baseline Index.
4. Update `docs/project-plan.md` and `docs/external-review-handoff.md` only when the release-readiness posture or external-review packet changes.
5. Keep superseded baselines visible until the new evidence has its own raw artifacts, checksums or immutable workflow links, and residual-risk wording.
