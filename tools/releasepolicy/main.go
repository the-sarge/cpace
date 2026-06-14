package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type finding struct {
	path string
	msg  string
}

func main() {
	repoRoot := flag.String("repo-root", "../..", "repository root")
	flag.Parse()

	findings, err := checkRepo(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "release policy checker failed: %v\n", err)
		os.Exit(2)
	}
	if len(findings) > 0 {
		sort.Slice(findings, func(i, j int) bool {
			if findings[i].path == findings[j].path {
				return findings[i].msg < findings[j].msg
			}
			return findings[i].path < findings[j].path
		})
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "%s: %s\n", f.path, f.msg)
		}
		os.Exit(1)
	}
	fmt.Println("release policy checker passed")
}

func checkRepo(repoRoot string) ([]finding, error) {
	workflowPath := filepath.Join(repoRoot, ".github", "workflows", "release.yml")
	workflow, err := loadYAML(workflowPath)
	if err != nil {
		return nil, err
	}
	var findings []finding
	findings = append(findings, checkWorkflow(workflowPath, workflow)...)
	findings = append(findings, checkAllowedSigners(repoRoot)...)
	findings = append(findings, checkRequiredScripts(repoRoot)...)
	return findings, nil
}

func loadYAML(path string) (*yaml.Node, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(in, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(doc.Content) != 1 {
		return nil, fmt.Errorf("%s: expected one YAML document", path)
	}
	return doc.Content[0], nil
}

func checkWorkflow(path string, root *yaml.Node) []finding {
	c := checker{path: path}
	c.checkDuplicateKeys("$", root)
	if len(c.findings) > 0 {
		return c.findings
	}
	c.checkRoot(root)
	c.checkTriggers(root)
	c.checkTopPermissions(root)
	jobs := mapping(root, "jobs")
	if jobs == nil {
		c.fail("jobs", "missing jobs")
		return c.findings
	}
	c.checkJobSet(jobs)
	c.checkAcceptedJobs(jobs)
	c.checkActionPins(jobs)
	c.checkCheckoutCredentials(jobs)
	c.checkNoRunExpressions(jobs)
	return c.findings
}

type checker struct {
	path     string
	findings []finding
}

func (c *checker) fail(nodePath, msg string) {
	c.findings = append(c.findings, finding{path: c.path + ":" + nodePath, msg: msg})
}

func (c *checker) checkDuplicateKeys(nodePath string, n *yaml.Node) {
	if n == nil {
		return
	}
	switch n.Kind {
	case yaml.MappingNode:
		seen := map[string]bool{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			key := n.Content[i]
			value := n.Content[i+1]
			valuePath := yamlPath(nodePath, key.Value)
			if key.Kind == yaml.ScalarNode {
				if seen[key.Value] {
					c.fail(valuePath, "duplicate YAML key")
				}
				seen[key.Value] = true
			}
			c.checkDuplicateKeys(valuePath, value)
		}
	case yaml.SequenceNode:
		for idx, item := range n.Content {
			c.checkDuplicateKeys(fmt.Sprintf("%s[%d]", nodePath, idx), item)
		}
	}
}

func (c *checker) checkRoot(root *yaml.Node) {
	policy := acceptedReleasePolicy
	if root.Kind != yaml.MappingNode {
		c.fail("$", "workflow must be a mapping")
	}
	if !sameStringSet(mapKeys(root), policy.rootKeys) {
		c.fail("$", "workflow root keys must exactly match the accepted release workflow")
	}
	expectExactScalar(c, "name", mapping(root, "name"), policy.workflowName)
	expectExactScalars(c, "env", mapping(root, "env"), policy.env)
	expectExactScalars(c, "concurrency", mapping(root, "concurrency"), policy.concurrency)
}

func (c *checker) checkTriggers(root *yaml.Node) {
	policy := acceptedReleasePolicy
	on := mapping(root, "on")
	if on == nil {
		c.fail("on", "missing trigger block")
		return
	}
	if !sameStringSet(mapKeys(on), policy.triggerKeys) {
		c.fail("on", "trigger block must contain only push.tags v* and workflow_dispatch")
	}
	push := mapping(on, "push")
	if push == nil {
		c.fail("on.push", "missing push trigger")
	} else {
		if !sameStringSet(mapKeys(push), policy.pushKeys) {
			c.fail("on.push", "push trigger must contain only tags")
		}
		if !sameStringSet(seqStrings(mapping(push, "tags")), policy.pushTags) {
			c.fail("on.push.tags", "push trigger must contain only v* tags")
		}
	}
	if mapping(on, "workflow_dispatch") == nil {
		c.fail("on.workflow_dispatch", "missing workflow_dispatch trigger")
	}
}

func (c *checker) checkTopPermissions(root *yaml.Node) {
	policy := acceptedReleasePolicy
	permissions := mapping(root, "permissions")
	if permissions == nil {
		c.fail("permissions", "missing top-level permissions")
		return
	}
	expectExactScalar(c, "permissions.contents", mapping(permissions, "contents"), policy.topPermission["contents"])
	for _, key := range mapKeys(permissions) {
		if _, ok := policy.topPermission[key]; !ok {
			c.fail("permissions."+key, "unexpected top-level permission")
		}
	}
}

func (c *checker) checkJobSet(jobs *yaml.Node) {
	want := acceptedReleasePolicy.jobNames()
	for _, name := range want {
		if mapping(jobs, name) == nil {
			c.fail("jobs."+name, "missing required job")
		}
	}
	for _, name := range mapKeys(jobs) {
		if !contains(want, name) {
			c.fail("jobs."+name, "unexpected job in release workflow")
		}
	}
}

func (c *checker) checkAcceptedJobs(jobs *yaml.Node) {
	for _, policy := range acceptedReleasePolicy.jobs {
		job := mapping(jobs, policy.name)
		if job == nil {
			continue
		}
		c.checkAcceptedJob("jobs."+policy.name, job, policy)
	}
}

func (c *checker) checkAcceptedJob(jobPath string, job *yaml.Node, policy releaseJobPolicy) {
	expectOnlyKeys(c, jobPath, job, []string{"name", "if", "needs", "runs-on", "timeout-minutes", "permissions", "outputs", "steps"})
	expectExactScalar(c, jobPath+".name", mapping(job, "name"), policy.displayName)
	expectExactScalar(c, jobPath+".runs-on", mapping(job, "runs-on"), policy.runsOn)
	expectExactScalar(c, jobPath+".timeout-minutes", mapping(job, "timeout-minutes"), policy.timeoutMinutes)
	expectExactScalar(c, jobPath+".if", mapping(job, "if"), policy.ifCond)

	gotNeeds := needs(job)
	switch {
	case policy.needs == nil && mapping(job, "needs") != nil:
		c.fail(jobPath+".needs", "job must not declare needs")
	case !sameStringSet(gotNeeds, policy.needs):
		c.fail(jobPath+".needs", "needs must exactly match: "+strings.Join(policy.needs, ", "))
	}

	permissions := mapping(job, "permissions")
	if policy.permissions == nil {
		if permissions != nil {
			c.fail(jobPath+".permissions", "job must inherit top-level contents: read and must not declare permissions")
		}
	} else {
		expectExactScalars(c, jobPath+".permissions", permissions, policy.permissions)
	}

	outputs := mapping(job, "outputs")
	if policy.outputs == nil {
		if outputs != nil {
			c.fail(jobPath+".outputs", "job must not declare outputs")
		}
	} else {
		expectExactScalars(c, jobPath+".outputs", outputs, policy.outputs)
	}

	jobSteps := steps(job)
	gotStepIdentities := make([]string, 0, len(jobSteps))
	for _, step := range jobSteps {
		gotStepIdentities = append(gotStepIdentities, stepIdentity(step))
	}
	wantStepIdentities := policy.stepIdentities()
	if !sameStringSlice(gotStepIdentities, wantStepIdentities) {
		c.fail(jobPath+".steps", fmt.Sprintf("steps must exactly match accepted release policy: got %q, want %q", gotStepIdentities, wantStepIdentities))
		return
	}
	for idx, stepPolicy := range policy.steps {
		if idx >= len(jobSteps) {
			break
		}
		c.checkAcceptedStep(fmt.Sprintf("%s.steps[%d]", jobPath, idx), jobSteps[idx], stepPolicy)
	}
}

func (c *checker) checkAcceptedStep(stepPath string, step *yaml.Node, policy releaseStepPolicy) {
	expectOnlyKeys(c, stepPath, step, []string{"name", "id", "if", "uses", "run", "with", "env", "shell", "continue-on-error"})
	expectOptionalExactScalar(c, stepPath+".name", mapping(step, "name"), policy.name)
	expectOptionalExactScalar(c, stepPath+".id", mapping(step, "id"), policy.id)
	expectOptionalExactScalar(c, stepPath+".if", mapping(step, "if"), policy.ifCond)
	expectOptionalExactScalar(c, stepPath+".shell", mapping(step, "shell"), policy.shell)
	expectOptionalExactScalar(c, stepPath+".continue-on-error", mapping(step, "continue-on-error"), policy.continueOnError)

	usesNode := mapping(step, "uses")
	if policy.usesPrefix == "" {
		if usesNode != nil {
			c.fail(stepPath+".uses", "unexpected uses")
		}
	} else {
		uses, ok := scalarValue(usesNode)
		switch {
		case usesNode == nil:
			c.fail(stepPath+".uses", "uses must start with "+policy.usesPrefix)
		case !ok:
			c.fail(stepPath+".uses", "uses must be a scalar")
		case !strings.HasPrefix(uses, policy.usesPrefix):
			c.fail(stepPath+".uses", "uses must start with "+policy.usesPrefix)
		}
	}

	runNode := mapping(step, "run")
	if policy.runLines == nil {
		if runNode != nil {
			c.fail(stepPath+".run", "unexpected run")
		}
	} else {
		run, ok := scalarValue(runNode)
		switch {
		case runNode == nil:
			c.fail(stepPath+".run", "missing run")
		case !ok:
			c.fail(stepPath+".run", "run must be a scalar")
		default:
			requireExactScriptLines(c, stepPath+".run", run, policy.runLines)
		}
	}

	expectOptionalExactScalars(c, stepPath+".with", mapping(step, "with"), policy.with)
	expectOptionalExactScalars(c, stepPath+".env", mapping(step, "env"), policy.env)
}

func (c *checker) checkActionPins(jobs *yaml.Node) {
	for _, ref := range actionUses(jobs) {
		action, version, ok := strings.Cut(ref.uses, "@")
		if !ok || action == "" || !shaRef.MatchString(version) {
			c.fail(ref.path+".uses", "action must be pinned by 40-character SHA")
		}
	}
}

func (c *checker) checkCheckoutCredentials(jobs *yaml.Node) {
	for _, jobName := range mapKeys(jobs) {
		job := mapping(jobs, jobName)
		if _, ok := acceptedReleasePolicy.job(jobName); ok {
			continue
		}
		for idx, step := range steps(job) {
			uses := scalar(mapping(step, "uses"))
			if !strings.HasPrefix(uses, "actions/checkout@") {
				continue
			}
			want := map[string]string{"persist-credentials": "false"}
			expectExactScalars(c, fmt.Sprintf("jobs.%s.steps[%d].with", jobName, idx), mapping(step, "with"), want)
		}
	}
}

func (c *checker) checkNoRunExpressions(jobs *yaml.Node) {
	for _, ref := range runSteps(jobs) {
		if strings.Contains(ref.run, "${{") {
			c.fail(ref.path+".run", "run steps must not interpolate GitHub expression contexts; pass values through env")
		}
	}
}

func checkAllowedSigners(repoRoot string) []finding {
	path := filepath.Join(repoRoot, ".github", "allowed_signers")
	in, err := os.ReadFile(path)
	if err != nil {
		return []finding{{path: path, msg: err.Error()}}
	}
	if strings.ReplaceAll(string(in), "\r\n", "\n") != acceptedReleasePolicy.expectedSigners {
		return []finding{{path: path, msg: "allowed_signers must exactly match the accepted maintainer signing keys"}}
	}
	return nil
}

func checkRequiredScripts(repoRoot string) []finding {
	var findings []finding
	for _, path := range acceptedReleasePolicy.requiredScripts {
		full := filepath.Join(repoRoot, path)
		info, err := os.Stat(full)
		switch {
		case errors.Is(err, os.ErrNotExist):
			findings = append(findings, finding{path: full, msg: "missing required release helper"})
		case err != nil:
			findings = append(findings, finding{path: full, msg: err.Error()})
		case info.IsDir():
			findings = append(findings, finding{path: full, msg: "expected file, got directory"})
		case info.Mode().Perm()&0111 == 0:
			findings = append(findings, finding{path: full, msg: "required release helper must be executable"})
		}
	}
	return findings
}

func expectScalars(c *checker, nodePath string, n *yaml.Node, want map[string]string) bool {
	if n == nil {
		c.fail(nodePath, "missing mapping")
		return false
	}
	if n.Kind != yaml.MappingNode {
		c.fail(nodePath, "must be a mapping")
		return false
	}
	for key, expected := range want {
		expectExactScalar(c, nodePath+"."+key, mapping(n, key), expected)
	}
	return true
}

func expectExactScalars(c *checker, nodePath string, n *yaml.Node, want map[string]string) {
	if !expectScalars(c, nodePath, n, want) {
		return
	}
	for _, key := range mapKeys(n) {
		if _, ok := want[key]; !ok {
			c.fail(nodePath+"."+key, "unexpected key")
		}
	}
}

func expectExactScalar(c *checker, nodePath string, n *yaml.Node, want string) {
	got, ok := scalarValue(n)
	if n != nil && !ok {
		c.fail(nodePath, "must be a scalar")
		return
	}
	if got != want {
		c.fail(nodePath, fmt.Sprintf("got %q, want %q", got, want))
	}
}

func expectOptionalExactScalar(c *checker, nodePath string, n *yaml.Node, want string) {
	if want == "" {
		if n != nil {
			c.fail(nodePath, "unexpected value")
		}
		return
	}
	got, ok := scalarValue(n)
	if n != nil && !ok {
		c.fail(nodePath, "must be a scalar")
		return
	}
	if got != want {
		c.fail(nodePath, fmt.Sprintf("got %q, want %q", got, want))
	}
}

func expectOptionalExactScalars(c *checker, nodePath string, n *yaml.Node, want map[string]string) {
	if want == nil {
		if n != nil {
			c.fail(nodePath, "unexpected mapping")
		}
		return
	}
	expectExactScalars(c, nodePath, n, want)
}

func expectOnlyKeys(c *checker, nodePath string, n *yaml.Node, want []string) {
	if n == nil || n.Kind != yaml.MappingNode {
		c.fail(nodePath, "missing mapping")
		return
	}
	for _, key := range mapKeys(n) {
		if !contains(want, key) {
			c.fail(nodePath+"."+key, "unexpected key")
		}
	}
}

func requireExactScriptLines(c *checker, nodePath, run string, want []string) {
	lines := scriptLines(run)
	if sameStringSlice(lines, want) {
		return
	}
	c.fail(nodePath, fmt.Sprintf("script lines must exactly match accepted release policy: got %q, want %q", lines, want))
}

func scriptLines(run string) []string {
	var out []string
	for _, line := range strings.Split(run, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

type actionRef struct {
	path string
	uses string
}

func actionUses(jobs *yaml.Node) []actionRef {
	var out []actionRef
	for _, jobName := range mapKeys(jobs) {
		job := mapping(jobs, jobName)
		for idx, step := range steps(job) {
			if uses := scalar(mapping(step, "uses")); uses != "" {
				out = append(out, actionRef{path: fmt.Sprintf("jobs.%s.steps[%d]", jobName, idx), uses: uses})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}

type runRef struct {
	path string
	run  string
}

func runSteps(jobs *yaml.Node) []runRef {
	var out []runRef
	for _, jobName := range mapKeys(jobs) {
		job := mapping(jobs, jobName)
		for idx, step := range steps(job) {
			if run := scalar(mapping(step, "run")); run != "" {
				out = append(out, runRef{path: fmt.Sprintf("jobs.%s.steps[%d]", jobName, idx), run: run})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}

func stepIdentity(step *yaml.Node) string {
	if name := scalar(mapping(step, "name")); name != "" {
		return name
	}
	if uses := scalar(mapping(step, "uses")); uses != "" {
		action, _, _ := strings.Cut(uses, "@")
		return "uses:" + action
	}
	return "unnamed"
}

func steps(job *yaml.Node) []*yaml.Node {
	n := mapping(job, "steps")
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	return n.Content
}

func needs(job *yaml.Node) []string {
	n := mapping(job, "needs")
	switch {
	case n == nil:
		return nil
	case n.Kind == yaml.ScalarNode:
		return []string{n.Value}
	case n.Kind == yaml.SequenceNode:
		return seqStrings(n)
	default:
		return nil
	}
}

func mapping(n *yaml.Node, key string) *yaml.Node {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

func scalar(n *yaml.Node) string {
	value, _ := scalarValue(n)
	return value
}

func scalarValue(n *yaml.Node) (string, bool) {
	if n == nil || n.Kind != yaml.ScalarNode {
		return "", false
	}
	return n.Value, true
}

func seqStrings(n *yaml.Node) []string {
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	out := make([]string, 0, len(n.Content))
	for _, item := range n.Content {
		out = append(out, scalar(item))
	}
	return out
}

func mapKeys(n *yaml.Node) []string {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	out := make([]string, 0, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		out = append(out, n.Content[i].Value)
	}
	sort.Strings(out)
	return out
}

func yamlPath(parent, child string) string {
	if parent == "" || parent == "$" {
		return child
	}
	return parent + "." + child
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func sameStringSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for _, item := range got {
		if !contains(want, item) {
			return false
		}
	}
	for _, item := range want {
		if !contains(got, item) {
			return false
		}
	}
	return true
}

func sameStringSlice(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
