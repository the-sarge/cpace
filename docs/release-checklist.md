# Release Checklist

This checklist is for future prerelease or production-readiness candidates. It
does not make the current package production-ready.

## 1. Freeze Candidate Scope

- Identify the exact candidate commit.
- Confirm whether the release is a prerelease snapshot or a
  production-readiness candidate.
- Confirm no untracked protocol, parser/framing, dependency, toolchain, fuzz
  harness, or package-profile documentation changes are missing from the
  candidate.
- Update `CHANGELOG.md`.

## 2. Local Validation

Run from the candidate commit:

```sh
task docs:check
task quick
task check
```

For release-oriented changes, also run:

```sh
go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...
govulncheck -test -show verbose ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 ./...
```

Record command versions and results in the relevant evidence docs if the
candidate is making a stronger release-readiness claim.

## 3. Evidence Bundle

Before refreshing summary docs, choose the evidence mode for the candidate:

- immutable GitHub Actions or manual workflow artifacts tied to the exact
  commit;
- committed raw transcripts under `docs/evidence/<candidate-or-date>/` with
  `SHA256SUMS`;
- both, when practical.

For exact-candidate releases and toolchain-security refreshes, prefer raw
artifacts over handwritten summaries. Follow `docs/evidence/README.md` for the
bundle shape, required metadata, checksum verification, and optional detached
signature handling. Summary docs may explain evidence, but the evidence packet
should preserve the raw logs or immutable workflow links.

Before each release, capture the active `refs/tags/v*` tag-authority ruleset JSON as recurring release evidence. Record `gh api /repos/the-sarge/cpace/rulesets` and `gh api /repos/the-sarge/cpace/rulesets/16048307`, then confirm ruleset `16048307` is active, covers creation/update/deletion for `refs/tags/v*`, has `bypass_actors: []`, and reports `current_user_can_bypass: never`. The 2026-06-10 baseline is committed under `docs/evidence/tagruleset-20260610/`, and the 2026-06-19 exact-candidate capture is committed under `docs/evidence/f7efa6a-20260619/`; each release needs a fresh capture because GitHub ruleset state is admin-mutable.

## 4. Long Fuzz Evidence

Run every registered fuzz target from `.github/fuzz-targets.json` against the
exact candidate commit. Use maintainer-controlled machines or trusted
main-only/manual workflows.

Recommended shape for stable evidence runs:

```sh
FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz
```

Record host, platform, Go version, Task version, command, start/end UTC,
target count, candidate commit, result, and residual risk in
`docs/fuzz-evidence.md`.

For release candidates and toolchain-security refreshes, preserve raw logs,
workflow artifact links, or both according to the evidence-bundle policy. For
lighter external-review refreshes, committed summaries are acceptable when they
do not make a stronger release-readiness claim, but prefer raw artifacts when
collecting them is low-friction.

If wrapping the command to capture timestamps and logs, avoid shell built-in
names such as zsh's read-only `status`; use a variable such as `rc` for the
command exit code.

## 5. Evidence Refresh

Refresh the evidence docs when the candidate changes security-relevant state:

- `docs/evidence-baseline.md` for the current pinned baseline, candidate target, and stale-trigger index.
- `docs/dependency-review.md` for dependency, vulnerability, and gosec results.
- `docs/fuzz-evidence.md` for fuzz campaign evidence.
- `docs/security-spec-audit.md` for review of `docs/security-assessment.md` and
  `docs/spec-matrix.md` against the exact candidate.
- `docs/project-plan.md` for release-readiness status and remaining blockers.

When a Go toolchain update triggers the refresh, run the vector-stability lane
from `docs/evidence/README.md` under the old and new Go toolchains when both
are available. Preserve the raw logs and explicitly record whether the vector
results are bit-identical. If the previous toolchain is unavailable, state that
limitation instead of implying bit-identical behavior from current-toolchain
tests alone.

## 6. SCA, SAST, And VEX Review

