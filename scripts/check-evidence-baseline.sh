#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

(
  cd "$repo_root/tools/evidencebaseline"
  # Tests read repository files outside this module, beyond Go's test-cache tracking.
  go test -count=1 ./...
  go run . --repo-root "$repo_root"
)
