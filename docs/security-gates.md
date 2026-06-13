# Security Gates Policy

This policy defines the threshold for SCA and SAST findings, how violations are
handled before release, and which automation currently enforces or reports
those checks.

## Current Automation

Pull-request gates:

- required PR `Check` runs tests for code changes and docs validation for
  docs-only changes;
- DCO validation checks every PR commit for a `Signed-off-by` trailer;
- `Dependency Gate` runs SCA tooling on pull requests and must pass before
  merge once the status check is enabled in branch protection;
- `SAST Gate` runs SAST tooling on pull requests and must pass before merge
  once the status check is enabled in branch protection.

Additional automated signal:

- CodeQL runs on Go pull requests, pushes to `main`, schedule, and manual
  dispatch;
- Staticcheck Advisory runs on Go pull requests, schedule, and manual dispatch;
- Vulnerability Scan runs `govulncheck` on schedule and manual dispatch;
- Gosec Advisory runs on schedule and manual dispatch and uploads SARIF to Code
  Scanning;
- Release Validation verifies signed annotated tag refs, runs tests, race tests, `govulncheck`, and `gosec`, generates and validates the CycloneDX SBOM, emits the SBOM attestation bundle, and publishes the GitHub Release only for `v*` tag pushes; manual dispatch is a non-publishing tag rehearsal, and branch dispatch fails closed with an explanatory job.

The `Dependency Gate` and `SAST Gate` workflows intentionally do not use
workflow path filters. Required checks must report a status on every pull
request; otherwise a skipped required workflow can block review with no result.

## SCA Threshold

SCA covers dependency vulnerability, malicious-dependency, and license findings
from `govulncheck`, GitHub Dependency Review, Dependabot/GitHub dependency
alerts, manual dependency review, and any future dependency-review workflow.

The following findings are violations:

- any reachable vulnerability reported by `govulncheck`;
- any critical or high severity vulnerability in a direct or transitive
  dependency, unless it is declared non-exploitable in `docs/vex.md`;
- any dependency believed to be malicious, typosquatted, abandoned in a way that
  creates security risk, or unexpectedly introduced;
- any dependency license that is unknown or incompatible with the project's
  BSD-3-Clause distribution goals.

Violations should be fixed by upgrading, replacing, removing, or isolating the
dependency. If a vulnerability does not affect this project, record the
non-exploitability rationale in `docs/vex.md` instead of silently ignoring it.

The required `Dependency Gate` status check runs GitHub Dependency Review,
`go mod verify`, and `govulncheck -test ./...` on pull requests. It blocks
newly introduced high or critical vulnerable dependencies, disallowed licenses
for new dependencies, module download-integrity failures, and reachable
vulnerabilities found by `govulncheck`.

## SAST Threshold

SAST covers CodeQL, `gosec`, ast-grep security rules, and manual review of
security-sensitive code paths. Staticcheck is treated as quality and
maintainability signal unless a maintainer determines that a finding affects
security behavior.

The following findings are violations:

- any CodeQL high or critical severity security alert;
- any CodeQL medium severity alert in CPace protocol, parser/framing,
  randomness, key-derivation, session-lifecycle, dependency, or release-process
  code;
- any `gosec` finding in package code, tests, examples, or release workflows
  unless reviewed and documented as a false positive or non-exploitable;
- any ast-grep rule violation;
- any manual review finding that could affect authentication, key derivation,
  message parsing, context binding, release integrity, or secret handling.

Violations should be fixed before merge when they affect the changed code, and
must be fixed before release unless they are documented as false positives or
non-exploitable. Suppressions should be narrow, reviewable, and linked to the
evidence explaining why they are safe.

The required `SAST Gate` status check runs `gosec -tests ./...` on pull
requests and blocks merge when gosec reports a security weakness. Same-repo
runs also upload SARIF to Code Scanning for triage; fork PRs skip SARIF upload
to avoid granting write permissions to untrusted code.

## Pre-Release Policy

Before any release tag, apply the release checklist and confirm that:

- `govulncheck` and `gosec` pass in Release Validation;
- signed-tag verification, SBOM generation/validation, SBOM attestation bundle preparation, and release publishing gates pass in Release Validation;
- Code Scanning has no unexpected open CodeQL or gosec alerts;
- advisory security scans have no unresolved violations under this policy;
- dependency-review evidence is current for the release candidate;
- any non-affecting vulnerability is accounted for in `docs/vex.md`.

Do not publish a release with unresolved SCA or SAST violations unless a
reviewable VEX or suppression rationale explains why the finding does not
affect the release.
