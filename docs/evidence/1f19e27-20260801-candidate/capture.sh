#!/bin/sh
set -eu

frozen_source=1f19e278112fa037890848ed6c086addeffdca4e
gate_worktree=${1:?usage: capture.sh <clean-gate-worktree>}
packet_path=docs/evidence/1f19e27-20260801-candidate

gate_commit=$(git -C "$gate_worktree" rev-parse HEAD)
gate_tree=$(git -C "$gate_worktree" rev-parse 'HEAD^{tree}')
if ! git -C "$gate_worktree" merge-base --is-ancestor "$frozen_source" HEAD; then
	echo "gate commit is not a descendant of frozen source $frozen_source" >&2
	exit 1
fi
if [ -n "$(git -C "$gate_worktree" status --porcelain=v1)" ]; then
	echo "gate worktree is dirty" >&2
	exit 1
fi

tool_root=$(mktemp -d "${TMPDIR:-/tmp}/cpace-237-tools.XXXXXX")
tool_bin=$tool_root/bin
mkdir "$tool_bin"
trap 'rm -rf "$tool_root"' EXIT HUP INT TERM

run() {
	display=$1
	shift
	printf '\n$ %s\n' "$display"
	start_time=$(date +%s)
	printf '# start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	set +e
	"$@" 2>&1
	rc=$?
	set -e
	end_time=$(date +%s)
	printf '# end_utc=%s duration_seconds=%s exit=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$((end_time - start_time))" "$rc"
	if [ "$rc" -ne 0 ]; then
		exit "$rc"
	fi
}

postfreeze_scope() {
	changed_paths=$(git diff --name-only "$frozen_source"..HEAD) || return 1
	printf '# changed_paths_since_frozen_source\n%s\n' "$changed_paths"
	unexpected_paths=$(printf '%s\n' "$changed_paths" | sed -e '/^AGENTS\.md$/d' -e '/^CHANGELOG\.md$/d' -e '/^DEV-JOURNAL\.md$/d' -e '/^docs\//d' -e '/^$/d') || return 1
	if [ -n "$unexpected_paths" ]; then
		printf 'unexpected post-freeze paths:\n%s\n' "$unexpected_paths" >&2
		return 1
	fi
	printf '# postfreeze_scope=documentation-and-evidence-only\n'
}

verify_manifest_links() {
	while read -r key bundle expected; do
		manifest_bundle=$(sed -n "s/^${key}_bundle=//p" "$packet_path/packet-manifest.txt") || return 1
		manifest_digest=$(sed -n "s/^${key}_sha256sums_sha256=//p" "$packet_path/packet-manifest.txt") || return 1
		if [ "$manifest_bundle" != "$bundle" ] || [ "$manifest_digest" != "$expected" ]; then
			echo "manifest mismatch for $key" >&2
			return 1
		fi
		actual=$(shasum -a 256 "$bundle/SHA256SUMS" | awk '{print $1}') || return 1
		printf '%s expected=%s actual=%s\n' "$bundle" "$expected" "$actual"
		if [ "$actual" != "$expected" ]; then
			return 1
		fi
		(cd "$bundle" && shasum -a 256 -c SHA256SUMS) || return 1
	done <<'EOF'
dependency_sast docs/evidence/1f19e27-20260731 f5bbef2b529dc7b05642ce3d67297394f3b28d50129cfd7f2cbe60ea7e02ce67
protocol_vector docs/evidence/1f19e27-20260731-protocol 5a696b17949d0343d6940d784f452861b95f2beaaf74f08d3cff8dbad1dc4ce0
paired_fuzz docs/evidence/1f19e27-20260731-fuzz ac5f7df2fe4bbc8863efcd0f5580347b921675eee3ec9dbdaddf32c628173878
github_controls docs/evidence/1f19e27-20260801-github 81ac3508d548357e779ea9efb94c80a37209eaa127bf016db8d4ce38ec29bc99
EOF
}

