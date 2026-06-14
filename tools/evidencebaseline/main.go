package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

type finding struct {
	path string
	msg  string
}

type baselineRow struct {
	lane        string
	rawCell     string
	summaryCell string
}

var (
	backtickRef = regexp.MustCompile("`([^`]+)`")
	sha256Hex   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

const summaryDocsManifestRef = "docs/evidence-baseline-summary-docs.txt"

func main() {
	repoRoot := flag.String("repo-root", "../..", "repository root")
	listSummaryDocs := flag.Bool("list-summary-docs", false, "list summary docs parsed from the evidence baseline")
	writeSummaryDocs := flag.Bool("write-summary-docs", false, "rewrite the summary-doc manifest from the evidence baseline")
	flag.Parse()

	if *listSummaryDocs && *writeSummaryDocs {
		fmt.Fprintln(os.Stderr, "evidence baseline checker failed: --list-summary-docs and --write-summary-docs are mutually exclusive")
		os.Exit(2)
	}
	if *listSummaryDocs || *writeSummaryDocs {
		refs, findings, err := summaryDocsFromBaseline(*repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "evidence baseline checker failed: %v\n", err)
			os.Exit(2)
		}
		if len(findings) > 0 {
			sortFindings(findings)
			for _, f := range findings {
				fmt.Fprintf(os.Stderr, "%s: %s\n", f.path, f.msg)
			}
			os.Exit(1)
		}
		if *writeSummaryDocs {
			if err := writeSummaryDocsManifest(*repoRoot, refs); err != nil {
				fmt.Fprintf(os.Stderr, "evidence baseline checker failed: %v\n", err)
				os.Exit(2)
			}
			return
		}
		for _, ref := range refs {
			fmt.Println(ref)
		}
		return
	}

	findings, err := checkRepo(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evidence baseline checker failed: %v\n", err)
		os.Exit(2)
	}
	if len(findings) > 0 {
		sortFindings(findings)
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "%s: %s\n", f.path, f.msg)
		}
		os.Exit(1)
	}
	fmt.Println("evidence baseline checker passed")
}

func checkRepo(repoRoot string) ([]finding, error) {
	baselinePath := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	if _, info, pathFindings, err := lstatRepoRelative(repoRoot, "docs/evidence-baseline.md", "docs/evidence-baseline.md"); err != nil {
		return nil, err
	} else if len(pathFindings) > 0 {
		return pathFindings, nil
	} else if info.Mode()&fs.ModeSymlink != 0 {
		return []finding{{path: "docs/evidence-baseline.md", msg: "evidence baseline file must not be a symlink"}}, nil
	} else if !info.Mode().IsRegular() {
		return []finding{{path: "docs/evidence-baseline.md", msg: "evidence baseline file must be a regular file"}}, nil
	}

	rows, findings, err := parseBaselineIndex(baselinePath)
	if err != nil {
		return nil, err
	}

	referencedBundles := map[string]bool{}
	summaryDocSet := map[string]bool{}
	for _, row := range rows {
		rawRefs, rawFindings := evidenceRefs(row.rawCell, baselinePath+":"+row.lane)
		findings = append(findings, rawFindings...)
		for _, ref := range rawRefs {
			_, info, pathFindings, err := lstatRepoRelative(repoRoot, ref, baselinePath+":"+row.lane)
			if len(pathFindings) > 0 {
				findings = append(findings, pathFindings...)
				continue
			}
			switch {
			case errors.Is(err, fs.ErrNotExist):
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced raw artifact does not exist: " + ref})
				continue
			case err != nil:
				return nil, err
			case info.Mode()&fs.ModeSymlink != 0:
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced raw artifact is a symlink: " + ref})
			}
			bundle := evidenceBundleForRef(ref, info)
			if bundle != "" {
				referencedBundles[bundle] = true
			}
		}

		docRefs, docFindings := summaryRefs(row.summaryCell, baselinePath+":"+row.lane)
		findings = append(findings, docFindings...)
		for _, ref := range docRefs {
			summaryDocSet[ref] = true
			_, info, pathFindings, err := lstatRepoRelative(repoRoot, ref, baselinePath+":"+row.lane)
			if len(pathFindings) > 0 {
				findings = append(findings, pathFindings...)
				continue
			}
			switch {
			case errors.Is(err, fs.ErrNotExist):
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced summary doc does not exist: " + ref})
			case err != nil:
				return nil, err
			case info.Mode()&fs.ModeSymlink != 0:
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced summary doc is a symlink: " + ref})
			case info.IsDir():
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced summary doc is a directory: " + ref})
			}
		}
	}
	findings = append(findings, checkSummaryDocsManifest(repoRoot, sortedKeys(summaryDocSet))...)

	allBundles, bundleFindings, err := discoverEvidenceBundles(repoRoot)
	if err != nil {
		return nil, err
	}
	findings = append(findings, bundleFindings...)
	for _, bundle := range allBundles {
		referencedBundles[bundle] = true
	}

	for bundle := range referencedBundles {
		findings = append(findings, checkBundle(repoRoot, bundle)...)
	}
	sortFindings(findings)
	return findings, nil
}

