# CI Policy

This repository treats CI as release evidence for an unaudited crypto package.
Required pull-request CI stays narrow because fork PRs run untrusted code on
hosted runners.

## When Tests Run

Local validation uses `Taskfile.yml` as the command facade:

- `task docs:check` validates tracked Markdown and whitespace.
- `task quick` runs Go formatting checks, docs validation, CI change-classifier, Go-tool catalogue, and fuzz-input validator smoke tests, and `go test ./...`; it requires `jq` for fuzz-registry validation.
- `task check` runs docs validation, release-helper, CI change-classifier, Go-tool catalogue, and fuzz-input validator smoke tests, release policy validation, nested release-policy and evidence-checker linting, evidence baseline validation, tests, race tests, the lint-config drift check, formatting/import checks, the curated `golangci-lint` analyzer set, ast-grep rules, and `govulncheck`; it requires `jq` for fuzz-registry and CycloneDX SBOM JSON validation.
- `task lint:golangci` runs the pinned `golangci-lint` analyzer set on the root module. It is the single owner of `gosec`, `staticcheck`, and `go vet` coverage for that module, and `task lint` (and therefore `task check`) now depends on it. `GOLANGCI_ARGS` is the only supported way for a lane to add scan-scope or output flags, mirroring the `GOSEC` variable on `task gosec`.
- `task lint:config-check` verifies that `.golangci.yml` still matches the portfolio Go lint standard v1.0.0 by sha256. `.golangci.yml` is a verbatim copy of the canonical file in `the-sarge/infra` (`config/golangci/.golangci.yml`, tag `golangci-standard/v1.0.0`) and must never be edited locally; per-file `//nolint:<linter> // reason` is the only escape valve. `task lint` and `task lint:golangci` both depend on this check, so `task check` enforces it.
- `task fuzz` runs every target from the fuzz-target registry (`.github/fuzz-targets.json`, with target function, package, and OSS-Fuzz-compatible binary name) using the caller-provided `FUZZTIME`, `PARALLEL`, `FUZZ_RACE`, `GOMAXPROCS`, and `FUZZ_TEST_PARALLEL` settings. Both this local task and Autoscaled Fuzz enforce those settings through `scripts/validate-fuzz-inputs.sh`; only the workflow supplies `FUZZ_MAX_WALL_MINUTES=240`, so maintainer-controlled local long fuzzing remains available. `go test ./...` fails if the registry drifts from the defined fuzz functions or OSS-Fuzz-compatible build lines.

Pinned Go CLI modules used by tracked workflows and the Taskfile Staticcheck and golangci-lint defaults are catalogued in `scripts/go-tool-versions.sh` and invoked through `scripts/go-tool.sh`. Updating one catalogue entry changes the version used by both local Taskfile defaults and CI; in particular, `task release:lint`, `task evidence:lint`, and the CI nested-module lint steps use the same Staticcheck pin.

## Analyzer Ownership

`gosec` and `staticcheck` run inside `golangci-lint` for the root module rather
than as standalone lanes. The pinned `golangci-lint v2.13.1` vendors
`gosec v2.28.0` and `honnef.co/go/tools v0.8.0`, so the consolidated lane is at
least as current as the retired standalone pins. The canonical `.golangci.yml`
enables `gosec` with `run.tests: true` and `staticcheck` with
`all,-QF1011,-ST1000`; the only intentional relaxation against the retired
standalone scans is that `G301`, `G302`, and `G306` are dropped in `_test.go`
files, where file and directory permission findings describe test fixtures. All
other `gosec` rules still apply to test code.

One triage detail changed with the SARIF source: `golangci-lint` emits results
with the linter name as the SARIF `ruleId`, so Code Scanning groups findings
under `gosec` rather than under individual rule ids such as `G401`. The specific
rule id stays in the alert message text.

Two lanes deliberately remain outside `golangci-lint`:

- `govulncheck` stays standalone. It queries the Go vulnerability database
  rather than analysing source, so it is not a linter and its findings do not
  come from the curated analyzer set.
- `tools/releasepolicy` and `tools/evidencebaseline` are separate Go modules and
  therefore outside the root `golangci-lint` run. They keep `go vet` plus
  Staticcheck through `task release:lint` and `task evidence:lint`, which is why
  the Staticcheck pin is still catalogued.

The tag-triggered `Gosec` job in Release Validation is still a standalone
`gosec` scan. It is snapshotted by the accepted ADR-0007 release policy and
feeds signed release evidence, so moving it onto the consolidated lane is
tracked as separate release-policy work.

Repository CI runs on these events:

- Pull requests to `main`: required `Check` runs for every PR. Every run sets up Go, runs the release policy checker and its module tests, and runs `go vet` and Staticcheck for `tools/releasepolicy`. Code changes also run `go test ./...` and the evidence baseline validator. Docs-only PRs run whitespace and Markdown validation without root-module tests unless they touch `docs/evidence-baseline.md`, `docs/evidence-baseline-summary-docs.txt`, `docs/evidence/**`, or any summary doc listed in `docs/evidence-baseline-summary-docs.txt`, in which case the job also runs the evidence baseline validator. The DCO workflow checks every PR commit for a `Signed-off-by` trailer.
  `Dependency Gate` runs blocking SCA tooling, and `SAST Gate` runs blocking
  `golangci-lint`, which carries `gosec` and `staticcheck`.
