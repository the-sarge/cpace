package main

import "regexp"

var shaRef = regexp.MustCompile(`^[0-9a-f]{40}$`)

// The checker intentionally snapshots the accepted ADR-0007 release workflow.
// Benign workflow edits may need lockstep updates here so release-critical
// shell and permission controls do not silently drift.
const (
	tagGuard            = "github.ref_type == 'tag' && startsWith(github.ref, 'refs/tags/v')"
	unsupportedRefGuard = "github.event_name == 'workflow_dispatch' && !(" + tagGuard + ")"
	publishGuard        = "github.event_name == 'push' && startsWith(github.ref, 'refs/tags/v')"
)

type releasePolicy struct {
	workflowName  string
	rootKeys      []string
	env           map[string]string
	concurrency   map[string]string
	triggerKeys   []string
	pushKeys      []string
	pushTags      []string
	topPermission map[string]string

	jobs            []releaseJobPolicy
	requiredScripts []string
	expectedSigners string
}

type releaseJobPolicy struct {
	name           string
	displayName    string
	runsOn         string
	timeoutMinutes string
	ifCond         string
	needs          []string
	permissions    map[string]string
	outputs        map[string]string
	steps          []releaseStepPolicy
}

type releaseStepPolicy struct {
	identity        string
	name            string
	usesPrefix      string
	runLines        []string
	with            map[string]string
	env             map[string]string
	id              string
	ifCond          string
	shell           string
	continueOnError string
}