func summaryDocsFromBaseline(repoRoot string) ([]string, []finding, error) {
	baselinePath := filepath.Join(repoRoot, "docs", "evidence-baseline.md")
	rows, findings, err := parseBaselineIndex(baselinePath)
	if err != nil {
		return nil, nil, err
	}
	summaryDocSet := map[string]bool{}
	for _, row := range rows {
		refs, refFindings := summaryRefs(row.summaryCell, baselinePath+":"+row.lane)
		findings = append(findings, refFindings...)
		for _, ref := range refs {
			summaryDocSet[ref] = true
		}
	}
	return sortedKeys(summaryDocSet), findings, nil
}

func checkSummaryDocsManifest(repoRoot string, want []string) []finding {
	got, findings, err := readSummaryDocsManifest(repoRoot)
	if len(findings) > 0 {
		return findings
	}
	if errors.Is(err, fs.ErrNotExist) {
		return []finding{{path: summaryDocsManifestRef, msg: "summary-doc manifest is missing; run (cd tools/evidencebaseline && go run . --repo-root ../.. --write-summary-docs)"}}
	}
	if err != nil {
		return []finding{{path: summaryDocsManifestRef, msg: err.Error()}}
	}
	if !sameStringSlice(got, want) {
		return []finding{{path: summaryDocsManifestRef, msg: fmt.Sprintf("summary-doc manifest got %q, want %q; run (cd tools/evidencebaseline && go run . --repo-root ../.. --write-summary-docs)", got, want)}}
	}
	return nil
}

func readSummaryDocsManifest(repoRoot string) ([]string, []finding, error) {
	full, info, pathFindings, err := lstatRepoRelative(repoRoot, summaryDocsManifestRef, summaryDocsManifestRef)
	if len(pathFindings) > 0 {
		return nil, pathFindings, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return nil, []finding{{path: summaryDocsManifestRef, msg: "summary-doc manifest must not be a symlink"}}, nil
	}
	if !info.Mode().IsRegular() {
		return nil, []finding{{path: summaryDocsManifestRef, msg: "summary-doc manifest must be a regular file"}}, nil
	}
	in, err := os.ReadFile(full)
	if err != nil {
		return nil, nil, err
	}
	return parseSummaryDocsManifest(string(in))
}

func parseSummaryDocsManifest(content string) ([]string, []finding, error) {
	var refs []string
	var findings []finding
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		ref := scanner.Text()
		if ref == "" || strings.HasPrefix(ref, "#") {
			continue
		}
		if strings.TrimSpace(ref) != ref {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", summaryDocsManifestRef, lineNo), msg: "summary-doc manifest entries must not have leading or trailing whitespace"})
			continue
		}
		refs = append(refs, ref)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return refs, findings, nil
}

