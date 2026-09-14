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
  want_release_policy_changed=${6:-false}
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
  grep -Fxq "release_policy_changed=$want_release_policy_changed" "$out" || {
    echo "$name: expected release_policy_changed=$want_release_policy_changed" >&2
    cat "$out" >&2
    exit 1
  }
}

assert_classification release-workflow ".github/workflows/release.yml" false false false true
assert_classification ordinary-doc "docs/governance.md" true false false
assert_classification evidence-baseline "docs/evidence-baseline.md" true true false
assert_classification evidence-summary-doc-manifest "docs/evidence-baseline-summary-docs.txt" true true false
assert_classification evidence-bundle "docs/evidence/go1264-20260611/local-analysis.log" true true false
assert_classification summary-doc "docs/dependency-review.md" true true false
assert_classification evidence-checker "tools/evidencebaseline/main.go" false true true
assert_classification evidence-checker-wrapper "scripts/check-evidence-baseline.sh" false true true
assert_classification classifier-script "scripts/classify-check-changes.sh" false true false true
assert_classification classifier-test "scripts/test-ci-classifier.sh" false true false true

for path in \
  .github/workflows/sast-gate.yml .github/allowed_signers .github/syft-release.yaml \
  Taskfile.yml scripts/check-release-policy.sh scripts/go-tool.sh \
  scripts/go-tool-versions.sh scripts/release-tag-policy.sh scripts/release-metadata.sh \
  scripts/release-tag-metadata.sh scripts/validate-cyclonedx-sbom.sh scripts/extract-release-notes.sh \
  tools/releasepolicy/main.go tools/releasepolicy/go.mod tools/releasepolicy/go.sum \
  tools/releasepolicy/testdata/example.yml
do
  assert_classification release-input "$path" false false false true
done
assert_classification unrelated-code "cpace.go" false false false
assert_classification unrelated-workflow ".github/workflows/nightly-fuzz.yml" false false false
assert_classification empty "" false false false
assert_classification mixed "docs/governance.md
.github/workflows/release.yml
.github/workflows/release.yml" false false false true

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

# Run the real Taskfile while replacing external Git input and child task
# execution. This keeps dispatch coverage fast and avoids recursive quick runs.
real_task=$(command -v task || true)
if [ -z "$real_task" ]; then
  # The hosted Check job has Go but does not preinstall Task.
  GOBIN="$tmpdir/task-bin" "$repo_root/scripts/go-tool.sh" install task
  real_task="$tmpdir/task-bin/task"
fi
dispatch_repo="$tmpdir/dispatch-repo"
dispatch_bin="$tmpdir/dispatch-bin"
mkdir -p "$dispatch_repo/scripts" "$dispatch_repo/docs" "$dispatch_bin"
cp "$repo_root/Taskfile.yml" "$dispatch_repo/Taskfile.yml"
cp "$repo_root/scripts/classify-check-changes.sh" "$dispatch_repo/scripts/"
cp "$repo_root/docs/evidence-baseline-summary-docs.txt" "$dispatch_repo/docs/"
cat >"$dispatch_bin/git" <<'EOF'
#!/bin/sh
set -eu
case "$1" in
  rev-parse) exit 0 ;;
  diff)
    case "$*" in
      *...HEAD) printf '%s\n' "$CPACE_TEST_CHANGED_PATHS" ;;
    esac
    ;;
  ls-files) ;;
  *) echo "unexpected git command: $*" >&2; exit 1 ;;
esac
EOF
cat >"$dispatch_bin/task" <<'EOF'
#!/bin/sh
set -eu
printf '%s\n' "$*" >>"$CPACE_TEST_TASK_LOG"
if [ "$1" = "${CPACE_TEST_FAIL_TASK:-}" ]; then
  exit 42
fi
EOF
chmod +x "$dispatch_bin/git" "$dispatch_bin/task"

assert_dispatch() {
  name=$1
  CPACE_TEST_CHANGED_PATHS=$2
  want_tasks=$3
  CPACE_TEST_FAIL_TASK=${4:-}
  CPACE_TEST_TASK_LOG="$tmpdir/dispatch.tasks"
  export CPACE_TEST_CHANGED_PATHS CPACE_TEST_TASK_LOG CPACE_TEST_FAIL_TASK
  : >"$CPACE_TEST_TASK_LOG"
  status=0
  PATH="$dispatch_bin:$PATH" BASE_REF=origin/main "$real_task" --dir "$dispatch_repo" check:changed >"$tmpdir/dispatch.out" 2>&1 || status=$?
  got_tasks=$(cat "$CPACE_TEST_TASK_LOG")
  if [ "$got_tasks" != "$want_tasks" ]; then
    echo "$name: tasks got [$got_tasks] want [$want_tasks]" >&2
    cat "$tmpdir/dispatch.out" >&2
    exit 1
  fi
  if { [ -z "$CPACE_TEST_FAIL_TASK" ] && [ "$status" -ne 0 ]; } ||
     { [ -n "$CPACE_TEST_FAIL_TASK" ] && [ "$status" -eq 0 ]; }; then
    echo "$name: unexpected dispatch status $status" >&2
    cat "$tmpdir/dispatch.out" >&2
    exit 1
  fi
}

assert_dispatch release-workflow ".github/workflows/release.yml" "quick
release:policy
release:lint"

assert_dispatch unrelated-code "cpace.go" "quick"
assert_dispatch empty "" "quick"
assert_dispatch ordinary-doc "docs/governance.md" "docs:check
ci:go-tools"
assert_dispatch evidence-doc "docs/evidence-baseline.md" "docs:check
ci:go-tools
evidence:baseline"
assert_dispatch mixed "tools/evidencebaseline/main.go
.github/workflows/release.yml" "quick
evidence:baseline
evidence:lint
release:policy
release:lint"
assert_dispatch policy-failure ".github/workflows/release.yml" "quick
release:policy" release:policy
assert_dispatch lint-failure ".github/workflows/release.yml" "quick
release:policy
release:lint" release:lint

echo "CI change classifier tests passed"
