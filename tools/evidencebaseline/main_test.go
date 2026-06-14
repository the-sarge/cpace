package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

func TestSummaryDocClassifierRefsMatchBaselineParser(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	baselinePath := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	if _, err := os.Stat(baselinePath); err != nil {
		if os.IsNotExist(err) {
			t.Skip("repository evidence baseline is not present")
		}
		t.Fatal(err)
	}

	rows, findings, err := parseBaselineIndex(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("repository Baseline Index parser findings: %#v", findings)
	}

	wantSet := map[string]bool{}
	for _, row := range rows {
		refs, refFindings := summaryRefs(row.summaryCell, baselinePath+":"+row.lane)
		if len(refFindings) > 0 {
			t.Fatalf("repository summary ref findings: %#v", refFindings)
		}
		for _, ref := range refs {
			wantSet[ref] = true
		}
	}
	want := sortedKeys(wantSet)

	cmd := exec.Command(filepath.Join(repoRoot, "scripts", "classify-check-changes.sh"), "--list-summary-docs")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	got := outputLines(out)
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("summary doc parser drift:\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
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

func TestEvidenceBaselineRejectsMissingSummaryDocsManifest(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	remove(t, filepath.Join(repoRoot, summaryDocsManifestRef))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "summary-doc manifest is missing")
}

func TestEvidenceBaselineRejectsStaleSummaryDocsManifest(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, summaryDocsManifestRef), "docs/dependency-review.md\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "summary-doc manifest got")
	requireFinding(t, findings, "docs/fuzz-evidence.md")
}

func TestEvidenceBaselineRejectsSummaryDocsManifestWhitespace(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, summaryDocsManifestRef), strings.Join([]string{
		"# Generated from docs/evidence-baseline.md by tools/evidencebaseline --write-summary-docs.",
		" docs/dependency-review.md ",
		"docs/fuzz-evidence.md",
		"",
	}, "\n"))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, summaryDocsManifestRef+":2")
	requireFinding(t, findings, "leading or trailing whitespace")
}

func TestEvidenceBaselineWriteSummaryDocsFlagRoundTrip(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	writeFile(t, filepath.Join(repoRoot, summaryDocsManifestRef), "stale\n")

	cmd := exec.Command("go", "run", ".", "--repo-root", repoRoot, "--write-summary-docs")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run --write-summary-docs failed: %v\n%s", err, out)
	}

	content, err := os.ReadFile(filepath.Join(repoRoot, summaryDocsManifestRef))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != fixtureSummaryDocsManifest() {
		t.Fatalf("summary-doc manifest content:\ngot:\n%s\nwant:\n%s", content, fixtureSummaryDocsManifest())
	}
	findings := checkSummaryDocsManifest(repoRoot, []string{"docs/dependency-review.md", "docs/fuzz-evidence.md"})
	if len(findings) > 0 {
		t.Fatalf("expected regenerated manifest to pass, got %#v", findings)
	}
}

func TestEvidenceBaselineRejectsMutuallyExclusiveSummaryDocFlags(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--list-summary-docs", "--write-summary-docs")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected mutually exclusive flags to fail, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got:\n%s", out)
	}
}

func TestEvidenceBaselineRejectsSymlinkedSummaryDoc(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside-summary.md")
	writeFile(t, outside, "# Outside Summary\n")
	summary := filepath.Join(repoRoot, "docs", "dependency-review.md")
	remove(t, summary)
	if err := os.Symlink(outside, summary); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "referenced summary doc is a symlink")
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

func TestEvidenceBaselineRejectsChecksumEntryUnderSymlinkedParentBeforeHashing(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside")
	writeFile(t, filepath.Join(outside, "secret.log"), "external secret\n")
	link := filepath.Join(repoRoot, "docs", "evidence", "candidate", "nested")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	appendSHA256SUMS(t, filepath.Join(repoRoot, "docs", "evidence", "candidate"), "nested/secret.log", "different content\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "symlinked parent")
	rejectFinding(t, findings, "hash mismatch")
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

func TestEvidenceBaselineRejectsSymlinkedEvidenceDirectory(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside-evidence", "candidate")
	writeFile(t, filepath.Join(outside, "README.md"), "# Outside Evidence\n")
	writeFile(t, filepath.Join(outside, "analysis.log"), "analysis\n")
	writeFile(t, filepath.Join(outside, "fuzz.log"), "fuzz\n")
	writeSHA256SUMS(t, outside, "analysis.log", "fuzz.log")

	evidenceDir := filepath.Join(repoRoot, "docs", "evidence")
	removeAll(t, evidenceDir)
	if err := os.Symlink(filepath.Join(repoRoot, "outside-evidence"), evidenceDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "docs/evidence")
	requireFinding(t, findings, "symlink")
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

func TestEvidenceBaselineRejectsSymlinkedChecksumSignature(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	outside := filepath.Join(repoRoot, "outside-SHA256SUMS.sig")
	writeFile(t, outside, "outside\n")
	link := filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS.sig")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "SHA256SUMS.sig")
	requireFinding(t, findings, "control file must not be a symlink")
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

func TestEvidenceBaselineAllowsNonRepoBacktickURLContainingDocsPath(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	baseline := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	replaceInFile(t, baseline, "`docs/evidence/candidate/analysis.log`", "`docs/evidence/candidate/analysis.log`, `https://example.invalid/archive/docs/run`")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("expected URL token to be ignored, got %#v", findings)
	}
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

func TestEvidenceBaselineParserRejectsMissingSeparatorWithoutSkippingFirstRow(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	baseline := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	remove(t, filepath.Join(repoRoot, "docs", "dependency-review.md"))
	replaceInFile(t, baseline, "| --- | --- | --- | --- | --- |\n", "")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "Baseline Index separator")
	requireFinding(t, findings, "referenced summary doc does not exist")
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
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), strings.Repeat("z", 64)+"  analysis.log\n")

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "checksum hash must be 64 lowercase hex characters")
}

func TestEvidenceBaselineRejectsBinaryModeChecksumEntry(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	sum := sha256.Sum256([]byte("analysis\n"))
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), fmt.Sprintf("%x *analysis.log\n", sum))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "text-mode SHA256SUMS format")
}

func TestEvidenceBaselineRejectsChecksumPathWithSpaces(t *testing.T) {
	repoRoot := validFixtureRepo(t)
	sum := sha256.Sum256([]byte("analysis\n"))
	writeFile(t, filepath.Join(repoRoot, "docs", "evidence", "candidate", "SHA256SUMS"), fmt.Sprintf("%x  analysis log\n", sum))

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "checksum path must not contain whitespace")
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
	writeFile(t, filepath.Join(repoRoot, summaryDocsManifestRef), fixtureSummaryDocsManifest())
	return repoRoot
}

func fixtureSummaryDocsManifest() string {
	return strings.Join([]string{
		"# Generated from docs/evidence-baseline.md by tools/evidencebaseline --write-summary-docs.",
		"docs/dependency-review.md",
		"docs/fuzz-evidence.md",
		"",
	}, "\n")
}

func outputLines(out []byte) []string {
	text := strings.TrimSuffix(string(out), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
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

func rejectFinding(t *testing.T, findings []finding, unwanted string) {
	t.Helper()
	for _, finding := range findings {
		if strings.Contains(finding.path, unwanted) || strings.Contains(finding.msg, unwanted) {
			t.Fatalf("unexpected finding containing %q; got %#v", unwanted, findings)
		}
	}
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
