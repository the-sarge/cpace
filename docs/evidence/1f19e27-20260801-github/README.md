# CPace v0.1.3 GitHub Release-Control Evidence

Frozen candidate source: `1f19e278112fa037890848ed6c086addeffdca4e`

Live `main` at capture: `fa4ea296daae82e920ae6bb410e9bcb7c0061641`

Capture interval: 2026-08-01 02:54:40–02:55:00 UTC

This bundle records the mutable GitHub-side controls, alerts, public Scorecard result, and scheduled-workflow status observed immediately before the `v0.1.3` publication work. It is point-in-time prerelease evidence for issue #236, not a production-readiness claim and not a substitute for the exact-candidate dependency, protocol, or paired long-fuzz bundles.

## Results and Disposition

### Tag Authority

Repository ruleset `16048307`, `Protect release tags`, was active for `refs/tags/v*`. Its rules restricted creation, update, and deletion; `bypass_actors` was empty; and `current_user_can_bypass` was `never`. The ruleset's `updated_at` remained `2026-06-10T23:19:18.094-04:00`, so no GitHub-recorded ruleset mutation occurred after the previous baseline capture.

### Main Protection and Required Checks

`main` was protected with strict required status checks, enforcement for administrators, required conversation resolution, force pushes disabled, and deletion disabled. The exact required contexts were `Check`, `DCO`, `Dependency Gate`, and `SAST Gate`, matching `docs/ci-policy.md` without an extra or missing required context.

The frozen candidate source commit is an intermediate commit in PR #239, not the PR head, and GitHub reports zero direct check suites, check runs, or commit statuses for that intermediate SHA. PR #239's checked head `ff73e9ebcbdabfc949e8905caba526cec5057f38` contains the candidate plus only policy/documentation files (`AGENTS.md`, `CHANGELOG.md`, and `docs/*.md`); all four required contexts concluded successfully on that head before protected merge. This bundle therefore does not relabel the intermediate candidate SHA as directly checked: it preserves the limitation and ties the protected merge evidence to the exact checked descendant.

### Alerts and Repository Signal

The open Code Scanning, Dependabot, and secret-scanning alert API responses were all empty arrays. Repository metadata reported a public, active, unarchived repository with secret scanning, secret-scanning push protection, and Dependabot security updates enabled. Empty alert sets are the expected result at capture; they do not prove absence of undiscovered defects.

The current OpenSSF Scorecard API result was 7.4, dated `2026-07-28T10:16:55Z` and evaluated at commit `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6`; the corresponding scheduled workflow run `30350120458` succeeded. Publicly visible deductions included `Maintained=0` because the repository was less than 90 days old, `Code-Review=0` because no approved changesets were detected in the sampled window, and `Branch-Protection=3` because Scorecard prefers approvers and CODEOWNERS review beyond this repository's frozen required-check policy. These are recorded maintenance signals, not silently treated as release-gate passes, and issue #236 makes no policy reopen.

### Scheduled and Background Workflows

- Vulnerability Scan run `30257725792` and Gosec Advisory run `30260278649` both succeeded on schedule at `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6`. They predate the frozen source candidate, so the exact-candidate local `govulncheck` and test-inclusive gosec bundle in `docs/evidence/1f19e27-20260731/` remains the release evidence for those tools.
- The latest CodeQL run at capture, push run `30679756195`, succeeded at live `main` `fa4ea296daae82e920ae6bb410e9bcb7c0061641`; the recent run history also preserves the successful scheduled CodeQL signal.
- Cross-Platform Smoke succeeded on both macOS and Windows for PR #239's checked head in run `30607298901`; the latest scheduled run in the captured history, `30437960712`, also succeeded.
- Nightly Fuzz run `30618250589` attempt 1 ran all 14 jobs but concluded failure when `FuzzInitiatorFinishWithFuzzedMessageB` reported `context deadline exceeded` at 302.05 seconds, just beyond the five-minute fuzz budget; the corpus-upload path was skipped. A failed-job rerun produced attempt 2, which ran the same target for 302.474 seconds, reported `PASS`, and made the workflow conclude successfully. The attempt-specific job logs are committed because the immutable run URL displays the latest attempt by default.
- Autoscaled Fuzz run `30339889549` completed amd64 successfully but cancelled arm64 after 24 hours queued. The next scheduled run, `30433278581`, recovered with successful amd64 and arm64 jobs. Run `30524006270` again completed amd64 and cancelled arm64 after a 24-hour queue wait; at capture, run `30614624510` had completed amd64 successfully while arm64 remained queued. These are self-hosted runner-availability limitations rather than fuzz-test failures. The full two-architecture success is supporting scheduled signal only; the exact-candidate paired long-fuzz evidence remains `docs/evidence/1f19e27-20260731-fuzz/`.