func writeSummaryDocsManifest(repoRoot string, refs []string) error {
	content := "# Generated from docs/evidence-baseline.md by tools/evidencebaseline --write-summary-docs.\n"
	for _, ref := range refs {
		content += ref + "\n"
	}
	path := filepath.Join(repoRoot, summaryDocsManifestRef)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func parseBaselineIndex(path string) ([]baselineRow, []finding, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(in)))
	var inSection bool
	var tableLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			if inSection {
				break
			}
			inSection = strings.TrimSpace(line) == "## Baseline Index"
			continue
		}
		if inSection && strings.HasPrefix(strings.TrimSpace(line), "|") {
			tableLines = append(tableLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if len(tableLines) == 0 {
		return nil, []finding{{path: path, msg: "Baseline Index table is missing or empty"}}, nil
	}

	header := splitTableRow(tableLines[0])
	wantHeader := []string{"Evidence lane", "Pinned baseline", "Raw artifacts", "Summary docs", "Freshness rule"}
	if !sameStringSlice(header, wantHeader) {
		return nil, []finding{{path: path, msg: fmt.Sprintf("Baseline Index header got %q, want %q", header, wantHeader)}}, nil
	}

	var rows []baselineRow
	var findings []finding
	seenLane := map[string]bool{}
	dataStart := 2
	if len(tableLines) < 2 {
		findings = append(findings, finding{path: path, msg: "Baseline Index separator is missing"})
		dataStart = 1
	} else {
		separator := splitTableRow(tableLines[1])
		if !validTableSeparator(separator, len(wantHeader)) {
			findings = append(findings, finding{path: path, msg: fmt.Sprintf("Baseline Index separator got %q, want %d Markdown separator columns", separator, len(wantHeader))})
			dataStart = 1
		}
	}
	for _, line := range tableLines[dataStart:] {
		cells := splitTableRow(line)
		if len(cells) != len(wantHeader) {
			findings = append(findings, finding{path: path, msg: "Baseline Index row has wrong column count: " + line})
			continue
		}
		lane := cells[0]
		if lane == "" {
			findings = append(findings, finding{path: path, msg: "Baseline Index row has empty evidence lane"})
			continue
		}
		if seenLane[lane] {
			findings = append(findings, finding{path: path + ":" + lane, msg: "duplicate evidence lane"})
		}
		seenLane[lane] = true
		rows = append(rows, baselineRow{
			lane:        lane,
			rawCell:     cells[2],
			summaryCell: cells[3],
		})
	}
	if len(rows) == 0 {
		findings = append(findings, finding{path: path, msg: "Baseline Index contains no evidence rows"})
	}
	return rows, findings, nil
}

func validTableSeparator(cells []string, wantCols int) bool {
	if len(cells) != wantCols {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		cell = strings.TrimPrefix(cell, ":")
		cell = strings.TrimSuffix(cell, ":")
		if len(cell) < 3 {
			return false
		}
		for _, r := range cell {
			if r != '-' {
				return false
			}
		}
	}
	return true
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimSpace(part))
	}
	return out
}

func evidenceRefs(cell, findingPath string) ([]string, []finding) {
	return filteredBacktickRefs(cell, findingPath, func(ref string) bool {
		return strings.HasPrefix(ref, "docs/evidence/")
	})
}

func summaryRefs(cell, findingPath string) ([]string, []finding) {
	return filteredBacktickRefs(cell, findingPath, func(ref string) bool {
		return strings.HasPrefix(ref, "docs/") && strings.HasSuffix(ref, ".md")
	})
}

func filteredBacktickRefs(cell, findingPath string, keep func(string) bool) ([]string, []finding) {
	matches := backtickRef.FindAllStringSubmatch(cell, -1)
	var out []string
	var findings []finding
	for _, match := range matches {
		raw := strings.TrimSpace(match[1])
		if raw == "" {
			continue
		}

		ref := trimLeadingDotSlash(raw)
		if !keep(ref) {
			if looksLikeUnsafeRepoRef(ref) {
				findings = append(findings, finding{path: findingPath, msg: "unsafe baseline ref: " + raw})
			}
			continue
		}
		if !safeRepoRelativeRef(ref) {
			findings = append(findings, finding{path: findingPath, msg: "unsafe baseline ref: " + raw})
			continue
		}
		if !slices.Contains(out, ref) {
			out = append(out, ref)
		}
	}
	return out, findings
}

