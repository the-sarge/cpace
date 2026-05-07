# Governance Policy

This project treats repository administration, release authority, CI secrets,
and security-report access as sensitive resources.

## Sensitive Resources

Sensitive resources include:

- write, maintain, admin, or security-manager permissions on the GitHub
  repository;
- branch protection, tag protection, rulesets, environments, and required
  status checks;
- GitHub Actions secrets, environment secrets, OIDC trust configuration, and
  deploy keys;
- release, tag-signing, package-publishing, or advisory-publishing authority;
- access to private vulnerability reports, embargoed issues, or non-public
  security review material.

## Escalated Permission Review

Before granting any code collaborator escalated permission to a sensitive
resource, a maintainer must review:

- the collaborator's identity and project relationship;
- the specific task that requires access;
- the minimum permission level and duration needed for that task;
- whether the collaborator has agreed to the DCO and project security policy;
- whether the account uses strong authentication appropriate for the access;
- how the access will be audited, expired, or revoked.

Access should be least-privilege and time-limited when possible. Do not grant
admin, release, secret, or security-advisory access merely because a
collaborator has contributed code.

## Ongoing Review

Review escalated permissions at least annually, after a maintainer or
collaborator role changes, and after any suspected account, secret, CI, or
release-process incident. Revoke access promptly when the original need no
longer exists.

If the project adds package publishing, automated release signing, or
self-hosted CI, document the sensitive resource owner and review evidence before
granting access.
