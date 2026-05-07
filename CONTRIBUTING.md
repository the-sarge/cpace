# Contributing

This repository is an auditable draft implementation of
`draft-irtf-cfrg-cpace-21` for `CPACE-RISTR255-SHA512`. It has not had
independent cryptographic review and is not production-ready.

## Security Issues

Do not open a public issue for a vulnerability, suspected secret leak, private
exploit path, or embargoed finding. Follow `SECURITY.md` and report privately
by email or GitHub private vulnerability reporting.

Public issues are appropriate for non-sensitive bugs, documentation gaps,
external review questions, and release-readiness tracking.

## Contribution Scope

Current work is release readiness. The public API and package-profile decisions
are frozen for review unless a new finding reopens one. Before proposing API,
wire-format, protocol, dependency, or fuzz-harness changes, explain which
release-readiness gap or review finding the change addresses.

Useful contributions include:

- documentation fixes that improve integration clarity;
- review findings against `docs/external-review-handoff.md`;
- tests or evidence that support the existing draft-21 profile;
- CI, dependency, or release-process fixes that preserve the repository's
  least-privilege and signed-release policy.

## Pull Requests

Keep pull requests narrow. A release-readiness PR should include:

- the gap or finding being closed;
- the exact commit, command, workflow, or review artifact used as evidence;
- residual risk or follow-up that remains;
- documentation updates when release posture changes.

Every commit must certify the Developer Certificate of Origin in `DCO` with a
`Signed-off-by` trailer. Use:

```sh
git commit -s
```

If you already made the latest commit, amend it with:

```sh
git commit --amend --no-edit -s
```

For a multi-commit PR, every commit needs its own signoff. GitHub web-based
commits are configured to add signoffs automatically.

Run the appropriate local validation before opening a PR:

```sh
task docs:check
task quick
task check
```

For release-oriented changes, also follow `docs/ci-policy.md` and record
evidence in the relevant docs.

## Review Expectations

Treat CPace behavior, binary framing, context/identity binding, randomness,
memory handling, and release evidence as security-relevant. If a change affects
protocol behavior, parser/framing behavior, dependencies, toolchain, fuzz
harnesses, or package-profile docs, assume the release evidence needs to be
refreshed before any production-readiness claim.
