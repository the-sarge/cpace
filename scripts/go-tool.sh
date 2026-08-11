#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$repo_root/scripts/go-tool-versions.sh"

usage() {
  echo "usage: $0 install|run TOOL [ARG ...]" >&2
  exit 2
}

[ "$#" -ge 2 ] || usage
mode=$1
tool_name=$2
shift 2

if ! cpace_go_tool_resolve "$tool_name"; then
  echo "unknown Go tool: $tool_name" >&2
  exit 2
fi

case "$mode" in
  install)
    [ "$#" -eq 0 ] || usage
    exec go install "${cpace_go_tool_module}@${cpace_go_tool_version}"
    ;;
  run)
    exec go run "${cpace_go_tool_module}@${cpace_go_tool_version}" "$@"
    ;;
  *)
    usage
    ;;
esac
