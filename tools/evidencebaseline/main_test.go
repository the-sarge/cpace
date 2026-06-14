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
	if _, err := os.Stat(filepath.Join(repoRoot, "docs", "evidence-baseline.md")); err != nil {
		if os.IsNotExist(err) {
			t.Skip("repository evidence baseline is not present")
		}
		t.Fatal(err)
	}
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

func TestEvidenceBaselineRejectsNestedUncoveredRawFile(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "nested", "uncovered.log"), "not covered\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "nested/uncovered.log")
	requireFinding(t, findings, "not covered by SHA256SUMS")
}

func TestEvidenceBaselineRejectsSymlinkedChecksumEntry(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside.log")
	writeFile(t, outside, "outside\n")
	link := filepath.Join(repoRoot, "docs", "evidence", "candidate", "linked.log")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	appendSHA256SUMS(t, filepath.Join(repoRoot, "docs", "evidence", "candidate"), "linked.log", "outside\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "checksum references symlink")
}

func TestEvidenceBaselineRejectsSymlinkedBundleRoot(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside-evidence")
	writeFile(t, filepath.Join(outside, "README.md"), "# Outside Evidence\n")
	writeFile(t, filepath.Join(outside, "analysis.log"), "analysis\n")
	writeFile(t, filepath.Join(outside, "fuzz.log"), "fuzz\n")
	writeSHA256SUMS(t, outside, "analysis.log", "fuzz.log")

	bundle := filepath.Join(repoRoot, "docs", "evidence", "candidate")
	removeAll(t, bundle)
	if err := os.Symlink(outside, bundle); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "evidence bundle root must not be a symlink")
}

func TestEvidenceBaselineRejectsSymlinkedControlFiles(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"readme", "README.md"},
		{"checksums", "SHA256SUMS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := validFixtureRepo(t)
			outside := filepath.Join(repoRoot, "outside-"+tt.file)
			writeFile(t, outside, "outside\n")
			link := filepath.Join(repoRoot, "docs", "evidence", "candidate", tt.file)
			remove(t, link)
			if err := os.Symlink(outside, link); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}

			findings, err := checkRepo(repoRoot)
			if err != nil {
				t.Fatal(err)
			}
			requireFinding(t, findings, tt.file)
			requireFinding(t, findings, "control file must not be a symlink")
		})
	}
}

func TestEvidenceBaselineRejectsUnsafeBaselineRef(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	baseline := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	content, err := os.ReadFile(baseline)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(content), "docs/evidence/candidate/analysis.log", "docs/evidence/candidate/../outside.log", 1)
	writeFile(t, baseline, updated)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "unsafe baseline ref")
}

func TestEvidenceBaselineParserRejectsMalformedBaselineHeader(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	baseline := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	replaceInFile(t, baseline, "| Evidence lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |", "| Lane | Pinned baseline | Raw artifacts | Summary docs | Freshness rule |")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "Baseline Index header got")
}

func TestEvidenceBaselineParserRejectsDuplicateLane(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	baseline := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	replaceInFile(t, baseline, "| Fuzzing | `abc123` |", "| Dependency review | `abc123` |")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "duplicate evidence lane")
}

func TestEvidenceBaselineParserRejectsSummaryDocDirectory(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	summary := filepath.Join(repoRoot, "docs", "dependency-review.md")
	remove(t, summary)
	if err := os.Mkdir(summary, 0o755); err != nil {
		t.Fatal(err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "referenced summary doc is a directory")
}

func TestEvidenceBaselineIgnoresLocalDSStore(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", ".DS_Store"), "local metadata\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("expected .DS_Store to be ignored, got %#v", findings)
	}
}

func TestEvidenceBaselineRejectsHiddenRawFile(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", ".hidden.log"), "hidden evidence\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, ".hidden.log")
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

func TestEvidenceBaselineRejectsMalformedChecksumHash(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), "zzzz  analysis.log\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "checksum hash must be 64 lowercase hex characters")
}

func TestEvidenceBaselineRejectsDuplicateChecksumPath(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	appendSHA256SUMS(t, filepath.Join(repoRoot, "docs", "evidence", "candidate"), "analysis.log", "analysis\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "duplicate checksum path")
}

func TestEvidenceBaselineRejectsEmptyChecksumFile(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), "\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "contains no checksum entries")
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

func appendSHA256SUMS(t *testing.T, dir, name, content string) {
	t.Helper()
	sum := sha256.Sum256([]byte(content))
	appendFile(t, filepath.Join(dir, "SHA256SUMS"), fmt.Sprintf("%x  %s\n", sum, name))
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

func replaceInFile(t *testing.T, path, old, new string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(content), old, new, 1)
	if updated == string(content) {
		t.Fatalf("did not find %q in %s", old, path)
	}
	writeFile(t, path, updated)
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func removeAll(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
}
