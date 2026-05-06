# Security Policy

Report security issues privately before opening a public issue. Email
`the-sarge@the-sarge.com`, or use GitHub private vulnerability reporting for
`the-sarge/cpace` if it is enabled.

This repository is an unaudited implementation of an active Internet-Draft:
`draft-irtf-cfrg-cpace-21`, published April 23, 2026. Do not describe it as
production-ready until independent cryptographic review is complete.

## Supported Versions

No production-ready version is supported yet. Until the release bar in
`docs/security-assessment.md` is satisfied, tags must remain in the `v0.x`
range and should be treated as draft implementation snapshots.

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
