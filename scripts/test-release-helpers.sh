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

"$repo_root/scripts/extract-release-notes.sh" "$changelog" v1.2.3 >"$tmpdir/notes.txt"
grep -q 'Release note one' "$tmpdir/notes.txt"

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

assert_tag_metadata v1.0.0 false true
assert_tag_metadata v1.0.0-rc.1 true false
assert_tag_metadata v0.1.3 true false

if "$repo_root/scripts/release-tag-metadata.sh" 'v1#foo' >"$tmpdir/tag-hash.out" 2>"$tmpdir/tag-hash.err"; then
  echo "unsafe hash tag unexpectedly succeeded" >&2
  exit 1
fi

if "$repo_root/scripts/release-tag-metadata.sh" 'v1/foo' >"$tmpdir/tag-slash.out" 2>"$tmpdir/tag-slash.err"; then
  echo "unsafe slash tag unexpectedly succeeded" >&2
  exit 1
fi

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

if command -v syft >/dev/null 2>&1; then
  real_sbom="$tmpdir/cpace-syft.cdx.json"
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

echo "release helper smoke tests passed"
