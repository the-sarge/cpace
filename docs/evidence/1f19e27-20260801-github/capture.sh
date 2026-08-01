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
repo_root=$(git -C "$script_dir" rev-parse --show-toplevel)
capture_impl_head=$(git -C "$repo_root" rev-parse HEAD)
capture_impl_branch=$(git -C "$repo_root" branch --show-current)
capture_worktree_status=$(git -C "$repo_root" status --porcelain=v1 --untracked-files=all)

if [ -e "$output_dir" ]; then
	if [ ! -d "$output_dir" ]; then
		echo "output path is not a directory: $output_dir" >&2
		exit 1
	fi
	if [ -n "$(find "$output_dir" -mindepth 1 -maxdepth 1 -print -quit)" ]; then
		echo "refusing to write to non-empty output directory: $output_dir" >&2
		exit 1
	fi
else
	mkdir -p "$output_dir"
fi
cp "$script_path" "$output_dir/capture.sh"

capture_run() {
	gh api "repos/$repo/actions/runs/$1" > "$output_dir/$2.json"
	gh api "repos/$repo/actions/runs/$1/jobs?per_page=100" > "$output_dir/$2-jobs.json"
}

capture_start_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)
{
	printf 'capture_start_utc=%s\n' "$capture_start_utc"
	printf 'repository=%s\n' "$repo"
	printf 'candidate=%s\n' "$candidate"
	printf 'pr_number=%s\n' "$pr_number"
	printf 'pr_head=%s\n' "$pr_head"
	printf 'tag_ruleset_id=%s\n' "$tag_ruleset_id"
	printf 'nightly_recovery_run=%s\n' "$nightly_recovery_run"
	printf 'capture_implementation_head=%s\n' "$capture_impl_head"
	printf 'capture_implementation_branch=%s\n' "$capture_impl_branch"
	if [ -n "$capture_worktree_status" ]; then
		printf 'capture_worktree_state=dirty\n'
		printf 'capture_worktree_status_begin\n%s\ncapture_worktree_status_end\n' "$capture_worktree_status"
	else
		printf 'capture_worktree_state=clean\n'
	fi
	printf 'hostname=%s\n' "$(hostname)"
	printf 'operating_system=%s\n' "$(uname -s)"
	printf 'kernel_release=%s\n' "$(uname -r)"
	printf 'architecture=%s\n' "$(uname -m)"
	printf 'git_version=%s\n' "$(git --version)"
	printf 'go_version=%s\n' "$(go version)"
	printf 'gh_version=%s\n' "$(gh --version | sed -n '1p')"
	printf 'curl_version=%s\n' "$(curl --version | sed -n '1p')"
	printf 'jq_version=%s\n' "$(jq --version)"
	printf 'shasum_version=%s\n' "$(shasum --version | sed -n '1p')"
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
capture_run 30350120458 scorecard-run-30350120458
gh api "repos/$repo/actions/workflows/vuln.yml/runs?per_page=20" > "$output_dir/vulnerability-runs.json"
capture_run 30257725792 vulnerability-run-30257725792
gh api "repos/$repo/actions/workflows/gosec.yml/runs?per_page=20" > "$output_dir/gosec-runs.json"
capture_run 30260278649 gosec-run-30260278649
gh api "repos/$repo/actions/workflows/codeql.yml/runs?per_page=20" > "$output_dir/codeql-runs.json"
capture_run 30679756195 codeql-run-30679756195
gh api "repos/$repo/actions/workflows/cross-platform.yml/runs?per_page=20" > "$output_dir/cross-platform-runs.json"
capture_run 30437960712 cross-platform-run-30437960712
capture_run 30607298901 cross-platform-run-30607298901

gh api "repos/$repo/actions/workflows/nightly-fuzz.yml/runs?per_page=20" > "$output_dir/nightly-fuzz-runs.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/1" > "$output_dir/nightly-fuzz-30618250589-attempt-1.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/1/jobs?per_page=100" > "$output_dir/nightly-fuzz-30618250589-attempt-1-jobs.json"
gh api --allow-escape-sequences "repos/$repo/actions/jobs/$nightly_failed_job/logs" | sed 's/ $//' > "$output_dir/nightly-fuzz-30618250589-attempt-1-failed-job.log"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/2" > "$output_dir/nightly-fuzz-30618250589-attempt-2.json"
gh api "repos/$repo/actions/runs/$nightly_recovery_run/attempts/2/jobs?per_page=100" > "$output_dir/nightly-fuzz-30618250589-attempt-2-jobs.json"
gh api --allow-escape-sequences "repos/$repo/actions/jobs/$nightly_recovery_job/logs" | sed 's/ $//' > "$output_dir/nightly-fuzz-30618250589-attempt-2-recovery-job.log"

gh api "repos/$repo/actions/workflows/autoscaled-fuzz.yml/runs?per_page=20" > "$output_dir/autoscaled-fuzz-runs.json"
for run_id in 30339889549 30433278581 30524006270 30614624510; do
	capture_run "$run_id" "autoscaled-fuzz-$run_id"
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
printf 'capture_exit=%s\n' 0 >> "$output_dir/capture-metadata.txt"

artifact_files=$(cd "$output_dir" && find . -type f ! -name SHA256SUMS -print | sed 's#^\./##' | LC_ALL=C sort)
(cd "$output_dir" && shasum -a 256 $artifact_files > SHA256SUMS)
