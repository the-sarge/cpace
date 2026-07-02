package cpace

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type fuzzTargetRegistryEntry struct {
	Target  string `json:"target"`
	Package string `json:"package"`
	Binary  string `json:"binary"`
}

type ossFuzzBuildTarget struct {
	Module string
	Target string
	Binary string
	Line   string
}

var fuzzTargetBinaryPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func TestFuzzTargetRegistrySchema(t *testing.T) {
	entries := readFuzzTargetRegistry(t)
	seenTargets := make(map[string]struct{}, len(entries))
	seenBinaries := make(map[string]struct{}, len(entries))

	for i, entry := range entries {
		if entry.Target == "" {
			t.Errorf("entry %d has empty target", i)
		}
		if entry.Package != "." {
			t.Errorf("entry %d target %q package = %q, want .", i, entry.Target, entry.Package)
		}
		if entry.Binary == "" {
			t.Errorf("entry %d target %q has empty binary", i, entry.Target)
		} else if !fuzzTargetBinaryPattern.MatchString(entry.Binary) {
			t.Errorf("entry %d target %q binary = %q, want match %s", i, entry.Target, entry.Binary, fuzzTargetBinaryPattern)
		}

		if _, ok := seenTargets[entry.Target]; ok {
			t.Errorf("duplicate target %q", entry.Target)
		}
		seenTargets[entry.Target] = struct{}{}

		if _, ok := seenBinaries[entry.Binary]; ok {
			t.Errorf("duplicate binary %q", entry.Binary)
		}
		seenBinaries[entry.Binary] = struct{}{}
	}
}

func TestFuzzTargetRegistryMatchesDefinedTargets(t *testing.T) {
	entries := readFuzzTargetRegistry(t)
	registeredTargets := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		registeredTargets[entry.Target] = struct{}{}
	}

	definedTargets := discoverDefinedFuzzTargets(t)
	if len(definedTargets) == 0 {
		t.Fatal("no func FuzzXxx(f *testing.F) definitions found")
	}

	if !reflect.DeepEqual(sortedKeys(registeredTargets), sortedKeys(definedTargets)) {
		t.Fatalf("registered fuzz targets do not match defined fuzz targets\nregistered: %v\ndefined:    %v", sortedKeys(registeredTargets), sortedKeys(definedTargets))
	}
}

func TestFuzzTargetRegistryMatchesOSSFuzzBuild(t *testing.T) {
	entries := readFuzzTargetRegistry(t)
	buildTargets := readOSSFuzzBuildTargets(t)
	if len(buildTargets) == 0 {
		t.Fatal("ossfuzz/build.sh has no compile_native_go_fuzzer lines")
	}

	module := readModulePath(t)
	for i, target := range buildTargets {
		if target.Module != module {
			t.Errorf("ossfuzz/build.sh compile line %d module = %q, want %q", i, target.Module, module)
		}
	}

	wantLines := make([]string, 0, len(entries))
	for _, entry := range entries {
		wantLines = append(wantLines, fmt.Sprintf("compile_native_go_fuzzer %s %s %s", module, entry.Target, entry.Binary))
	}

	gotLines := make([]string, 0, len(buildTargets))
	for _, target := range buildTargets {
		gotLines = append(gotLines, target.Line)
	}

	if !reflect.DeepEqual(gotLines, wantLines) {
		t.Fatalf("ossfuzz/build.sh compile lines do not match .github/fuzz-targets.json\nwant:\n%s\ngot:\n%s", strings.Join(wantLines, "\n"), strings.Join(gotLines, "\n"))
	}
}

func readFuzzTargetRegistry(tb testing.TB) []fuzzTargetRegistryEntry {
	tb.Helper()

	data, err := os.ReadFile(".github/fuzz-targets.json")
	if err != nil {
		tb.Fatalf("read .github/fuzz-targets.json: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var entries []fuzzTargetRegistryEntry
	if err := decoder.Decode(&entries); err != nil {
		tb.Fatalf("parse .github/fuzz-targets.json: %v", err)
	}
	if len(entries) == 0 {
		tb.Fatal(".github/fuzz-targets.json contains no targets")
	}

	return entries
}

func discoverDefinedFuzzTargets(tb testing.TB) map[string]struct{} {
	tb.Helper()

	files, err := filepath.Glob("*_test.go")
	if err != nil {
		tb.Fatalf("list *_test.go files: %v", err)
	}

	targets := make(map[string]struct{})
	fset := token.NewFileSet()
	for _, name := range files {
		parsed, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			tb.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Fuzz") {
				continue
			}
			if hasTestingFSignature(fn) {
				targets[fn.Name.Name] = struct{}{}
			}
		}
	}

	return targets
}

func hasTestingFSignature(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	if fn.Type.Results != nil && len(fn.Type.Results.List) != 0 {
		return false
	}
	param := fn.Type.Params.List[0]
	if len(param.Names) != 1 {
		return false
	}
	star, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "testing" && selector.Sel.Name == "F"
}

func readOSSFuzzBuildTargets(tb testing.TB) []ossFuzzBuildTarget {
	tb.Helper()

	data, err := os.ReadFile("ossfuzz/build.sh")
	if err != nil {
		tb.Fatalf("read ossfuzz/build.sh: %v", err)
	}

	var targets []ossFuzzBuildTarget
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "compile_native_go_fuzzer ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 {
			tb.Fatalf("ossfuzz/build.sh compile line %q has %d fields, want 4", line, len(fields))
		}
		targets = append(targets, ossFuzzBuildTarget{
			Module: fields[1],
			Target: fields[2],
			Binary: fields[3],
			Line:   line,
		})
	}
	if err := scanner.Err(); err != nil {
		tb.Fatalf("scan ossfuzz/build.sh: %v", err)
	}

	return targets
}

func readModulePath(tb testing.TB) string {
	tb.Helper()

	data, err := os.ReadFile("go.mod")
	if err != nil {
		tb.Fatalf("read go.mod: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		tb.Fatalf("scan go.mod: %v", err)
	}
	tb.Fatal("go.mod has no module line")
	return ""
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
