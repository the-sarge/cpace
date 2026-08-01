#!/bin/sh
set -eu

repo=the-sarge/cpace
candidate=1f19e278112fa037890848ed6c086addeffdca4e
pr_number=239
pr_head=ff73e9ebcbdabfc949e8905caba526cec5057f38
tag_ruleset_id=16048307
nightly_recovery_run=30618250589
nightly_failed_job=91116437432
nightly_recovery_job=91315441946
output_dir=${1:?usage: capture.sh <output-directory>}
script_dir=$(CDPATH= cd "$(dirname "$0")" && pwd)
script_path=$script_dir/$(basename "$0")

mkdir -p "$output_dir"
if [ ! "$script_path" -ef "$output_dir/capture.sh" ]; then
	cp "$script_path" "$output_dir/capture.sh"
fi

artifacts='repo.json main-branch.json rulesets-list.json ruleset-16048307.json main-branch-protection.json candidate-commit.json candidate-check-suites.json candidate-check-runs.json candidate-combined-status.json pr-239.json pr-239-check-runs.json candidate-to-pr-239-head-compare.json code-scanning-open-alerts.json dependabot-open-alerts.json secret-scanning-open-alerts.json scorecard-api.json scorecard-runs.json scorecard-run-30350120458.json vulnerability-runs.json gosec-runs.json codeql-runs.json cross-platform-runs.json nightly-fuzz-runs.json nightly-fuzz-30618250589-attempt-1.json nightly-fuzz-30618250589-attempt-1-jobs.json nightly-fuzz-30618250589-attempt-1-failed-job.log nightly-fuzz-30618250589-attempt-2.json nightly-fuzz-30618250589-attempt-2-jobs.json nightly-fuzz-30618250589-attempt-2-recovery-job.log autoscaled-fuzz-runs.json autoscaled-fuzz-30339889549.json autoscaled-fuzz-30339889549-jobs.json autoscaled-fuzz-30433278581.json autoscaled-fuzz-30433278581-jobs.json autoscaled-fuzz-30524006270.json autoscaled-fuzz-30524006270-jobs.json autoscaled-fuzz-30614624510.json autoscaled-fuzz-30614624510-jobs.json capture-metadata.txt'
for artifact in $artifacts; do
	if [ -e "$output_dir/$artifact" ]; then
		echo "refusing to overwrite $output_dir/$artifact" >&2
		exit 1
	fi
done

capture_start_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)
{
	printf 'capture_start_utc=%s\n' "$capture_start_utc"
	printf 'repository=%s\n' "$repo"
	printf 'candidate=%s\n' "$candidate"
	printf 'pr_number=%s\n' "$pr_number"
	printf 'pr_head=%s\n' "$pr_head"
	printf 'tag_ruleset_id=%s\n' "$tag_ruleset_id"
	printf 'nightly_recovery_run=%s\n' "$nightly_recovery_run"
	gh --version | sed -n '1p'
} > "$output_dir/capture-metadata.txt"

gh api "repos/$repo" > "$output_dir/repo.json"
gh api "repos/$repo/branches/main" > "$output_dir/main-branch.json"
gh api "repos/$repo/rulesets" > "$output_dir/rulesets-list.json"
gh api "repos/$repo/rulesets/$tag_ruleset_id" > "$output_dir/ruleset-16048307.json"
gh api "repos/$repo/branches/main/protection" > "$output_dir/main-branch-protection.json"

gh api "repos/$repo/commits/$candidate" > "$output_dir/candidate-commit.json"
gh api "repos/$repo/commits/$candidate/check-suites?per_page=100" > "$output_dir/candidate-check-suites.json"
gh api "repos/$repo/commits/$candidate/check-runs?per_page=100" > "$output_dir/candidate-check-runs.json"
gh api "repos/$repo/commits/$candidate/status" > "$output_dir/candidate-combined-status.json"
gh api "repos/$repo/pulls/$pr_number" > "$output_dir/pr-239.json"
gh api "repos/$repo/commits/$pr_head/check-runs?per_page=100" > "$output_dir/pr-239-check-runs.json"
gh api "repos/$repo/compare/$candidate...$pr_head" > "$output_dir/candidate-to-pr-239-head-compare.json"

gh api "repos/$repo/code-scanning/alerts?state=open&per_page=100" > "$output_dir/code-scanning-open-alerts.json"
gh api "repos/$repo/dependabot/alerts?state=open&per_page=100" > "$output_dir/dependabot-open-alerts.json"
gh api "repos/$repo/secret-scanning/alerts?state=open&per_page=100" > "$output_dir/secret-scanning-open-alerts.json"

