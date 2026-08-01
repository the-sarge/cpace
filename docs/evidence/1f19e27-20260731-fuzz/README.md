# CPace v0.1.3 Paired Long-Fuzz Evidence

Candidate commit: `1f19e278112fa037890848ed6c086addeffdca4e`

Capture interval: 2026-07-31 08:29:08–23:50:10 UTC

Candidate state: clean detached worktree before each campaign and after each recorded run

Hosts and platforms: `m1mini.local`, macOS 15.7.7 (24G720), Darwin 24.6.0, `darwin/arm64`; `iMacPro.local`, macOS 15.7.7 (24G720), Darwin 24.6.0, `darwin/amd64`

Toolchain and runner tools: Go 1.26.5, Task 3.52.0, jq 1.7.1-apple, Git 2.39.5 (Apple Git-154)

This bundle records paired maintainer-machine long fuzz campaigns for all 14 registered fuzz targets against the exact frozen `v0.1.3` candidate, plus same-host triage for the Intel campaign's `FuzzMessageARoundTrip` deadline miss. It is project-side prerelease evidence, not a production-readiness claim.

## Contents

| File | Description |
| --- | --- |
| `capture.sh` | Reproduction script used by both all-target campaigns; it refuses a mismatched or dirty candidate worktree, records host/tool metadata, checks registry/build-input drift, runs registry preflights and `go test -count=1 ./...`, runs `task fuzz`, records timing/exit state, and checks for new fuzz artifacts. |
| `m1mini-setup.log` | ARM setup transcript cloning the exact detached candidate worktree. |
| `m1mini-campaign.log` | ARM all-target campaign transcript. |
| `m1mini-final-status.txt` | ARM wrapper return code and finish timestamp. |
| `imacpro-setup.log` | Intel setup transcript cloning the exact detached candidate worktree. |
| `imacpro-campaign.log` | Intel all-target campaign transcript, including the `FuzzMessageARoundTrip` deadline miss. |
| `imacpro-final-status.txt` | Intel all-target wrapper return code and finish timestamp. |
| `imacpro-FuzzMessageARoundTrip-targeted-rerun.log` | Same-host Intel targeted rerun transcript for `FuzzMessageARoundTrip`. |
| `imacpro-FuzzMessageARoundTrip-targeted-rerun-final-status.txt` | Same-host targeted rerun wrapper return code and finish timestamp. |
| `imacpro-FuzzMessageARoundTrip-targeted-rerun-launch-status.txt` | Same-host targeted rerun launch directory, PID, and start timestamp. |
| `imacpro-FuzzMessageARoundTrip-targeted-rerun-path-failure.log` | Earlier targeted-rerun wrapper attempt that failed before fuzzing because the non-interactive `nohup` environment did not have `go` on `PATH`. |
| `imacpro-FuzzMessageARoundTrip-targeted-rerun-path-failure-launch-status.txt` | Launch metadata for the PATH-failure wrapper attempt. |
| `SHA256SUMS` | SHA-256 digests for the reproduction script, raw transcripts, and status captures. |

## Results and Disposition

- Both setup logs show clean detached clones at candidate commit `1f19e278112fa037890848ed6c086addeffdca4e`.
- Both all-target campaign transcripts record host/platform metadata, Go 1.26.5, Task 3.52.0, jq 1.7.1-apple, Git 2.39.5, clean candidate state, SHA-256 hashes for `.github/fuzz-targets.json`, `fuzz_registry_test.go`, `fuzz_test.go`, `ossfuzz/build.sh`, and `Taskfile.yml`, the full 14-target registry, focused registry drift tests, `go test -count=1 ./...`, and a clean preflight result before fuzzing.
- The ARM campaign on `m1mini.local` ran `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz` from `2026-07-31T08:29:15Z` through `2026-07-31T22:29:31Z`, took 50,416 seconds, passed all 14 registered targets, exited `0`, produced `new_fuzz_artifacts=none`, and left the worktree clean.
- The Intel all-target campaign on `iMacPro.local` ran the same command from `2026-07-31T08:29:19Z` through `2026-07-31T22:29:39Z`, took 50,420 seconds, passed 13 registered targets, and ended nonzero with `FuzzMessageARoundTrip` reporting `context deadline exceeded` at `3600.11s`; it exited `201`, produced `new_fuzz_artifacts=none`, and left the worktree clean.
- The Intel failure is classified as a fuzz shutdown/deadline miss at the configured one-hour target boundary, not as an input-triggered crash, because the transcript recorded no failing corpus artifact and no new `testdata/fuzz` artifact.
- The same-host Intel targeted rerun then ran `FUZZ_RACE=0 GOMAXPROCS=4 go test -timeout=0 -fuzz=FuzzMessageARoundTrip -fuzztime=1h .` from `2026-07-31T22:50:09Z` through `2026-07-31T23:50:10Z`, took 3,601 seconds, passed with `ok github.com/the-sarge/cpace 3600.480s`, exited `0`, produced `new_fuzz_artifacts=none`, and left the worktree clean.
- The first targeted-rerun wrapper launch at `2026-07-31T22:49:12Z` failed before any fuzz command ran because `go` was not on `PATH` in the non-interactive `nohup` environment. The corrected rerun used an explicit `/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin` PATH; the failed wrapper is preserved only as infrastructure context.
- No crash, failing corpus, dirty-worktree residue, registry drift, build-metadata drift, or unreproduced target finding remains open from this campaign set. The evidence nevertheless records the Intel all-target non-pass explicitly; the targeted rerun explains it but does not rewrite the original all-target campaign as clean.

## Verification

On macOS:

```sh
cd docs/evidence/1f19e27-20260731-fuzz
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/1f19e27-20260731-fuzz
sha256sum -c SHA256SUMS
```

To perform an equivalent all-target procedural recapture, run the following command from the repository root after preparing a clean detached worktree at the candidate commit. The new output will contain live timestamps, paths, execution counts, and fuzz scheduling details, so it is not expected to be byte-identical to the committed transcripts and must not overwrite these hash-covered artifacts.

```sh
docs/evidence/1f19e27-20260731-fuzz/capture.sh /path/to/clean/candidate-worktree > /tmp/cpace-235-recapture.log 2>&1
```

## Residual Limitations

The paired all-target campaigns exercised the exact clean candidate on ARM and Intel maintainer machines, but the Intel all-target campaign ended nonzero on a one-hour `FuzzMessageARoundTrip` deadline miss. The same-host one-hour targeted rerun passed and found no new artifacts, which supports treating the Intel all-target failure as a fuzz shutdown/deadline miss rather than a reproducible input finding; it does not convert the original Intel all-target transcript into a clean all-target pass.

`FUZZ_RACE=0` intentionally leaves race instrumentation to the separate release validation lanes so the long campaign can spend its budget on input-space exploration. These maintainer-machine campaigns do not replace hosted continuous fuzzing, scheduled autoscaled fuzzing, external review, independent cryptographic review, or the remaining release-control evidence lanes.

No `SHA256SUMS.sig` is included because no release-authorized signing key was used during branch implementation. The committed hashes provide tamper detection within repository history; the future signed release tag remains the release trust root.
