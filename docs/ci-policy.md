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

Repository CI runs on these events:

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
  Autoscaled Fuzz, CodeQL, Staticcheck Advisory, Scorecard, and
  cross-platform smoke workflows provide background and release-posture signal.
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
runners through both `workflow_dispatch` and scheduled triggers. `Autoscaled
Fuzz` validates inputs on a GitHub-hosted preflight job, then runs fuzzing on the
self-hosted GARM `cpace-garm-linux-fuzz` runner label through scheduled triggers
and trusted main-branch manual dispatch. These lanes provide scheduled drift
detection, Code Scanning history, and fuzz regression signal in addition to the
PR gates.

Manual `Dependency Gate` dispatch runs module verification and `govulncheck`;
GitHub Dependency Review runs only on pull requests because it compares the PR
dependency diff against the base branch.

The hosted scheduled fuzz lane is a short 5-minute-per-target regression run.
It can catch crashes and upload new failure corpus files, but it is not
long-fuzz release evidence by itself.

The autoscaled fuzz lane is a longer 20-minute-per-target background run. It
defaults to `FUZZ_RACE=1` because it does not run `task check` before fuzzing,
so scheduled runs provide their own race-instrumented fuzz coverage. Trusted
main-branch manual dispatch can set `FUZZ_RACE=0` for targeted non-race runs.
Its default `PARALLEL=2` and `GOMAXPROCS=4` settings assume a runner with at
least eight vCPUs and enough memory for two concurrent race-enabled fuzz
processes; reduce those values if the autoscaled runner class is smaller. The
preflight job rejects manual inputs unless `FUZZTIME` matches `[0-9]+[smh]`,
`PARALLEL` is a positive integer, `FUZZ_RACE` is `0` or `1`, and
`ceil(targets/PARALLEL) * FUZZTIME` stays below the 240-minute fuzz job timeout.

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

The current self-hosted lane is `Autoscaled Fuzz`, which uses the
`infra-autoscale-cpace-fuzz-linux` runner label. Its job-level guard skips the
checked-in fuzz job except for scheduled runs and manual dispatches from
`refs/heads/main`. Treat that guard as workflow hygiene and defense in depth:
the trust boundary is that fork PRs cannot schedule or dispatch this workflow,
and manual dispatch requires repository write access.

The autoscaled runner image must provide a POSIX/GNU userland and a working C
compiler for Linux race-detector fuzz builds. At minimum the workflow checks
for `bash`, `find`, `jq`, `mktemp`, `sed`, `sort`, `touch`, `xargs`, and a C
compiler (`cc`, `gcc`, or `clang`) before reporting the fuzz plan or invoking
`task fuzz`. Go and Task are installed by the workflow itself.

Any additional self-hosted lane must either be ephemeral with one job per
runner instance, or restricted to trusted `main`-only scheduled and manual
workflows. Long fuzzing may run on maintainer-controlled machines only through
manual, scheduled-main, or ephemeral-runner workflows.
