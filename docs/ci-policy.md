# CI Policy

This repository treats CI as release evidence for an unaudited crypto package.
Required pull-request CI stays narrow because fork PRs run untrusted code on
hosted runners.

## When Tests Run

Local validation uses `Taskfile.yml` as the command facade:

- `task docs:check` validates tracked Markdown and whitespace.
- `task quick` runs Go formatting checks, docs validation, and `go test ./...`.
- `task check` runs docs validation, tests, race tests, formatting/import
  checks, `go vet`, Staticcheck, ast-grep rules, and `govulncheck`.
- `task fuzz` runs every fuzz target in `.github/fuzz-targets.json` with the
  caller-provided `FUZZTIME`, `PARALLEL`, and `FUZZ_RACE` settings.

Hosted CI runs on these events:

- Pull requests to `main`: required `Check` runs for every PR. Code changes run
  `go test ./...`; docs-only PRs run whitespace and Markdown validation. The
  DCO workflow checks every PR commit for a `Signed-off-by` trailer.
  `Dependency Gate` runs blocking SCA tooling, and `SAST Gate` runs blocking
  `gosec`.
- Pull requests that touch Go code or Go module files: CodeQL and Staticcheck
  Advisory run as background signal.
- Pushes to `main`: required `Check` runs again, and CodeQL analyzes the main
  branch.
- Scheduled or manual runs: Vulnerability Scan, Gosec Advisory, Nightly Fuzz,
  CodeQL, Staticcheck Advisory, Scorecard, and cross-platform smoke workflows
  provide background and release-posture signal.
- Release tags matching `v*`: Release Validation runs tests, race tests,
  `govulncheck`, and `gosec` with SARIF upload.

Maintainer-controlled long fuzzing is run outside the required PR gate and
recorded in `docs/fuzz-evidence.md` when it supports a release-readiness claim.
For exact release candidates and toolchain-security refreshes, preserve raw
logs, transcripts, or immutable workflow artifacts with checksums under
`docs/evidence/` or link to the immutable workflow artifact from the evidence
docs.

## PR Gates

The intended required PR gates are:

- `Check` in `.github/workflows/ci.yml`. It runs on GitHub-hosted Ubuntu
  runners with read-only repository permissions. Code changes run
  `go test ./...`; docs-only PRs run whitespace and Markdown validation without
  setting up Go.
- `DCO` in `.github/workflows/dco.yml`. It checks every PR commit for a
  `Signed-off-by` trailer.
- `Dependency Gate` in `.github/workflows/dependency-gate.yml`. It runs GitHub
  Dependency Review, `go mod verify`, and `govulncheck -test ./...`.
- `SAST Gate` in `.github/workflows/sast-gate.yml`. It runs blocking
  `gosec -tests ./...` and uploads SARIF for same-repository runs.

`Dependency Gate` and `SAST Gate` must be listed in branch protection before
the project treats OSPS-VM-05.03 and OSPS-VM-06.02 as satisfied.

Keep required lanes short, deterministic, and least-privilege. New security or
analysis tools should start as background signal before being considered for a
required gate.

## Background Signal

`Vulnerability Scan`, `Gosec Advisory`, and `Nightly Fuzz` run on GitHub-hosted
runners through both `workflow_dispatch` and scheduled triggers. These lanes
provide scheduled drift detection, Code Scanning history, and fuzz regression
signal in addition to the PR gates.

Manual `Dependency Gate` dispatch runs module verification and `govulncheck`;
GitHub Dependency Review runs only on pull requests because it compares the PR
dependency diff against the base branch.

The scheduled fuzz lane is a short 5-minute-per-target regression run. It can
catch crashes and upload new failure corpus files, but it is not long-fuzz
release evidence by itself.

## Long Fuzzing And Release Evidence

Release-oriented changes should still run the full local gate, dependency
review, SCA/SAST gates, advisory security scans, and maintainer-controlled long
fuzzing before a release tag. Record exact evidence in the project evidence
docs: commit SHA, command or workflow, fuzz duration, target count, and
residual risk. Raw or immutable artifacts are required for exact release
candidates and recommended for external-review refreshes when they are cheap to
capture.

Release tags should remain signed annotated tags. Downstream consumers should
be able to verify each release tag with `git verify-tag`.

## Self-Hosted Runners

GitHub-hosted runners handle untrusted PR validation. Self-hosted runners must
not run code from untrusted fork PRs.

If self-hosted capacity is added later, it must either be ephemeral with one
job per runner instance, or restricted to trusted `main`-only scheduled and
manual workflows. Long fuzzing may run on maintainer-controlled machines only
through manual, scheduled-main, or ephemeral-runner workflows.