## Workflow Provenance

This README is the workflow-link summary, and `docs/evidence-baseline.md` points to this bundle. GitHub Actions run pages and logs are mutable or deletable according to repository retention and administrator actions; no permanence is claimed, and the capture did not establish a fixed retention period. The committed history, run, and job JSON files below are the fallback for every cited run; the Nightly Fuzz fallback also includes the attempt-specific failed and recovery job logs.

| Workflow | Run, event, ref, and exact SHA | Expected jobs or checks | Committed fallback |
| --- | --- | --- | --- |
| `.github/workflows/scorecard.yml` | [30350120458](https://github.com/the-sarge/cpace/actions/runs/30350120458); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Scorecard` | `scorecard-runs.json`, `scorecard-run-30350120458.json`, `scorecard-run-30350120458-jobs.json`, and `scorecard-api.json` |
| `.github/workflows/vuln.yml` | [30257725792](https://github.com/the-sarge/cpace/actions/runs/30257725792); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Govulncheck` | `vulnerability-runs.json`, `vulnerability-run-30257725792.json`, and `vulnerability-run-30257725792-jobs.json` |
| `.github/workflows/gosec.yml` | [30260278649](https://github.com/the-sarge/cpace/actions/runs/30260278649); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Advisory Gosec` | `gosec-runs.json`, `gosec-run-30260278649.json`, and `gosec-run-30260278649-jobs.json` |
| `.github/workflows/codeql.yml` | [30679756195](https://github.com/the-sarge/cpace/actions/runs/30679756195); `push`; `main`; `fa4ea296daae82e920ae6bb410e9bcb7c0061641` | `Analyze` | `codeql-runs.json`, `codeql-run-30679756195.json`, and `codeql-run-30679756195-jobs.json` |
| `.github/workflows/cross-platform.yml` | [30437960712](https://github.com/the-sarge/cpace/actions/runs/30437960712); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `macos-latest` and `windows-latest` | `cross-platform-runs.json`, `cross-platform-run-30437960712.json`, and `cross-platform-run-30437960712-jobs.json` |
| `.github/workflows/cross-platform.yml` | [30607298901](https://github.com/the-sarge/cpace/actions/runs/30607298901); `pull_request`; `release/v0.1.3-231-232`; `ff73e9ebcbdabfc949e8905caba526cec5057f38` | `macos-latest` and `windows-latest` | `cross-platform-runs.json`, `cross-platform-run-30607298901.json`, and `cross-platform-run-30607298901-jobs.json` |
| `.github/workflows/nightly-fuzz.yml` | [30618250589 attempt 1 and attempt 2](https://github.com/the-sarge/cpace/actions/runs/30618250589); `schedule`; `main`; `228ebfcaea4a15432a7289f2299dfacb9248910d` | `Prepare fuzz matrix` and fourteen `Fuzz (…)` matrix jobs | `nightly-fuzz-runs.json`, attempt-specific run/job JSON, and attempt-specific failed/recovery job logs |
| `.github/workflows/autoscaled-fuzz.yml` | [30339889549](https://github.com/the-sarge/cpace/actions/runs/30339889549); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Validate fuzz inputs`, `Autoscaled fuzz (amd64)`, and `Autoscaled fuzz (arm64)` | `autoscaled-fuzz-runs.json` and `autoscaled-fuzz-30339889549{,-jobs}.json` |
| `.github/workflows/autoscaled-fuzz.yml` | [30433278581](https://github.com/the-sarge/cpace/actions/runs/30433278581); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Validate fuzz inputs`, `Autoscaled fuzz (amd64)`, and `Autoscaled fuzz (arm64)` | `autoscaled-fuzz-runs.json` and `autoscaled-fuzz-30433278581{,-jobs}.json` |
| `.github/workflows/autoscaled-fuzz.yml` | [30524006270](https://github.com/the-sarge/cpace/actions/runs/30524006270); `schedule`; `main`; `741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6` | `Validate fuzz inputs`, `Autoscaled fuzz (amd64)`, and `Autoscaled fuzz (arm64)` | `autoscaled-fuzz-runs.json` and `autoscaled-fuzz-30524006270{,-jobs}.json` |
| `.github/workflows/autoscaled-fuzz.yml` | [30614624510](https://github.com/the-sarge/cpace/actions/runs/30614624510); `schedule`; `main`; `228ebfcaea4a15432a7289f2299dfacb9248910d` | `Validate fuzz inputs`, `Autoscaled fuzz (amd64)`, and `Autoscaled fuzz (arm64)` | `autoscaled-fuzz-runs.json` and `autoscaled-fuzz-30614624510{,-jobs}.json` |

## Contents

| Files | Description |
| --- | --- |
| `capture.sh`, `capture-metadata.txt` | Non-overwriting reproduction script plus repository and candidate identities, implementation commit and clean/dirty state, host/OS/architecture, relevant tool versions, main head, return code, and UTC capture boundaries. |
| `repo.json`, `main-branch.json` | Public repository configuration and the protected live `main` head at capture. |
| `rulesets-list.json`, `ruleset-16048307.json`, `main-branch-protection.json` | Raw tag-ruleset inventory/detail and branch-protection policy. |
| `candidate-*.json`, `pr-239*.json`, `candidate-to-pr-239-head-compare.json` | Candidate identity and empty direct-check state, PR #239 identity/checks, and the candidate-to-checked-head documentation-only comparison. |
| `*-open-alerts.json` | Raw open Code Scanning, Dependabot, and secret-scanning alert responses. |
| `scorecard-api.json`, `scorecard-runs.json`, `scorecard-run-30350120458*.json` | Current Scorecard service result and supporting GitHub workflow run/job state. |
| `vulnerability-*.json`, `gosec-*.json`, `codeql-*.json`, `cross-platform-*.json` | Recent workflow histories plus run/job detail for every cited scheduled or background security/cross-platform run. |
| `nightly-fuzz-*.json`, `nightly-fuzz-*.log` | Nightly Fuzz history plus attempt-specific run, job, failure, and recovery records. |
| `autoscaled-fuzz-*.json` | Autoscaled Fuzz history and detailed job state for the two cancellations, the full architecture recovery, and the live queued run. |
| `SHA256SUMS` | SHA-256 digests for the capture script, metadata, raw API responses, and attempt-specific logs. |

## Verification

On macOS:

```sh
cd docs/evidence/1f19e27-20260801-github
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/1f19e27-20260801-github
sha256sum -c SHA256SUMS
```

To perform an equivalent live recapture without overwriting this bundle, run the script with a new empty directory. Live API responses, workflow state, timestamps, and Scorecard results can legitimately differ and require a new disposition.

```sh
capture_dir=$(mktemp -d)
docs/evidence/1f19e27-20260801-github/capture.sh "$capture_dir"
(cd "$capture_dir" && shasum -a 256 -c SHA256SUMS)
```

## Residual Limitations

The GitHub ruleset, branch protection, alert state, repository configuration, Scorecard service result, and scheduled workflow state are admin- or service-mutable after the capture boundary. Checksums protect the committed byte sequence against unnoticed change within the repository history; they do not authenticate GitHub as the origin. No `SHA256SUMS.sig` is included because no release-authorized signing key was used during branch implementation; the future signed release tag remains the release trust root.

The latest Scorecard, scheduled Vulnerability Scan, scheduled Gosec Advisory, and scheduled cross-platform run were not evaluated at the frozen candidate source SHA. Their commit identities are explicit above and in the raw files. Exact-candidate dependency/SAST, protocol, vector, and paired long-fuzz claims remain pinned to their separate clean-worktree bundles. The final signed-tag Release Validation and post-release Code Scanning recheck remain publication work under issue #237.

At capture, the latest autoscaled arm64 job remained queued even though its amd64 peer had passed; the most recent complete two-architecture success was run `30433278581`. This availability signal does not invalidate the separate exact-candidate paired long-fuzz evidence, but it must not be described as a current all-green autoscaled run.
