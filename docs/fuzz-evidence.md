# Fuzz Evidence

Date: 2026-05-06

Target module: `github.com/the-sarge/cpace`

Evidence commit: `06f21c51645f54e2b7bde7c5b538479463be5d0e`

Registered fuzz targets: 14 from `.github/fuzz-targets.json`

## Commands

- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=10s PARALLEL=2 task fuzz`
- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz`

## Local Smoke Run

Host: `mbp128.local`

Platform: `darwin/arm64`

Toolchain: Go 1.26.2, Task 3.50.0

Command: `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=10s PARALLEL=2 task fuzz`

Result: all 14 registered fuzz targets passed.

Purpose: local smoke evidence after the LEB128 parser cleanup,
`FuzzProtocolMismatch` guard fix, and `task fuzz` timeout adjustment. This is
not the release-bar long-fuzz evidence by itself.

## Long Runs

| Host | Platform | Started | Finished | Command | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | `2026-05-06T12:35:36Z` | `2026-05-06T13:31:42Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets |
| `iMacPro.local` | `darwin/amd64` | `2026-05-06T12:35:36Z` | `2026-05-06T13:31:44Z` | `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` | PASS: all 14 targets |

Each long run uses the same 14-target registry. With 14 targets and
`PARALLEL=2`, each run executes seven target batches. The recorded wall-clock
duration is about 56 minutes, matching seven 8-minute batches and confirming
that every target ran the full `FUZZTIME=8m`, above the five-minute release-bar
minimum. `FUZZ_RACE=0` leaves race detection to `task check` and uses the long
campaign for input-space exploration. `GOMAXPROCS=4` keeps each fuzzing
subprocess from oversubscribing the shared host.

## Residual Risk

The long runs are release-readiness evidence for the exact code commit above.
Repeat them if parser, protocol, fuzz harness, dependency, or toolchain changes
land before a release tag. Longer continuous fuzzing and OSS-Fuzz remain later
investigations.

Earlier local and remote attempts included 4-hour local `m4max` runs, 15-minute
remote ARM/Intel runs, and race-enabled 8-minute remote runs on intermediate
commits such as `97993dee7354ab306705920d369a9a2b20fafc32` and
`c93ec8e2e4b09d62e1d368489cb86ddcca0ed4d8`. Those attempts failed, were
interrupted, or covered superseded commits. Some Intel-only intermediate runs
passed, but the paired ARM/Intel runs did not pass until the final
`06f21c51645f54e2b7bde7c5b538479463be5d0e` evidence commit with
`FUZZ_RACE=0`. The earlier attempts produced no crash corpus. They led to the
`task fuzz` `-timeout=0` and `FUZZ_RACE=0` long-campaign adjustments in this
evidence commit and are treated as runner-budget or exploratory attempts, not
passing release evidence.
