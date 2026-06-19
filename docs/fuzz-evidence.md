# Fuzz Evidence

Date: 2026-05-08 through 2026-06-19

Target module: `github.com/the-sarge/cpace`

Evidence code commit: `f7efa6a963a954952b1ecad3f46530f13799fe89`

Superseded candidate commit: `2e09774f171dde8c62763d6e35a258b0fef88801`

Supplemental tag commit: `4e661bc1f925ebedf1f270668129d85bab73e468`
(`v0.1.2`)

Registered fuzz targets: 14 from `.github/fuzz-targets.json`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether these pinned fuzz runs are fresh for the latest release candidate.

## Command

- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz` (current Go 1.26.4 baseline runs)
- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=2 task fuzz` (earlier campaigns below)

## Go 1.26.4 Exact-Candidate Paired Long Runs

These paired maintainer-machine runs refresh all 14 registered targets under Go 1.26.4 at commit `f7efa6a963a954952b1ecad3f46530f13799fe89`, covering the accepted-ADR implementation sequence, issue #80, PR #199's Go fix modernization, and PR #200's development-journal merge. Raw logs, final status captures, setup/preflight files, and SHA-256 digests are committed under `docs/evidence/f7efa6a-20260619/`.

| Host | Platform | Toolchain | Started | Finished | Result |
| --- | --- | --- | --- | --- | --- |
| `m1mini.local` | `darwin/arm64` | Go 1.26.4, Task 3.51.1 | `2026-06-19T03:45:46Z` | `2026-06-19T17:45:59Z` | PASS: all 14 targets, `RC=0` |
| `iMacPro.local` | `darwin/amd64` | Go 1.26.4, Task 3.51.1 | `2026-06-19T03:45:46Z` | `2026-06-19T17:45:59Z` | PASS: all 14 targets, `RC=0` |

With `PARALLEL=1` the 14 targets run sequentially, so the expected wall clock is fourteen hours per host; both hosts match that schedule. The wrapper logs embed paired-run launcher start/end observations, detached worktree commit, clean status, command, final `All 14 fuzz targets passed` marker, and return code. The identical start/end values across hosts are not used as independent clock-skew evidence; host independence is corroborated by distinct platform/toolchain preflight captures, run directories, setup logs, and detached wrapper PIDs. `FUZZ_RACE=0` leaves race instrumentation to the explicit non-cached `go test -race -count=1 ./...` lane and `task check` in the exact-candidate local analysis transcript.

## Go 1.26.4 Baseline Paired Long Runs

