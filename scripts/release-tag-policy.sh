#!/bin/sh

release_tag_semver_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$'

release_tag_is_supported() {
  release_tag=$1
  case "$release_tag" in
    *'
'*)
      return 1
      ;;
  esac
  printf '%s\n' "$release_tag" | grep -Eq "$release_tag_semver_re"
}

release_tag_require_supported() {
  release_tag=$1
  if release_tag_is_supported "$release_tag"; then
    return 0
  fi
  echo "unsupported release tag: $release_tag" >&2
  echo "expected vMAJOR.MINOR.PATCH with an optional SemVer prerelease suffix" >&2
  return 1
}
