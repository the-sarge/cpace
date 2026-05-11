# v0.1.2 Supplemental Soak Fuzz Transcripts

Date: 2026-05-09 through 2026-05-11

Target tag: `v0.1.2`

Target commit: `4e661bc1f925ebedf1f270668129d85bab73e468`

These files preserve supplemental maintainer-machine fuzz transcripts for the
signed `v0.1.2` prerelease tag. They supplement the package-code candidate
evidence summarized in [`../../fuzz-evidence.md`](../../fuzz-evidence.md).

## Files

| File | Contents |
| --- | --- |
| `fuzz-m4mini-4h.log` | Raw `task fuzz` wrapper log from `m4mini.local` for the paired 4-hour-per-target ARM soak. |
| `fuzz-imacpro-4h.log` | Raw `task fuzz` wrapper log from `iMacPro.local` for the paired 4-hour-per-target Intel soak. This transcript records an end-of-target `FuzzProtocolConsistency` deadline failure and nonzero wrapper return code. |
| `fuzz-imacpro-protocol-consistency-rerun.log` | Raw targeted rerun transcript from `iMacPro.local` for `FuzzProtocolConsistency` with `-fuzztime=4h`; this rerun passed. |
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

The all-target soak used `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=4h PARALLEL=2 task
fuzz`. With 14 registered targets and `PARALLEL=2`, the expected wall-clock
duration is roughly 28 hours.

The `m4mini.local` soak completed cleanly with all 14 targets passing. The
`iMacPro.local` all-target soak completed its wall-clock schedule but returned
nonzero because `FuzzProtocolConsistency` reported `context deadline exceeded`
at its 4-hour boundary. A same-host targeted 4-hour rerun of
`FuzzProtocolConsistency` passed cleanly. Treat this bundle as supplemental
evidence plus recovery evidence, not as two clean paired all-target soak passes.

These are maintainer-machine transcripts committed for review traceability with
SHA-256 digests. They are not independent third-party attestations. No detached
signature is included for `SHA256SUMS`.
