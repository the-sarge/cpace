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
. "$script_dir/release-metadata.sh"

release_metadata_write "$1"
