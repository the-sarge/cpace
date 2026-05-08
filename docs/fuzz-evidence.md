# Fuzz Evidence

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Evidence code commit: `737bc56ffba81e2df5e9caa0df1ff180bfdb594b`

Registered fuzz targets: 14 from `.github/fuzz-targets.json`

## Command

- `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=2 task fuzz`

## Candidate Paired Long Runs

The Go 1.26.3 security release changed the toolchain baseline after the earlier
Go 1.26.2 evidence. These paired maintainer-machine runs refresh all registered
targets under Go 1.26.3 on ARM and Intel hosts.

| Host | Platform | Toolchain | Started | Finished | Result |
| --- | --- | --- | --- | --- | --- |
| `m4mini.local` | `darwin/arm64` | Go 1.26.3, Task 3.50.0 | `2026-05-08T09:09:50Z` | `2026-05-08T16:09:59Z` | PASS: all 14 targets, `RC=0` |
| `iMacPro.local` | `darwin/amd64` | Go 1.26.3, Task 3.50.0 | `2026-05-08T09:09:50Z` | `2026-05-08T16:10:10Z` | PASS: all 14 targets, `RC=0` |

With 14 targets, `PARALLEL=2`, and `FUZZTIME=1h`, the command executes seven
one-hour target batches. The recorded wall-clock duration on each host matches
that expected schedule and is consistent with each target receiving the
configured `FUZZTIME=1h`. `FUZZ_RACE=0` leaves race detection to `task check`
and uses the long campaign for input-space exploration. `GOMAXPROCS=4` keeps
each fuzzing subprocess from oversubscribing the host.

Raw maintainer-machine logs and SHA-256 digests are committed under
`docs/evidence/go1263-20260508/`. The `task fuzz` logs preserve host, commit,
Go version, Task version, command, timestamps, return code, and the synthesized
per-target PASS set, but they do not preserve Go's per-target fuzz counter
output. `fuzz-worktree-status.log` records the detached `HEAD`, clean
`git status --short --branch`, successful `git diff --exit-code --stat`, Go
version, Task version, and `go env GOOS GOARCH` for both fuzz worktrees.

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

The previous Go 1.26.2 refresh remains useful historical signal, but it has
been superseded by the paired Go 1.26.3 candidate runs above.

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

The 2026-05-08 Go 1.26.3 candidate runs refresh paired ARM/Intel long-fuzz
evidence for package-code commit
`737bc56ffba81e2df5e9caa0df1ff180bfdb594b`. They do not replace continuous
fuzzing or upstream OSS-Fuzz coverage. The Go 1.26 `go fix` modernization
touches `crypto.go` and `framing.go`, so these runs become historical after
that modernization merges. Repeat long fuzzing if parser, protocol, fuzz
harness, dependency, or toolchain changes land before a release tag. Exact
release candidates still need evidence recorded against the exact candidate
commit.

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
