# CPace v0.1.3 Dependency and SAST Evidence

Candidate commit: `1f19e278112fa037890848ed6c086addeffdca4e`

Capture interval: 2026-07-31 06:58:34–06:58:39 UTC

Candidate state: clean detached worktree before and after capture

Host and platform: `mbp128.local`, macOS 26.6 (25G72), Darwin 25.6.0, `arm64`

Toolchain and scanners: Go 1.26.5, `govulncheck@v1.3.0` with vulnerability database updated 2026-07-27 20:14:16 UTC, and `gosec@v2.26.1` built with Go 1.26.5. The gosec binary's embedded Go module metadata records the pinned `github.com/securego/gosec/v2 v2.26.1` version and checksum.

This bundle records the dependency inventory, module-integrity verification, test-aware vulnerability analysis, and test-inclusive pinned gosec scan for the exact frozen `v0.1.3` candidate. It is one release-evidence lane for an explicitly non-production-ready prerelease, not a production-readiness claim.

## Contents

| File | Description |
| --- | --- |
| `capture.sh` | Reproduction script pinned to the candidate and scanner versions; it refuses a mismatched or dirty candidate worktree. |
| `local-analysis.log` | Raw command transcript with candidate state, host/platform, tool versions and embedded module metadata, dependency versions and sums, module graph, dependency-license hashes and texts, module verification, govulncheck output, gosec output, command timestamps, and exit codes. |
| `SHA256SUMS` | SHA-256 digests for the reproduction script and raw transcript. |

## Results and Disposition

- The runtime module graph contains only the expected direct dependency `github.com/gtank/ristretto255 v0.2.0` and indirect dependency `filippo.io/edwards25519 v1.2.0`; their module and `go.mod` sums match `go.sum`, and `go mod verify` reports `all modules verified`.
- Both dependency license files are unchanged BSD-3-Clause texts with hashes preserved in the transcript. No unexpected dependency, unknown license, or incompatible license was found.
- `govulncheck -test -show verbose ./...` scanned the package, its tests, both dependencies, and the Go 1.26.5 standard library and reported `No vulnerabilities found`.
- `gosec@v2.26.1 -tests ./...` reported per-analysis-pass totals of 38 files and 8,266 lines, no suppressions, and zero issues. Because gosec analyzes the 13 non-test files twice under `-tests`, those totals represent 25 distinct root-module files and 6,787 distinct lines.
- The captured manual dependency/license review, `govulncheck`, and test-inclusive gosec results produced no finding requiring a fix, suppression, or `docs/vex.md` entry under their applicable `docs/security-gates.md` thresholds. The completed point-in-time GitHub release-control bundle records empty Code Scanning, Dependabot, and secret-scanning alert sets, Scorecard 7.4, ast-grep through the final local gate, and the passing `Dependency Gate` and other required contexts on PR #239's documentation-only checked descendant; the frozen source SHA itself remains without direct checks. Signed-tag Release Validation and post-SARIF Code Scanning remain publication work.

## Verification

On macOS:

```sh
cd docs/evidence/1f19e27-20260731
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/1f19e27-20260731
sha256sum -c SHA256SUMS
```

To perform an equivalent procedural recapture, run the following command from the repository root. The new output will contain live timestamps, temporary paths, and vulnerability-database metadata, so it is not expected to be byte-identical to the committed transcript and must not overwrite that hash-covered artifact.

```sh
docs/evidence/1f19e27-20260731/capture.sh /path/to/clean/candidate-worktree > /tmp/cpace-233-recapture.log 2>&1
```

## Residual Limitations

The captured `./...` inventory and scans cover the root `github.com/the-sarge/cpace` module and exclude the separate modules under `tools/`. This lane is complemented by the separate exact-candidate Capslock, security/spec, and vector-stability evidence in `docs/evidence/1f19e27-20260731-protocol/`, the paired long-fuzz evidence in `docs/evidence/1f19e27-20260731-fuzz/`, and the completed point-in-time GitHub alert, release-control, and Scorecard evidence in `docs/evidence/1f19e27-20260801-github/`; the assembled packet is `docs/evidence/1f19e27-20260801-candidate/`. Signed-tag Release Validation and post-SARIF Code Scanning remain publication work. A later dependency, toolchain, parser/framing, protocol, security-relevant code, or package-profile change invalidates this lane and requires recapture against the newly selected candidate.

No `SHA256SUMS.sig` is included because no release-authorized signing key was used during branch implementation. The committed hashes provide tamper detection within repository history; the future signed release tag remains the release trust root.
