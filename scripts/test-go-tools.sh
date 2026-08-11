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

direct_go_tool_pattern='(^|[[:space:]])go[[:space:]]+(install|run)([[:space:]]|$)'
scan_direct_go_tool_commands() {
  grep -En "$direct_go_tool_pattern" "$@"
}

fixture_number=0
for command in \
  'go install example.com/tool@v1.2.3' \
  'go  install example.com/tool@latest' \
  'go run -mod=readonly example.com/tool@main' \
  'go install -v example.com/tool@deadbeef'; do
  fixture_number=$((fixture_number + 1))
  printf '%s\n' "$command" >"$tmpdir/direct-command-$fixture_number"
  if ! scan_direct_go_tool_commands "$tmpdir/direct-command-$fixture_number" >/dev/null; then
    echo "direct Go tool guard missed: $command" >&2
    exit 1
  fi
done
printf 'go\tinstall example.com/tool@latest\n' >"$tmpdir/direct-command-tab"
if ! scan_direct_go_tool_commands "$tmpdir/direct-command-tab" >/dev/null; then
  echo "direct Go tool guard missed tab-separated command" >&2
  exit 1
fi

if scan_direct_go_tool_commands \
  "$repo_root/Taskfile.yml" \
  "$repo_root"/.github/workflows/*.yml \
  "$repo_root/docs/release-checklist.md" \
  "$repo_root"/tools/releasepolicy/*.go >"$tmpdir/raw-pins.out"; then
  echo "active configuration contains a direct go install/run command:" >&2
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