func trimLeadingDotSlash(ref string) string {
	for strings.HasPrefix(ref, "./") {
		ref = strings.TrimPrefix(ref, "./")
	}
	return ref
}

func looksLikeUnsafeRepoRef(ref string) bool {
	if safeRepoRelativeRef(ref) {
		return false
	}
	candidate := strings.TrimLeft(ref, "/")
	return strings.HasPrefix(candidate, "docs/")
}

func safeRepoRelativeRef(ref string) bool {
	return safeSlashPath(ref, true)
}

func safeBundleRelativePath(ref string) bool {
	return safeSlashPath(ref, false)
}

func safeSlashPath(ref string, allowTrailingSlash bool) bool {
	if ref == "" || strings.Contains(ref, "\\") || strings.Contains(ref, "\x00") || strings.Contains(ref, ":") || strings.HasPrefix(ref, "/") {
		return false
	}
	if strings.HasSuffix(ref, "/") {
		if !allowTrailingSlash {
			return false
		}
		ref = strings.TrimSuffix(ref, "/")
		if ref == "" {
			return false
		}
	}
	for _, part := range strings.Split(ref, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func evidenceBundleForRef(ref string, info os.FileInfo) string {
	parts := strings.Split(strings.Trim(ref, "/"), "/")
	if len(parts) < 3 || parts[0] != "docs" || parts[1] != "evidence" {
		return ""
	}
	if info.IsDir() || info.Mode()&fs.ModeSymlink != 0 || strings.HasSuffix(ref, "/") {
		return strings.Join(parts[:3], "/")
	}
	if len(parts) >= 4 {
		return strings.Join(parts[:3], "/")
	}
	return ""
}

func lstatRepoRelative(repoRoot, ref, findingPath string) (string, fs.FileInfo, []finding, error) {
	if !safeRepoRelativeRef(ref) {
		return "", nil, []finding{{path: findingPath, msg: "unsafe repository path: " + ref}}, nil
	}

	clean := strings.TrimSuffix(ref, "/")
	parts := strings.Split(clean, "/")
	full := repoRoot
	for i, part := range parts {
		full = filepath.Join(full, part)
		info, err := os.Lstat(full)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && i < len(parts)-1 {
				parent := strings.Join(parts[:i+1], "/")
				return "", nil, []finding{{path: findingPath, msg: "repository path parent does not exist: " + parent}}, nil
			}
			return full, nil, nil, err
		}
		if i < len(parts)-1 {
			parent := strings.Join(parts[:i+1], "/")
			if info.Mode()&fs.ModeSymlink != 0 {
				return "", nil, []finding{{path: findingPath, msg: "repository path contains symlinked parent: " + parent}}, nil
			}
			if !info.IsDir() {
				return "", nil, []finding{{path: findingPath, msg: "repository path parent is not a directory: " + parent}}, nil
			}
		}
		if i == len(parts)-1 {
			return full, info, nil, nil
		}
	}
	return "", nil, []finding{{path: findingPath, msg: "repository path is empty: " + ref}}, nil
}

func discoverEvidenceBundles(repoRoot string) ([]string, []finding, error) {
	const rootRef = "docs/evidence"
	root, info, pathFindings, err := lstatRepoRelative(repoRoot, rootRef, rootRef)
	if err != nil {
		return nil, nil, err
	}
	if len(pathFindings) > 0 {
		return nil, pathFindings, nil
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return nil, []finding{{path: rootRef, msg: "evidence directory must not be a symlink"}}, nil
	}
	if !info.IsDir() {
		return nil, []finding{{path: rootRef, msg: "evidence directory is not a directory"}}, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	var bundles []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
			continue
		}
		bundles = append(bundles, "docs/evidence/"+entry.Name())
	}
	sort.Strings(bundles)
	return bundles, nil, nil
}

