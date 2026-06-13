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
