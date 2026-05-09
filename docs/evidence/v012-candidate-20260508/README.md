# v0.1.2 Candidate Evidence Transcripts

Date: 2026-05-08

Package-code baseline: `2e09774f171dde8c62763d6e35a258b0fef88801`

These files preserve the raw local evidence used by the v0.1.2 candidate
refresh summaries in [`../../dependency-review.md`](../../dependency-review.md),
[`../../fuzz-evidence.md`](../../fuzz-evidence.md), and
[`../../capslock-report.md`](../../capslock-report.md).

## Files

| File | Contents |
| --- | --- |
| `fuzz-m4mini.log` | Raw `task fuzz` wrapper log from `m4mini.local` for the paired one-hour ARM run. |
| `fuzz-imacpro.log` | Raw `task fuzz` wrapper log from `iMacPro.local` for the paired one-hour Intel run. |
| `local-analysis.log` | Local clean-worktree transcript for Go version, clean status, module list, `task check`, `go mod verify`, `govulncheck`, pinned `gosec`, and Capslock commands. |
| `SHA256SUMS` | SHA-256 digests for the transcript files above. |

## Verification

On macOS:

```sh
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
sha256sum -c SHA256SUMS
```

## Notes

The fuzz and local-analysis logs are maintainer-machine transcripts. They are
committed for review traceability with SHA-256 digests, but they are not an
independent third-party attestation. The package remains unaudited and not
production-ready.
