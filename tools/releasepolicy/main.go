package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var shaRef = regexp.MustCompile(`^[0-9a-f]{40}$`)

// The checker intentionally snapshots the accepted ADR-0007 release workflow.
// Benign workflow edits may need lockstep updates here so release-critical
// shell and permission controls do not silently drift.
const (
	tagGuard            = "github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v')"
	unsupportedRefGuard = "github.event_name == 'workflow_dispatch' && !(" + tagGuard + ")"
	publishGuard        = "github.event_name == 'push' && startsWith(github.ref, 'refs/tags/v')"
	expectedSigners     = `the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFDxEpP8Q6LERBcA5//zwD5dBisHL7uHQsFa+TTibRXC
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFF32/OwUJwQ/8OX5i2VNBO8oZf6B8l07U/R5n1rj0z6
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILlg3QNI+Zsnt6pR2Aip97Ak7VOajBeo+AlhIGfDYlPk
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILvdes5QNqI3PpKK6ksX6FtlL4LQgkq61AGflWVqoV0L
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJaEbAxjr0LjcZKsqfUvrHDZJVmvL/AEIg+WSQGt+75v
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBc0CLdNLHpbdkrEf/WLR3YH8oHyxsvSeaCwQ6MvlW4q
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEd3JYo6vayWkANtsMbPx81ilaiq7a4oPpW6A0uD6TkF
`
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
	if root.Kind != yaml.MappingNode {
		c.fail("$", "workflow must be a mapping")
	}
	if !sameStringSet(mapKeys(root), []string{"name", "on", "permissions", "env", "concurrency", "jobs"}) {
		c.fail("$", "workflow root keys must exactly match the accepted release workflow")
	}
	if scalar(mapping(root, "name")) != "Release Validation" {
		c.fail("name", "workflow name must be Release Validation")
	}
	expectExactScalars(c, "env", mapping(root, "env"), map[string]string{
		"GOTOOLCHAIN": "local",
	})
	expectExactScalars(c, "concurrency", mapping(root, "concurrency"), map[string]string{
		"group":              "release-${{ github.ref }}",
		"cancel-in-progress": "false",
	})
}

func (c *checker) checkTriggers(root *yaml.Node) {
	on := mapping(root, "on")
	if on == nil {
		c.fail("on", "missing trigger block")
		return
	}
	if !sameStringSet(mapKeys(on), []string{"push", "workflow_dispatch"}) {
		c.fail("on", "trigger block must contain only push.tags v* and workflow_dispatch")
	}
	push := mapping(on, "push")
	if push == nil {
		c.fail("on.push", "missing push trigger")
	} else {
		if !sameStringSet(mapKeys(push), []string{"tags"}) {
			c.fail("on.push", "push trigger must contain only tags")
		}
		if !sameStringSet(seqStrings(mapping(push, "tags")), []string{"v*"}) {
			c.fail("on.push.tags", "push trigger must contain only v* tags")
		}
	}
	if mapping(on, "workflow_dispatch") == nil {
		c.fail("on.workflow_dispatch", "missing workflow_dispatch trigger")
	}
}

func (c *checker) checkTopPermissions(root *yaml.Node) {
	permissions := mapping(root, "permissions")
	if permissions == nil {
		c.fail("permissions", "missing top-level permissions")
		return
	}
	if scalar(mapping(permissions, "contents")) != "read" {
		c.fail("permissions.contents", "top-level contents permission must be read")
	}
	for _, key := range mapKeys(permissions) {
		if key != "contents" {
			c.fail("permissions."+key, "unexpected top-level permission")
		}
	}
}

