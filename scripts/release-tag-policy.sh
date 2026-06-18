#!/bin/sh

# Sourced by release helpers; keep this file side-effect-free for caller shells.
release_tag_policy_loaded=1
release_tag_semver_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$'

release_tag_is_supported() {
  case "$1" in
    *'
'*)
      return 1
      ;;
  esac
  printf '%s\n' "$1" | grep -Eq "$release_tag_semver_re"
}

release_tag_require_supported() {
  if release_tag_is_supported "$1"; then
    return 0
  fi
  echo "unsupported release tag: $1" >&2
  echo "expected vMAJOR.MINOR.PATCH with an optional SemVer prerelease suffix" >&2
  return 1
}
