#!/bin/sh
set -eu

candidate=1f19e278112fa037890848ed6c086addeffdca4e
historical_audit=f7efa6a963a954952b1ecad3f46530f13799fe89
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

tool_root=$(mktemp -d "${TMPDIR:-/tmp}/cpace-234-tools.XXXXXX")
vector_root=$(mktemp -d "${TMPDIR:-/tmp}/cpace-234-vectors.XXXXXX")
tool_bin=$tool_root/bin
mkdir "$tool_bin"
trap 'rm -rf "$tool_root" "$vector_root"' EXIT HUP INT TERM

vector_regex='Test(StringUtilitiesDraftVectors|IRTranscriptDraftVectorFlow|EmbeddedDraft.*|RistrettoDraft21Vectors|CoreDraft21Vectors|ScalarMultVFYDraftInvalidVectors|BuildCIWireStability|WireFormatPrefixByte)$'
protocol_regex='Test(ConfirmedExchangeAndExport|SessionPeerMetadata|SessionClose.*|SessionValueCopiesShareCloseState|NilReceiverMethods|FinishCleanupDoesNotAliasReturnedSessions|ClearOnFinishFailurePaths|SessionISKSurvivesCoreClear|SingleUseState.*|InputValidation|InputFieldSizeLimits|ProtocolAllowsEmptyLocalAssociatedData|ProtocolAllowsEmptySessionIDWithCompatibilityFlag|ProtocolRejectsAsymmetricSessionID|RoleLocalIdentityReversalFailsConfirmation|TranscriptLockingMismatches|ProtocolAbortsOnInvalidRistrettoEncoding|InitiatorAbortsOnInvalidResponderShare|PeerShare.*|WireLengthRejectionIsMessageNotPeerShare|ScalarMultVFY.*|MessageFramingCatalogue.*|BuildCIWireStability|WireFormatPrefixByte)$'
vector_1264_raw=$vector_root/go1.26.4.json
vector_1265_raw=$vector_root/go1.26.5.json
vector_1264_normalized=$vector_root/go1.26.4.normalized.json
vector_1265_normalized=$vector_root/go1.26.5.normalized.json
protocol_raw=$vector_root/protocol-audit.log

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

run_to_file() {
	display=$1
	output=$2
	shift 2
	printf '\n$ %s\n' "$display"
	printf '# start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	set +e
	"$@" >"$output" 2>&1
	rc=$?
	set -e
	cat "$output"
	printf '# end_utc=%s exit=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$rc"
	if [ "$rc" -ne 0 ]; then
		exit "$rc"
	fi
}

normalize_vector_json() {
	jq -c '
		del(.Time, .Elapsed)
		| if has("Output") then
			.Output |= (
				gsub("\\([0-9]+(\\.[0-9]+)?s\\)"; "(DURATION)")
				| gsub("\\t[0-9]+(\\.[0-9]+)?s\\n$"; "\\tDURATION\\n")
			)
		else
			.
		end
	' "$1" >"$2"
}

reject_empty_selection() {
	log=$1
	if grep -F "no tests to run" "$log" >/dev/null; then
		echo "test selection matched no tests: $log" >&2
		return 1
	fi
}

require_json_test_names() {
	log=$1
	shift
	reject_empty_selection "$log"
	for test_name in "$@"; do
		if ! grep -F '"Action":"run"' "$log" | grep -F "\"Test\":\"$test_name\"" >/dev/null; then
			echo "required test missing from transcript: $test_name" >&2
			return 1
		fi
	done
}

require_text_test_names() {
	log=$1
	shift
	reject_empty_selection "$log"
	for test_name in "$@"; do
		if ! grep -F -x "=== RUN   $test_name" "$log" >/dev/null; then
			echo "required test missing from transcript: $test_name" >&2
			return 1
		fi
	done
}

compare_vector_results() {
	hash_1264=$(shasum -a 256 "$vector_1264_normalized" | awk '{print $1}')
	hash_1265=$(shasum -a 256 "$vector_1265_normalized" | awk '{print $1}')
	lines_1264=$(wc -l <"$vector_1264_normalized" | tr -d ' ')
	lines_1265=$(wc -l <"$vector_1265_normalized" | tr -d ' ')
	printf 'go1.26.4 normalized_sha256=%s lines=%s\n' "$hash_1264" "$lines_1264"
	printf 'go1.26.5 normalized_sha256=%s lines=%s\n' "$hash_1265" "$lines_1265"
	cmp -s "$vector_1264_normalized" "$vector_1265_normalized"
	printf 'vector_result=bit-identical\n'
}

cd "$candidate_worktree"

printf '# cpace v0.1.3 toolchain, vector, capability, and protocol evidence\n'
printf '# candidate=%s\n' "$candidate"
printf '# historical_audit=%s\n' "$historical_audit"
printf '# candidate_worktree=%s\n' "$candidate_worktree"
printf '# transcript_start_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# tool_bin=%s\n' "$tool_bin"
printf '# capture_script=docs/evidence/1f19e27-20260731-protocol/capture.sh\n'

