package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	acceptedGosecWorkflowCommand = "task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'"
	acceptedTaskInstallCommand   = "scripts/go-tool.sh install task"
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

func checkGosecWorkflowTaskRouting(repoRoot, filename, jobName string) ([]finding, error) {
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
	var gosecSteps []*yaml.Node
	gosecStepIndex := -1
	var taskInstallIndices []int
	for idx, step := range steps(job) {
		if scalar(mapping(step, "id")) == "gosec" {
			gosecSteps = append(gosecSteps, step)
			gosecStepIndex = idx
		}
		if scalar(mapping(step, "run")) == acceptedTaskInstallCommand {
			taskInstallIndices = append(taskInstallIndices, idx)
		}
	}
	if len(gosecSteps) != 1 {
		c.fail("jobs."+jobName+".steps", fmt.Sprintf("gosec lane must contain exactly one id gosec step got %d", len(gosecSteps)))
		return c.findings, nil
	}
	if got := scalar(mapping(gosecSteps[0], "run")); got != acceptedGosecWorkflowCommand {
		c.fail("jobs."+jobName+".steps.gosec.run", fmt.Sprintf("gosec lane command got %q want %q", got, acceptedGosecWorkflowCommand))
	}
	if len(taskInstallIndices) != 1 || taskInstallIndices[0] >= gosecStepIndex {
		c.fail("jobs."+jobName+".steps", "gosec lane must install task before scanning")
	}
	return c.findings, nil
}

func checkAdvisoryGosecFailurePolicy(repoRoot string) ([]finding, error) {
	path := filepath.Join(repoRoot, ".github", "workflows", "gosec.yml")
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	job := mapping(mapping(root, "jobs"), "gosec")
	if scalar(mapping(job, "continue-on-error")) == "true" {
		return nil, nil
	}
	return []finding{{path: path + ":jobs.gosec.continue-on-error", msg: "advisory gosec job must remain non-blocking"}}, nil
}

func checkSASTGosecFailurePolicy(repoRoot string) ([]finding, error) {
	path := filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml")
	root, err := loadYAML(path)
	if err != nil {
		return nil, err
	}
	job := mapping(mapping(root, "jobs"), "sast-gate")
	if continueOnError := mapping(job, "continue-on-error"); continueOnError != nil && scalar(continueOnError) != "false" {
		return []finding{{path: path + ":jobs.sast-gate.continue-on-error", msg: "SAST gosec job must remain blocking"}}, nil
	}
	scanIndex := -1
	reportIndex := -1
	var reportStep *yaml.Node
	for idx, step := range steps(job) {
		if scalar(mapping(step, "id")) == "gosec" {
			scanIndex = idx
		}
		if scalar(mapping(step, "name")) == "Report gosec result" {
			reportIndex = idx
			reportStep = step
		}
	}
	if reportStep == nil || scalar(mapping(reportStep, "if")) != "steps.gosec.outcome == 'failure'" || scalar(mapping(reportStep, "run")) != "exit 1" {
		return []finding{{path: path + ":jobs.sast-gate.steps", msg: "SAST gosec report step must fail the job when the scan fails"}}, nil
	}
	if continueOnError := mapping(reportStep, "continue-on-error"); continueOnError != nil && scalar(continueOnError) != "false" {
		return []finding{{path: path + ":jobs.sast-gate.steps.Report gosec result.continue-on-error", msg: "SAST gosec report step must remain blocking"}}, nil
	}
	if scanIndex < 0 || reportIndex <= scanIndex {
		return []finding{{path: path + ":jobs.sast-gate.steps", msg: "SAST gosec report step must follow the scan"}}, nil
	}
	return nil, nil
}
