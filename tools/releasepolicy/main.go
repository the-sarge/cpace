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
	c.checkRoot(root)
	c.checkTriggers(root)
	c.checkTopPermissions(root)
	jobs := mapping(root, "jobs")
	if jobs == nil {
		c.fail("jobs", "missing jobs")
		return c.findings
	}
	c.checkJobSet(jobs)
	c.checkStepSets(jobs)
	c.checkJobPermissions(jobs)
	c.checkUnsupportedRefJob(jobs)
	c.checkVerifyTagJob(jobs)
	c.checkNeeds(jobs)
	c.checkActionPins(jobs)
	c.checkCheckoutCredentials(jobs)
	c.checkNoRunExpressions(jobs)
	c.checkValidationJobs(jobs)
	c.checkSBOMJob(jobs)
	c.checkSBOMAttestationJob(jobs)
	c.checkReleaseJob(jobs)
	return c.findings
}

type checker struct {
	path     string
	findings []finding
}

func (c *checker) fail(nodePath, msg string) {
	c.findings = append(c.findings, finding{path: c.path + ":" + nodePath, msg: msg})
}

func (c *checker) checkRoot(root *yaml.Node) {
	policy := acceptedReleasePolicy
	if root.Kind != yaml.MappingNode {
		c.fail("$", "workflow must be a mapping")
	}
	if !sameStringSet(mapKeys(root), policy.rootKeys) {
		c.fail("$", "workflow root keys must exactly match the accepted release workflow")
	}
	if scalar(mapping(root, "name")) != policy.workflowName {
		c.fail("name", "workflow name must be Release Validation")
	}
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
	if scalar(mapping(permissions, "contents")) != policy.topPermission["contents"] {
		c.fail("permissions.contents", "top-level contents permission must be read")
	}
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

func (c *checker) checkStepSets(jobs *yaml.Node) {
	for _, jobName := range mapKeys(jobs) {
		job := mapping(jobs, jobName)
		expected, ok := acceptedReleasePolicy.job(jobName)
		if !ok || job == nil {
			continue
		}
		var got []string
		for _, step := range steps(job) {
			got = append(got, stepIdentity(step))
		}
		want := expected.stepIdentities()
		if !sameStringSlice(got, want) {
			c.fail("jobs."+jobName+".steps", fmt.Sprintf("steps must exactly match accepted release policy: got %q, want %q", got, want))
		}
	}
}

func (c *checker) checkJobPermissions(jobs *yaml.Node) {
	for _, policy := range acceptedReleasePolicy.jobs {
		job := mapping(jobs, policy.name)
		if job == nil {
			continue
		}
		permissions := mapping(job, "permissions")
		if policy.permissions == nil {
			if permissions != nil {
				c.fail("jobs."+policy.name+".permissions", "job must inherit top-level contents: read and must not declare permissions")
			}
			continue
		}
		expectExactScalars(c, "jobs."+policy.name+".permissions", permissions, policy.permissions)
	}
}

func (c *checker) checkUnsupportedRefJob(jobs *yaml.Node) {
	policy := acceptedJobPolicy("unsupported-ref")
	job := mapping(jobs, "unsupported-ref")
	if job == nil {
		return
	}
	ifCond := scalar(mapping(job, "if"))
	if ifCond != policy.ifCond {
		c.fail("jobs.unsupported-ref.if", "unsupported-ref job must be limited to non-v* workflow_dispatch refs")
	}
	steps := steps(job)
	if len(steps) != 1 {
		c.fail("jobs.unsupported-ref.steps", "unsupported-ref job should have one explanatory failing step")
		return
	}
	run := scalar(mapping(steps[0], "run"))
	requireExactScriptLines(c, "jobs.unsupported-ref.steps[0].run", run, acceptedStepPolicy("unsupported-ref", "Explain unsupported ref").runLines)
}

func (c *checker) checkVerifyTagJob(jobs *yaml.Node) {
	policy := acceptedJobPolicy("verify-tag")
	job := mapping(jobs, "verify-tag")
	if job == nil {
		return
	}
	if needs(job) != nil {
		c.fail("jobs.verify-tag.needs", "verify-tag must not depend on another job")
	}
	if scalar(mapping(job, "if")) != policy.ifCond {
		c.fail("jobs.verify-tag.if", "verify-tag must run only for signed v* tag refs")
	}
	outputs := mapping(job, "outputs")
	expectExactScalars(c, "jobs.verify-tag.outputs", outputs, policy.outputs)
	verifyStep := stepByName(job, "Verify tag object and signature")
	if verifyStep == nil {
		c.fail("jobs.verify-tag.steps", "missing tag verification step")
		return
	}
	run := scalar(mapping(verifyStep, "run"))
	requireExactScriptLines(c, "jobs.verify-tag.steps.Verify tag object and signature.run", run, acceptedStepPolicy("verify-tag", "Verify tag object and signature").runLines)
	metaStep := stepByName(job, "Validate release tag metadata")
	if metaStep == nil {
		c.fail("jobs.verify-tag.steps", "missing release tag metadata step")
	} else {
		requireExactScriptLines(c, "jobs.verify-tag.steps.Validate release tag metadata.run", scalar(mapping(metaStep, "run")), acceptedStepPolicy("verify-tag", "Validate release tag metadata").runLines)
	}
}

func (c *checker) checkNeeds(jobs *yaml.Node) {
	for _, policy := range acceptedReleasePolicy.jobs {
		job := mapping(jobs, policy.name)
		if job == nil {
			continue
		}
		got := needs(job)
		if !sameStringSet(got, policy.needs) {
			c.fail("jobs."+policy.name+".needs", "needs must exactly match: "+strings.Join(policy.needs, ", "))
		}
		if policy.name != "unsupported-ref" && scalar(mapping(job, "if")) != policy.ifCond {
			c.fail("jobs."+policy.name+".if", "job must run only for signed v* tag refs")
		}
	}
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
		policy, ok := acceptedReleasePolicy.job(jobName)
		for idx, step := range steps(job) {
			uses := scalar(mapping(step, "uses"))
			if !strings.HasPrefix(uses, "actions/checkout@") {
				continue
			}
			want := map[string]string{"persist-credentials": "false"}
			if ok {
				checkout, ok := policy.stepByIdentity("uses:actions/checkout")
				if !ok {
					c.fail(fmt.Sprintf("jobs.%s.steps[%d]", jobName, idx), "checkout step is not part of accepted release policy")
					continue
				}
				want = checkout.with
			}
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

func (c *checker) checkValidationJobs(jobs *yaml.Node) {
	check := mapping(jobs, "check")
	if check != nil {
		runTests := stepByName(check, "Run tests")
		if runTests == nil {
			c.fail("jobs.check.steps", "missing Run tests step")
		} else {
			requireExactScriptLines(c, "jobs.check.steps.Run tests.run", scalar(mapping(runTests, "run")), acceptedStepPolicy("check", "Run tests").runLines)
		}
	}
	race := mapping(jobs, "race")
	if race != nil {
		runRace := stepByName(race, "Run race tests")
		if runRace == nil {
			c.fail("jobs.race.steps", "missing Run race tests step")
		} else {
			requireExactScriptLines(c, "jobs.race.steps.Run race tests.run", scalar(mapping(runRace, "run")), acceptedStepPolicy("race", "Run race tests").runLines)
		}
	}
	vuln := mapping(jobs, "vuln")
	if vuln != nil {
		requireRunStep(c, vuln, "jobs.vuln", "Install task", acceptedStepPolicy("vuln", "Install task").runLines)
		requireRunStep(c, vuln, "jobs.vuln", "Install govulncheck", acceptedStepPolicy("vuln", "Install govulncheck").runLines)
		requireRunStep(c, vuln, "jobs.vuln", "Run vulnerability scan", acceptedStepPolicy("vuln", "Run vulnerability scan").runLines)
	}
	gosec := mapping(jobs, "gosec")
	if gosec != nil {
		requireRunStep(c, gosec, "jobs.gosec", "Install task", acceptedStepPolicy("gosec", "Install task").runLines)
		requireRunStep(c, gosec, "jobs.gosec", "Install gosec", acceptedStepPolicy("gosec", "Install gosec").runLines)
		scanPolicy := acceptedStepPolicy("gosec", "Run gosec scan")
		scan := requireRunStep(c, gosec, "jobs.gosec", "Run gosec scan", scanPolicy.runLines)
		if scan != nil {
			if scalar(mapping(scan, "id")) != scanPolicy.id {
				c.fail("jobs.gosec.steps.Run gosec scan.id", "gosec scan step id must be gosec")
			}
			if scalar(mapping(scan, "continue-on-error")) != scanPolicy.continueOnError {
				c.fail("jobs.gosec.steps.Run gosec scan.continue-on-error", "gosec scan must continue so the report step can fail explicitly")
			}
		}
		upload := stepByName(gosec, "Upload gosec SARIF to code scanning")
		uploadPolicy := acceptedStepPolicy("gosec", "Upload gosec SARIF to code scanning")
		if upload == nil {
			c.fail("jobs.gosec.steps", "missing gosec SARIF upload step")
		} else {
			if scalar(mapping(upload, "if")) != uploadPolicy.ifCond {
				c.fail("jobs.gosec.steps.Upload gosec SARIF to code scanning.if", "gosec SARIF upload guard changed")
			}
			if !strings.HasPrefix(scalar(mapping(upload, "uses")), uploadPolicy.usesPrefix) {
				c.fail("jobs.gosec.steps.Upload gosec SARIF to code scanning.uses", "gosec SARIF upload must use github/codeql-action/upload-sarif")
			}
			expectExactScalars(c, "jobs.gosec.steps.Upload gosec SARIF to code scanning.with", mapping(upload, "with"), uploadPolicy.with)
		}
		reportPolicy := acceptedStepPolicy("gosec", "Report gosec result")
		report := requireRunStep(c, gosec, "jobs.gosec", "Report gosec result", reportPolicy.runLines)
		if report != nil && scalar(mapping(report, "if")) != reportPolicy.ifCond {
			c.fail("jobs.gosec.steps.Report gosec result.if", "gosec failure report guard changed")
		}
	}
}

func requireRunStep(c *checker, job *yaml.Node, jobPath, name string, want []string) *yaml.Node {
	step := stepByName(job, name)
	if step == nil {
		c.fail(jobPath+".steps", "missing "+name+" step")
		return nil
	}
	requireExactScriptLines(c, jobPath+".steps."+name+".run", scalar(mapping(step, "run")), want)
	return step
}

func acceptedJobPolicy(name string) releaseJobPolicy {
	job, ok := acceptedReleasePolicy.job(name)
	if !ok {
		panic("release policy catalogue missing job " + name)
	}
	return job
}

func acceptedStepPolicy(jobName, stepName string) releaseStepPolicy {
	job := acceptedJobPolicy(jobName)
	step, ok := job.step(stepName)
	if !ok {
		panic("release policy catalogue missing step " + jobName + "/" + stepName)
	}
	return step
}

func (c *checker) checkSBOMJob(jobs *yaml.Node) {
	policy := acceptedJobPolicy("sbom")
	job := mapping(jobs, "sbom")
	if job == nil {
		return
	}
	outputs := mapping(job, "outputs")
	for _, name := range policy.requiredOutputs {
		if scalar(mapping(outputs, name)) == "" {
			c.fail("jobs.sbom.outputs."+name, "missing SBOM output")
		}
	}
	gen := stepByName(job, "Generate CycloneDX SBOM")
	genPolicy := acceptedStepPolicy("sbom", "Generate CycloneDX SBOM")
	if gen == nil {
		c.fail("jobs.sbom.steps", "missing SBOM generation step")
		return
	}
	if !strings.HasPrefix(scalar(mapping(gen, "uses")), genPolicy.usesPrefix) {
		c.fail("jobs.sbom.steps.Generate CycloneDX SBOM.uses", "SBOM generation must use anchore/sbom-action")
	}
	expectExactScalars(c, "jobs.sbom.steps.Generate CycloneDX SBOM.with", mapping(gen, "with"), genPolicy.with)
	validate := stepByName(job, "Validate SBOM and compute checksum")
	validatePolicy := acceptedStepPolicy("sbom", "Validate SBOM and compute checksum")
	if validate == nil {
		c.fail("jobs.sbom.steps", "missing SBOM validation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom.steps.Validate SBOM and compute checksum.run", scalar(mapping(validate, "run")), validatePolicy.runLines)
		expectExactScalars(c, "jobs.sbom.steps.Validate SBOM and compute checksum.env", mapping(validate, "env"), validatePolicy.env)
	}
	upload := stepByName(job, "Upload SBOM artifact")
	uploadPolicy := acceptedStepPolicy("sbom", "Upload SBOM artifact")
	if upload == nil {
		c.fail("jobs.sbom.steps", "missing SBOM artifact upload step")
	} else {
		if !strings.HasPrefix(scalar(mapping(upload, "uses")), uploadPolicy.usesPrefix) {
			c.fail("jobs.sbom.steps.Upload SBOM artifact.uses", "SBOM artifact upload must use actions/upload-artifact")
		}
		expectExactScalars(c, "jobs.sbom.steps.Upload SBOM artifact.with", mapping(upload, "with"), uploadPolicy.with)
	}
}

func (c *checker) checkSBOMAttestationJob(jobs *yaml.Node) {
	job := mapping(jobs, "sbom-attestation")
	if job == nil {
		return
	}
	download := stepByName(job, "Download SBOM artifact")
	downloadPolicy := acceptedStepPolicy("sbom-attestation", "Download SBOM artifact")
	if download == nil {
		c.fail("jobs.sbom-attestation.steps", "missing SBOM artifact download step")
	} else {
		if !strings.HasPrefix(scalar(mapping(download, "uses")), downloadPolicy.usesPrefix) {
			c.fail("jobs.sbom-attestation.steps.Download SBOM artifact.uses", "SBOM download must use actions/download-artifact")
		}
		expectExactScalars(c, "jobs.sbom-attestation.steps.Download SBOM artifact.with", mapping(download, "with"), downloadPolicy.with)
	}
	validate := stepByName(job, "Validate downloaded SBOM")
	validatePolicy := acceptedStepPolicy("sbom-attestation", "Validate downloaded SBOM")
	if validate == nil {
		c.fail("jobs.sbom-attestation.steps", "missing downloaded SBOM validation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom-attestation.steps.Validate downloaded SBOM.run", scalar(mapping(validate, "run")), validatePolicy.runLines)
		expectExactScalars(c, "jobs.sbom-attestation.steps.Validate downloaded SBOM.env", mapping(validate, "env"), validatePolicy.env)
	}
	attest := stepByName(job, "Attest SBOM")
	attestPolicy := acceptedStepPolicy("sbom-attestation", "Attest SBOM")
	if attest == nil {
		c.fail("jobs.sbom-attestation.steps", "missing SBOM attestation step")
		return
	}
	if !strings.HasPrefix(scalar(mapping(attest, "uses")), attestPolicy.usesPrefix) {
		c.fail("jobs.sbom-attestation.steps.Attest SBOM.uses", "SBOM attestation must use actions/attest")
	}
	expectExactScalars(c, "jobs.sbom-attestation.steps.Attest SBOM.with", mapping(attest, "with"), attestPolicy.with)
	prepare := stepByName(job, "Prepare Sigstore bundle asset")
	preparePolicy := acceptedStepPolicy("sbom-attestation", "Prepare Sigstore bundle asset")
	if prepare == nil {
		c.fail("jobs.sbom-attestation.steps", "missing Sigstore bundle preparation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom-attestation.steps.Prepare Sigstore bundle asset.run", scalar(mapping(prepare, "run")), preparePolicy.runLines)
		expectExactScalars(c, "jobs.sbom-attestation.steps.Prepare Sigstore bundle asset.env", mapping(prepare, "env"), preparePolicy.env)
	}
	upload := stepByName(job, "Upload release assets artifact")
	uploadPolicy := acceptedStepPolicy("sbom-attestation", "Upload release assets artifact")
	if upload == nil {
		c.fail("jobs.sbom-attestation.steps", "missing release assets upload step")
	} else {
		if !strings.HasPrefix(scalar(mapping(upload, "uses")), uploadPolicy.usesPrefix) {
			c.fail("jobs.sbom-attestation.steps.Upload release assets artifact.uses", "release assets upload must use actions/upload-artifact")
		}
		expectExactScalars(c, "jobs.sbom-attestation.steps.Upload release assets artifact.with", mapping(upload, "with"), uploadPolicy.with)
	}
}

func (c *checker) checkReleaseJob(jobs *yaml.Node) {
	job := mapping(jobs, "release")
	if job == nil {
		return
	}
	download := stepByName(job, "Download prepared release assets")
	downloadPolicy := acceptedStepPolicy("release", "Download prepared release assets")
	if download == nil {
		c.fail("jobs.release.steps", "missing prepared release assets download step")
	} else {
		if !strings.HasPrefix(scalar(mapping(download, "uses")), downloadPolicy.usesPrefix) {
			c.fail("jobs.release.steps.Download prepared release assets.uses", "release asset download must use actions/download-artifact")
		}
		expectExactScalars(c, "jobs.release.steps.Download prepared release assets.with", mapping(download, "with"), downloadPolicy.with)
	}
	prepare := stepByName(job, "Prepare release notes and assets")
	preparePolicy := acceptedStepPolicy("release", "Prepare release notes and assets")
	if prepare == nil {
		c.fail("jobs.release.steps", "missing release preparation step")
	} else {
		run := scalar(mapping(prepare, "run"))
		requireExactScriptLines(c, "jobs.release.steps.Prepare release notes and assets.run", run, preparePolicy.runLines)
		expectExactScalars(c, "jobs.release.steps.Prepare release notes and assets.env", mapping(prepare, "env"), preparePolicy.env)
	}
	publish := stepByName(job, "Publish GitHub Release")
	publishPolicy := acceptedStepPolicy("release", "Publish GitHub Release")
	if publish == nil {
		c.fail("jobs.release.steps", "missing publishing step")
		return
	}
	if scalar(mapping(publish, "if")) != publishPolicy.ifCond {
		c.fail("jobs.release.steps.Publish GitHub Release.if", "publishing must be limited to v* tag pushes")
	}
	if scalar(mapping(publish, "shell")) != publishPolicy.shell {
		c.fail("jobs.release.steps.Publish GitHub Release.shell", "publishing must use bash")
	}
	run := scalar(mapping(publish, "run"))
	requireExactScriptLines(c, "jobs.release.steps.Publish GitHub Release.run", run, publishPolicy.runLines)
	expectExactScalars(c, "jobs.release.steps.Publish GitHub Release.env", mapping(publish, "env"), publishPolicy.env)
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

func expectScalars(c *checker, nodePath string, n *yaml.Node, want map[string]string) {
	if n == nil {
		c.fail(nodePath, "missing mapping")
		return
	}
	for key, expected := range want {
		if got := scalar(mapping(n, key)); got != expected {
			c.fail(nodePath+"."+key, fmt.Sprintf("got %q, want %q", got, expected))
		}
	}
}

func expectExactScalars(c *checker, nodePath string, n *yaml.Node, want map[string]string) {
	expectScalars(c, nodePath, n, want)
	if n == nil {
		return
	}
	for _, key := range mapKeys(n) {
		if _, ok := want[key]; !ok {
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

func stepByName(job *yaml.Node, name string) *yaml.Node {
	for _, step := range steps(job) {
		if scalar(mapping(step, "name")) == name {
			return step
		}
	}
	return nil
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
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
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