curl -fsSL "https://api.scorecard.dev/projects/github.com/$repo" > "$output_dir/scorecard-api.json"
gh api "repos/$repo/actions/workflows/scorecard.yml/runs?per_page=20" > "$output_dir/scorecard-runs.json"
gh api "repos/$repo/actions/runs/30350120458" > "$output_dir/scorecard-run-30350120458.json"
gh api "repos/$repo/actions/workflows/vuln.yml/runs?per_page=20" > "$output_dir/vulnerability-runs.json"
gh api "repos/$repo/actions/workflows/gosec.yml/runs?per_page=20" > "$output_dir/gosec-runs.json"
gh api "repos/$repo/actions/workflows/codeql.yml/runs?per_page=20" > "$output_dir/codeql-runs.json"
gh api "repos/$repo/actions/workflows/cross-platform.yml/runs?per_page=20" > "$output_dir/cross-platform-runs.json"

gh api "repos/$repo/actions/workflows/nightly-fuzz.yml/runs?per_page=20" > "$output_dir/nightly-fuzz-runs.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/1" > "$output_dir/nightly-fuzz-30618250589-attempt-1.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/1/jobs?per_page=100" > "$output_dir/nightly-fuzz-30618250589-attempt-1-jobs.json"
gh api --allow-escape-sequences "repos/$repo/actions/jobs/$nightly_failed_job/logs" | sed 's/ $//' > "$output_dir/nightly-fuzz-30618250589-attempt-1-failed-job.log"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/2" > "$output_dir/nightly-fuzz-30618250589-attempt-2.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/2/jobs?per_page=100" > "$output_dir/nightly-fuzz-30618250589-attempt-2-jobs.json"
gh api --allow-escape-sequences "repos/$repo/actions/jobs/$nightly_recovery_job/logs" | sed 's/ $//' > "$output_dir/nightly-fuzz-30618250589-attempt-2-recovery-job.log"

gh api "repos/$repo/actions/workflows/autoscaled-fuzz.yml/runs?per_page=20" > "$output_dir/autoscaled-fuzz-runs.json"
for run_id in 30339889549 30433278581 30524006270 30614624510; do
	gh api "repos/$repo/actions/runs/$run_id" > "$output_dir/autoscaled-fuzz-$run_id.json"
	gh api "repos/$repo/actions/runs/$run_id/jobs?per_page=100" > "$output_dir/autoscaled-fuzz-$run_id-jobs.json"
done

for artifact in "$output_dir"/*.json; do
	jq -e . "$artifact" >/dev/null
done

jq -e '.enforcement == "active" and .target == "tag" and .conditions.ref_name.include == ["refs/tags/v*"] and .bypass_actors == [] and .current_user_can_bypass == "never" and ([.rules[].type] | sort) == (["creation", "deletion", "update"] | sort)' "$output_dir/ruleset-16048307.json" >/dev/null
jq -e '.required_status_checks.strict == true and ([.required_status_checks.checks[].context] | sort) == (["Check", "DCO", "Dependency Gate", "SAST Gate"] | sort)' "$output_dir/main-branch-protection.json" >/dev/null
jq -e '.total_count == 0' "$output_dir/candidate-check-suites.json" >/dev/null
jq -e '.total_count == 0' "$output_dir/candidate-check-runs.json" >/dev/null
jq -e '.check_runs as $runs | ["Check", "DCO", "Dependency Gate", "SAST Gate"] | all(. as $required | any($runs[]; .name == $required and .conclusion == "success"))' "$output_dir/pr-239-check-runs.json" >/dev/null
jq -e 'all(.files[]; (.filename | endswith(".md")) or .filename == "AGENTS.md")' "$output_dir/candidate-to-pr-239-head-compare.json" >/dev/null
jq -e 'length == 0' "$output_dir/code-scanning-open-alerts.json" >/dev/null
jq -e 'length == 0' "$output_dir/dependabot-open-alerts.json" >/dev/null
jq -e 'length == 0' "$output_dir/secret-scanning-open-alerts.json" >/dev/null
jq -e '.score == 7.4 and .repo.commit == "741be4ab58ec0cd86e4cb2dcc1da39de8d6348d6"' "$output_dir/scorecard-api.json" >/dev/null
jq -e '.run_attempt == 1 and .conclusion == "failure"' "$output_dir/nightly-fuzz-30618250589-attempt-1.json" >/dev/null
jq -e '.run_attempt == 2 and .conclusion == "success"' "$output_dir/nightly-fuzz-30618250589-attempt-2.json" >/dev/null

capture_end_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)
printf 'main_head=%s\n' "$(jq -r '.commit.sha' "$output_dir/main-branch.json")" >> "$output_dir/capture-metadata.txt"
printf 'capture_end_utc=%s\n' "$capture_end_utc" >> "$output_dir/capture-metadata.txt"

(cd "$output_dir" && shasum -a 256 capture.sh $artifacts > SHA256SUMS)
