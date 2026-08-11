#!/bin/sh
set -eu

if [ "$#" -gt 0 ]; then
  registry=$1
else
  repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
  registry="$repo_root/.github/fuzz-targets.json"
fi

fail() {
  printf 'Fuzz input validation failed: %s\n' "$1" >&2
  exit 1
}

validate_positive_integer() {
  name=$1
  value=$2
  case "$value" in
    ''|*[!0-9]*) fail "$name must match ^[0-9]+$ (got: '$value')" ;;
  esac
  case "$value" in
    *[1-9]*) ;;
    *) fail "$name must be positive (got: '$value')" ;;
  esac
  if [ "${#value}" -gt 6 ]; then
    fail "$name value is too large (got: '$value')"
  fi
}

decimal_value() {
  value=$1
  while [ "${value#0}" != "$value" ]; do
    value=${value#0}
  done
  printf '%s\n' "${value:-0}"
}

fuzztime=${FUZZTIME:-}
[ -n "$fuzztime" ] || fail "FUZZTIME is required"
case "$fuzztime" in
  *[smh])
    fuzztime_value=${fuzztime%?}
    fuzztime_unit=${fuzztime#"$fuzztime_value"}
    ;;
  *) fail "FUZZTIME must match ^[0-9]+[smh]$ (got: '$fuzztime')" ;;
esac
case "$fuzztime_value" in
  ''|*[!0-9]*) fail "FUZZTIME must match ^[0-9]+[smh]$ (got: '$fuzztime')" ;;
esac
case "$fuzztime_value" in
  *[1-9]*) ;;
  *) fail "FUZZTIME must be positive (got: '$fuzztime')" ;;
esac
if [ "${#fuzztime_value}" -gt 6 ]; then
  fail "FUZZTIME value is too large (got: '$fuzztime')"
fi

parallel=${PARALLEL:-}
[ -n "$parallel" ] || fail "PARALLEL is required"
validate_positive_integer PARALLEL "$parallel"

fuzz_race=${FUZZ_RACE:-}
case "$fuzz_race" in
  0|1) ;;
  *) fail "FUZZ_RACE must match ^[01]$ (got: '$fuzz_race')" ;;
esac

gomaxprocs=${GOMAXPROCS:-}
if [ -n "$gomaxprocs" ]; then
  validate_positive_integer GOMAXPROCS "$gomaxprocs"
fi

fuzz_test_parallel=${FUZZ_TEST_PARALLEL:-}
if [ -n "$fuzz_test_parallel" ]; then
  validate_positive_integer FUZZ_TEST_PARALLEL "$fuzz_test_parallel"
fi

max_wall_minutes=${FUZZ_MAX_WALL_MINUTES:-}
if [ -n "$max_wall_minutes" ]; then
  validate_positive_integer FUZZ_MAX_WALL_MINUTES "$max_wall_minutes"
fi

command -v jq >/dev/null 2>&1 || fail "jq is required to validate the fuzz target registry"
[ -f "$registry" ] || fail "fuzz target registry not found: $registry"
if ! jq -e . "$registry" >/dev/null; then
  fail "$registry contains malformed JSON"
fi
target_count=$(jq -er 'if type == "array" then length else empty end' "$registry") || fail "$registry must be a JSON array"
if [ "$target_count" -lt 1 ]; then
  fail "$registry contains no targets"
fi

if [ -n "$max_wall_minutes" ]; then
  fuzztime_number=$(decimal_value "$fuzztime_value")
  parallel_number=$(decimal_value "$parallel")
  max_wall_minutes_number=$(decimal_value "$max_wall_minutes")
  case "$fuzztime_unit" in
    s) fuzztime_seconds=$fuzztime_number ;;
    m) fuzztime_seconds=$((fuzztime_number * 60)) ;;
    h) fuzztime_seconds=$((fuzztime_number * 3600)) ;;
  esac
  waves=$(((target_count + parallel_number - 1) / parallel_number))
  estimated_seconds=$((waves * fuzztime_seconds))
  timeout_seconds=$((max_wall_minutes_number * 60))
  if [ "$estimated_seconds" -ge "$timeout_seconds" ]; then
    estimated_minutes=$(((estimated_seconds + 59) / 60))
    fail "ceil(targets/PARALLEL) * FUZZTIME must stay under $max_wall_minutes_number minutes (got about ${estimated_minutes}m)"
  fi
fi

printf 'Validated FUZZTIME=%s, PARALLEL=%s, FUZZ_RACE=%s, GOMAXPROCS=%s, FUZZ_TEST_PARALLEL=%s, TARGETS=%s\n' \
  "$fuzztime" "$parallel" "$fuzz_race" "${gomaxprocs:-default}" "${fuzz_test_parallel:-default}" "$target_count"
