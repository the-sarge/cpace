#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
baseline="$repo_root/docs/evidence-baseline.md"

summary_doc_refs() {
  [ -f "$baseline" ] || return 0
  refs=$(
    awk -F '|' '
    /^## / {
      if (in_section) {
        exit
      }
      in_section = ($0 == "## Baseline Index")
      next
    }
    in_section && /^[[:space:]]*\|/ {
      table_row++
      if (table_row <= 2) {
        next
      }
      cell = $5
      while (match(cell, /`[^`]+`/)) {
        ref = substr(cell, RSTART + 1, RLENGTH - 2)
        while (sub(/^\.\//, "", ref)) {}
        if (ref ~ /^docs\/.*\.md$/) {
          print ref
        }
        cell = substr(cell, RSTART + RLENGTH)
      }
    }
  ' "$baseline"
  ) || return $?
  [ -z "$refs" ] || printf '%s\n' "$refs" | sort -u
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

while IFS= read -r path; do
  [ -n "$path" ] || continue
  changed=true

  case "$path" in
    docs/evidence-baseline.md|docs/evidence/*)
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