func (c *checker) checkJobSet(jobs *yaml.Node) {
	want := []string{"unsupported-ref", "verify-tag", "check", "race", "vuln", "gosec", "sbom", "sbom-attestation", "release"}
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
	want := map[string][]string{
		"unsupported-ref": {
			"Explain unsupported ref",
		},
		"verify-tag": {
			"uses:actions/checkout",
			"Verify tag object and signature",
			"Validate release tag metadata",
		},
		"check": {
			"uses:actions/checkout",
			"Set up Go",
			"Report Go environment",
			"Run tests",
		},
		"race": {
			"uses:actions/checkout",
			"Set up Go",
			"Report Go environment",
			"Run race tests",
		},
		"vuln": {
			"uses:actions/checkout",
			"Set up Go",
			"Report Go environment",
			"Install task",
			"Install govulncheck",
			"Run vulnerability scan",
		},
		"gosec": {
			"uses:actions/checkout",
			"Set up Go",
			"Report Go environment",
			"Install task",
			"Install gosec",
			"Run gosec scan",
			"Upload gosec SARIF to code scanning",
			"Report gosec result",
		},
		"sbom": {
			"uses:actions/checkout",
			"Generate CycloneDX SBOM",
			"Validate SBOM and compute checksum",
			"Upload SBOM artifact",
		},
		"sbom-attestation": {
			"uses:actions/checkout",
			"Download SBOM artifact",
			"Validate downloaded SBOM",
			"Attest SBOM",
			"Prepare Sigstore bundle asset",
			"Upload release assets artifact",
		},
		"release": {
			"uses:actions/checkout",
			"Download prepared release assets",
			"Prepare release notes and assets",
			"Publish GitHub Release",
		},
	}
	for _, jobName := range mapKeys(jobs) {
		job := mapping(jobs, jobName)
		expected, ok := want[jobName]
		if !ok || job == nil {
			continue
		}
		var got []string
		for _, step := range steps(job) {
			got = append(got, stepIdentity(step))
		}
		if !sameStringSlice(got, expected) {
			c.fail("jobs."+jobName+".steps", fmt.Sprintf("steps must exactly match accepted release policy: got %q, want %q", got, expected))
		}
	}
}

func (c *checker) checkJobPermissions(jobs *yaml.Node) {
	want := map[string]map[string]string{
		"gosec": {
			"contents":        "read",
			"security-events": "write",
		},
		"sbom-attestation": {
			"contents":     "read",
			"id-token":     "write",
			"attestations": "write",
		},
		"release": {
			"contents": "write",
		},
	}
	for _, jobName := range []string{"unsupported-ref", "verify-tag", "check", "race", "vuln", "gosec", "sbom", "sbom-attestation", "release"} {
		job := mapping(jobs, jobName)
		if job == nil {
			continue
		}
		permissions := mapping(job, "permissions")
		expected, hasOverride := want[jobName]
		if !hasOverride {
			if permissions != nil {
				c.fail("jobs."+jobName+".permissions", "job must inherit top-level contents: read and must not declare permissions")
			}
			continue
		}
		expectExactScalars(c, "jobs."+jobName+".permissions", permissions, expected)
	}
}

func (c *checker) checkUnsupportedRefJob(jobs *yaml.Node) {
	job := mapping(jobs, "unsupported-ref")
	if job == nil {
		return
	}
	ifCond := scalar(mapping(job, "if"))
	if ifCond != unsupportedRefGuard {
		c.fail("jobs.unsupported-ref.if", "unsupported-ref job must be limited to non-v* workflow_dispatch refs")
	}
	steps := steps(job)
	if len(steps) != 1 {
		c.fail("jobs.unsupported-ref.steps", "unsupported-ref job should have one explanatory failing step")
		return
	}
	run := scalar(mapping(steps[0], "run"))
	requireExactScriptLines(c, "jobs.unsupported-ref.steps[0].run", run, []string{
		`echo "Release Validation workflow_dispatch runs must target a signed v* tag ref."`,
		`echo "Use the regular CI workflows for branch validation."`,
		`exit 1`,
	})
}

