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
	repoRoot := mustFixtureRepo(t)
	if err := os.Remove(filepath.Join(repoRoot, "docs/dependency-review.md")); err != nil {
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
