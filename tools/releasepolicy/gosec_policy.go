package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	acceptedGosecWorkflowCommand = "task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'"
	acceptedSASTWorkflowCommand  = "task lint:golangci GOLANGCI_LINT=golangci-lint " +
		"GOLANGCI_ARGS='--output.text.path stdout --output.sarif.path golangci.sarif'"
	acceptedTaskInstallCommand = "scripts/go-tool.sh install task"
)

func checkGosecTaskPolicy(repoRoot string) ([]finding, error) {
	path := filepath.Join(repoRoot, "Taskfile.yml")
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	c := checker{path: path}
	c.checkDuplicateKeys("$", root)
	if len(c.findings) > 0 {
		return c.findings, nil
	}

	task := mapping(mapping(root, "tasks"), "gosec")
	commands := mapping(task, "cmds")
	if commands == nil || commands.Kind != yaml.SequenceNode || len(commands.Content) != 1 {
		c.fail("tasks.gosec.cmds", "gosec task must contain exactly one command")
		return c.findings, nil
	}
	command := scalar(commands.Content[0])
	if !strings.HasPrefix(command, "{{.GOSEC}} ") || !strings.HasSuffix(command, " ./...") {
		c.fail("tasks.gosec.cmds[0]", fmt.Sprintf("gosec task command %q must own scan policy between {{.GOSEC}} and ./...", command))
	}
	return c.findings, nil
}

func checkScanWorkflowTaskRouting(repoRoot, filename, jobName, stepID, wantCommand string) ([]finding, error) {
	path := filepath.Join(repoRoot, ".github", "workflows", filename)
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	c := checker{path: path}
	c.checkDuplicateKeys("$", root)
	if len(c.findings) > 0 {
		return c.findings, nil
	}

	job := mapping(mapping(root, "jobs"), jobName)
	var scanSteps []*yaml.Node
	scanStepIndex := -1
	var taskInstallIndices []int
	for idx, step := range steps(job) {
		if scalar(mapping(step, "id")) == stepID {
			scanSteps = append(scanSteps, step)
			scanStepIndex = idx
		}
		if scalar(mapping(step, "run")) == acceptedTaskInstallCommand {
			taskInstallIndices = append(taskInstallIndices, idx)
		}
	}
	if len(scanSteps) != 1 {
		c.fail("jobs."+jobName+".steps", fmt.Sprintf("scan lane must contain exactly one id %s step got %d", stepID, len(scanSteps)))
		return c.findings, nil
	}
	if got := scalar(mapping(scanSteps[0], "run")); got != wantCommand {
		c.fail("jobs."+jobName+".steps."+stepID+".run", fmt.Sprintf("scan lane command got %q want %q", got, wantCommand))
	}
	if len(taskInstallIndices) != 1 || taskInstallIndices[0] >= scanStepIndex {
		c.fail("jobs."+jobName+".steps", "scan lane must install task before scanning")
	}
	return c.findings, nil
}

// checkGolangciTaskPolicy keeps the Taskfile the single owner of golangci-lint
// scan-scope and output flags, mirroring the gosec task policy. Workflow lanes
// may only pass GOLANGCI_LINT and GOLANGCI_ARGS through the task facade.
func checkGolangciTaskPolicy(repoRoot string) ([]finding, error) {
	path := filepath.Join(repoRoot, "Taskfile.yml")
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	c := checker{path: path}
	c.checkDuplicateKeys("$", root)
	if len(c.findings) > 0 {
		return c.findings, nil
	}

	task := mapping(mapping(root, "tasks"), "lint:golangci")
	commands := mapping(task, "cmds")
	if commands == nil || commands.Kind != yaml.SequenceNode || len(commands.Content) != 2 {
		c.fail("tasks.lint:golangci.cmds", "lint:golangci task must contain the config check and exactly one lint command")
		return c.findings, nil
	}
	command := scalar(commands.Content[1])
	if !strings.Contains(command, "{{.GOLANGCI_LINT}} run {{.GOLANGCI_ARGS}}") ||
		!strings.Contains(command, `"{{.GO_TOOL}}" run golangci-lint run {{.GOLANGCI_ARGS}}`) {
		c.fail("tasks.lint:golangci.cmds[1]", fmt.Sprintf("lint:golangci command %q must route both branches through {{.GOLANGCI_ARGS}}", command))
	}
	return c.findings, nil
}

func checkSASTFailurePolicy(repoRoot string) ([]finding, error) {
	path := filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml")
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	job := mapping(mapping(root, "jobs"), "sast-gate")
	if continueOnError := mapping(job, "continue-on-error"); continueOnError != nil && scalar(continueOnError) != "false" {
		return []finding{{path: path + ":jobs.sast-gate.continue-on-error", msg: "SAST job must remain blocking"}}, nil
	}
	scanIndex := -1
	reportIndex := -1
	var reportStep *yaml.Node
	for idx, step := range steps(job) {
		if scalar(mapping(step, "id")) == "sast" {
			scanIndex = idx
		}
		if scalar(mapping(step, "name")) == "Report golangci-lint result" {
			reportIndex = idx
			reportStep = step
		}
	}
	if reportStep == nil || scalar(mapping(reportStep, "if")) != "steps.sast.outcome == 'failure'" || scalar(mapping(reportStep, "run")) != "exit 1" {
		return []finding{{path: path + ":jobs.sast-gate.steps", msg: "SAST report step must fail the job when the scan fails"}}, nil
	}
	if continueOnError := mapping(reportStep, "continue-on-error"); continueOnError != nil && scalar(continueOnError) != "false" {
		return []finding{{path: path + ":jobs.sast-gate.steps.Report golangci-lint result.continue-on-error", msg: "SAST report step must remain blocking"}}, nil
	}
	if scanIndex < 0 || reportIndex <= scanIndex {
		return []finding{{path: path + ":jobs.sast-gate.steps", msg: "SAST report step must follow the scan"}}, nil
	}
	return nil, nil
}
