#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

changelog="$tmpdir/CHANGELOG.md"
cat >"$changelog" <<'EOF'
# Changelog

## Unreleased

- Work in progress.

## v1.2.3 - 2026-06-13

- Release note one.
- Release note two.

## v1.2.2 - 2026-06-12

- Prior release.

## v0.0.1 - 2026-06-11

## v0.0.0 - 2026-06-10

- Older release.
EOF

assert_release_tag_supported() {
  tag=$1
  sh -c '. "$1"; release_tag_is_supported "$2"' sh "$repo_root/scripts/release-tag-policy.sh" "$tag"
}

assert_release_tag_rejected() {
  tag=$1
  if sh -c '. "$1"; release_tag_is_supported "$2"' sh "$repo_root/scripts/release-tag-policy.sh" "$tag"; then
    echo "release tag policy unexpectedly accepted: $tag" >&2
    exit 1
  fi
}

assert_release_tag_supported v1.2.3
assert_release_tag_supported v1.2.3-rc.1
assert_release_tag_supported v0.1.3
assert_release_tag_rejected 1.2.3
assert_release_tag_rejected v01.0.0
assert_release_tag_rejected v1.0.0-rc..1
assert_release_tag_rejected v1/foo
assert_release_tag_rejected "$(printf 'v1.0.0\nlatest=true')"

assert_helper_path_reuses_release_tag_policy() {
  helper=$1
  helper_path=$2

  if ! grep -Fq '. "$script_dir/release-tag-policy.sh"' "$helper_path"; then
    echo "$helper does not source scripts/release-tag-policy.sh" >&2
    exit 1
  fi
  assert_helper_path_defines_no_local_release_tag_policy_functions "$helper" "$helper_path"
  if grep -Fq 'release_tag_semver_re=' "$helper_path" || grep -Fq '^v(0|' "$helper_path"; then
    echo "$helper redefines the release tag SemVer policy" >&2
    exit 1
  fi
}

assert_helper_path_defines_no_local_release_tag_policy_functions() {
  helper=$1
  helper_path=$2

  if grep -Eq '(^|[[:space:];{&|])release_tag_[A-Za-z0-9_]+[[:space:]]*\([[:space:]]*\)' "$helper_path" ||
    grep -Eq '(^|[[:space:];{&|])function[[:space:]]+release_tag_[A-Za-z0-9_]+' "$helper_path"; then
    echo "$helper defines a local release tag policy function" >&2
    exit 1
  fi
}

assert_helper_reuses_release_tag_policy() {
  helper=$1
  assert_helper_path_reuses_release_tag_policy "$helper" "$repo_root/$helper"
}

assert_helper_rejects_release_tag_policy_function_shadow() {
  helper=$1
  function_name=$2
  function_definition=$3
  assert_helper_path_rejects_release_tag_policy_function_shadow "$helper" "$repo_root/$helper" "$function_name" "$function_definition"
}

assert_helper_path_rejects_release_tag_policy_function_shadow() {
  helper=$1
  helper_path=$2
  function_name=$3
  function_definition=$4
  shadow_helper="$tmpdir/$(basename -- "$helper")-$function_name-shadow.sh"
  injected=false

  while IFS= read -r line || [ -n "$line" ]; do
    printf '%s\n' "$line"
    case "$line" in
      *'. "$script_dir/release-tag-policy.sh"'*)
        printf '%s\n' "$function_definition"
        injected=true
        ;;
    esac
  done <"$helper_path" >"$shadow_helper"

  if [ "$injected" != true ]; then
    echo "injection anchor not found in $helper" >&2
    exit 1
  fi

  if ( assert_helper_path_reuses_release_tag_policy "$helper with local $function_name" "$shadow_helper" ) >"$shadow_helper.out" 2>"$shadow_helper.err"; then
    echo "$helper unexpectedly allowed local $function_name definition: $function_definition" >&2
    exit 1
  fi
  if ! grep -q 'defines a local release tag policy function' "$shadow_helper.err"; then
    echo "$helper rejected local $function_name definition with an unexpected diagnostic: $function_definition" >&2
    cat "$shadow_helper.err" >&2
    exit 1
  fi
}

