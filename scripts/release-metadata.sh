#!/bin/sh

# Sourced by release helpers; keep this file side-effect-free for caller shells.

release_metadata_write() (
  if [ "$#" -ne 1 ]; then
    echo "usage: release_metadata_write vMAJOR.MINOR.PATCH[-PRERELEASE]" >&2
    exit 2
  fi
  if [ "${release_tag_policy_loaded:-}" != 1 ]; then
    echo "release metadata requires scripts/release-tag-policy.sh" >&2
    exit 2
  fi

  release_metadata_tag=$1
  if ! release_tag_require_supported "$release_metadata_tag"; then
    exit 1
  fi

  release_metadata_version=${release_metadata_tag#v}
  release_metadata_major=${release_metadata_version%%.*}
  release_metadata_prerelease=false
  release_metadata_latest=true

  case "$release_metadata_version" in
    *-*)
      release_metadata_prerelease=true
      release_metadata_latest=false
      ;;
  esac

  if [ "$release_metadata_major" = "0" ]; then
    release_metadata_prerelease=true
    release_metadata_latest=false
  fi

  printf 'release-tag=%s\n' "$release_metadata_tag"
  printf 'sbom-file=cpace-%s.cdx.json\n' "$release_metadata_tag"
  printf 'prerelease=%s\n' "$release_metadata_prerelease"
  printf 'latest=%s\n' "$release_metadata_latest"
)
