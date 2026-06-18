# Reviewer Outreach Note

Use this note as a starting point when asking for external review. Adjust the
recipient, scope, and time expectations before sending.

## Short Note

Hello,

I am looking for external review of a Go CPace draft implementation:

`https://github.com/the-sarge/cpace`

The package implements only `CPACE-RISTR255-SHA512` from
`draft-irtf-cfrg-cpace-21`, in initiator-responder mode with mandatory explicit
key confirmation. It is explicitly marked as an auditable draft implementation,
not production-ready cryptographic software.

The review handoff is here:

`https://github.com/the-sarge/cpace/blob/main/docs/external-review-handoff.md`

The threat model is here:

`https://github.com/the-sarge/cpace/blob/main/docs/threat-model.md`

The most useful review focus would be:

- package-owned context-info construction and role-local identity input;
- binary wire framing and field-size limits;
- empty-session-ID policy and integration guidance;
- scalar sampling, invalid-point handling, confirmation, exporter, and session
  lifecycle claims;
- whether the CI, dependency, fuzz, and signed-release evidence is sufficient
  for an auditable prerelease while independent cryptographic review remains a
  blocker.

The public tracking issues for those review areas are:

- `https://github.com/the-sarge/cpace/issues/29`
- `https://github.com/the-sarge/cpace/issues/30`
- `https://github.com/the-sarge/cpace/issues/31`

The latest public prerelease is `v0.1.2`, an SSH-signed annotated tag with the
external-review packet and refreshed Go 1.26.3 evidence. It does not claim
production readiness.

Please report private vulnerabilities through the channels in `SECURITY.md`.
Public review notes that are not sensitive can be opened as GitHub issues.

Thank you.

## Reviewer Packet

Primary files for review:

- `README.md`
- `SECURITY.md`
- `docs/external-review-handoff.md`
- `docs/security-assessment.md`
- `docs/spec-matrix.md`
- `docs/integration-guidance.md`
- `docs/ci-policy.md`
- `docs/dependency-review.md`
- `docs/fuzz-evidence.md`
- `docs/security-spec-audit.md`
- `docs/capslock-report.md`
- `docs/performance.md`

Open release blockers:

- external review of package-owned framing, context-info construction, and
  profile choices;
- independent cryptographic review;
- exact-candidate dependency review, long fuzzing, and security/spec audit after
  review-driven changes.