assert_helper_rejects_release_tag_policy_function_shadow_forms() {
  helper=$1
  function_name=$2

  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "$function_name() { return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "$function_name( ) { return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "$function_name() # local override
{ return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "$function_name( ) # local override
{ return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "function $function_name() { return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "function $function_name { return 0; }"
  assert_helper_rejects_release_tag_policy_function_shadow "$helper" "$function_name" "true && $function_name() { return 0; }"
}

assert_helper_rejects_release_tag_policy_function_shadow_after_reformatted_source() {
  helper=$1
  function_name=$2
  helper_path="$tmpdir/$(basename -- "$helper")-$function_name-reformatted-source.sh"

  awk '
    $0 == ". \"$script_dir/release-tag-policy.sh\"" {
      print "  " $0 " # required policy module"
      next
    }
    {
      print
    }
  ' "$repo_root/$helper" >"$helper_path"

  assert_helper_path_rejects_release_tag_policy_function_shadow "$helper with reformatted policy source" "$helper_path" "$function_name" "$function_name() { return 0; }"
}

assert_helper_reuses_release_metadata_module() {
  helper=$1
  helper_path="$repo_root/$helper"

  if ! grep -Fq '. "$script_dir/release-metadata.sh"' "$helper_path"; then
    echo "$helper does not source scripts/release-metadata.sh" >&2
    exit 1
  fi
  if grep -Fq 'prerelease=false' "$helper_path" || grep -Fq 'latest=true' "$helper_path"; then
    echo "$helper redefines release metadata derivation" >&2
    exit 1
  fi
}

assert_release_tag_policy_preserves_caller_names() {
  function_name=$1
  sh -c '
    . "$1"
    release_tag=before
    tag=before
    version=before
    major=before
    prerelease=before
    latest=before
    "$2" v1.2.3 >/dev/null
    test "$release_tag:$tag:$version:$major:$prerelease:$latest" = before:before:before:before:before:before
  ' sh "$repo_root/scripts/release-tag-policy.sh" "$function_name"
}

assert_release_metadata_module_preserves_caller_names() {
  sh -c '
    . "$1"
    . "$2"
    release_tag=before
    tag=before
    version=before
    major=before
    prerelease=before
    latest=before
    sbom_file=before
    release_metadata_tag=before
    release_metadata_version=before
    release_metadata_major=before
    release_metadata_prerelease=before
    release_metadata_latest=before
    release_metadata_write v1.2.3 >/dev/null
    test "$release_tag:$tag:$version:$major:$prerelease:$latest:$sbom_file:$release_metadata_tag:$release_metadata_version:$release_metadata_major:$release_metadata_prerelease:$release_metadata_latest" = before:before:before:before:before:before:before:before:before:before:before:before
  ' sh "$repo_root/scripts/release-tag-policy.sh" "$repo_root/scripts/release-metadata.sh"
}

assert_helper_reuses_release_tag_policy scripts/extract-release-notes.sh
assert_helper_reuses_release_tag_policy scripts/release-tag-metadata.sh
assert_helper_reuses_release_tag_policy scripts/validate-cyclonedx-sbom.sh
assert_helper_path_defines_no_local_release_tag_policy_functions scripts/release-metadata.sh "$repo_root/scripts/release-metadata.sh"
assert_helper_rejects_release_tag_policy_function_shadow_forms scripts/release-tag-metadata.sh release_tag_is_supported
assert_helper_rejects_release_tag_policy_function_shadow_forms scripts/release-tag-metadata.sh release_tag_require_supported
assert_helper_rejects_release_tag_policy_function_shadow_after_reformatted_source scripts/release-tag-metadata.sh release_tag_is_supported
assert_helper_rejects_release_tag_policy_function_shadow scripts/release-tag-metadata.sh release_tag_policy_is_supported 'release_tag_policy_is_supported() { return 0; }'
assert_helper_rejects_release_tag_policy_function_shadow scripts/release-tag-metadata.sh release_tag_policy_require_supported_for_metadata 'release_tag_policy_require_supported_for_metadata() { return 0; }'
assert_helper_reuses_release_metadata_module scripts/release-tag-metadata.sh

assert_release_tag_policy_preserves_caller_names release_tag_is_supported
assert_release_tag_policy_preserves_caller_names release_tag_require_supported
assert_release_metadata_module_preserves_caller_names

"$repo_root/scripts/extract-release-notes.sh" "$changelog" v1.2.3 >"$tmpdir/notes.txt"
grep -q 'Release note one' "$tmpdir/notes.txt"

if "$repo_root/scripts/extract-release-notes.sh" "$changelog" 1.2.3 >"$tmpdir/untagged-notes.txt" 2>"$tmpdir/untagged-notes.err"; then
  echo "untagged release notes unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/extract-release-notes.sh" "$changelog" v1.2 >"$tmpdir/short-tag-notes.txt" 2>"$tmpdir/short-tag-notes.err"; then
  echo "short release tag notes unexpectedly succeeded" >&2
  exit 1
fi

multiline_tag=$(printf 'v1.2.3\nv9.9.9')
if "$repo_root/scripts/extract-release-notes.sh" "$changelog" "$multiline_tag" >"$tmpdir/multiline-tag-notes.txt" 2>"$tmpdir/multiline-tag-notes.err"; then
  echo "multiline release tag notes unexpectedly succeeded" >&2
  exit 1
fi
grep -q 'unsupported release tag' "$tmpdir/multiline-tag-notes.err"

invalid_tag_changelog="$tmpdir/INVALID_TAG_CHANGELOG.md"
cat >"$invalid_tag_changelog" <<'EOF'
# Changelog

## 1.2.3 - 2026-06-13

- Invalid unprefixed tag notes.
EOF

if "$repo_root/scripts/extract-release-notes.sh" "$invalid_tag_changelog" 1.2.3 >"$tmpdir/invalid-tag-notes.txt" 2>"$tmpdir/invalid-tag-notes.err"; then
  echo "invalid release tag notes unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/extract-release-notes.sh" "$changelog" v9.9.9 >"$tmpdir/missing.txt" 2>"$tmpdir/missing.err"; then
  echo "missing release notes unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/extract-release-notes.sh" "$changelog" v0.0.1 >"$tmpdir/empty.txt" 2>"$tmpdir/empty.err"; then
  echo "empty release notes unexpectedly succeeded" >&2
  exit 1
fi

prerelease_changelog="$tmpdir/PRERELEASE_CHANGELOG.md"
cat >"$prerelease_changelog" <<'EOF'
# Changelog

## Unreleased

- Work in progress.

## v1.2.3-rc.1 - 2026-06-13

- Release candidate note.

## v1.2.2 - 2026-06-12

- Prior release.
EOF

if "$repo_root/scripts/extract-release-notes.sh" "$prerelease_changelog" v1.2.3 >"$tmpdir/stable-from-rc.txt" 2>"$tmpdir/stable-from-rc.err"; then
  echo "stable tag unexpectedly matched prerelease notes" >&2
  exit 1
fi

"$repo_root/scripts/extract-release-notes.sh" "$prerelease_changelog" v1.2.3-rc.1 >"$tmpdir/prerelease-notes.txt"
grep -q 'Release candidate note' "$tmpdir/prerelease-notes.txt"

assert_tag_metadata() {
  tag=$1
  expected_prerelease=$2
  expected_latest=$3
  metadata="$tmpdir/tag-$tag.env"

  "$repo_root/scripts/release-tag-metadata.sh" "$tag" >"$metadata"
  grep -Fxq "release-tag=$tag" "$metadata"
  grep -Fxq "sbom-file=cpace-$tag.cdx.json" "$metadata"
  grep -Fxq "prerelease=$expected_prerelease" "$metadata"
  grep -Fxq "latest=$expected_latest" "$metadata"
}

assert_release_metadata_module() {
  tag=$1
  expected_prerelease=$2
  expected_latest=$3
  metadata="$tmpdir/module-tag-$tag.env"

  sh -c '. "$1"; . "$2"; release_metadata_write "$3"' sh "$repo_root/scripts/release-tag-policy.sh" "$repo_root/scripts/release-metadata.sh" "$tag" >"$metadata"
  grep -Fxq "release-tag=$tag" "$metadata"
  grep -Fxq "sbom-file=cpace-$tag.cdx.json" "$metadata"
  grep -Fxq "prerelease=$expected_prerelease" "$metadata"
  grep -Fxq "latest=$expected_latest" "$metadata"
}

assert_release_metadata_module_rejects_unsupported_tag() {
  if sh -c '. "$1"; . "$2"; release_metadata_write "$3"' sh "$repo_root/scripts/release-tag-policy.sh" "$repo_root/scripts/release-metadata.sh" v01.0.0 >"$tmpdir/module-invalid-tag.out" 2>"$tmpdir/module-invalid-tag.err"; then
    echo "release metadata module unexpectedly accepted unsupported tag" >&2
    exit 1
  fi
  grep -q 'unsupported release tag' "$tmpdir/module-invalid-tag.err"
}

assert_release_metadata_module_requires_sourced_policy() {
  path_stub_dir="$tmpdir/release-metadata-path-stub"
  mkdir "$path_stub_dir"
  printf '#!/bin/sh\nexit 0\n' >"$path_stub_dir/release_tag_policy_require_supported_for_metadata"
  chmod +x "$path_stub_dir/release_tag_policy_require_supported_for_metadata"

  set +e
  PATH="$path_stub_dir:$PATH" sh -c '. "$1"; release_metadata_write v01.0.0' sh "$repo_root/scripts/release-metadata.sh" >"$tmpdir/module-missing-policy.out" 2>"$tmpdir/module-missing-policy.err"
  status=$?
  set -e
  if [ "$status" -ne 2 ]; then
    echo "release metadata module missing-policy status got $status want 2" >&2
    exit 1
  fi
  if [ -s "$tmpdir/module-missing-policy.out" ]; then
    echo "release metadata module emitted metadata without sourced policy" >&2
    exit 1
  fi
  grep -q 'release metadata requires scripts/release-tag-policy.sh' "$tmpdir/module-missing-policy.err"
}

assert_release_metadata_module_rejects_spoofed_policy_marker() {
  path_stub_dir="$tmpdir/release-metadata-spoofed-marker-path-stub"
  mkdir "$path_stub_dir"
  printf '#!/bin/sh\nexit 0\n' >"$path_stub_dir/release_tag_policy_require_supported_for_metadata"
  chmod +x "$path_stub_dir/release_tag_policy_require_supported_for_metadata"

  set +e
  release_tag_policy_metadata_check_ran=1 PATH="$path_stub_dir:$PATH" sh -c '. "$1"; release_metadata_write v01.0.0' sh "$repo_root/scripts/release-metadata.sh" >"$tmpdir/module-spoofed-marker.out" 2>"$tmpdir/module-spoofed-marker.err"
  status=$?
  set -e
  if [ "$status" -ne 2 ]; then
    echo "release metadata module spoofed-marker status got $status want 2" >&2
    exit 1
  fi
  if [ -s "$tmpdir/module-spoofed-marker.out" ]; then
    echo "release metadata module emitted metadata with spoofed policy marker" >&2
    exit 1
  fi
  grep -q 'release metadata requires scripts/release-tag-policy.sh' "$tmpdir/module-spoofed-marker.err"
}

assert_release_metadata_module_rejects_unset_policy_function() {
  path_stub_dir="$tmpdir/release-metadata-unset-function-path-stub"
  mkdir "$path_stub_dir"
  printf '#!/bin/sh\nexit 0\n' >"$path_stub_dir/release_tag_policy_require_supported_for_metadata"
  chmod +x "$path_stub_dir/release_tag_policy_require_supported_for_metadata"

  set +e
  PATH="$path_stub_dir:$PATH" sh -c '. "$1"; . "$2"; unset -f release_tag_policy_require_supported_for_metadata; release_metadata_write v01.0.0' sh "$repo_root/scripts/release-tag-policy.sh" "$repo_root/scripts/release-metadata.sh" >"$tmpdir/module-unset-function.out" 2>"$tmpdir/module-unset-function.err"
  status=$?
  set -e
  if [ "$status" -ne 2 ]; then
    echo "release metadata module unset-function status got $status want 2" >&2
    exit 1
  fi
  if [ -s "$tmpdir/module-unset-function.out" ]; then
    echo "release metadata module emitted metadata after policy function was unset" >&2
    exit 1
  fi
  grep -q 'release metadata requires scripts/release-tag-policy.sh' "$tmpdir/module-unset-function.err"
}

assert_tag_metadata v1.0.0 false true
assert_tag_metadata v1.0.0-rc.1 true false
assert_tag_metadata v0.1.3 true false
assert_release_metadata_module v1.0.0 false true
assert_release_metadata_module v1.0.0-rc.1 true false
assert_release_metadata_module v0.1.3 true false
assert_release_metadata_module_rejects_unsupported_tag
assert_release_metadata_module_requires_sourced_policy
assert_release_metadata_module_rejects_spoofed_policy_marker
assert_release_metadata_module_rejects_unset_policy_function

if "$repo_root/scripts/release-tag-metadata.sh" 'v01.0.0' >"$tmpdir/tag-leading-zero.out" 2>"$tmpdir/tag-leading-zero.err"; then
  echo "leading-zero tag unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/release-tag-metadata.sh" 'v1.0.0-rc..1' >"$tmpdir/tag-empty-prerelease.out" 2>"$tmpdir/tag-empty-prerelease.err"; then
  echo "empty prerelease tag component unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/release-tag-metadata.sh" 'v1#foo' >"$tmpdir/tag-hash.out" 2>"$tmpdir/tag-hash.err"; then
  echo "unsafe hash tag unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/release-tag-metadata.sh" 'v1/foo' >"$tmpdir/tag-slash.out" 2>"$tmpdir/tag-slash.err"; then
  echo "unsafe slash tag unexpectedly succeeded" >&2
  exit 1
fi

multiline_metadata_tag=$(printf 'v1.0.0\nlatest=true')
if "$repo_root/scripts/release-tag-metadata.sh" "$multiline_metadata_tag" >"$tmpdir/tag-multiline.out" 2>"$tmpdir/tag-multiline.err"; then
  echo "multiline metadata tag unexpectedly succeeded" >&2
  exit 1
fi
grep -q 'unsupported release tag' "$tmpdir/tag-multiline.err"

sbom="$tmpdir/cpace-v1.2.3.cdx.json"
cat >"$sbom" <<'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "metadata": {
    "component": {
      "name": "github.com/the-sarge/cpace"
    }
  },
  "components": [
    {
      "type": "library",
      "name": "github.com/gtank/ristretto255",
      "purl": "pkg:golang/github.com/gtank/ristretto255@v0.2.0"
    },
    {
      "type": "library",
      "name": "filippo.io/edwards25519",
      "purl": "pkg:golang/filippo.io/edwards25519@v1.2.0"
    }
  ]
}
EOF

"$repo_root/scripts/validate-cyclonedx-sbom.sh" "$sbom"

assert_sbom_filename_rejects_unsupported_release_tag() {
  sbom_name=$1
  rejected_sbom="$tmpdir/$sbom_name"

  cp "$sbom" "$rejected_sbom"
  if "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$rejected_sbom" >"$tmpdir/$sbom_name.out" 2>"$tmpdir/$sbom_name.err"; then
    echo "unsupported-release-tag SBOM filename unexpectedly succeeded: $sbom_name" >&2
    exit 1
  fi
  grep -q 'SBOM filename must use a supported release tag' "$tmpdir/$sbom_name.err"
}

assert_sbom_filename_rejects_unsupported_release_tag cpace-v01.0.0.cdx.json
assert_sbom_filename_rejects_unsupported_release_tag cpace-v1.2.cdx.json

wrong_name_sbom="$tmpdir/other-v1.2.3.cdx.json"
cp "$sbom" "$wrong_name_sbom"
if "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$wrong_name_sbom" >"$tmpdir/wrong-name-sbom.out" 2>"$tmpdir/wrong-name-sbom.err"; then
  echo "wrongly named SBOM unexpectedly succeeded" >&2
  exit 1
fi

newline_name_sbom="$tmpdir/$(printf 'cpace-v1.2.3\nignored.cdx.json')"
cp "$sbom" "$newline_name_sbom"
if "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$newline_name_sbom" >"$tmpdir/newline-name-sbom.out" 2>"$tmpdir/newline-name-sbom.err"; then
  echo "newline-bearing SBOM filename unexpectedly succeeded" >&2
  exit 1
fi
grep -q 'SBOM filename must use a supported release tag' "$tmpdir/newline-name-sbom.err"

substring_sbom="$tmpdir/cpace-v1.2.4.cdx.json"
cat >"$substring_sbom" <<'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "metadata": {
    "component": {
      "name": "github.com/the-sarge/cpace"
    }
  },
  "components": [
    {
      "type": "library",
      "name": "github.com/example/not-github.com/gtank/ristretto255"
    },
    {
      "type": "library",
      "name": "github.com/example/not-filippo.io/edwards25519"
    }
  ]
}
EOF

if "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$substring_sbom" >"$tmpdir/substring-sbom.out" 2>"$tmpdir/substring-sbom.err"; then
  echo "substring-only SBOM module matches unexpectedly succeeded" >&2
  exit 1
fi

if command -v syft >/dev/null 2>&1; then
  real_sbom="$tmpdir/cpace-v9.9.9.cdx.json"
  (cd "$repo_root" && syft dir:. -o "cyclonedx-json@1.5=$real_sbom" >/dev/null)
  "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$real_sbom"
else
  echo "syft not found; skipping optional real Syft SBOM validation"
fi

bad_sbom="$tmpdir/bad.cdx.json"
cat >"$bad_sbom" <<'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "components": []
}
EOF

if "$repo_root/scripts/validate-cyclonedx-sbom.sh" "$bad_sbom" >"$tmpdir/bad.out" 2>"$tmpdir/bad.err"; then
  echo "invalid SBOM unexpectedly succeeded" >&2
  exit 1
fi

"$repo_root/scripts/check-release-policy.sh"

echo "release helper smoke tests passed"
