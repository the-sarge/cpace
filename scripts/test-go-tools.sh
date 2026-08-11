#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM

mkdir "$tmpdir/bin"
cat >"$tmpdir/bin/go" <<'EOF'
#!/bin/sh
set -eu
printf '%s\n' "$*"
EOF
chmod +x "$tmpdir/bin/go"

run_tool() {
  PATH="$tmpdir/bin:$PATH" "$repo_root/scripts/go-tool.sh" "$@"
}

assert_install() {
  tool_name=$1
  module_path=$2
  output=$(run_tool install "$tool_name")
  case "$output" in
    "install ${module_path}@v"*) ;;
    *)
      echo "$tool_name: unexpected install command: $output" >&2
      exit 1
      ;;
  esac
}

assert_install task github.com/go-task/task/v3/cmd/task
assert_install govulncheck golang.org/x/vuln/cmd/govulncheck
assert_install gosec github.com/securego/gosec/v2/cmd/gosec
assert_install staticcheck honnef.co/go/tools/cmd/staticcheck
assert_install golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint
assert_install actionlint github.com/rhysd/actionlint/cmd/actionlint

run_output=$(run_tool run staticcheck ./...)
case "$run_output" in
  "run honnef.co/go/tools/cmd/staticcheck@v"*" ./...") ;;
  *)
    echo "staticcheck run: unexpected command: $run_output" >&2
    exit 1
    ;;
esac

if run_tool install unknown >"$tmpdir/unknown.out" 2>"$tmpdir/unknown.err"; then
  echo "unknown tool unexpectedly succeeded" >&2
  exit 1
fi
grep -Fq "unknown Go tool: unknown" "$tmpdir/unknown.err"

pin_count=$(grep -c '^ *cpace_go_tool_version=v' "$repo_root/scripts/go-tool-versions.sh")
if [ "$pin_count" -ne 6 ]; then
  echo "tool catalogue contains $pin_count version pins, want 6" >&2
  exit 1
fi

direct_selector_pattern='go (install|run) [^[:space:]]+@[^[:space:]]+'
for selector in '@v1.2.3' '@latest' '@main' '@deadbeef'; do
  printf 'go install example.com/tool%s\n' "$selector" >"$tmpdir/direct-selector"
  if ! grep -Eq "$direct_selector_pattern" "$tmpdir/direct-selector"; then
    echo "direct selector guard missed $selector" >&2
    exit 1
  fi
done

if git -C "$repo_root" grep -nE "$direct_selector_pattern" -- Taskfile.yml .github/workflows docs/release-checklist.md tools/releasepolicy >"$tmpdir/raw-pins.out"; then
  echo "active configuration contains direct Go tool pins:" >&2
  cat "$tmpdir/raw-pins.out" >&2
  exit 1
fi

grep -Fq 'GO_TOOL:' "$repo_root/Taskfile.yml"
grep -Fq '"{{.GO_TOOL}}" run staticcheck' "$repo_root/Taskfile.yml"
grep -Fq '"{{.GO_TOOL}}" run golangci-lint' "$repo_root/Taskfile.yml"
grep -Fq 'scripts/go-tool.sh run govulncheck' "$repo_root/docs/release-checklist.md"
grep -Fq 'scripts/go-tool.sh run gosec' "$repo_root/docs/release-checklist.md"

for workflow in actionlint.yml staticcheck.yml golangci-lint.yml; do
  workflow_path="$repo_root/.github/workflows/$workflow"
  grep -Fq -- "- 'scripts/go-tool-versions.sh'" "$workflow_path"
  grep -Fq -- "- 'scripts/go-tool.sh'" "$workflow_path"
done

echo "Go tool helper tests passed"
