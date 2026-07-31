#!/bin/sh
set -eu

candidate=1f19e278112fa037890848ed6c086addeffdca4e
candidate_worktree=${1:?usage: capture.sh <candidate-worktree>}

actual_candidate=$(git -C "$candidate_worktree" rev-parse HEAD)
if [ "$actual_candidate" != "$candidate" ]; then
	echo "candidate mismatch: got $actual_candidate, want $candidate" >&2
	exit 1
fi
if [ -n "$(git -C "$candidate_worktree" status --porcelain=v1)" ]; then
	echo "candidate worktree is dirty" >&2
	exit 1
fi

tool_root=$(mktemp -d "${TMPDIR:-/tmp}/cpace-233-tools.XXXXXX")
tool_bin=$tool_root/bin
mkdir "$tool_bin"
trap 'rm -rf "$tool_root"' EXIT HUP INT TERM

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

module_inventory() {
	go list -m -f '{{if .Main}}{{.Path}} (main){{else}}{{.Path}} {{.Version}} sum={{.Sum}} go_mod_sum={{.GoModSum}}{{end}}' all
}

license_inventory() {
	for module in filippo.io/edwards25519 github.com/gtank/ristretto255; do
		module_dir=$(go list -m -f '{{.Dir}}' "$module")
		module_version=$(go list -m -f '{{.Version}}' "$module")
		license_file=$module_dir/LICENSE
		printf '%s %s\n' "$module" "$module_version"
		shasum -a 256 "$license_file" | sed "s|$license_file|LICENSE|"
		sed -n '1,120p' "$license_file"
	done
}

cd "$candidate_worktree"

printf '# cpace v0.1.3 dependency, vulnerability, and SAST evidence\n'
printf '# candidate=%s\n' "$candidate"
printf '# candidate_worktree=%s\n' "$candidate_worktree"
printf '# transcript_start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# tool_bin=%s\n' "$tool_bin"
printf '# capture_script=docs/evidence/1f19e27-20260731/capture.sh\n'

run "hostname" hostname
run "sw_vers" sw_vers
run "uname -a" uname -a
run "git --version" git --version
run "git rev-parse HEAD" git rev-parse HEAD
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# candidate_worktree_clean=true\n'
run "go version" go version
run "go env GOOS GOARCH GOVERSION GOTOOLCHAIN" go env GOOS GOARCH GOVERSION GOTOOLCHAIN
run "GOBIN=$tool_bin go install golang.org/x/vuln/cmd/govulncheck@v1.3.0" env GOBIN="$tool_bin" go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
run "GOBIN=$tool_bin go install github.com/securego/gosec/v2/cmd/gosec@v2.26.1" env GOBIN="$tool_bin" go install github.com/securego/gosec/v2/cmd/gosec@v2.26.1
run "$tool_bin/govulncheck -version" "$tool_bin/govulncheck" -version
run "go version -m $tool_bin/govulncheck" go version -m "$tool_bin/govulncheck"
run "go version -m $tool_bin/gosec" go version -m "$tool_bin/gosec"
run "go mod verify" go mod verify
run "go list -m all" go list -m all
run "module_inventory" module_inventory
run "go mod graph" go mod graph
run "license_inventory" license_inventory
run "$tool_bin/govulncheck -test -show verbose ./..." "$tool_bin/govulncheck" -test -show verbose ./...
run "$tool_bin/gosec -tests ./..." "$tool_bin/gosec" -tests ./...
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# candidate_worktree_clean_after=true\n'
printf '# transcript_end_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# result=pass\n'