func checkBundle(repoRoot, bundle string) []finding {
	var findings []finding

	bundlePath, info, pathFindings, err := lstatRepoRelative(repoRoot, bundle, bundle)
	if len(pathFindings) > 0 {
		return pathFindings
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return []finding{{path: bundle, msg: "evidence bundle root does not exist"}}
	case err != nil:
		return []finding{{path: bundle, msg: err.Error()}}
	case info.Mode()&fs.ModeSymlink != 0:
		return []finding{{path: bundle, msg: "evidence bundle root must not be a symlink"}}
	case !info.IsDir():
		return []finding{{path: bundle, msg: "evidence bundle root is not a directory"}}
	}

	readmeFindings, _ := validateBundleControlFile(bundlePath, bundle, "README.md")
	findings = append(findings, readmeFindings...)

	sigFindings := validateOptionalBundleControlFile(bundlePath, bundle, "SHA256SUMS.sig")
	findings = append(findings, sigFindings...)

	sumPath := filepath.Join(bundlePath, "SHA256SUMS")
	sumFileFindings, ok := validateBundleControlFile(bundlePath, bundle, "SHA256SUMS")
	findings = append(findings, sumFileFindings...)
	if !ok {
		return findings
	}
	entries, sumFindings, err := parseSHA256SUMS(sumPath)
	if err != nil {
		findings = append(findings, finding{path: bundle + "/SHA256SUMS", msg: err.Error()})
		return findings
	}
	findings = append(findings, sumFindings...)
	if len(entries) == 0 {
		findings = append(findings, finding{path: bundle + "/SHA256SUMS", msg: "contains no checksum entries"})
	}

	covered := map[string]bool{}
	for _, entry := range entries {
		covered[entry.path] = true
		findings = append(findings, verifySHA256Entry(bundlePath, bundle, entry)...)
	}

	rawFiles, rawFindings, err := bundleRawFiles(bundlePath, bundle)
	findings = append(findings, rawFindings...)
	if err != nil {
		findings = append(findings, finding{path: bundle, msg: err.Error()})
		return findings
	}
	for _, raw := range rawFiles {
		if !covered[raw] {
			findings = append(findings, finding{path: bundle + "/" + raw, msg: "raw evidence file is not covered by SHA256SUMS"})
		}
	}
	return findings
}

func validateBundleControlFile(bundlePath, bundle, name string) ([]finding, bool) {
	path := filepath.Join(bundlePath, name)
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return []finding{{path: bundle, msg: "evidence bundle is missing " + name}}, false
	case err != nil:
		return []finding{{path: bundle + "/" + name, msg: err.Error()}}, false
	case info.Mode()&fs.ModeSymlink != 0:
		return []finding{{path: bundle + "/" + name, msg: "evidence bundle control file must not be a symlink"}}, false
	case !info.Mode().IsRegular():
		return []finding{{path: bundle + "/" + name, msg: "evidence bundle control file must be a regular file"}}, false
	default:
		return nil, true
	}
}

func validateOptionalBundleControlFile(bundlePath, bundle, name string) []finding {
	path := filepath.Join(bundlePath, name)
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return []finding{{path: bundle + "/" + name, msg: err.Error()}}
	case info.Mode()&fs.ModeSymlink != 0:
		return []finding{{path: bundle + "/" + name, msg: "evidence bundle control file must not be a symlink"}}
	case !info.Mode().IsRegular():
		return []finding{{path: bundle + "/" + name, msg: "evidence bundle control file must be a regular file"}}
	default:
		return nil
	}
}

type checksumEntry struct {
	hash string
	path string
}

