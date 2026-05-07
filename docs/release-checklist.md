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

## 3. Long Fuzz Evidence

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

## 4. Evidence Refresh

Refresh the evidence docs when the candidate changes security-relevant state:

- `docs/dependency-review.md` for dependency, vulnerability, and gosec results.
- `docs/fuzz-evidence.md` for fuzz campaign evidence.
- `docs/security-spec-audit.md` for review of `docs/security-assessment.md` and
  `docs/spec-matrix.md` against the exact candidate.
- `docs/project-plan.md` for release-readiness status and remaining blockers.

## 5. SCA, SAST, And VEX Review

Apply the thresholds in `docs/security-gates.md` before tagging. Do not publish
a release with unresolved SCA or SAST violations unless the finding is fixed,
declared non-exploitable in `docs/vex.md`, or otherwise suppressed with a
documented rationale that reviewers can inspect.

## 6. Release Asset Scope And SBOM

Current releases do not attach compiled software assets. The canonical source
release artifact is the repository content reachable from the signed annotated
tag.

If a future release adds compiled assets, each compiled asset must be delivered
with a software bill of materials. The SBOM may be one release-level SBOM that
maps every compiled asset to its components, or one SBOM per asset. Release
notes must identify the SBOM location, and the asset verification instructions
must cover the asset hashes or signed manifest described in
`docs/release-verification.md`.

Do not publish a release containing compiled assets until the SBOM is generated
and attached or otherwise made available from the release notes.

## 7. Source Repository Scope

Current releases are built from this single source repository.

If a future release is made from multiple source repositories, each subproject
must enforce security requirements that are at least as strict as this
repository's requirements for DCO signoff, required tests, branch and tag
protection, dependency review, SAST/SCA handling, secret management, release
signing, and vulnerability disclosure. Record the subproject repositories and
their evidence before tagging.

## 8. GitHub Checks

Before tagging, confirm:

- required `Check` is passing on `main`;
- CodeQL has no unexpected open alerts;
- advisory gosec and vulnerability-scan lanes have no unresolved findings;
- Scorecard results are current enough for the release posture being claimed;
- branch and tag protections are active.

## 9. Signed Tag

Create a signed annotated tag:

```sh
git tag -s vX.Y.Z -m "vX.Y.Z"
git verify-tag vX.Y.Z
git push origin vX.Y.Z
```

Do not force-update a release tag. If a tag is wrong, document the mistake and
cut a new tag.

## 10. Release Validation

After pushing the tag, wait for the tag-triggered Release Validation workflow.
It must pass:

- `Check`
- `Race`
- `Govulncheck`
- `Gosec` with SARIF upload

Confirm Code Scanning has no unexpected open alerts after SARIF ingestion.

## 11. Publish Release

Create the GitHub release or prerelease. Notes should state:

- whether the release is production-ready;
- the supported CPace draft, suite, and mode;
- whether Go API, protocol behavior, dependencies, or release evidence changed;
- validation workflow run URL;
- signed-tag verification expectation and a link to
  `docs/release-verification.md`;
- SBOM location for any compiled release assets;
- remaining blockers, if any.

Keep tags in the `v0.x` range until independent review is complete and the
release bar in `docs/security-assessment.md` is satisfied.
