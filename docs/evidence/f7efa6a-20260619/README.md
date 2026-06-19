# CPace Evidence Bundle: f7efa6a-20260619

Candidate commit: `f7efa6a963a954952b1ecad3f46530f13799fe89`

Date: 2026-06-19 UTC

This bundle refreshes the local dependency/vulnerability/SAST, Capslock, security/spec audit support, paired long-fuzz, tag-ruleset, GitHub status, Scorecard, and vector-stability evidence for the exact candidate commit. It is evidence for an auditable release candidate, not a production-readiness claim; external review and independent cryptographic review remain release blockers.

## Contents

| File | Description |
| --- | --- |
| `local-analysis.log` | Clean-worktree local transcript from an isolated detached clone: host/tool versions, `go mod verify`, `go list -m all`, `govulncheck -test -show verbose ./...`, pinned `gosec@v2.26.1`, Capslock normal and verbose output, pinned Staticcheck, explicit non-cached race test, `task docs:check`, `task quick`, and `task check`. Host-local `GOPATH` and `GOMODCACHE` path values were omitted as non-material local path data. |
| `vector-stability.log` | Draft/vector test stability under the default Go 1.26.4 toolchain and `GOTOOLCHAIN=go1.26.3`. |
| `fuzz-m1mini.log`, `fuzz-imacpro.log` | Final paired long-fuzz transcripts for all 14 registered fuzz targets with `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz`; both ended `RC=0` with `All 14 fuzz targets passed`. |
| `fuzz-*-preflight.txt`, `fuzz-*-setup.log`, `fuzz-*-command.txt`, `fuzz-*-pid`, `fuzz-*-final-status.txt` | Fuzz host setup, command, detached wrapper PID, and compact final status captures. The identical start/finish values are paired-run launcher observations, not independent clock-skew evidence. |
| `rulesets-list.json`, `ruleset-16048307.json`, `ruleset-16048307-verify.json`, `tagruleset-capture.log` | Fresh GitHub tag-ruleset capture and verification for active `refs/tags/v*` create/update/delete protection with no bypass actors and `current_user_can_bypass: never`. |
| `github-runs-for-candidate.json`, `github-status-capture.log` | GitHub Actions status capture for the exact candidate commit. |
| `github-code-scanning-open-alerts.json`, `github-dependabot-open-alerts.json` | Open alert captures; both arrays were empty at capture time. |
| `github-scorecard-20260619-run.json`, `github-scorecard-runs.json` | Fresh OpenSSF Scorecard workflow dispatch result and recent Scorecard run list. |
| `SHA256SUMS` | SHA-256 digests for the files in this bundle. |

## Verification

On macOS:

```sh
cd docs/evidence/f7efa6a-20260619
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/f7efa6a-20260619
sha256sum -c SHA256SUMS
```
