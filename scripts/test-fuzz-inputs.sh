#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
validator="$repo_root/scripts/validate-fuzz-inputs.sh"
fixture_dir=$(mktemp -d)
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM

valid_registry="$fixture_dir/valid.json"
empty_registry="$fixture_dir/empty.json"
invalid_registry="$fixture_dir/invalid.json"
non_array_registry="$fixture_dir/non-array.json"
printf '%s\n' '[{"target":"FuzzOne"},{"target":"FuzzTwo"},{"target":"FuzzThree"}]' >"$valid_registry"
printf '%s\n' '[]' >"$empty_registry"
printf '%s\n' '{' >"$invalid_registry"
printf '%s\n' '{}' >"$non_array_registry"
mkdir "$fixture_dir/no-jq"

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

assert_passes() {
  description=$1
  shift
  if ! "$@" >"$fixture_dir/output" 2>&1; then
    printf 'Expected pass: %s\n' "$description" >&2
    cat "$fixture_dir/output" >&2
    exit 1
  fi
}

assert_fails() {
  description=$1
  expected=$2
  shift 2
  if "$@" >"$fixture_dir/output" 2>&1; then
    fail "Expected failure: $description"
  fi
  if ! grep -F "$expected" "$fixture_dir/output" >/dev/null; then
    printf 'Wrong failure for %s; expected text: %s\n' "$description" "$expected" >&2
    cat "$fixture_dir/output" >&2
    exit 1
  fi
}

run_validator() {
  registry=$1
  shift
  (
    unset FUZZTIME PARALLEL FUZZ_RACE GOMAXPROCS FUZZ_TEST_PARALLEL FUZZ_MAX_WALL_MINUTES
    export FUZZTIME=10m PARALLEL=2 FUZZ_RACE=0
    for assignment in "$@"; do
      export "$assignment"
    done
    "$validator" "$registry"
  )
}

run_without_jq() {
  (
    export FUZZTIME=10m PARALLEL=2 FUZZ_RACE=0
    PATH="$fixture_dir/no-jq" /bin/sh "$validator" "$valid_registry"
  )
}

assert_passes "required inputs" run_validator "$valid_registry"
assert_passes "optional worker limits" run_validator "$valid_registry" GOMAXPROCS=4 FUZZ_TEST_PARALLEL=2
assert_passes "duration units" run_validator "$valid_registry" FUZZTIME=1s
assert_passes "uncapped local long fuzz" run_validator "$valid_registry" FUZZTIME=999999h PARALLEL=1
assert_passes "just below workflow wall cap" run_validator "$valid_registry" FUZZTIME=59m PARALLEL=2 FUZZ_MAX_WALL_MINUTES=120

assert_fails "missing FUZZTIME" "FUZZTIME is required" run_validator "$valid_registry" FUZZTIME=
assert_fails "invalid FUZZTIME shape" "FUZZTIME must match" run_validator "$valid_registry" FUZZTIME=10ms
assert_fails "zero FUZZTIME" "FUZZTIME must be positive" run_validator "$valid_registry" FUZZTIME=0m
assert_fails "oversized FUZZTIME" "FUZZTIME value is too large" run_validator "$valid_registry" FUZZTIME=1000000m
assert_fails "zero PARALLEL" "PARALLEL must be positive" run_validator "$valid_registry" PARALLEL=0
assert_fails "non-numeric PARALLEL" "PARALLEL must match" run_validator "$valid_registry" PARALLEL=two
assert_fails "oversized PARALLEL" "PARALLEL value is too large" run_validator "$valid_registry" PARALLEL=1000000
assert_fails "invalid FUZZ_RACE" "FUZZ_RACE must match" run_validator "$valid_registry" FUZZ_RACE=true
assert_fails "zero GOMAXPROCS" "GOMAXPROCS must be positive" run_validator "$valid_registry" GOMAXPROCS=0
assert_fails "oversized GOMAXPROCS" "GOMAXPROCS value is too large" run_validator "$valid_registry" GOMAXPROCS=1000000
assert_fails "zero FUZZ_TEST_PARALLEL" "FUZZ_TEST_PARALLEL must be positive" run_validator "$valid_registry" FUZZ_TEST_PARALLEL=0
assert_fails "oversized FUZZ_TEST_PARALLEL" "FUZZ_TEST_PARALLEL value is too large" run_validator "$valid_registry" FUZZ_TEST_PARALLEL=1000000
assert_fails "invalid wall cap" "FUZZ_MAX_WALL_MINUTES must be positive" run_validator "$valid_registry" FUZZ_MAX_WALL_MINUTES=0
assert_fails "wall cap reached" "must stay under 120 minutes" run_validator "$valid_registry" FUZZTIME=60m PARALLEL=2 FUZZ_MAX_WALL_MINUTES=120
assert_fails "empty registry" "contains no targets" run_validator "$empty_registry"
assert_fails "invalid registry" "contains malformed JSON" run_validator "$invalid_registry"
assert_fails "non-array registry" "must be a JSON array" run_validator "$non_array_registry"
assert_fails "missing registry" "fuzz target registry not found" run_validator "$fixture_dir/missing.json"
assert_fails "missing jq" "jq is required" run_without_jq

printf '%s\n' "Fuzz input validator tests passed"