These paired maintainer-machine runs are historical signal for commit `933ece246e6170b11e838395bf36f852cba0cd02` — the combined-trigger baseline (go1.26.4 toolchain security release of 2026-06-02 plus the PR #73 package-code changes) shared with the dependency/Capslock/audit refresh in `docs/evidence/go1264-20260611/`. They served as the pre-refactor fuzz baseline for the ADR-0001 core-extraction sequence and are superseded by the exact-candidate paired long runs above. The ARM leg ran on `mbp128.local` rather than `m4mini.local` to avoid contention with the scheduled autoscaled-fuzz window.

| Host | Platform | Toolchain | Started | Finished | Result |
| --- | --- | --- | --- | --- | --- |
| `mbp128.local` | `darwin/arm64` | Go 1.26.4, Task 3.51.1 | `2026-06-11T07:13:34Z` | `2026-06-11T21:13:55Z` | PASS: all 14 targets |
| `iMacPro.local` | `darwin/amd64` | Go 1.26.4, Task 3.51.1 | `2026-06-11T07:15:10Z` | `2026-06-11T21:15:33Z` | PASS: all 14 targets |

With `PARALLEL=1` the 14 targets run sequentially, so the expected wall clock is fourteen hours per host; both hosts match it to within a minute. Unlike the `v012-candidate` wrapper logs, these raw `task fuzz` transcripts do not embed timestamps or a return code: start times come from the separate worktree-status captures (detached worktree at the baseline commit, clean status, toolchain versions), finish times are the final log writes observed at copy time, and the `All 14 fuzz targets passed` line is the success marker the task script prints only when no per-target `.fail` files exist. Raw logs, status captures, and SHA-256 digests are committed under `docs/evidence/go1264-20260611/`.

**Supersession note:** ADR-0003 and the later accepted-ADR implementation sequence landed after these 2026-06-11 campaigns. The exact-candidate paired long runs above are the covering campaign for the post-implementation shape.

## ADR-0001 Interim Gate (Non-Evidence)

This local interim gate does not replace or advance the pinned long-fuzz evidence above. It records only the ADR-0001 build-sequence check on the feature branch; it is superseded by the 2026-06-19 exact-candidate paired long campaign.

| Host | Platform | Toolchain | Commit | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- | --- | --- |
| local worktree | `darwin/arm64` | Go 1.26.4, Task 3.51.1 | `7aa79e4a40304a14610df36d0bd906fd6c7e3a24` | `2026-06-12T04:08:47Z` | `2026-06-12T05:04:52Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets, `rc=0` |

## ADR-0002 Interim Gate (Non-Evidence)

This local interim gate does not replace or advance the pinned long-fuzz evidence above. It records only the ADR-0002 build-sequence check on the feature branch; it is superseded by the 2026-06-19 exact-candidate paired long campaign.

| Host | Platform | Toolchain | Commit | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `mbp128.local` | `darwin/arm64` | Go 1.26.4, Task 3.51.1 | `03a35cc3dac55b3dfa798c218bf335761cc7502c` | `2026-06-12T14:48:13Z` | `2026-06-12T15:44:29Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets, `rc=0` |

## Supplemental v0.1.2 Tag Soak

After the signed `v0.1.2` prerelease tag was cut at commit
`4e661bc1f925ebedf1f270668129d85bab73e468`, paired maintainer-machine soak
runs exercised the same 14-target registry with `FUZZTIME=4h` and `PARALLEL=2`.
That schedule gives roughly 28 hours of wall-clock runtime per host.

| Host | Platform | Scope | Started | Finished | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | all 14 targets, `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=4h PARALLEL=2 task fuzz` | `2026-05-09T04:42:35Z` | `2026-05-10T08:42:44Z` | PASS: all 14 targets, `rc=0` |
| `iMacPro.local` | `darwin/amd64` | all 14 targets, `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=4h PARALLEL=2 task fuzz` | `2026-05-09T04:42:35Z` | `2026-05-10T08:42:46Z` | NON-PASS: 13 targets passed; `FuzzProtocolConsistency` reported `context deadline exceeded` at `14400.12s`, `rc=201` |
| `iMacPro.local` | `darwin/amd64` | targeted recovery rerun, `FUZZ_RACE=0 GOMAXPROCS=4 go test -timeout=0 -fuzz=FuzzProtocolConsistency -fuzztime=4h .` | `2026-05-10T23:44:41Z` | `2026-05-11T03:44:44Z` | PASS: `FuzzProtocolConsistency`, `676273567` execs, no new interesting inputs, `rc=0` |

The targeted Intel rerun supports treating the all-target Intel failure as an
isolated fuzz shutdown/deadline miss rather than an input-triggered crash. It
does not change the raw all-target Intel transcript into a clean all-target
pass. No failing corpus file was present in the Intel scratch worktrees when
checked after the failed all-target run and after the targeted rerun.

Raw maintainer-machine transcripts and SHA-256 digests are committed under
`docs/evidence/v012-soak-20260509/`.

## Candidate Paired Long Runs

These runs remain useful historical signal, but they have been superseded by
the Go 1.26.4 baseline runs above because PR #73 touched `api.go`, `crypto.go`,
and `session.go` and the toolchain moved to go1.26.4.

The Go 1.26 `go fix` modernization touched `crypto.go` and `framing.go` after
the earlier Go 1.26.3 evidence. These paired maintainer-machine runs refresh all
registered targets under Go 1.26.3 on ARM and Intel hosts for the v0.1.2
package-code candidate.

| Host | Platform | Toolchain | Started | Finished | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | Go 1.26.3, Task 3.50.0 | `2026-05-08T19:05:29Z` | `2026-05-09T02:05:36Z` | PASS: all 14 targets, `RC=0` |
| `iMacPro.local` | `darwin/amd64` | Go 1.26.3, Task 3.50.0 | `2026-05-08T19:05:28Z` | `2026-05-09T02:05:38Z` | PASS: all 14 targets, `RC=0` |

With 14 targets, `PARALLEL=2`, and `FUZZTIME=1h`, the command executes seven
one-hour target batches. The recorded wall-clock duration on each host matches
that expected schedule and is consistent with each target receiving the
configured `FUZZTIME=1h`. `FUZZ_RACE=0` leaves race detection to `task check`
and uses the long campaign for input-space exploration. `GOMAXPROCS=4` keeps
each fuzzing subprocess from oversubscribing the host.

Raw maintainer-machine logs and SHA-256 digests are committed under
`docs/evidence/v012-candidate-20260508/`. The `task fuzz` logs preserve host,
commit, clean detached worktree status, successful `git diff --exit-code
--stat`, Go version, Task version, `go env GOOS GOARCH`, command, timestamps,
return code, and the synthesized per-target PASS set, but they do not preserve
Go's per-target fuzz counter output.

Both logs recorded the full target pass set:

```text
=== PASS: FuzzDecodeMessageA ===
=== PASS: FuzzDecodeMessageB ===
=== PASS: FuzzDecodeMessageC ===
=== PASS: FuzzDraftInvalidVectorJSONLoader ===
=== PASS: FuzzDraftVectorJSONLoader ===
=== PASS: FuzzInitiatorFinishWithFuzzedMessageB ===
=== PASS: FuzzMessageARoundTrip ===
=== PASS: FuzzMessageBRoundTrip ===
=== PASS: FuzzMessageCRoundTrip ===
=== PASS: FuzzProtocolConsistency ===
=== PASS: FuzzProtocolMismatch ===
=== PASS: FuzzRespondWithFuzzedMessageA ===
=== PASS: FuzzResponderFinishWithFuzzedMessageC ===
=== PASS: FuzzScalarMultVFY ===
All 14 fuzz targets passed
```

## Historical Long Runs

The previous Go 1.26.3 refresh remains useful historical signal, but it has
been superseded by the paired v0.1.2 candidate runs above because PR #45 touched
`crypto.go` and `framing.go`.

| Host | Platform | Commit | Toolchain | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | `737bc56ffba81e2df5e9caa0df1ff180bfdb594b` | Go 1.26.3, Task 3.50.0 | `2026-05-08T09:09:50Z` | `2026-05-08T16:09:59Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=2 task fuzz` | PASS: all 14 targets |
| `iMacPro.local` | `darwin/amd64` | `737bc56ffba81e2df5e9caa0df1ff180bfdb594b` | Go 1.26.3, Task 3.50.0 | `2026-05-08T09:09:50Z` | `2026-05-08T16:10:10Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=2 task fuzz` | PASS: all 14 targets |

The previous Go 1.26.2 refresh also remains useful historical signal, but it
has been superseded by later paired Go 1.26.3 candidate runs.

| Host | Platform | Commit | Toolchain | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `mbp128.local` | `darwin/arm64` | `955855b58424a8868d318096149be615bb3989da` | Go 1.26.2, Task 3.50.0 | `2026-05-08T02:34:04Z` | `2026-05-08T03:30:17Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets |

The earlier paired ARM/Intel long campaign remains historical evidence for
commit `06f21c51645f54e2b7bde7c5b538479463be5d0e`. It does not cover the PR #40
fuzz-harness refactor or OSS-Fuzz seed-guard changes.

| Host | Platform | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | `2026-05-06T12:35:36Z` | `2026-05-06T13:31:42Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets |
| `iMacPro.local` | `darwin/amd64` | `2026-05-06T12:35:36Z` | `2026-05-06T13:31:44Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets |

An earlier no-race split campaign on commit
`07ff1e9265c2e003e6dc7d37754c8b2185f03286` ran each then-registered fuzz
target for four hours. That commit had seven registered fuzz targets, before
the later expansion to 14 targets.

| Host | Platform | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | `2026-05-05T14:02:10Z` | `2026-05-06T06:02:14Z` | `FUZZ_RACE=0 FUZZTIME=4h` split target run | PASS: `FuzzDecodeMessageA`, `FuzzDecodeMessageB`, `FuzzDecodeMessageC`, `FuzzDraftVectorJSONLoader` |
| `iMacPro.local` | `darwin/amd64` | `2026-05-05T14:02:35Z` | `2026-05-06T02:02:42Z` | `FUZZ_RACE=0 FUZZTIME=4h` split target run | PASS: `FuzzDraftInvalidVectorJSONLoader`, `FuzzProtocolConsistency`, `FuzzProtocolMismatch` |

## Residual Risk

The 2026-06-19 Go 1.26.4 exact-candidate runs are the current paired ARM/Intel long-fuzz evidence, pinned to `f7efa6a963a954952b1ecad3f46530f13799fe89`. The 2026-06-11 Go 1.26.4 baseline runs cover the superseded PR #73-era baseline, the earlier Go 1.26.3 candidate runs cover the superseded candidate `2e09774f171dde8c62763d6e35a258b0fef88801`, and the 2026-05-09 through 2026-05-11 supplemental soak records stronger fuzz duration against the signed `v0.1.2` tag commit `4e661bc1f925ebedf1f270668129d85bab73e468`, with a clean ARM all-target soak, an Intel all-target soak that ended nonzero on `FuzzProtocolConsistency`, and a clean same-host targeted rerun of that target. These runs do not replace continuous fuzzing or upstream OSS-Fuzz coverage. Repeat long fuzzing if parser, protocol, fuzz harness, dependency, or toolchain changes land before a future release tag. Production-readiness still requires completion of the remaining release blockers.

The 4-hour split campaign is strong historical signal for the seven-target
registry at commit `07ff1e9265c2e003e6dc7d37754c8b2185f03286`, but it is not
exact-candidate evidence for the current 14-target harness. Longer continuous
fuzzing and OSS-Fuzz upstream review/merge monitoring remain ongoing
release-readiness work.

Earlier race-enabled remote attempts included `FUZZTIME=8h` on `m4mini.local`
and `FUZZTIME=24h` on `iMacPro.local` for commit
`07ff1e9265c2e003e6dc7d37754c8b2185f03286`; both were terminated with exit
status 143 before producing all-target pass evidence. Other intermediate runs
on commits such as `97993dee7354ab306705920d369a9a2b20fafc32` and
`c93ec8e2e4b09d62e1d368489cb86ddcca0ed4d8` failed, were interrupted, or
covered superseded commits. The earlier failed/interrupted attempts produced no
crash corpus. They led to the `task fuzz` `-timeout=0` and `FUZZ_RACE=0`
long-campaign adjustments and are treated as runner-budget or exploratory
attempts, not passing release evidence.
