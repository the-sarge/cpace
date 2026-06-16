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
sbom_name=${sbom##*/}
semver_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?$'

if [ ! -s "$sbom" ]; then
  echo "SBOM not found or empty: $sbom" >&2
  exit 1
fi

case "$sbom_name" in
  cpace-v*.cdx.json)
    sbom_tag=${sbom_name#cpace-}
    sbom_tag=${sbom_tag%.cdx.json}
    ;;
  *)
    echo "SBOM filename must be cpace-vMAJOR.MINOR.PATCH[-PRERELEASE].cdx.json: $sbom_name" >&2
    exit 1
    ;;
esac

if ! printf '%s\n' "$sbom_tag" | grep -Eq "$semver_re"; then
  echo "SBOM filename must use a supported release tag: $sbom_name" >&2
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
    def exact_module:
      . == $module or
      . == ("pkg:golang/" + $module) or
      startswith("pkg:golang/" + $module + "@");

    def candidate_strings:
      [
        .metadata.component?,
        .components[]?
      ]
      | map(select(type == "object"))
      | map(.name?, .purl?, .["bom-ref"]?)
      | map(select(type == "string"));

    any(candidate_strings[]; exact_module)
  ' "$sbom" >/dev/null || {
    echo "SBOM is missing expected Go module entry: $module" >&2
    echo "If the release-relevant module graph changed intentionally, update scripts/validate-cyclonedx-sbom.sh and scripts/test-release-helpers.sh together." >&2
    exit 1
  }
done
