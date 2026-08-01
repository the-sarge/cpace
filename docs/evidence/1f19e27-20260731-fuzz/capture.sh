#!/bin/sh
set -eu

candidate=1f19e278112fa037890848ed6c086addeffdca4e
candidate_worktree=${1:?usage: capture.sh <candidate-worktree>}
campaign_command='FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz'

actual_candidate=$(git -C "$candidate_worktree" rev-parse HEAD)
if [ "$actual_candidate" != "$candidate" ]; then
	echo "candidate mismatch: got $actual_candidate, want $candidate" >&2
	exit 1
fi
if [ -n "$(git -C "$candidate_worktree" status --porcelain=v1)" ]; then
	echo "candidate worktree is dirty" >&2
	exit 1
fi

run() {
	display=$1
	shift
	printf '\n$ %s\n' "$display"
	printf '# start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	set +e
	"$@" 2>&1
	rc=$?
	set -e
	printf '# end_utc=%s exit=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$rc"
	if [ "$rc" -ne 0 ]; then
		exit "$rc"
	fi
}

cd "$candidate_worktree"
unset FUZZ_TEST_PARALLEL
export GOTOOLCHAIN=go1.26.5

printf '# cpace v0.1.3 paired long-fuzz evidence\n'
printf '# candidate=%s\n' "$candidate"
printf '# candidate_worktree=%s\n' "$candidate_worktree"
printf '# transcript_start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# capture_script=docs/evidence/1f19e27-20260731-fuzz/capture.sh\n'

run "hostname" hostname
run "sw_vers" sw_vers
run "uname -a" uname -a
run "git --version" git --version
run "go version" go version
run "go env GOOS GOARCH GOVERSION GOTOOLCHAIN GOFLAGS GOENV" go env GOOS GOARCH GOVERSION GOTOOLCHAIN GOFLAGS GOENV
run "task --version" task --version
run "jq --version" jq --version
run "git rev-parse HEAD" git rev-parse HEAD
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# candidate_worktree_clean=true\n'
run "shasum -a 256 .github/fuzz-targets.json fuzz_registry_test.go fuzz_test.go ossfuzz/build.sh Taskfile.yml" shasum -a 256 .github/fuzz-targets.json fuzz_registry_test.go fuzz_test.go ossfuzz/build.sh Taskfile.yml
run "jq -r '.[] | [.target, .package, .binary] | @tsv' .github/fuzz-targets.json" jq -r '.[] | [.target, .package, .binary] | @tsv' .github/fuzz-targets.json
registered_targets=$(jq -r 'length' .github/fuzz-targets.json)
printf '# registered_targets=%s\n' "$registered_targets"
run "go test -v -run '^TestFuzzTargetRegistry(Schema|MatchesDefinedTargets|MatchesOSSFuzzBuild)$' -count=1 ." go test -v -run '^TestFuzzTargetRegistry(Schema|MatchesDefinedTargets|MatchesOSSFuzzBuild)$' -count=1 .
run "go test -count=1 ./..." go test -count=1 ./...
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# preflight_result=pass\n'

campaign_marker=.fuzz-campaign-start
touch "$campaign_marker"
campaign_start_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)
campaign_start_epoch=$(date +%s)
printf '\n$ %s\n' "$campaign_command"
printf '# campaign_start_utc=%s\n' "$campaign_start_utc"
printf '# environment=GOTOOLCHAIN=go1.26.5 FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 FUZZ_TEST_PARALLEL=unset\n'
set +e
FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=1 task fuzz 2>&1
campaign_rc=$?
set -e
campaign_end_epoch=$(date +%s)
campaign_end_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)
campaign_duration_seconds=$((campaign_end_epoch - campaign_start_epoch))
printf '# campaign_end_utc=%s\n' "$campaign_end_utc"
printf '# campaign_duration_seconds=%s\n' "$campaign_duration_seconds"
printf '# campaign_exit=%s\n' "$campaign_rc"

new_artifacts=$(find . -type f -path '*/testdata/fuzz/*/*' -newer "$campaign_marker" -print | sort)
rm "$campaign_marker"
if [ -n "$new_artifacts" ]; then
	printf '# new_fuzz_artifacts=present\n%s\n' "$new_artifacts"
else
	printf '# new_fuzz_artifacts=none\n'
fi
run "git status --porcelain=v1" git status --porcelain=v1

printf '# transcript_end_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
if [ "$campaign_rc" -eq 0 ]; then
	printf '# result=pass\n'
else
	printf '# result=non-pass\n'
fi
exit "$campaign_rc"
