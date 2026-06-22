# CI Policy

This repository treats CI as release evidence for an unaudited crypto package.
Required pull-request CI stays narrow because fork PRs run untrusted code on
hosted runners.

## When Tests Run

Local validation uses `Taskfile.yml` as the command facade:

- `task docs:check` validates tracked Markdown and whitespace.
- `task quick` runs Go formatting checks, docs validation, and `go test ./...`.
- `task check` runs docs validation, release-helper smoke tests, evidence baseline validation, nested evidence-checker linting, tests, race tests, formatting/import checks, `go vet`, Staticcheck, ast-grep rules, and `govulncheck`; it requires `jq` for CycloneDX SBOM JSON validation.
- `task lint:golangci` runs a pinned, curated advisory `golangci-lint` analyzer set; it is not part of the required local gate.
- `task fuzz` runs every fuzz target in `.github/fuzz-targets.json` with the
  caller-provided `FUZZTIME`, `PARALLEL`, `FUZZ_RACE`, `GOMAXPROCS`, and
  `FUZZ_TEST_PARALLEL` settings.

Repository CI runs on these events:

- Pull requests to `main`: required `Check` runs for every PR. Code changes set up Go, run `go test ./...`, and run the evidence baseline validator. Docs-only PRs run whitespace and Markdown validation without Go unless they touch `docs/evidence-baseline.md`, `docs/evidence-baseline-summary-docs.txt`, `docs/evidence/**`, or any summary doc listed in `docs/evidence-baseline-summary-docs.txt`, in which case the job also sets up Go and runs the evidence baseline validator. The DCO workflow checks every PR commit for a `Signed-off-by` trailer.
  `Dependency Gate` runs blocking SCA tooling, and `SAST Gate` runs blocking
  `gosec`.
- Pull requests that touch Go code or Go module files: CodeQL and Staticcheck
  Advisory run as background signal. Pull requests that touch Go code, Go
  module files, `Taskfile.yml`, `.golangci.yml`, or the GolangCI-Lint workflow
  also run GolangCI-Lint Advisory.
- Pushes to `main`: required `Check` runs again, and CodeQL analyzes the main
  branch.
- Scheduled or manual runs: Vulnerability Scan, Gosec Advisory, GolangCI-Lint
  Advisory, Nightly Fuzz, Autoscaled Fuzz, CodeQL, Staticcheck Advisory, Scorecard, and
  cross-platform smoke workflows provide background and release-posture signal.
- Release tags matching `v*`: Release Validation verifies the signed annotated tag first, runs tests, race tests, `govulncheck`, and `gosec` with SARIF upload, then generates, validates, attests, and publishes the GitHub Release with SBOM assets. `v0.x` and SemVer prerelease tags are published as GitHub prereleases and are explicitly not marked latest.

Maintainer-controlled long fuzzing is run outside the required PR gate and
recorded in `docs/fuzz-evidence.md` when it supports a release-readiness claim.
For exact release candidates and toolchain-security refreshes, preserve raw
logs, transcripts, or immutable workflow artifacts with checksums under
`docs/evidence/` or link to the immutable workflow artifact from the evidence
docs.

## PR Gates

The intended required PR gates are:

- `Check` in `.github/workflows/ci.yml`. It runs on GitHub-hosted Ubuntu runners with read-only repository permissions. Code changes set up Go, run `go test ./...`, and run the evidence baseline validator. Docs-only PRs run whitespace and Markdown validation without Go unless they touch `docs/evidence-baseline.md`, `docs/evidence-baseline-summary-docs.txt`, `docs/evidence/**`, or any summary doc listed in `docs/evidence-baseline-summary-docs.txt`, in which case the job also sets up Go and runs the evidence baseline validator.
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

`Vulnerability Scan`, `Gosec Advisory`, `GolangCI-Lint Advisory`, and `Nightly
Fuzz` run on GitHub-hosted runners through both `workflow_dispatch` and scheduled
triggers. `Autoscaled Fuzz` validates inputs on a GitHub-hosted preflight job,
then runs fuzzing on the self-hosted GARM `cpace-garm-linux-fuzz-arm64` and
`cpace-garm-linux-fuzz-amd64` runner labels through scheduled triggers and
trusted main-branch manual dispatch. These lanes provide scheduled drift
detection, Code Scanning history, and fuzz regression signal in addition to the
PR gates.

Manual `Dependency Gate` dispatch runs module verification and `govulncheck`;
GitHub Dependency Review runs only on pull requests because it compares the PR
dependency diff against the base branch.

The hosted scheduled fuzz lane is a short 5-minute-per-target regression run.
It can catch crashes and upload new failure corpus files, but it is not
long-fuzz release evidence by itself.