func (c *checker) checkVerifyTagJob(jobs *yaml.Node) {
	job := mapping(jobs, "verify-tag")
	if job == nil {
		return
	}
	if needs(job) != nil {
		c.fail("jobs.verify-tag.needs", "verify-tag must not depend on another job")
	}
	if scalar(mapping(job, "if")) != tagGuard {
		c.fail("jobs.verify-tag.if", "verify-tag must run only for signed v* tag refs")
	}
	outputs := mapping(job, "outputs")
	expectExactScalars(c, "jobs.verify-tag.outputs", outputs, map[string]string{
		"release-tag": "${{ steps.release-tag.outputs.release-tag }}",
		"sbom-file":   "${{ steps.release-tag.outputs.sbom-file }}",
		"prerelease":  "${{ steps.release-tag.outputs.prerelease }}",
		"latest":      "${{ steps.release-tag.outputs.latest }}",
	})
	verifyStep := stepByName(job, "Verify tag object and signature")
	if verifyStep == nil {
		c.fail("jobs.verify-tag.steps", "missing tag verification step")
		return
	}
	run := scalar(mapping(verifyStep, "run"))
	requireExactScriptLines(c, "jobs.verify-tag.steps.Verify tag object and signature.run", run, []string{
		`git fetch --force origin "refs/tags/$GITHUB_REF_NAME:refs/tags/$GITHUB_REF_NAME"`,
		`tag_type="$(git cat-file -t "$GITHUB_REF_NAME")"`,
		`printf '%s\n' "$tag_type"`,
		`test "$tag_type" = tag`,
		`git config gpg.ssh.allowedSignersFile .github/allowed_signers`,
		`git verify-tag "$GITHUB_REF_NAME"`,
	})
	metaStep := stepByName(job, "Validate release tag metadata")
	if metaStep == nil {
		c.fail("jobs.verify-tag.steps", "missing release tag metadata step")
	} else {
		requireExactScriptLines(c, "jobs.verify-tag.steps.Validate release tag metadata.run", scalar(mapping(metaStep, "run")), []string{
			`scripts/release-tag-metadata.sh "$GITHUB_REF_NAME" >> "$GITHUB_OUTPUT"`,
		})
	}
}

