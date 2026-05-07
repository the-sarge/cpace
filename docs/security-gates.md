# Security Gates Policy

This policy defines the threshold for SCA and SAST findings, how violations are
handled before release, and which automation currently enforces or reports
those checks.

## Current Automation

Current automated signal:

- required PR `Check` runs tests for code changes and docs validation for
  docs-only changes;
- DCO validation checks every PR commit for a `Signed-off-by` trailer;
- CodeQL runs on Go pull requests, pushes to `main`, schedule, and manual
  dispatch;
- Staticcheck Advisory runs on Go pull requests, schedule, and manual dispatch;
- Vulnerability Scan runs `govulncheck` on schedule and manual dispatch;
- Gosec Advisory runs on schedule and manual dispatch and uploads SARIF to Code
  Scanning;
- Release Validation runs tests, race tests, `govulncheck`, and `gosec` for
  `v*` tags and manual dispatch.

The repository does not yet claim that every codebase change is blocked by a
dedicated SCA/SAST security gate. Promoting those checks to required PR gates is
a separate branch-protection and workflow decision.

## SCA Threshold

SCA covers dependency vulnerability, malicious-dependency, and license findings
from `govulncheck`, Dependabot/GitHub dependency alerts, manual dependency
review, and any future dependency-review workflow.

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

## Pre-Release Policy

Before any release tag, apply the release checklist and confirm that:

- `govulncheck` and `gosec` pass in Release Validation;
- Code Scanning has no unexpected open CodeQL or gosec alerts;
- advisory security scans have no unresolved violations under this policy;
- dependency-review evidence is current for the release candidate;
- any non-affecting vulnerability is accounted for in `docs/vex.md`.

Do not publish a release with unresolved SCA or SAST violations unless a
reviewable VEX or suppression rationale explains why the finding does not
affect the release.
