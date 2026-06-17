#!/bin/sh
set -eu

usage() {
  echo "usage: $0 vMAJOR.MINOR.PATCH[-PRERELEASE]" >&2
}

if [ "$#" -ne 1 ]; then
  usage
  exit 2
fi

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$script_dir/release-tag-policy.sh"

tag=$1

if ! release_tag_require_supported "$tag"; then
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