- Pull requests that touch Go code or Go module files: CodeQL runs as background
  signal.
- Pushes to `main`: required `Check` runs again, and CodeQL analyzes the main
  branch.
- Scheduled or manual runs: Vulnerability Scan, Nightly Fuzz, Autoscaled Fuzz,
  CodeQL, Scorecard, and cross-platform smoke workflows provide background and
  release-posture signal.
- Release tags matching `v*`: Release Validation verifies the signed annotated tag first, runs tests, race tests, `govulncheck`, and `gosec` with SARIF upload, then generates, validates, attests, and publishes the GitHub Release with SBOM assets. `v0.x` and SemVer prerelease tags are published as GitHub prereleases and are explicitly not marked latest.

Maintainer-controlled long fuzzing is run outside the required PR gate and
recorded in `docs/fuzz-evidence.md` when it supports a release-readiness claim.
For exact release candidates and toolchain-security refreshes, preserve raw
logs, transcripts, or immutable workflow artifacts with checksums under
`docs/evidence/` or link to the immutable workflow artifact from the evidence
docs.

## PR Gates

The intended required PR gates are:

- `Check` in `.github/workflows/ci.yml`. It runs on GitHub-hosted Ubuntu runners with read-only repository permissions. Every run sets up Go, runs the release policy checker and its module tests, and runs `go vet` and Staticcheck for `tools/releasepolicy`. Code changes also run `go test ./...` and the evidence baseline validator. Docs-only PRs run whitespace and Markdown validation without root-module tests unless they touch `docs/evidence-baseline.md`, `docs/evidence-baseline-summary-docs.txt`, `docs/evidence/**`, or any summary doc listed in `docs/evidence-baseline-summary-docs.txt`, in which case the job also runs the evidence baseline validator.
- `DCO` in `.github/workflows/dco.yml`. It checks every PR commit for a
  `Signed-off-by` trailer.
- `Dependency Gate` in `.github/workflows/dependency-gate.yml`. It runs GitHub
  Dependency Review, `go mod verify`, and `govulncheck -test ./...`.
- `SAST Gate` in `.github/workflows/sast-gate.yml`. It runs the blocking
  curated `golangci-lint` analyzer set, which includes `gosec` over test code,
  uploads the resulting SARIF to Code Scanning under the `sast-gate` category
  for same-repository runs, and then runs the ast-grep structural rules.

`Dependency Gate` and `SAST Gate` must be listed in branch protection before
the project treats OSPS-VM-05.03 and OSPS-VM-06.02 as satisfied.

The 2026-08-01 release-control capture in `docs/evidence/1f19e27-20260801-github/` confirmed that strict `main` protection required exactly `Check`, `DCO`, `Dependency Gate`, and `SAST Gate`, matching this policy, with administrator enforcement and conversation resolution enabled and force pushes and deletion disabled. The frozen source candidate is an intermediate commit in PR #239 with no direct GitHub check suites; all four required contexts passed on the PR's documentation-only checked descendant before protected merge.

Keep required lanes short, deterministic, and least-privilege. New security or
analysis tools should start as background signal before being considered for a
required gate.

## Background Signal

`Vulnerability Scan` and `Nightly Fuzz` run on GitHub-hosted runners through
both `workflow_dispatch` and scheduled triggers. `Autoscaled Fuzz` validates inputs on a GitHub-hosted preflight job,
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

The autoscaled fuzz lane is a longer 10-minute-per-target background run. Scheduled runs default to `FUZZ_RACE=0`, `PARALLEL=1`, `GOMAXPROCS=1`, and `FUZZ_TEST_PARALLEL=1` so the self-hosted Mac remains responsive while fuzzing runs. Race coverage remains owned by `task check` and can be requested on this lane only by trusted main-branch manual dispatch. The shared validator rejects inputs unless `FUZZTIME` matches `[0-9]+[smh]`, `PARALLEL`, `GOMAXPROCS`, and `FUZZ_TEST_PARALLEL` are positive integers, `FUZZ_RACE` is `0` or `1`, the fuzz-target registry is a nonempty JSON array, and, when the workflow bound is supplied, `ceil(targets/PARALLEL) * FUZZTIME` stays below the 240-minute fuzz job timeout.

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

The primary tag-authority control is the active GitHub repository ruleset `16048307` on `refs/tags/v*`, covering creation, update, and deletion with no routine bypass actors. That state is admin-mutable GitHub configuration rather than repository content, so each release must capture fresh ruleset JSON before tagging and document any break-glass change. The current `v0.1.3` point-in-time capture is `docs/evidence/1f19e27-20260801-github/`; the 2026-06-10 and 2026-06-19 bundles remain historical, and none of these captures is a permanent claim for future tags.

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
