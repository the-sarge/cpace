#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

classify() {
  printf '%s\n' "$1" | "$repo_root/scripts/classify-check-changes.sh"
}

assert_classification() {
  name=$1
  input=$2
  want_docs_only=$3
  want_evidence_changed=$4
  want_evidence_checker_changed=$5
  out="$tmpdir/$name.out"

  classify "$input" >"$out"
  grep -Fxq "docs_only=$want_docs_only" "$out" || {
    echo "$name: expected docs_only=$want_docs_only" >&2
    cat "$out" >&2
    exit 1
  }
  grep -Fxq "evidence_changed=$want_evidence_changed" "$out" || {
    echo "$name: expected evidence_changed=$want_evidence_changed" >&2
    cat "$out" >&2
    exit 1
  }
  grep -Fxq "evidence_checker_changed=$want_evidence_checker_changed" "$out" || {
    echo "$name: expected evidence_checker_changed=$want_evidence_checker_changed" >&2
    cat "$out" >&2
    exit 1
  }
}

assert_classification ordinary-doc "docs/governance.md" true false false
assert_classification evidence-baseline "docs/evidence-baseline.md" true true false
assert_classification evidence-summary-doc-manifest "docs/evidence-baseline-summary-docs.txt" true true false
assert_classification evidence-bundle "docs/evidence/go1264-20260611/local-analysis.log" true true false
assert_classification summary-doc "docs/dependency-review.md" true true false
assert_classification evidence-checker "tools/evidencebaseline/main.go" false true true
assert_classification evidence-checker-wrapper "scripts/check-evidence-baseline.sh" false true true
assert_classification classifier-script "scripts/classify-check-changes.sh" false true false
assert_classification classifier-test "scripts/test-ci-classifier.sh" false true false

rename_repo="$tmpdir/rename-repo"
mkdir "$rename_repo"
git -C "$rename_repo" -c init.defaultBranch=main init -q
git -C "$rename_repo" config user.email "ci-classifier@example.invalid"
git -C "$rename_repo" config user.name "CI Classifier Test"
mkdir -p "$rename_repo/docs"
printf '# Dependency Review\n' >"$rename_repo/docs/dependency-review.md"
git -C "$rename_repo" add docs/dependency-review.md
git -C "$rename_repo" commit -q -m "seed docs"
git -C "$rename_repo" mv docs/dependency-review.md docs/dependency-review-renamed.md
git -C "$rename_repo" diff --name-only --no-renames HEAD >"$tmpdir/rename.paths"
"$repo_root/scripts/classify-check-changes.sh" <"$tmpdir/rename.paths" >"$tmpdir/rename.out"
grep -Fxq "docs_only=true" "$tmpdir/rename.out"
grep -Fxq "evidence_changed=true" "$tmpdir/rename.out"

fakebin="$tmpdir/fakebin"
mkdir "$fakebin"
printf '#!/bin/sh\nexit 42\n' >"$fakebin/awk"
chmod +x "$fakebin/awk"
PATH="$fakebin:$PATH" "$repo_root/scripts/classify-check-changes.sh" --list-summary-docs >"$tmpdir/no-awk.out"
grep -Fxq "docs/dependency-review.md" "$tmpdir/no-awk.out"
grep -Fxq "docs/fuzz-evidence.md" "$tmpdir/no-awk.out"

missing_repo="$tmpdir/missing-manifest-repo"
mkdir -p "$missing_repo/scripts" "$missing_repo/docs"
cp "$repo_root/scripts/classify-check-changes.sh" "$missing_repo/scripts/classify-check-changes.sh"
chmod +x "$missing_repo/scripts/classify-check-changes.sh"

set +e
"$missing_repo/scripts/classify-check-changes.sh" --list-summary-docs >"$tmpdir/missing-list.out" 2>"$tmpdir/missing-list.err"
status=$?
set -e
if [ "$status" -ne 2 ]; then
  echo "missing-list: expected status 2, got $status" >&2
  cat "$tmpdir/missing-list.out" >&2
  cat "$tmpdir/missing-list.err" >&2
  exit 1
fi
grep -Fq "missing summary-doc manifest" "$tmpdir/missing-list.err" || {
  echo "missing-list: expected missing-manifest stderr" >&2
  cat "$tmpdir/missing-list.err" >&2
  exit 1
}

set +e
printf 'docs/dependency-review.md\n' | "$missing_repo/scripts/classify-check-changes.sh" >"$tmpdir/missing-stdin.out" 2>"$tmpdir/missing-stdin.err"
status=$?
set -e
if [ "$status" -ne 2 ]; then
  echo "missing-stdin: expected status 2, got $status" >&2
  cat "$tmpdir/missing-stdin.out" >&2
  cat "$tmpdir/missing-stdin.err" >&2
  exit 1
fi
grep -Fq "missing summary-doc manifest" "$tmpdir/missing-stdin.err" || {
  echo "missing-stdin: expected missing-manifest stderr" >&2
  cat "$tmpdir/missing-stdin.err" >&2
  exit 1
}

echo "CI change classifier tests passed"
