#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
summary_docs_manifest="$repo_root/docs/evidence-baseline-summary-docs.txt"

summary_doc_refs() {
  [ -f "$summary_docs_manifest" ] || {
    echo "missing summary-doc manifest: docs/evidence-baseline-summary-docs.txt" >&2
    return 2
  }
  while IFS= read -r ref || [ -n "$ref" ]; do
    case "$ref" in
      ""|\#*) continue ;;
    esac
    printf '%s\n' "$ref"
  done <"$summary_docs_manifest" | sort -u
}

case "${1:-}" in
  "")
    ;;
  --list-summary-docs)
    summary_doc_refs
    exit 0
    ;;
  *)
    echo "usage: $0 [--list-summary-docs]" >&2
    exit 2
    ;;
esac

summary_docs="
$(summary_doc_refs)
"

is_summary_doc_ref() {
  case "$summary_docs" in
    *"
$1
"*) return 0 ;;
    *) return 1 ;;
  esac
}

changed=false
docs_only=true
evidence_changed=false
evidence_checker_changed=false
release_policy_changed=false

while IFS= read -r path; do
  [ -n "$path" ] || continue
  changed=true

  case "$path" in
    .github/workflows/release.yml|.github/workflows/sast-gate.yml|\
    .github/allowed_signers|.github/syft-release.yaml|Taskfile.yml|\
    scripts/check-release-policy.sh|tools/releasepolicy/*|\
    scripts/classify-check-changes.sh|scripts/test-ci-classifier.sh|\
    scripts/go-tool.sh|scripts/go-tool-versions.sh|scripts/release-tag-policy.sh|\
    scripts/release-metadata.sh|scripts/release-tag-metadata.sh|\
    scripts/validate-cyclonedx-sbom.sh|scripts/extract-release-notes.sh)
      release_policy_changed=true
      ;;
  esac

  case "$path" in
    docs/evidence-baseline.md|docs/evidence-baseline-summary-docs.txt|docs/evidence/*)
      evidence_changed=true
      ;;
    scripts/check-evidence-baseline.sh|tools/evidencebaseline/*)
      evidence_changed=true
      evidence_checker_changed=true
      ;;
    scripts/classify-check-changes.sh|scripts/test-ci-classifier.sh)
      evidence_changed=true
      ;;
  esac

  if is_summary_doc_ref "$path"; then
    evidence_changed=true
  fi

  case "$path" in
    *.md|docs/*) ;;
    *) docs_only=false ;;
  esac
done

if [ "$changed" = "false" ]; then
  docs_only=false
fi

printf 'docs_only=%s\n' "$docs_only"
printf 'evidence_changed=%s\n' "$evidence_changed"
printf 'evidence_checker_changed=%s\n' "$evidence_checker_changed"
printf 'release_policy_changed=%s\n' "$release_policy_changed"
