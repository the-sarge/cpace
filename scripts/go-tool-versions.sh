#!/bin/sh

cpace_go_tool_resolve() {
  case "$1" in
    actionlint)
      cpace_go_tool_module=github.com/rhysd/actionlint/cmd/actionlint
      cpace_go_tool_version=v1.7.12
      ;;
    golangci-lint)
      cpace_go_tool_module=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
      cpace_go_tool_version=v2.12.2
      ;;
    gosec)
      cpace_go_tool_module=github.com/securego/gosec/v2/cmd/gosec
      cpace_go_tool_version=v2.26.1
      ;;
    govulncheck)
      cpace_go_tool_module=golang.org/x/vuln/cmd/govulncheck
      cpace_go_tool_version=v1.3.0
      ;;
    staticcheck)
      cpace_go_tool_module=honnef.co/go/tools/cmd/staticcheck
      cpace_go_tool_version=v0.7.0
      ;;
    task)
      cpace_go_tool_module=github.com/go-task/task/v3/cmd/task
      cpace_go_tool_version=v3.50.0
      ;;
    *)
      return 1
      ;;
  esac
}
