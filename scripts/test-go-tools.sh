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

scan_direct_go_module_selectors() {
  grep -En '[[:alnum:]_.-]+\.[[:alpha:]][[:alnum:]-]*/[^[:space:]]+@[^[:space:]]+' "$@"
}

scan_direct_go_tool_commands() {
  found=1
  for path in "$@"; do
    if awk '{ sub(/\\[[:space:]]*$/, ""); printf "%s ", $0 }' "$path" | grep -Eq '(^|[^[:alnum:]_])go[[:space:]]+((-C([[:space:]]+[^[:space:]]+|=[^[:space:]]+)|-[nx])[[:space:]]+)*(install|run)([[:space:]]|$)'; then
      printf '%s\n' "$path"
      found=0
    fi
  done
  return "$found"
}

fixture_number=0
for command in \
  'go install example.com/tool@v1.2.3' \
  'go  install example.com/tool@latest' \
  'go run -mod=readonly example.com/tool@main' \
  'go install -v example.com/tool@deadbeef'; do
  fixture_number=$((fixture_number + 1))
  printf '%s\n' "$command" >"$tmpdir/direct-command-$fixture_number"
  if ! scan_direct_go_module_selectors "$tmpdir/direct-command-$fixture_number" >/dev/null; then
    echo "direct Go module selector guard missed: $command" >&2
    exit 1
  fi
done
printf 'go\tinstall example.com/tool@latest\n' >"$tmpdir/direct-command-tab"
if ! scan_direct_go_module_selectors "$tmpdir/direct-command-tab" >/dev/null; then
  echo "direct Go module selector guard missed tab-separated command" >&2
  exit 1
fi
printf 'run: "go install example.com/tool@latest"\n' >"$tmpdir/direct-command-yaml"
printf 'runLines: []string{`go run example.com/tool@main`},\n' >"$tmpdir/direct-command-policy"
printf 'go \\\n  install example.com/tool@deadbeef\n' >"$tmpdir/direct-command-continuation"
printf '%s\n' 'run: >-' '  go install' '  example.com/tool@latest' >"$tmpdir/direct-command-folded-yaml"
printf '%s\n' 'go -C . install example.com/tool@latest' >"$tmpdir/direct-command-go-c"
printf '%s\n' 'go install "$MODULE@$VERSION"' >"$tmpdir/direct-command-variable"
printf '%s\n' 'run: go install example.com/tool@v1.2.3' >"$tmpdir/workflow.yaml"
for fixture in "$tmpdir/direct-command-yaml" "$tmpdir/direct-command-policy" "$tmpdir/direct-command-continuation" "$tmpdir/direct-command-folded-yaml" "$tmpdir/direct-command-go-c"; do
  if ! scan_direct_go_module_selectors "$fixture" >/dev/null; then
    echo "direct Go module selector guard missed formatted fixture: $fixture" >&2
    exit 1
  fi
done

for fixture in "$tmpdir"/direct-command-*; do
  if ! scan_direct_go_tool_commands "$fixture" >/dev/null; then
    echo "direct Go tool command guard missed fixture: $fixture" >&2
    exit 1
  fi
done
if ! scan_direct_go_tool_commands "$tmpdir/workflow.yaml" >/dev/null; then
  echo "direct Go tool command guard missed .yaml workflow fixture" >&2
  exit 1
fi

set -- \
  "$repo_root/Taskfile.yml" \
  "$repo_root/docs/release-checklist.md" \
  "$repo_root"/tools/releasepolicy/*.go
workflow_files=$(git -C "$repo_root" ls-files -- '.github/workflows/*.yml' '.github/workflows/*.yaml')
[ -n "$workflow_files" ] || {
  echo "no tracked GitHub workflows found" >&2
  exit 1
}
while IFS= read -r workflow; do
  set -- "$@" "$repo_root/$workflow"
done <<EOF
$workflow_files
EOF
: >"$tmpdir/raw-pins.out"
raw_tool_configuration=false
if scan_direct_go_tool_commands "$@" >>"$tmpdir/raw-pins.out"; then
  raw_tool_configuration=true
fi
if scan_direct_go_module_selectors "$@" >>"$tmpdir/raw-pins.out"; then
  raw_tool_configuration=true
fi
if [ "$raw_tool_configuration" = true ]; then
  echo "active configuration contains a direct Go tool command or module selector:" >&2
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