var acceptedReleasePolicy = releasePolicy{
	workflowName: "Release Validation",
	rootKeys:     []string{"name", "on", "permissions", "env", "concurrency", "jobs"},
	env: map[string]string{
		"GOTOOLCHAIN": "local",
	},
	concurrency: map[string]string{
		"group":              "release-${{ github.ref }}",
		"cancel-in-progress": "false",
	},
	triggerKeys: []string{"push", "workflow_dispatch"},
	pushKeys:    []string{"tags"},
	pushTags:    []string{"v*"},
	topPermission: map[string]string{
		"contents": "read",
	},
	jobs: []releaseJobPolicy{
		{
			name:           "unsupported-ref",
			displayName:    "Unsupported Dispatch Ref",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "2",
			ifCond:         unsupportedRefGuard,
			steps: []releaseStepPolicy{
				{
					name:     "Explain unsupported ref",
					identity: "Explain unsupported ref",
					runLines: []string{
						`echo "Release Validation workflow_dispatch runs must target a signed v* tag ref."`,
						`echo "Use the regular CI workflows for branch validation."`,
						`exit 1`,
					},
				},
			},
		},
		{
			name:           "verify-tag",
			displayName:    "Verify Signed Tag",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "5",
			ifCond:         tagGuard,
			outputs: map[string]string{
				"release-tag": "${{ steps.release-tag.outputs.release-tag }}",
				"sbom-file":   "${{ steps.release-tag.outputs.sbom-file }}",
				"prerelease":  "${{ steps.release-tag.outputs.prerelease }}",
				"latest":      "${{ steps.release-tag.outputs.latest }}",
			},
			steps: []releaseStepPolicy{
				checkoutStep(map[string]string{
					"fetch-depth":         "0",
					"persist-credentials": "false",
				}),
				{
					name:     "Verify tag object and signature",
					identity: "Verify tag object and signature",
					runLines: []string{
						`git fetch --force origin "refs/tags/$GITHUB_REF_NAME:refs/tags/$GITHUB_REF_NAME"`,
						`tag_type="$(git cat-file -t "$GITHUB_REF_NAME")"`,
						`printf '%s\n' "$tag_type"`,
						`test "$tag_type" = tag`,
						`git config gpg.ssh.allowedSignersFile .github/allowed_signers`,
						`git verify-tag "$GITHUB_REF_NAME"`,
					},
				},
				{
					name:     "Validate release tag metadata",
					identity: "Validate release tag metadata",
					id:       "release-tag",
					runLines: []string{
						`scripts/release-tag-metadata.sh "$GITHUB_REF_NAME" >> "$GITHUB_OUTPUT"`,
					},
				},
			},
		},
		{
			name:           "check",
			displayName:    "Check",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "5",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag"},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				setupGoStep(),
				reportGoEnvironmentStep(),
				{
					name:     "Run tests",
					identity: "Run tests",
					runLines: []string{`go test ./...`},
				},
			},
		},
		{
			name:           "race",
			displayName:    "Race",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag"},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				setupGoStep(),
				reportGoEnvironmentStep(),
				{
					name:     "Run race tests",
					identity: "Run race tests",
					runLines: []string{`go test -race ./...`},
				},
			},
		},
		{
			name:           "vuln",
			displayName:    "Govulncheck",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag"},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				setupGoStep(),
				reportGoEnvironmentStep(),
				{
					name:     "Install task",
					identity: "Install task",
					runLines: []string{`go install github.com/go-task/task/v3/cmd/task@v3.50.0`},
				},
				{
					name:     "Install govulncheck",
					identity: "Install govulncheck",
					runLines: []string{`go install golang.org/x/vuln/cmd/govulncheck@v1.3.0`},
				},
				{
					name:     "Run vulnerability scan",
					identity: "Run vulnerability scan",
					runLines: []string{`task vuln`},
				},
			},
		},
		{
			name:           "gosec",
			displayName:    "Gosec",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag"},
			permissions: map[string]string{
				"contents":        "read",
				"security-events": "write",
			},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				setupGoStep(),
				reportGoEnvironmentStep(),
				{
					name:     "Install task",
					identity: "Install task",
					runLines: []string{`go install github.com/go-task/task/v3/cmd/task@v3.50.0`},
				},
				{
					name:     "Install gosec",
					identity: "Install gosec",
					runLines: []string{`go install github.com/securego/gosec/v2/cmd/gosec@v2.26.1`},
				},
				{
					name:            "Run gosec scan",
					identity:        "Run gosec scan",
					runLines:        []string{`task gosec GOSEC='gosec -fmt sarif -out gosec.sarif'`},
					id:              "gosec",
					continueOnError: "true",
				},
				{
					name:       "Upload gosec SARIF to code scanning",
					identity:   "Upload gosec SARIF to code scanning",
					usesPrefix: "github/codeql-action/upload-sarif@",
					ifCond:     "always() && hashFiles('gosec.sarif') != ''",
					with: map[string]string{
						"sarif_file": "gosec.sarif",
						"category":   "gosec-release",
					},
				},
				{
					name:     "Report gosec result",
					identity: "Report gosec result",
					runLines: []string{`exit 1`},
					ifCond:   "steps.gosec.outcome == 'failure'",
				},
			},
		},
		{
			name:           "sbom",
			displayName:    "SBOM",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag", "check", "race", "vuln", "gosec"},
			outputs: map[string]string{
				"sbom-file":   "${{ steps.sbom-metadata.outputs.sbom-file }}",
				"sbom-sha256": "${{ steps.sbom-metadata.outputs.sbom-sha256 }}",
			},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				{
					name:       "Generate CycloneDX SBOM",
					identity:   "Generate CycloneDX SBOM",
					usesPrefix: "anchore/sbom-action@",
					with: map[string]string{
						"format":                "cyclonedx-json@1.5",
						"output-file":           "${{ needs.verify-tag.outputs.sbom-file }}",
						"syft-version":          "v1.45.1",
						"upload-artifact":       "false",
						"upload-release-assets": "false",
					},
				},
				{
					name:     "Validate SBOM and compute checksum",
					identity: "Validate SBOM and compute checksum",
					id:       "sbom-metadata",
					runLines: []string{
						`sbom_file="$SBOM_FILE"`,
						`scripts/validate-cyclonedx-sbom.sh "$sbom_file"`,
						`sbom_sha256="$(sha256sum "$sbom_file" | awk '{ print $1 }')"`,
						`{`,
						`echo "sbom-file=$sbom_file"`,
						`echo "sbom-sha256=$sbom_sha256"`,
						`} >> "$GITHUB_OUTPUT"`,
					},
					env: map[string]string{
						"SBOM_FILE": "${{ needs.verify-tag.outputs.sbom-file }}",
					},
				},
				{
					name:       "Upload SBOM artifact",
					identity:   "Upload SBOM artifact",
					usesPrefix: "actions/upload-artifact@",
					with: map[string]string{
						"name":              "release-sbom",
						"path":              "${{ needs.verify-tag.outputs.sbom-file }}",
						"if-no-files-found": "error",
					},
				},
			},
		},
		{
			name:           "sbom-attestation",
			displayName:    "SBOM Attestation",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"sbom"},
			permissions: map[string]string{
				"contents":     "read",
				"id-token":     "write",
				"attestations": "write",
			},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				{
					name:       "Download SBOM artifact",
					identity:   "Download SBOM artifact",
					usesPrefix: "actions/download-artifact@",
					with: map[string]string{
						"name": "release-sbom",
						"path": "dist",
					},
				},
				{
					name:     "Validate downloaded SBOM",
					identity: "Validate downloaded SBOM",
					runLines: []string{
						`scripts/validate-cyclonedx-sbom.sh "dist/$SBOM_FILE"`,
					},
					env: map[string]string{
						"SBOM_FILE": "${{ needs.sbom.outputs.sbom-file }}",
					},
				},
				{
					name:       "Attest SBOM",
					identity:   "Attest SBOM",
					usesPrefix: "actions/attest@",
					id:         "attest-sbom",
					with: map[string]string{
						"subject-path": "dist/${{ needs.sbom.outputs.sbom-file }}",
						"sbom-path":    "dist/${{ needs.sbom.outputs.sbom-file }}",
					},
				},
				{
					name:     "Prepare Sigstore bundle asset",
					identity: "Prepare Sigstore bundle asset",
					runLines: []string{
						`bundle_dst="dist/${SBOM_FILE}.sigstore.json"`,
						`cp "$ATTESTATION_BUNDLE_PATH" "$bundle_dst"`,
						`test -s "$bundle_dst"`,
					},
					env: map[string]string{
						"ATTESTATION_BUNDLE_PATH": "${{ steps.attest-sbom.outputs.bundle-path }}",
						"SBOM_FILE":               "${{ needs.sbom.outputs.sbom-file }}",
					},
				},
				{
					name:       "Upload release assets artifact",
					identity:   "Upload release assets artifact",
					usesPrefix: "actions/upload-artifact@",
					with: map[string]string{
						"name":              "release-assets",
						"path":              "dist/${{ needs.sbom.outputs.sbom-file }}\ndist/${{ needs.sbom.outputs.sbom-file }}.sigstore.json\n",
						"if-no-files-found": "error",
					},
				},
			},
		},
		{
			name:           "release",
			displayName:    "Release",
			runsOn:         "ubuntu-latest",
			timeoutMinutes: "10",
			ifCond:         tagGuard,
			needs:          []string{"verify-tag", "sbom", "sbom-attestation"},
			permissions: map[string]string{
				"contents": "write",
			},
			steps: []releaseStepPolicy{
				checkoutStep(nil),
				{
					name:       "Download prepared release assets",
					identity:   "Download prepared release assets",
					usesPrefix: "actions/download-artifact@",
					with: map[string]string{
						"name": "release-assets",
						"path": "dist",
					},
				},
				{
					name:     "Prepare release notes and assets",
					identity: "Prepare release notes and assets",
					runLines: []string{
						`sbom_path="dist/$SBOM_FILE"`,
						`bundle_path="${sbom_path}.sigstore.json"`,
						`test -s "$sbom_path"`,
						`test -s "$bundle_path"`,
						`scripts/validate-cyclonedx-sbom.sh "$sbom_path"`,
						`computed_sha256="$(sha256sum "$sbom_path" | awk '{ print $1 }')"`,
						`test "$computed_sha256" = "$SBOM_SHA256"`,
						`scripts/extract-release-notes.sh CHANGELOG.md "$RELEASE_TAG" > release-notes.md`,
						`run_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"`,
						`{`,
						`cat release-notes.md`,
						`printf "\n## Supply-chain artifacts\n\n"`,
						`printf -- "- Release validation: %s\n" "$run_url"`,
						"printf -- \"- SBOM: \\`%s\\`\\n\" \"$SBOM_FILE\"",
						"printf -- \"- SBOM SHA-256 (corruption detection only): \\`%s\\`\\n\" \"$computed_sha256\"",
						"printf -- \"- SBOM attestation bundle: \\`%s.sigstore.json\\`\\n\" \"$SBOM_FILE\"",
						"printf -- \"- Verification instructions: \\`docs/release-verification.md\\`\\n\"",
						`} > release-body.md`,
						`echo "Prepared release body and assets for $RELEASE_TAG."`,
					},
					env: map[string]string{
						"RELEASE_TAG": "${{ needs.verify-tag.outputs.release-tag }}",
						"SBOM_FILE":   "${{ needs.sbom.outputs.sbom-file }}",
						"SBOM_SHA256": "${{ needs.sbom.outputs.sbom-sha256 }}",
					},
				},
				{
					name:     "Publish GitHub Release",
					identity: "Publish GitHub Release",
					runLines: []string{
						`set -euo pipefail`,
						`tag="$RELEASE_TAG"`,
						`sbom_path="dist/$SBOM_FILE"`,
						`bundle_path="${sbom_path}.sigstore.json"`,
						`release_args=()`,
						`if [ "$RELEASE_PRERELEASE" = "true" ]; then`,
						`release_args+=(--prerelease)`,
						`fi`,
						`if [ "$RELEASE_LATEST" = "false" ]; then`,
						`release_args+=(--latest=false)`,
						`fi`,
						`if gh release view "$tag" --repo "$GITHUB_REPOSITORY" >/dev/null 2>&1; then`,
						`echo "GitHub Release already exists for $tag; refusing automated in-place asset replacement." >&2`,
						`echo "Delete the draft/release or repair it manually before rerunning Release Validation." >&2`,
						`exit 1`,
						`fi`,
						`gh release create "$tag" "$sbom_path" "$bundle_path" \`,
						`--repo "$GITHUB_REPOSITORY" \`,
						`--title "$tag" \`,
						`--notes-file release-body.md \`,
						`--verify-tag \`,
						`"${release_args[@]}"`,
					},
					env: map[string]string{
						"GH_TOKEN":           "${{ github.token }}",
						"RELEASE_TAG":        "${{ needs.verify-tag.outputs.release-tag }}",
						"RELEASE_PRERELEASE": "${{ needs.verify-tag.outputs.prerelease }}",
						"RELEASE_LATEST":     "${{ needs.verify-tag.outputs.latest }}",
						"SBOM_FILE":          "${{ needs.sbom.outputs.sbom-file }}",
					},
					ifCond: publishGuard,
					shell:  "bash",
				},
			},
		},
	},
	requiredScripts: []string{
		"scripts/release-tag-policy.sh",
		"scripts/release-tag-metadata.sh",
		"scripts/validate-cyclonedx-sbom.sh",
		"scripts/extract-release-notes.sh",
	},
	expectedSigners: `the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFDxEpP8Q6LERBcA5//zwD5dBisHL7uHQsFa+TTibRXC
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFF32/OwUJwQ/8OX5i2VNBO8oZf6B8l07U/R5n1rj0z6
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILlg3QNI+Zsnt6pR2Aip97Ak7VOajBeo+AlhIGfDYlPk
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILvdes5QNqI3PpKK6ksX6FtlL4LQgkq61AGflWVqoV0L
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJaEbAxjr0LjcZKsqfUvrHDZJVmvL/AEIg+WSQGt+75v
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBc0CLdNLHpbdkrEf/WLR3YH8oHyxsvSeaCwQ6MvlW4q
the-sarge@the-sarge.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEd3JYo6vayWkANtsMbPx81ilaiq7a4oPpW6A0uD6TkF
`,
}

func checkoutStep(with map[string]string) releaseStepPolicy {
	if with == nil {
		with = map[string]string{"persist-credentials": "false"}
	}
	return releaseStepPolicy{
		identity:   "uses:actions/checkout",
		usesPrefix: "actions/checkout@",
		with:       with,
	}
}

func setupGoStep() releaseStepPolicy {
	return releaseStepPolicy{
		name:       "Set up Go",
		identity:   "Set up Go",
		usesPrefix: "actions/setup-go@",
		with: map[string]string{
			"go-version-file": "go.mod",
			"cache":           "true",
		},
	}
}

func reportGoEnvironmentStep() releaseStepPolicy {
	return releaseStepPolicy{
		name:     "Report Go environment",
		identity: "Report Go environment",
		runLines: []string{
			`go version`,
			`go env GOTOOLCHAIN GOPROXY GOSUMDB`,
		},
	}
}

func (p releasePolicy) jobNames() []string {
	out := make([]string, 0, len(p.jobs))
	for _, job := range p.jobs {
		out = append(out, job.name)
	}
	return out
}

func (p releasePolicy) job(name string) (releaseJobPolicy, bool) {
	for _, job := range p.jobs {
		if job.name == name {
			return job, true
		}
	}
	return releaseJobPolicy{}, false
}

func (j releaseJobPolicy) stepIdentities() []string {
	out := make([]string, 0, len(j.steps))
	for _, step := range j.steps {
		out = append(out, step.identity)
	}
	return out
}
