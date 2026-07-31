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
- `gosec@v2.26.1 -tests ./...` scanned package code and tests and reported 38 files, 8,266 lines, no suppressions, and zero issues.
- The SCA and SAST scans produced no findings requiring a fix, suppression, or `docs/vex.md` entry under `docs/security-gates.md`.

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

To reproduce the capture from a clean detached worktree at the candidate commit:

```sh
docs/evidence/1f19e27-20260731/capture.sh /path/to/clean/candidate-worktree > local-analysis.log 2>&1
```

## Residual Limitations

This packet does not replace the still-pending exact-candidate Capslock, paired long-fuzz, security/spec and vector-stability, GitHub alert and release-control, Scorecard, or Release Validation lanes. A later dependency, toolchain, parser/framing, protocol, security-relevant code, or package-profile change invalidates this lane and requires recapture against the newly selected candidate.

No `SHA256SUMS.sig` is included because no release-authorized signing key was used during branch implementation. The committed hashes provide tamper detection within repository history; the future signed release tag remains the release trust root.