run "hostname" hostname
run "sw_vers" sw_vers
run "uname -a" uname -a
run "git --version" git --version
run "jq --version" jq --version
run "git rev-parse HEAD" git rev-parse HEAD
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# candidate_worktree_clean=true\n'
run "go version" go version
run "go env GOOS GOARCH GOVERSION GOTOOLCHAIN" go env GOOS GOARCH GOVERSION GOTOOLCHAIN
run "env GOTOOLCHAIN=go1.26.4 go version" env GOTOOLCHAIN=go1.26.4 go version
run "env GOTOOLCHAIN=go1.26.4 go env GOOS GOARCH GOVERSION GOTOOLCHAIN" env GOTOOLCHAIN=go1.26.4 go env GOOS GOARCH GOVERSION GOTOOLCHAIN
run "env GOTOOLCHAIN=go1.26.5 go version" env GOTOOLCHAIN=go1.26.5 go version
run "env GOTOOLCHAIN=go1.26.5 go env GOOS GOARCH GOVERSION GOTOOLCHAIN" env GOTOOLCHAIN=go1.26.5 go env GOOS GOARCH GOVERSION GOTOOLCHAIN
run "git log --oneline --no-decorate $historical_audit..$candidate" git log --oneline --no-decorate "$historical_audit..$candidate"
run "git diff --name-status $historical_audit..$candidate -- '*.go' go.mod go.sum" git diff --name-status "$historical_audit..$candidate" -- '*.go' go.mod go.sum
run "go doc -all ." go doc -all .
run_to_file "env GOTOOLCHAIN=go1.26.4 go test -json -v -run '$vector_regex' -count=1 ./..." "$vector_1264_raw" env GOTOOLCHAIN=go1.26.4 go test -json -v -run "$vector_regex" -count=1 ./...
run_to_file "env GOTOOLCHAIN=go1.26.5 go test -json -v -run '$vector_regex' -count=1 ./..." "$vector_1265_raw" env GOTOOLCHAIN=go1.26.5 go test -json -v -run "$vector_regex" -count=1 ./...
run "require_json_test_names go1.26.4 vectors" require_json_test_names "$vector_1264_raw" TestStringUtilitiesDraftVectors TestIRTranscriptDraftVectorFlow TestEmbeddedDraftVectorJSON TestEmbeddedDraftGeneratorJSON TestEmbeddedDraftConfirmationTagGoldens TestEmbeddedDraftInvalidVectorJSON TestRistrettoDraft21Vectors TestCoreDraft21Vectors TestScalarMultVFYDraftInvalidVectors TestBuildCIWireStability TestWireFormatPrefixByte
run "require_json_test_names go1.26.5 vectors" require_json_test_names "$vector_1265_raw" TestStringUtilitiesDraftVectors TestIRTranscriptDraftVectorFlow TestEmbeddedDraftVectorJSON TestEmbeddedDraftGeneratorJSON TestEmbeddedDraftConfirmationTagGoldens TestEmbeddedDraftInvalidVectorJSON TestRistrettoDraft21Vectors TestCoreDraft21Vectors TestScalarMultVFYDraftInvalidVectors TestBuildCIWireStability TestWireFormatPrefixByte
run "normalize_vector_json go1.26.4" normalize_vector_json "$vector_1264_raw" "$vector_1264_normalized"
run "normalize_vector_json go1.26.5" normalize_vector_json "$vector_1265_raw" "$vector_1265_normalized"
run "compare_vector_results" compare_vector_results
run "GOBIN=$tool_bin go install github.com/google/capslock/cmd/capslock@v0.3.2" env GOBIN="$tool_bin" go install github.com/google/capslock/cmd/capslock@v0.3.2
run "$tool_bin/capslock -version" "$tool_bin/capslock" -version
run "go version -m $tool_bin/capslock" go version -m "$tool_bin/capslock"
run "$tool_bin/capslock -packages ./..." "$tool_bin/capslock" -packages ./...
run "$tool_bin/capslock -packages ./... -output=verbose" "$tool_bin/capslock" -packages ./... -output=verbose
run_to_file "env GOTOOLCHAIN=go1.26.5 go test -v -run '$protocol_regex' -count=1 ./..." "$protocol_raw" env GOTOOLCHAIN=go1.26.5 go test -v -run "$protocol_regex" -count=1 ./...
run "require_text_test_names protocol audit" require_text_test_names "$protocol_raw" TestConfirmedExchangeAndExport TestSessionClose TestFinishCleanupDoesNotAliasReturnedSessions TestSingleUseStateCloseCleansAbandonedState TestInputValidation TestMessageFramingCatalogueRejectsMalformed TestPeerShareErrorsWrapErrAbort TestBuildCIWireStability TestWireFormatPrefixByte
run "git status --porcelain=v1" git status --porcelain=v1
run "git diff --exit-code --no-ext-diff" git diff --exit-code --no-ext-diff
printf '# candidate_worktree_clean_after=true\n'
printf '# transcript_end_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf '# result=pass\n'
