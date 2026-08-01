# Dependency Review

Date: 2026-07-31

Target module: `github.com/the-sarge/cpace`

Review commit: `1f19e278112fa037890848ed6c086addeffdca4e`

Review worktree: clean worktree at the review commit.

Toolchain: Go 1.26.5 (`darwin/arm64`)

Transcript: `docs/evidence/1f19e27-20260731/local-analysis.log`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned review is fresh for the latest release candidate.

Evidence status: current for the frozen `v0.1.3` candidate's dependency, vulnerability, and SAST/gosec lane. Other evidence lanes remain independently gated in `docs/evidence-baseline.md`.

Dependencies:

| Module | Version | Role | Notes |
| --- | --- | --- | --- |
| `github.com/gtank/ristretto255` | `v0.2.0` | Direct | Ristretto255 group and scalar operations. The package documents constant-time operations except variable-time APIs; this module does not call the variable-time APIs. |
| `filippo.io/edwards25519` | `v1.2.0` | Indirect | Pulled by `github.com/gtank/ristretto255`; pinned above the `v1.1.0` release noted by some SCA tools. |

## Commands

- `go version`
- `go mod verify`
- `go list -m all`
- `go list -m` with module and `go.mod` sums
- `go mod graph`
- dependency `LICENSE` inspection with SHA-256 hashes
- `govulncheck -version`
- `govulncheck -test -show verbose ./...`
- pinned `gosec@v2.26.1 -tests ./...`

## Results

`go version` reported:

```text
go version go1.26.5 darwin/arm64
```

`go list -m all` reported only the main module plus:

- `filippo.io/edwards25519 v1.2.0`
- `github.com/gtank/ristretto255 v0.2.0`

The module and `go.mod` sums match `go.sum`, the graph contains no unexpected module, and `go mod verify` reported `all modules verified`. Both dependency `LICENSE` files are BSD-3-Clause texts; their SHA-256 hashes and full texts are preserved in the transcript. No unknown or incompatible license was found.

`govulncheck -version` reported:

```text
Go: go1.26.5
Scanner: govulncheck@v1.3.0
DB: https://vuln.go.dev
DB updated: 2026-07-27 20:14:16 +0000 UTC
```

`govulncheck -test -show verbose ./...` scanned the package and its tests, the two dependency modules, and the Go 1.26.5 standard library. Result: no vulnerabilities found.

The pinned test-inclusive gosec command reported zero issues:

```text
Summary:
  Gosec  : dev
  Files  : 38
  Lines  : 8266
  Nosec  : 0
  Issues : 0
```

The `Gosec : dev` value is the upstream binary's version banner. The transcript's embedded Go module metadata independently records `github.com/securego/gosec/v2 v2.26.1` and its module checksum. The scanner-reported 38-file and 8,266-line totals are per-analysis-pass aggregates: under `-tests`, gosec analyzes the 13 non-test root-module files twice and the 12 root-module test files once, representing 25 distinct files and 6,787 distinct lines.

The captured manual dependency/license review, `govulncheck`, and test-inclusive gosec results produced no finding requiring a fix, suppression, or VEX record under their applicable `docs/security-gates.md` thresholds; `docs/vex.md` correctly remains empty for these results. This claim does not disposition CodeQL, ast-grep, GitHub Dependency Review, Dependabot, or other GitHub-alert results, which remain part of the pending release-control lane. This refresh covers the exact frozen candidate, including PR #219's Close/zeroization-path changes and the Go 1.26.5 toolchain update. Dependency versions did not change from the previous review.

## Residual Risk

The captured `./...` inventory and scans cover the root `github.com/the-sarge/cpace` module and exclude the separate modules under `tools/`. This lane does not replace the separate exact-candidate Capslock, security/spec, and vector-stability evidence in `docs/evidence/1f19e27-20260731-protocol/`, the paired long-fuzz evidence in `docs/evidence/1f19e27-20260731-fuzz/`, or the still-pending GitHub alert and release-control, Scorecard, and Release Validation lanes. Repeat this review against a newly selected exact candidate if any dependency, toolchain, parser/framing, protocol, security-relevant code, or package-profile change lands before release.