func parseSHA256SUMS(path string) ([]checksumEntry, []finding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var entries []checksumEntry
	var findings []finding
	seen := map[string]bool{}
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) >= 66 && line[64] == ' ' && line[65] == '*' {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum line must use text-mode SHA256SUMS format without a binary '*' path prefix"})
			continue
		}
		if len(line) < 67 || line[64:66] != "  " {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum line must use '<64 lowercase hex><two spaces><bundle-relative path>' format"})
			continue
		}
		hash, rel := line[:64], line[66:]
		if !sha256Hex.MatchString(hash) {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum hash must be 64 lowercase hex characters"})
		}
		if strings.ContainsAny(rel, " \t\r\n") {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum path must not contain whitespace"})
			continue
		}
		if strings.HasPrefix(rel, "*") {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum path must not start with a binary '*' path prefix"})
			continue
		}
		if !safeBundleRelativePath(rel) {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum path must be a safe bundle-relative path"})
			continue
		}
		if seen[rel] {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "duplicate checksum path: " + rel})
		}
		seen[rel] = true
		entries = append(entries, checksumEntry{hash: hash, path: rel})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return entries, findings, nil
}

func verifySHA256Entry(bundlePath, bundle string, entry checksumEntry) []finding {
	full, info, findings := lstatChecksumEntry(bundlePath, bundle, entry.path)
	if len(findings) > 0 {
		return findings
	}
	if !info.Mode().IsRegular() {
		return []finding{{path: bundle + "/" + entry.path, msg: "checksum references non-regular file"}}
	}

	file, err := os.Open(full)
	if err != nil {
		return []finding{{path: bundle + "/" + entry.path, msg: err.Error()}}
	}
	defer file.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return []finding{{path: bundle + "/" + entry.path, msg: err.Error()}}
	}
	got := hex.EncodeToString(sum.Sum(nil))
	if got != entry.hash {
		return []finding{{path: bundle + "/" + entry.path, msg: "SHA256SUMS hash mismatch"}}
	}
	return nil
}

func lstatChecksumEntry(bundlePath, bundle, rel string) (string, fs.FileInfo, []finding) {
	full := bundlePath
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		full = filepath.Join(full, part)
		info, err := os.Lstat(full)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return "", nil, []finding{{path: bundle + "/" + rel, msg: "checksum references missing file"}}
			}
			return "", nil, []finding{{path: bundle + "/" + rel, msg: err.Error()}}
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			if i < len(parts)-1 {
				parent := strings.Join(parts[:i+1], "/")
				return "", nil, []finding{{path: bundle + "/" + rel, msg: "checksum path contains symlinked parent: " + parent}}
			}
			return "", nil, []finding{{path: bundle + "/" + rel, msg: "checksum references symlink"}}
		}
		if i < len(parts)-1 && !info.IsDir() {
			parent := strings.Join(parts[:i+1], "/")
			return "", nil, []finding{{path: bundle + "/" + rel, msg: "checksum path parent is not a directory: " + parent}}
		}
		if i == len(parts)-1 {
			return full, info, nil
		}
	}
	return "", nil, []finding{{path: bundle + "/" + rel, msg: "checksum references missing file"}}
}

func bundleRawFiles(bundlePath, bundle string) ([]string, []finding, error) {
	var out []string
	var findings []finding
	err := filepath.WalkDir(bundlePath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == bundlePath {
			return nil
		}

		relOS, err := filepath.Rel(bundlePath, path)
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relOS)
		if ignoredBundleEntry(rel) {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			findings = append(findings, finding{path: bundle + "/" + rel, msg: "raw evidence entry must not be a symlink"})
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			findings = append(findings, finding{path: bundle + "/" + rel, msg: err.Error()})
			return nil
		}
		if !info.Mode().IsRegular() {
			findings = append(findings, finding{path: bundle + "/" + rel, msg: "raw evidence entry must be a regular file"})
			return nil
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(out)
	return out, findings, nil
}

func ignoredBundleEntry(rel string) bool {
	switch rel {
	case "README.md", "SHA256SUMS", "SHA256SUMS.sig":
		return true
	}
	return filepath.Base(rel) == ".DS_Store"
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

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func sortFindings(findings []finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].path == findings[j].path {
			return findings[i].msg < findings[j].msg
		}
		return findings[i].path < findings[j].path
	})
}
