#!/bin/sh
set -eu

usage() {
  echo "usage: $0 cpace-vX.Y.Z.cdx.json" >&2
}

if [ "$#" -ne 1 ]; then
  usage
  exit 2
fi

sbom=$1

if [ ! -s "$sbom" ]; then
  echo "SBOM not found or empty: $sbom" >&2
  exit 1
fi

command -v jq >/dev/null 2>&1 || {
  echo "jq not found; install jq to validate CycloneDX SBOMs" >&2
  exit 1
}

jq -e '.bomFormat == "CycloneDX" and .specVersion == "1.5"' "$sbom" >/dev/null || {
  echo "SBOM must be CycloneDX JSON 1.5" >&2
  exit 1
}

# Keep this expected set aligned with go.mod for release-relevant module graph entries that must appear in Syft's CycloneDX output.
for module in \
  github.com/the-sarge/cpace \
  github.com/gtank/ristretto255 \
  filippo.io/edwards25519
do
  jq -e --arg module "$module" '
    def candidate_strings:
      [
        .metadata.component.name?,
        .metadata.component.purl?,
        .metadata.component["bom-ref"]?,
        (.components[]? | .name?, .purl?, .["bom-ref"]?)
      ] | map(select(type == "string"));

    any(candidate_strings[]; . == $module or contains($module))
  ' "$sbom" >/dev/null || {
    echo "SBOM is missing expected Go module entry: $module" >&2
    echo "If the release-relevant module graph changed intentionally, update scripts/validate-cyclonedx-sbom.sh and scripts/test-release-helpers.sh together." >&2
    exit 1
  }
done
