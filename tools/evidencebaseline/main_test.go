package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentRepositoryEvidenceBaseline(t *testing.T) {
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

func TestEvidenceBaselineAcceptsValidFixture(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("expected clean fixture, got %#v", findings)
	}
}

func TestEvidenceBaselineRejectsMissingSummaryDoc(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	remove(t, filepath.Join(repoRoot, "docs", "dependency-review.md"))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "referenced summary doc does not exist")
}

func TestEvidenceBaselineRejectsMissingRawArtifact(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	remove(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "analysis.log"))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "referenced raw artifact does not exist")
}

func TestEvidenceBaselineRejectsMissingBundleFiles(t *testing.T) {
	tests := []struct {
		name string
		file string
		want string
	}{
		{"readme", "README.md", "missing README.md"},
		{"checksums", "SHA256SUMS", "missing SHA256SUMS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := validFixtureRepo(t)
			remove(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", tt.file))

			findings, err := checkRepo(repoRoot)
			if err != nil {
				t.Fatal(err)
			}
			requireFinding(t, findings, tt.want)
		})
	}
}

func TestEvidenceBaselineChecksUnreferencedBundles(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "historical", "README.md"), "# Historical\n")
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "historical", "old.log"), "old\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "missing SHA256SUMS")
}

func TestEvidenceBaselineRejectsBadChecksum(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "analysis.log"), "changed\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "hash mismatch")
}

func TestEvidenceBaselineRejectsUncoveredRawFile(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "uncovered.log"), "not covered\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "not covered by SHA256SUMS")
}

func TestEvidenceBaselineRejectsUnsafeChecksumPath(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	appendFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  ../outside.log\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "safe bundle-relative path")
}

func validFixtureRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "docs", "dependency-review.md"), "# Dependency Review\n")
	writeFile(t, filepath.Join(repoRoot, "docs", "fuzz-evidence.md"), "# Fuzz Evidence\n")
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "README.md"), "# Candidate Evidence\n")
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "analysis.log"), "analysis\n")
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "fuzz.log"), "fuzz\n")
	writeSHA256SUMS(t, filepath.Join(repoRoot, "docs", "evidence", "candidate"), "analysis.log", "fuzz.log")
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence-baseline.md"), strings.Join([]string{
		"# Evidence Baseline",
		"",
		"## Baseline Index",
		"",
		"| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |",
		"| --- | --- | --- | --- | --- |",
		"| Dependency review | `abc123` | `docs/evidence/candidate/analysis.log` | `docs/dependency-review.md` | Repeat on code change. |",
		"| Fuzzing | `abc123` | `docs/evidence/candidate/fuzz.log`, `docs/evidence/candidate/` | `docs/fuzz-evidence.md` | Repeat on parser change. |",
		"",
		"## Refresh Procedure",
		"",
		"Keep this short in fixtures.",
		"",
	}, "\n"))
	return repoRoot
}

func writeSHA256SUMS(t *testing.T, dir string, files ...string) {
	t.Helper()
	var lines []string
	for _, name := range files {
		in, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(in)
		lines = append(lines, fmt.Sprintf("%x  %s", sum, name))
	}
	writeFile(t, filepath.Join(dir, "SHA256SUMS"), strings.Join(lines, "\n")+"\n")
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendFile(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