Apply the thresholds in `docs/security-gates.md` before tagging. Do not publish
a release with unresolved SCA or SAST violations unless the finding is fixed,
declared non-exploitable in `docs/vex.md`, or otherwise suppressed with a
documented rationale that reviewers can inspect.

## 7. Release Asset Scope And SBOM

The canonical source release artifact remains the repository content reachable from the signed annotated tag. For signed `v*` tags, the GitHub Release must also include a CycloneDX JSON 1.5 SBOM named `cpace-<tag>.cdx.json` and the SBOM attestation bundle named `cpace-<tag>.cdx.json.sigstore.json`. Tags in the `v0.x` range and SemVer prerelease tags are published as GitHub prereleases and are not marked latest.

Do not publish a release until the Release Validation workflow has generated and validated the SBOM, attested it with GitHub artifact attestations, attached the SBOM and Sigstore bundle, and appended the SBOM SHA-256 checksum to the release body. The checksum is release-body corruption detection only; the SBOM's authenticity comes from the attached attestation bundle and `gh attestation verify`.

The pre-tag local gate remains the prevention layer because Go module proxy and checksum database entries may observe a pushed tag before tag-triggered CI finishes. A failed Release Validation run means cutting a superseding tag, not force-updating the failed tag.

No SLSA Build Level 3 provenance is generated for v1.0.0. ADR-0007 defers Level 3 source-only-module provenance to a later v1.x minor release after the SLSA generator path is validated.

## 8. Source Repository Scope

Current releases are built from this single source repository.

If a future release is made from multiple source repositories, each subproject
must enforce security requirements that are at least as strict as this
repository's requirements for DCO signoff, required tests, branch and tag
protection, dependency review, SAST/SCA handling, secret management, release
signing, and vulnerability disclosure. Record the subproject repositories and
their evidence before tagging.

## 9. GitHub Checks

Before tagging, confirm:

- required `Check` is passing on `main`;
- CodeQL has no unexpected open alerts;
- advisory gosec and vulnerability-scan lanes have no unresolved findings;
- Scorecard results are current enough for the release posture being claimed;
- branch and tag protections are active.

## 10. Signed Tag

Create a signed annotated tag:

```sh
git tag -s vX.Y.Z -m "vX.Y.Z"
git verify-tag vX.Y.Z
git push origin vX.Y.Z
```

Do not force-update a release tag. If a tag is wrong, document the mistake and
cut a new tag.

## 11. Release Validation

After pushing the tag, wait for the tag-triggered Release Validation workflow.
It must pass:

- `Verify Signed Tag`
- `Check`
- `Race`
- `Govulncheck`
- `Gosec` with SARIF upload
- `SBOM` generation, CycloneDX 1.5 validation, and checksum calculation
- `SBOM Attestation` with a GitHub/Sigstore bundle
- `Release` note extraction, asset preparation, and publishing on tag pushes

Confirm Code Scanning has no unexpected open alerts after SARIF ingestion. For rehearsals, `workflow_dispatch` must target a signed tag ref and must exercise verification, SBOM generation, attestation, release-note extraction, and asset preparation without publishing a GitHub Release; branch dispatches should fail with the workflow's unsupported-ref explanation.

## 12. Publish Release

The Release Validation workflow creates the GitHub release on publishing-eligible tag pushes and fails closed if a release for that tag already exists, so manual repair can happen without in-place asset replacement. Verify the published notes state:

- whether the release is production-ready;
- the supported CPace draft, suite, and mode;
- whether Go API, protocol behavior, dependencies, or release evidence changed;
- validation workflow run URL;
- signed-tag verification expectation and a link to
  `docs/release-verification.md`;
- SBOM asset name `cpace-<tag>.cdx.json`, SBOM SHA-256 checksum, and Sigstore bundle asset name `cpace-<tag>.cdx.json.sigstore.json`;
- SLSA Build Level 3 provenance deferral for v1.0.0, if the release notes discuss supply-chain artifact scope;
- remaining blockers, if any.

Keep tags in the `v0.x` range until independent review is complete and the
release bar in `docs/security-assessment.md` is satisfied.
