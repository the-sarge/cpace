# Security Policy

Report security issues privately before opening a public issue. Email
`the-sarge@the-sarge.com`, or use GitHub private vulnerability reporting for
`the-sarge/cpace` if it is enabled.

Do not include vulnerability details, suspected secret leaks, private exploit
paths, or embargoed findings in public issues or pull requests. Public issues
are appropriate for non-sensitive bugs, documentation gaps, external review
questions, and release-readiness tracking.

## Implementation Status

This repository is an unaudited implementation of an active Internet-Draft:
`draft-irtf-cfrg-cpace-21`, published April 23, 2026. Do not describe it as
production-ready until independent cryptographic review is complete.

## Coordinated Vulnerability Disclosure

The project follows coordinated vulnerability disclosure for confirmed or
suspected vulnerabilities.

Expected response timeline:

- Acknowledge private vulnerability reports within 7 calendar days.
- Provide initial triage or status within 14 calendar days.
- Provide status updates at least every 30 calendar days while a confirmed
  issue remains unresolved.
- Coordinate public disclosure timing with the reporter.
- Aim to publish a fix, mitigation, advisory, or documented rationale within 90
  calendar days for confirmed vulnerabilities, unless active exploitation,
  report complexity, or reviewer coordination requires a different timeline.

Reports should include the affected version or commit, reproduction steps or
proof-of-concept details when safe to share privately, expected impact, and any
requested embargo or coordination constraints.

## Secrets And Credentials

Project secrets and credentials must not be committed to the repository,
included in public issues or pull requests, pasted into logs, or stored in
plain text project files.

Storage guidelines:

- Store project automation secrets only in GitHub Actions secrets, GitHub
  environment secrets, or another maintainer-approved secret manager.
- Prefer short-lived credentials and OIDC-based federation over long-lived
  static tokens when a workflow or integration supports it.
- Keep local maintainer credentials outside the repository, protected by the
  operating system keychain, SSH agent, password manager, or equivalent local
  secret store.
- Do not add secrets to examples, tests, fixtures, fuzz corpora, release
  artifacts, or documentation.

Access guidelines:

- Grant access only to maintainers or automation that need the credential for a
  documented project purpose.
- Scope tokens to the minimum repository, permission, and lifetime that the task
  supports.
- Review workflow permissions before adding any credential use. Workflows should
  keep least-privilege `permissions:` blocks and avoid persisting checkout
  credentials unless a job explicitly needs write access.
- Public pull requests from forks must not receive project secrets.

Rotation guidelines:

- Rotate any credential immediately if it is suspected to be exposed,
  over-broad, unused, or no longer controlled by an active maintainer.
- Rotate long-lived project credentials at least annually when they cannot be
  replaced by short-lived or federated credentials.
- Revoke credentials promptly when a maintainer or integration no longer needs
  access.
- If a secret is exposed, revoke and replace it before publishing details, then
  assess whether a private vulnerability report or security advisory is needed.

## Supported Versions

No production-ready version is supported yet. Until the release bar in
`docs/security-assessment.md` is satisfied, tags must remain in the `v0.x`
range and should be treated as draft implementation snapshots.

Release support scope and duration:

- `v0.x` prereleases are supported only for security reporting, vulnerability
  triage, and best-effort fixes for the latest published prerelease.
- Older `v0.x` prereleases are superseded when a newer prerelease is published;
  security updates for an older `v0.x` prerelease end when the newer prerelease
  is published.
- No `v0.x` prerelease has a guaranteed maintenance duration or production
  support window.
- A production-ready support policy will be defined before any `v1.0.0`
  release or production-readiness claim.

## Security Finding Thresholds

Policies for SCA findings, SAST findings, malicious dependencies, pre-release
security gates, and VEX records are documented in `docs/security-gates.md` and
`docs/vex.md`.

## Release Readiness

Before any production-readiness claim, the project must complete long fuzzing,
dependency review, security/spec documentation audit, external review of
package-owned framing/profile choices, and independent cryptographic review.
Dependency, fuzz, and security/spec audit evidence is recorded in
`docs/dependency-review.md`, `docs/fuzz-evidence.md`, and
`docs/security-spec-audit.md`; remaining work is tracked in
`docs/project-plan.md`.

Supported scope for the initial implementation:

- `CPACE-RISTR255-SHA512`
- initiator-responder mode
- mandatory explicit key confirmation

Unsupported scope:

- symmetric mode
- draft revisions other than draft-21
- ciphersuites other than Ristretto255 with SHA-512

Outer protocol negotiation, downgrade protection, and application channel
binding are application responsibilities. See `docs/integration-guidance.md`
for integration guidance.