func (c *checker) checkNeeds(jobs *yaml.Node) {
	want := map[string][]string{
		"unsupported-ref":  nil,
		"verify-tag":       nil,
		"check":            {"verify-tag"},
		"race":             {"verify-tag"},
		"vuln":             {"verify-tag"},
		"gosec":            {"verify-tag"},
		"sbom":             {"verify-tag", "check", "race", "vuln", "gosec"},
		"sbom-attestation": {"sbom"},
		"release":          {"verify-tag", "sbom", "sbom-attestation"},
	}
	for jobName, expected := range want {
		job := mapping(jobs, jobName)
		if job == nil {
			continue
		}
		got := needs(job)
		if !sameStringSet(got, expected) {
			c.fail("jobs."+jobName+".needs", "needs must exactly match: "+strings.Join(expected, ", "))
		}
		if jobName != "unsupported-ref" && scalar(mapping(job, "if")) != tagGuard {
			c.fail("jobs."+jobName+".if", "job must run only for signed v* tag refs")
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
		for idx, step := range steps(job) {
			uses := scalar(mapping(step, "uses"))
			if !strings.HasPrefix(uses, "actions/checkout@") {
				continue
			}
			with := mapping(step, "with")
			want := map[string]string{"persist-credentials": "false"}
			if jobName == "verify-tag" {
				want["fetch-depth"] = "0"
			}
			expectExactScalars(c, fmt.Sprintf("jobs.%s.steps[%d].with", jobName, idx), with, want)
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
			requireExactScriptLines(c, "jobs.check.steps.Run tests.run", scalar(mapping(runTests, "run")), []string{`go test ./...`})
		}
	}
	race := mapping(jobs, "race")
	if race != nil {
		runRace := stepByName(race, "Run race tests")
		if runRace == nil {
			c.fail("jobs.race.steps", "missing Run race tests step")
		} else {
			requireExactScriptLines(c, "jobs.race.steps.Run race tests.run", scalar(mapping(runRace, "run")), []string{`go test -race ./...`})
		}
	}
	vuln := mapping(jobs, "vuln")
	if vuln != nil {
		requireRunStep(c, vuln, "jobs.vuln", "Install task", []string{`go install github.com/go-task/task/v3/cmd/task@v3.50.0`})
		requireRunStep(c, vuln, "jobs.vuln", "Install govulncheck", []string{`go install golang.org/x/vuln/cmd/govulncheck@v1.3.0`})
		requireRunStep(c, vuln, "jobs.vuln", "Run vulnerability scan", []string{`task vuln`})
	}
	gosec := mapping(jobs, "gosec")
	if gosec != nil {
		requireRunStep(c, gosec, "jobs.gosec", "Install task", []string{`go install github.com/go-task/task/v3/cmd/task@v3.50.0`})
		requireRunStep(c, gosec, "jobs.gosec", "Install gosec", []string{`go install github.com/securego/gosec/v2/cmd/gosec@v2.26.1`})
		scan := requireRunStep(c, gosec, "jobs.gosec", "Run gosec scan", []string{`task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'`})
		if scan != nil {
			if scalar(mapping(scan, "id")) != "gosec" {
				c.fail("jobs.gosec.steps.Run gosec scan.id", "gosec scan step id must be gosec")
			}
			if scalar(mapping(scan, "continue-on-error")) != "true" {
				c.fail("jobs.gosec.steps.Run gosec scan.continue-on-error", "gosec scan must continue so the report step can fail explicitly")
			}
		}
		upload := stepByName(gosec, "Upload gosec SARIF to code scanning")
		if upload == nil {
			c.fail("jobs.gosec.steps", "missing gosec SARIF upload step")
		} else {
			if scalar(mapping(upload, "if")) != "always() && hashFiles('gosec.sarif') != ''" {
				c.fail("jobs.gosec.steps.Upload gosec SARIF to code scanning.if", "gosec SARIF upload guard changed")
			}
			if !strings.HasPrefix(scalar(mapping(upload, "uses")), "github/codeql-action/upload-sarif@") {
				c.fail("jobs.gosec.steps.Upload gosec SARIF to code scanning.uses", "gosec SARIF upload must use github/codeql-action/upload-sarif")
			}
			expectExactScalars(c, "jobs.gosec.steps.Upload gosec SARIF to code scanning.with", mapping(upload, "with"), map[string]string{
				"sarif_file": "gosec.sarif",
				"category":   "gosec-release",
			})
		}
		report := requireRunStep(c, gosec, "jobs.gosec", "Report gosec result", []string{`exit 1`})
		if report != nil && scalar(mapping(report, "if")) != "steps.gosec.outcome == 'failure'" {
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

func (c *checker) checkSBOMJob(jobs *yaml.Node) {
	job := mapping(jobs, "sbom")
	if job == nil {
		return
	}
	outputs := mapping(job, "outputs")
	for _, name := range []string{"sbom-file", "sbom-sha256"} {
		if scalar(mapping(outputs, name)) == "" {
			c.fail("jobs.sbom.outputs."+name, "missing SBOM output")
		}
	}
	gen := stepByName(job, "Generate CycloneDX SBOM")
	if gen == nil {
		c.fail("jobs.sbom.steps", "missing SBOM generation step")
		return
	}
	if !strings.HasPrefix(scalar(mapping(gen, "uses")), "anchore/sbom-action@") {
		c.fail("jobs.sbom.steps.Generate CycloneDX SBOM.uses", "SBOM generation must use anchore/sbom-action")
	}
	with := mapping(gen, "with")
	expectExactScalars(c, "jobs.sbom.steps.Generate CycloneDX SBOM.with", with, map[string]string{
		"format":                "cyclonedx-json@1.5",
		"output-file":           "${{ needs.verify-tag.outputs.sbom-file }}",
		"syft-version":          "v1.45.1",
		"upload-artifact":       "false",
		"upload-release-assets": "false",
	})
	validate := stepByName(job, "Validate SBOM and compute checksum")
	if validate == nil {
		c.fail("jobs.sbom.steps", "missing SBOM validation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom.steps.Validate SBOM and compute checksum.run", scalar(mapping(validate, "run")), []string{
			`sbom_file="$SBOM_FILE"`,
			`scripts/validate-cyclonedx-sbom.sh "$sbom_file"`,
			`sbom_sha256="$(sha256sum "$sbom_file" | awk '{ print $1 }')"`,
			`{`,
			`echo "sbom-file=$sbom_file"`,
			`echo "sbom-sha256=$sbom_sha256"`,
			`} >> "$GITHUB_OUTPUT"`,
		})
		expectExactScalars(c, "jobs.sbom.steps.Validate SBOM and compute checksum.env", mapping(validate, "env"), map[string]string{
			"SBOM_FILE": "${{ needs.verify-tag.outputs.sbom-file }}",
		})
	}
	upload := stepByName(job, "Upload SBOM artifact")
	if upload == nil {
		c.fail("jobs.sbom.steps", "missing SBOM artifact upload step")
	} else {
		if !strings.HasPrefix(scalar(mapping(upload, "uses")), "actions/upload-artifact@") {
			c.fail("jobs.sbom.steps.Upload SBOM artifact.uses", "SBOM artifact upload must use actions/upload-artifact")
		}
		expectExactScalars(c, "jobs.sbom.steps.Upload SBOM artifact.with", mapping(upload, "with"), map[string]string{
			"name":              "release-sbom",
			"path":              "${{ needs.verify-tag.outputs.sbom-file }}",
			"if-no-files-found": "error",
		})
	}
}

func (c *checker) checkSBOMAttestationJob(jobs *yaml.Node) {
	job := mapping(jobs, "sbom-attestation")
	if job == nil {
		return
	}
	download := stepByName(job, "Download SBOM artifact")
	if download == nil {
		c.fail("jobs.sbom-attestation.steps", "missing SBOM artifact download step")
	} else {
		if !strings.HasPrefix(scalar(mapping(download, "uses")), "actions/download-artifact@") {
			c.fail("jobs.sbom-attestation.steps.Download SBOM artifact.uses", "SBOM download must use actions/download-artifact")
		}
		expectExactScalars(c, "jobs.sbom-attestation.steps.Download SBOM artifact.with", mapping(download, "with"), map[string]string{
			"name": "release-sbom",
			"path": "dist",
		})
	}
	validate := stepByName(job, "Validate downloaded SBOM")
	if validate == nil {
		c.fail("jobs.sbom-attestation.steps", "missing downloaded SBOM validation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom-attestation.steps.Validate downloaded SBOM.run", scalar(mapping(validate, "run")), []string{
			`scripts/validate-cyclonedx-sbom.sh "dist/$SBOM_FILE"`,
		})
		expectExactScalars(c, "jobs.sbom-attestation.steps.Validate downloaded SBOM.env", mapping(validate, "env"), map[string]string{
			"SBOM_FILE": "${{ needs.sbom.outputs.sbom-file }}",
		})
	}
	attest := stepByName(job, "Attest SBOM")
	if attest == nil {
		c.fail("jobs.sbom-attestation.steps", "missing SBOM attestation step")
		return
	}
	if !strings.HasPrefix(scalar(mapping(attest, "uses")), "actions/attest@") {
		c.fail("jobs.sbom-attestation.steps.Attest SBOM.uses", "SBOM attestation must use actions/attest")
	}
	with := mapping(attest, "with")
	expectExactScalars(c, "jobs.sbom-attestation.steps.Attest SBOM.with", with, map[string]string{
		"subject-path": "dist/${{ needs.sbom.outputs.sbom-file }}",
		"sbom-path":    "dist/${{ needs.sbom.outputs.sbom-file }}",
	})
	prepare := stepByName(job, "Prepare Sigstore bundle asset")
	if prepare == nil {
		c.fail("jobs.sbom-attestation.steps", "missing Sigstore bundle preparation step")
	} else {
		requireExactScriptLines(c, "jobs.sbom-attestation.steps.Prepare Sigstore bundle asset.run", scalar(mapping(prepare, "run")), []string{
			`bundle_dst="dist/${SBOM_FILE}.sigstore.json"`,
			`cp "$ATTESTATION_BUNDLE_PATH" "$bundle_dst"`,
			`test -s "$bundle_dst"`,
		})
		expectExactScalars(c, "jobs.sbom-attestation.steps.Prepare Sigstore bundle asset.env", mapping(prepare, "env"), map[string]string{
			"ATTESTATION_BUNDLE_PATH": "${{ steps.attest-sbom.outputs.bundle-path }}",
			"SBOM_FILE":               "${{ needs.sbom.outputs.sbom-file }}",
		})
	}
	upload := stepByName(job, "Upload release assets artifact")
	if upload == nil {
		c.fail("jobs.sbom-attestation.steps", "missing release assets upload step")
	} else {
		if !strings.HasPrefix(scalar(mapping(upload, "uses")), "actions/upload-artifact@") {
			c.fail("jobs.sbom-attestation.steps.Upload release assets artifact.uses", "release assets upload must use actions/upload-artifact")
		}
		expectExactScalars(c, "jobs.sbom-attestation.steps.Upload release assets artifact.with", mapping(upload, "with"), map[string]string{
			"name":              "release-assets",
			"path":              "dist/${{ needs.sbom.outputs.sbom-file }}\ndist/${{ needs.sbom.outputs.sbom-file }}.sigstore.json\n",
			"if-no-files-found": "error",
		})
	}
}

func (c *checker) checkReleaseJob(jobs *yaml.Node) {
	job := mapping(jobs, "release")
	if job == nil {
		return
	}
	download := stepByName(job, "Download prepared release assets")
	if download == nil {
		c.fail("jobs.release.steps", "missing prepared release assets download step")
	} else {
		if !strings.HasPrefix(scalar(mapping(download, "uses")), "actions/download-artifact@") {
			c.fail("jobs.release.steps.Download prepared release assets.uses", "release asset download must use actions/download-artifact")
		}
		expectExactScalars(c, "jobs.release.steps.Download prepared release assets.with", mapping(download, "with"), map[string]string{
			"name": "release-assets",
			"path": "dist",
		})
	}
	prepare := stepByName(job, "Prepare release notes and assets")
	if prepare == nil {
		c.fail("jobs.release.steps", "missing release preparation step")
	} else {
		run := scalar(mapping(prepare, "run"))
		requireExactScriptLines(c, "jobs.release.steps.Prepare release notes and assets.run", run, []string{
			`sbom_path="dist/$SBOM_FILE"`,
			`bundle_path="${sbom_path}.sigstore.json"`,
			`test -s "$sbom_path"`,
			`test -s "$bundle_path"`,
			`scripts/validate-cyclonedx-sbom.sh "$sbom_path"`,
			`computed_sha256="$(sha256sum "$sbom_path" | awk '{ print $1 }')"`,
			`test "$computed_sha256" = "$SBOM_SHA256"`,
			`scripts/extract-release-notes.sh CHANGELOG.md "$RELEASE_TAG" > release-notes.md`,
			`run_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"`,
			`{`,
			`cat release-notes.md`,
			`printf "\n## Supply-chain artifacts\n\n"`,
			`printf -- "- Release validation: %s\n" "$run_url"`,
			"printf -- \"- SBOM: \\`%s\\`\\n\" \"$SBOM_FILE\"",
			"printf -- \"- SBOM SHA-256 (corruption detection only): \\`%s\\`\\n\" \"$computed_sha256\"",
			"printf -- \"- SBOM attestation bundle: \\`%s.sigstore.json\\`\\n\" \"$SBOM_FILE\"",
			"printf -- \"- Verification instructions: \\`docs/release-verification.md\\`\\n\"",
			`} > release-body.md`,
			`echo "Prepared release body and assets for $RELEASE_TAG."`,
		})
		expectExactScalars(c, "jobs.release.steps.Prepare release notes and assets.env", mapping(prepare, "env"), map[string]string{
			"RELEASE_TAG": "${{ needs.verify-tag.outputs.release-tag }}",
			"SBOM_FILE":   "${{ needs.sbom.outputs.sbom-file }}",
			"SBOM_SHA256": "${{ needs.sbom.outputs.sbom-sha256 }}",
		})
	}
	publish := stepByName(job, "Publish GitHub Release")
	if publish == nil {
		c.fail("jobs.release.steps", "missing publishing step")
		return
	}
	if scalar(mapping(publish, "if")) != publishGuard {
		c.fail("jobs.release.steps.Publish GitHub Release.if", "publishing must be limited to v* tag pushes")
	}
	if scalar(mapping(publish, "shell")) != "bash" {
		c.fail("jobs.release.steps.Publish GitHub Release.shell", "publishing must use bash")
	}
	run := scalar(mapping(publish, "run"))
	requireExactScriptLines(c, "jobs.release.steps.Publish GitHub Release.run", run, []string{
		`set -euo pipefail`,
		`tag="$RELEASE_TAG"`,
		`sbom_path="dist/$SBOM_FILE"`,
		`bundle_path="${sbom_path}.sigstore.json"`,
		`release_args=()`,
		`if [ "$RELEASE_PRERELEASE" = "true" ]; then`,
		`release_args+=(--prerelease)`,
		`fi`,
		`if [ "$RELEASE_LATEST" = "false" ]; then`,
		`release_args+=(--latest=false)`,
		`fi`,
		`if gh release view "$tag" --repo "$GITHUB_REPOSITORY" >/dev/null 2>&1; then`,
		`echo "GitHub Release already exists for $tag; refusing automated in-place asset replacement." >&2`,
		`echo "Delete the draft/release or repair it manually before rerunning Release Validation." >&2`,
		`exit 1`,
		`fi`,
		`gh release create "$tag" "$sbom_path" "$bundle_path" \`,
		`--repo "$GITHUB_REPOSITORY" \`,
		`--title "$tag" \`,
		`--notes-file release-body.md \`,
		`--verify-tag \`,
		`"${release_args[@]}"`,
	})
	expectExactScalars(c, "jobs.release.steps.Publish GitHub Release.env", mapping(publish, "env"), map[string]string{
		"GH_TOKEN":           "${{ github.token }}",
		"RELEASE_TAG":        "${{ needs.verify-tag.outputs.release-tag }}",
		"RELEASE_PRERELEASE": "${{ needs.verify-tag.outputs.prerelease }}",
		"RELEASE_LATEST":     "${{ needs.verify-tag.outputs.latest }}",
		"SBOM_FILE":          "${{ needs.sbom.outputs.sbom-file }}",
	})
}

func checkAllowedSigners(repoRoot string) []finding {
	path := filepath.Join(repoRoot, ".github", "allowed_signers")
	in, err := os.ReadFile(path)
	if err != nil {
		return []finding{{path: path, msg: err.Error()}}
	}
	if strings.ReplaceAll(string(in), "\r\n", "\n") != expectedSigners {
		return []finding{{path: path, msg: "allowed_signers must exactly match the accepted maintainer signing keys"}}
	}
	return nil
}

func checkRequiredScripts(repoRoot string) []finding {
	var findings []finding
	for _, path := range []string{
		"scripts/release-tag-metadata.sh",
		"scripts/validate-cyclonedx-sbom.sh",
		"scripts/extract-release-notes.sh",
	} {
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
