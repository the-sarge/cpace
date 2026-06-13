#!/bin/sh
set -eu

usage() {
  echo "usage: $0 vMAJOR.MINOR.PATCH[-PRERELEASE]" >&2
}

if [ "$#" -ne 1 ]; then
  usage
  exit 2
fi

tag=$1
semver_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$'

if ! printf '%s\n' "$tag" | grep -Eq "$semver_re"; then
  echo "unsupported release tag: $tag" >&2
  echo "expected vMAJOR.MINOR.PATCH with an optional SemVer prerelease suffix" >&2
  exit 1
fi

version=${tag#v}
major=${version%%.*}
prerelease=false
latest=true

case "$version" in
  *-*)
    prerelease=true
    latest=false
    ;;
esac

if [ "$major" = "0" ]; then
  prerelease=true
  latest=false
fi

printf 'release-tag=%s\n' "$tag"
printf 'sbom-file=cpace-%s.cdx.json\n' "$tag"
printf 'prerelease=%s\n' "$prerelease"
printf 'latest=%s\n' "$latest"
