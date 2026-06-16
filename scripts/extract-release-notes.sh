#!/bin/sh
set -eu

usage() {
  echo "usage: $0 CHANGELOG.md vX.Y.Z" >&2
}

if [ "$#" -ne 2 ]; then
  usage
  exit 2
fi

changelog=$1
tag=$2
semver_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$'

if [ ! -f "$changelog" ]; then
  echo "changelog not found: $changelog" >&2
  exit 1
fi

if ! printf '%s\n' "$tag" | grep -Eq "$semver_re"; then
  echo "unsupported release tag: $tag" >&2
  echo "expected vMAJOR.MINOR.PATCH with an optional SemVer prerelease suffix" >&2
  exit 1
fi

awk -v tag="$tag" '
  function header_matches(line, body, suffix) {
    body = line
    sub(/^##[[:space:]]+/, "", body)
    sub(/[[:space:]]+$/, "", body)
    if (body == tag) {
      return 1
    }
    if (substr(body, 1, length(tag)) != tag) {
      return 0
    }
    suffix = substr(body, length(tag) + 1)
    return suffix ~ /^[[:space:]]+-[[:space:]]+.+/
  }

  /^##[[:space:]]+/ {
    if (in_section) {
      exit
    }
    if (header_matches($0)) {
      found = 1
      in_section = 1
      next
    }
  }

  in_section {
    if (!seen && $0 ~ /^[[:space:]]*$/) {
      next
    }
    seen = 1
    lines[++line_count] = $0
  }

  END {
    if (!found) {
      printf "release notes for %s were not found in %s\n", tag, FILENAME > "/dev/stderr"
      exit 1
    }

    while (line_count > 0 && lines[line_count] ~ /^[[:space:]]*$/) {
      line_count--
    }

    if (line_count == 0) {
      printf "release notes for %s are empty in %s\n", tag, FILENAME > "/dev/stderr"
      exit 1
    }

    for (i = 1; i <= line_count; i++) {
      print lines[i]
    }
  }
' "$changelog"
