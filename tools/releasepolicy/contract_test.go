package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestFindingPathsAreRepositoryRelative(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteReleasePolicyRepoFixture(t, repoRoot)
	if err := os.Remove(filepath.Join(repoRoot, "scripts/release-metadata.sh")); err != nil {
		t.Fatal(err)
	}
	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("missing regression finding")
	}
	for _, f := range findings {
		if strings.Contains(f.path, repoRoot) || strings.HasPrefix(f.path, "../") || filepath.IsAbs(f.path) {
			t.Fatalf("finding path got %q want repository-relative location", f.path)
		}
	}
}

func TestAcceptedReleasePolicyConstructionIsIndependent(t *testing.T) {
	workflow := mustReadCurrentWorkflow(t)
	policy := newAcceptedReleasePolicy()
	policy.env["GOTOOLCHAIN"] = "auto"
	policy.rootKeys[0] = "changed"
	policy.jobs[1].steps[0].with["persist-credentials"] = "true"
	if got := mustFindingsForWorkflowAgainstPolicy(t, workflow, policy); len(got) == 0 {
		t.Fatal("mutated policy did not reject the accepted workflow")
	}
	if got := mustFindingsForWorkflowAgainstPolicy(t, workflow, newAcceptedReleasePolicy()); len(got) != 0 {
		t.Fatalf("fresh policy findings got %#v want none", got)
	}
}

func TestReleaseStepIdentityDerivation(t *testing.T) {
	for _, tt := range []struct{ name, uses, want string }{
		{"Named action", "actions/checkout@", "Named action"},
		{"Run tests", "", "Run tests"},
		{"", "actions/checkout@", "uses:actions/checkout"},
		{"", "", "unnamed"},
	} {
		if got := (releaseJobPolicy{steps: []releaseStepPolicy{{name: tt.name, usesPrefix: tt.uses}}}).stepIdentities(); len(got) != 1 || got[0] != tt.want {
			t.Fatalf("step identity got %q want [%q]", got, tt.want)
		}
	}
}

func TestRequiredFileKinds(t *testing.T) {
	for _, tt := range []struct{ path, want string }{
		{"scripts/release-metadata.sh", "required release helper must be a regular file"},
		{".github/syft-release.yaml", "expected file, got directory"},
	} {
		t.Run(tt.path, func(t *testing.T) {
			root := t.TempDir()
			mustWriteReleasePolicyRepoFixture(t, root)
			full := filepath.Join(root, tt.path)
			if err := os.Remove(full); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(full, 0o755); err != nil {
				t.Fatal(err)
			}
			findings, err := checkRepo(root)
			if err != nil {
				t.Fatal(err)
			}
			assertFinding(t, findings, tt.path, tt.want)
		})
	}
}

func TestFindingSortOrder(t *testing.T) {
	findings := []finding{{path: "b", msg: "a"}, {path: "a", msg: "z"}, {path: "a", msg: "a"}}
	sortFindings(findings)
	want := []finding{{path: "a", msg: "a"}, {path: "a", msg: "z"}, {path: "b", msg: "a"}}
	if !slices.Equal(findings, want) {
		t.Fatalf("sorted findings got %#v want %#v", findings, want)
	}
}

func TestAssertFindingPairsPathAndMessage(t *testing.T) {
	if scenario := os.Getenv("CPACE_FINDING_ASSERT_SCENARIO"); scenario != "" {
		findings := []finding{{path: "expected", msg: "renamed"}, {path: "elsewhere", msg: "original"}}
		wantMsg := "original"
		if scenario == "path contains message" {
			findings = []finding{{path: "expected", msg: "renamed"}}
			wantMsg = "expected"
		}
		assertFinding(t, findings, "expected", wantMsg)
		return
	}
	for _, scenario := range []string{"separate findings", "path contains message"} {
		t.Run(scenario, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestAssertFindingPairsPathAndMessage$")
			cmd.Env = append(os.Environ(), "CPACE_FINDING_ASSERT_SCENARIO="+scenario)
			output, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "want path") {
				t.Fatalf("mismatched assertion got error %v output %s want assertion failure", err, output)
			}
		})
	}
	assertFinding(t, []finding{{path: "expected", msg: "original message"}}, "expected", "original")
}