compare_release_notes() {
	extracted=$tool_root/release-notes.txt
	scripts/extract-release-notes.sh CHANGELOG.md v0.1.3 >"$extracted" || return 1
	cmp "$extracted" "$packet_path/release-notes.txt" || return 1
}

candidate_packet_checksums() {
	(cd "$packet_path" && shasum -a 256 -c SHA256SUMS) || return 1
}

verify_release_metadata() {
	metadata=$(scripts/release-tag-metadata.sh v0.1.3) || return 1
	printf '%s\n' "$metadata"
	printf '%s\n' "$metadata" | grep -Fxq 'release-tag=v0.1.3' || return 1
	printf '%s\n' "$metadata" | grep -Fxq 'sbom-file=cpace-v0.1.3.cdx.json' || return 1
	printf '%s\n' "$metadata" | grep -Fxq 'prerelease=true' || return 1
	printf '%s\n' "$metadata" | grep -Fxq 'latest=false' || return 1
	grep -Fq 'bundle_dst="dist/${SBOM_FILE}.sigstore.json"' .github/workflows/release.yml || return 1
	printf 'attestation-bundle=cpace-v0.1.3.cdx.json.sigstore.json\n'
}

cd "$gate_worktree"

printf '# cpace v0.1.3 final pre-tag gate\n'
printf '# frozen_source_candidate=%s\n' "$frozen_source"
printf '# gate_commit=%s\n' "$gate_commit"
printf '# gate_tree=%s\n' "$gate_tree"
printf '# transcript_start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# tool_bin=%s\n' "$tool_bin"
printf '# capture_script=%s/capture.sh\n' "$packet_path"
printf '\n$ %s/capture.sh %s\n' "$packet_path" "$gate_worktree"

run "hostname" hostname
run "sw_vers" sw_vers
run "uname -a" uname -a
run "git --version" git --version
run "git rev-parse HEAD" git rev-parse HEAD
run "git rev-parse HEAD^{tree}" git rev-parse 'HEAD^{tree}'
run "git merge-base --is-ancestor $frozen_source HEAD" git merge-base --is-ancestor "$frozen_source" HEAD
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
run "postfreeze_scope" postfreeze_scope
run "go version" go version
run "go env GOOS GOARCH GOVERSION GOTOOLCHAIN" go env GOOS GOARCH GOVERSION GOTOOLCHAIN
run "task --version" task --version
run "jq --version" jq --version
run "cmark --version" cmark --version
run "verify_manifest_links" verify_manifest_links
run "candidate_packet_checksums" candidate_packet_checksums
run "compare_release_notes" compare_release_notes
run "verify_release_metadata" verify_release_metadata
run "scripts/check-release-policy.sh" scripts/check-release-policy.sh
run "task docs:check" task docs:check
run "task quick" task quick
run "task check" task check
run "go test -race ./..." go test -race ./...
run "go test -cover ./..." go test -cover ./...
run "go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./..." go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...
run "GOBIN=$tool_bin go install golang.org/x/vuln/cmd/govulncheck@v1.3.0" env GOBIN="$tool_bin" go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
run "$tool_bin/govulncheck -version" "$tool_bin/govulncheck" -version
run "$tool_bin/govulncheck -test -show verbose ./..." "$tool_bin/govulncheck" -test -show verbose ./...
run "go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -tests ./..." go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -tests ./...
run "GOBIN=$tool_bin go install github.com/google/capslock/cmd/capslock@v0.3.2" env GOBIN="$tool_bin" go install github.com/google/capslock/cmd/capslock@v0.3.2
run "$tool_bin/capslock -version" "$tool_bin/capslock" -version
run "$tool_bin/capslock -packages ./..." "$tool_bin/capslock" -packages ./...
run "$tool_bin/capslock -packages ./... -output=verbose" "$tool_bin/capslock" -packages ./... -output=verbose
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# transcript_end_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# result=pass\n'
