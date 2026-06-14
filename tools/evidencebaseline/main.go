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

func main() {
	repoRoot := flag.String("repo-root", "../..", "repository root")
	flag.Parse()

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
	rows, findings, err := parseBaselineIndex(baselinePath)
	if err != nil {
		return nil, err
	}

	referencedBundles := map[string]bool{}
	for _, row := range rows {
		rawRefs, rawFindings := evidenceRefs(row.rawCell, baselinePath+":"+row.lane)
		findings = append(findings, rawFindings...)
		for _, ref := range rawRefs {
			full := filepath.Join(repoRoot, filepath.FromSlash(ref))
			info, err := os.Lstat(full)
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
			full := filepath.Join(repoRoot, filepath.FromSlash(ref))
			info, err := os.Stat(full)
			switch {
			case errors.Is(err, fs.ErrNotExist):
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced summary doc does not exist: " + ref})
			case err != nil:
				return nil, err
			case info.IsDir():
				findings = append(findings, finding{path: baselinePath + ":" + row.lane, msg: "referenced summary doc is a directory: " + ref})
			}
		}
	}

	allBundles, err := discoverEvidenceBundles(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, bundle := range allBundles {
		referencedBundles[bundle] = true
	}

	for bundle := range referencedBundles {
		findings = append(findings, checkBundle(repoRoot, bundle)...)
	}
	sortFindings(findings)
	return findings, nil
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
	if len(tableLines) < 3 {
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
	for _, line := range tableLines[2:] {
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
			if looksLikeUnsafeKeptRef(ref, keep) {
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

func looksLikeUnsafeKeptRef(ref string, keep func(string) bool) bool {
	if safeRepoRelativeRef(ref) {
		return false
	}
	if strings.HasPrefix(ref, "/") && keep(strings.TrimLeft(ref, "/")) {
		return true
	}
	candidate := strings.TrimLeft(ref, "/")
	return strings.Contains(candidate, "docs/")
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

func discoverEvidenceBundles(repoRoot string) ([]string, error) {
	root := filepath.Join(repoRoot, "docs", "evidence")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
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
	return bundles, nil
}

func checkBundle(repoRoot, bundle string) []finding {
	bundlePath := filepath.Join(repoRoot, filepath.FromSlash(bundle))
	var findings []finding

	info, err := os.Lstat(bundlePath)
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
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum line must contain hash and relative path"})
			continue
		}
		hash, rel := fields[0], fields[1]
		if !sha256Hex.MatchString(hash) {
			findings = append(findings, finding{path: fmt.Sprintf("%s:%d", path, lineNo), msg: "checksum hash must be 64 lowercase hex characters"})
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
	full := filepath.Join(bundlePath, filepath.FromSlash(entry.path))
	info, err := os.Lstat(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []finding{{path: bundle + "/" + entry.path, msg: "checksum references missing file"}}
		}
		return []finding{{path: bundle + "/" + entry.path, msg: err.Error()}}
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return []finding{{path: bundle + "/" + entry.path, msg: "checksum references symlink"}}
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

func sortFindings(findings []finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].path == findings[j].path {
			return findings[i].msg < findings[j].msg
		}
		return findings[i].path < findings[j].path
	})
}
