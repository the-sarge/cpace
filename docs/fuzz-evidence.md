# Fuzz Evidence

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Evidence code commit: `955855b58424a8868d318096149be615bb3989da`

Registered fuzz targets: 14 from `.github/fuzz-targets.json`

## Command

- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz`

## Candidate Long Run

Host: `mbp128.local`

Platform: `darwin/arm64`

Toolchain: Go 1.26.2, Task 3.50.0

Started: `2026-05-08T02:34:04Z`

Finished: `2026-05-08T03:30:17Z`

Command: `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz`

Result: PASS: all 14 registered fuzz targets.

With 14 targets and `PARALLEL=2`, the command executes seven target batches.
The recorded wall-clock duration is about 56 minutes, matching seven 8-minute
batches and confirming that every target ran the full `FUZZTIME=8m`, above the
five-minute release-bar minimum. `FUZZ_RACE=0` leaves race detection to
`task check` and uses the long campaign for input-space exploration.
`GOMAXPROCS=4` keeps each fuzzing subprocess from oversubscribing the host.

Run note: the local timestamp-capture wrapper returned non-zero after the fuzz
task completed because it attempted to assign to zsh's read-only `status`
parameter. The `task fuzz` output itself recorded 14 `=== PASS:` target lines
and `All 14 fuzz targets passed`; the wrapper issue happened after the fuzz
campaign completed and is not treated as a fuzz failure. Future ad hoc
wrappers should use a variable such as `rc`, not `status`.

## Historical Long Runs

The previous paired ARM/Intel long campaign remains useful historical evidence,
but it is exact to commit `06f21c51645f54e2b7bde7c5b538479463be5d0e` and does
not cover the PR #40 fuzz-harness refactor or OSS-Fuzz seed-guard changes.

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

The 2026-05-08 candidate run refreshes long-fuzz evidence for the merged PR #40
code commit on local `darwin/arm64`. It is not paired Intel evidence and does
not replace continuous fuzzing. Repeat long fuzzing if parser, protocol, fuzz
harness, dependency, or toolchain changes land before a release tag. Rerun on
an Intel host if cross-architecture release evidence is required. This refresh
is sufficient for the current external-review packet; paired Intel evidence
would strengthen a future exact-release-candidate packet.

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