The autoscaled fuzz lane is a longer 10-minute-per-target background run.
Scheduled runs default to `FUZZ_RACE=0`, `PARALLEL=1`, `GOMAXPROCS=1`, and
`FUZZ_TEST_PARALLEL=1` so the self-hosted Mac remains responsive while fuzzing
runs. Race coverage remains owned by `task check` and can be requested on this
lane only by trusted main-branch manual dispatch. The preflight job rejects
manual inputs unless `FUZZTIME` matches `[0-9]+[smh]`, `PARALLEL`,
`GOMAXPROCS`, and `FUZZ_TEST_PARALLEL` are positive integers, `FUZZ_RACE` is
`0` or `1`, and `ceil(targets/PARALLEL) * FUZZTIME` stays below the 240-minute
fuzz job timeout.

## Long Fuzzing And Release Evidence

Release-oriented changes should still run the full local gate, dependency
review, SCA/SAST gates, advisory security scans, and maintainer-controlled long
fuzzing before a release tag. Record exact evidence in the project evidence
docs: commit SHA, command or workflow, fuzz duration, target count, and
residual risk. Raw or immutable artifacts are required for exact release
candidates and recommended for external-review refreshes when they are cheap to
capture.

Release tags should remain signed annotated tags. Downstream consumers should be able to verify each release tag with `git verify-tag`.

## Distribution Surface

The primary release trust root is the signed annotated `v*` tag. Release Validation verifies that tag against `.github/allowed_signers`, then treats the checked-out source tree as the release candidate. That CI verification catches maintainer mistakes such as lightweight tags, unsigned tags, or signatures outside the documented signer set, but it does not protect against a principal who can create, update, or delete a `v*` tag and thereby choose both the workflow definition and the checked-in signer file.

The primary tag-authority control is the active GitHub repository ruleset `16048307` on `refs/tags/v*`, covering creation, update, and deletion with no routine bypass actors. That state is admin-mutable GitHub configuration rather than repository content, so each release must capture fresh ruleset JSON before tagging and document any break-glass change. The committed 2026-06-10 evidence under `docs/evidence/tagruleset-20260610/` is historical baseline evidence, and the 2026-06-19 exact-candidate capture lives under `docs/evidence/f7efa6a-20260619/`; neither is a permanent release claim for future tags.

CI attests the generated SBOM asset, not the CPace protocol implementation, Go API, source archive, Go module proxy entry, or SLSA Build Level 3 provenance. The SBOM attestation binds `cpace-<tag>.cdx.json` to the GitHub Actions run through GitHub artifact attestations with the CycloneDX predicate type `https://cyclonedx.org/bom`; the attached `cpace-<tag>.cdx.json.sigstore.json` bundle is retained for verifiers that need the Sigstore bundle. The release-body SHA-256 checksum is only a corruption-detection convenience because the release body and assets share the mutable GitHub Release trust domain.

`anchore/sbom-action` is SHA-pinned and requests Syft `v1.45.1`, but the action downloads the Syft release binary from Anchore's GitHub releases at runtime. A compromised Syft download could falsify the SBOM, but the SBOM job has only read repository permissions and cannot make the signed tag authentic or mutate release publishing by itself. Checksum-pinning the Syft binary remains a possible follow-up if the project needs a stronger SBOM-generation toolchain claim.

## Self-Hosted Runners

GitHub-hosted runners handle untrusted PR validation. Self-hosted runners must
not run code from untrusted fork PRs.

The current self-hosted lane is `Autoscaled Fuzz`, which uses separate
`cpace-garm-linux-fuzz-arm64` and `cpace-garm-linux-fuzz-amd64` GitHub runner
labels. Its job-level guard skips the checked-in fuzz job except for scheduled
runs and manual dispatches from `refs/heads/main`. Treat that guard as workflow
hygiene and defense in depth: the trust boundary is that fork PRs cannot
schedule or dispatch this workflow, and manual dispatch requires repository
write access.

The autoscaled runner image must provide a POSIX/GNU userland and a working C
compiler for Linux race-detector fuzz builds. At minimum the workflow checks
for `bash`, `find`, `jq`, `mktemp`, `sed`, `sort`, `touch`, `xargs`, and a C
compiler (`cc`, `gcc`, or `clang`) before reporting the fuzz plan or invoking
`task fuzz`. Go and Task are installed by the workflow itself.

Any additional self-hosted lane must either be ephemeral with one job per
runner instance, or restricted to trusted `main`-only scheduled and manual
workflows. Long fuzzing may run on maintainer-controlled machines only through
manual, scheduled-main, or ephemeral-runner workflows.
