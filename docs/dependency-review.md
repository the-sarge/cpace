# Dependency Review

Date: 2026-06-19

Target module: `github.com/the-sarge/cpace`

Review commit: `f7efa6a963a954952b1ecad3f46530f13799fe89`

Review worktree: clean worktree at the review commit.

Toolchain: Go 1.26.4 (`darwin/arm64`)

Transcript: `docs/evidence/f7efa6a-20260619/local-analysis.log`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned review is fresh for the latest release candidate.

Dependencies:

| Module | Version | Role | Notes |
| --- | --- | --- | --- |
| `github.com/gtank/ristretto255` | `v0.2.0` | Direct | Ristretto255 group and scalar operations. The package documents constant-time operations except variable-time APIs; this module does not call the variable-time APIs. |
| `filippo.io/edwards25519` | `v1.2.0` | Indirect | Pulled by `github.com/gtank/ristretto255`; pinned above the `v1.1.0` release noted by some SCA tools. |

## Commands

- `go version`
- `go list -m all`
- `govulncheck -version`
- `govulncheck -test -show verbose ./...`
- `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 ./...`

## Results

`go version` reported:

```text
go version go1.26.4 darwin/arm64
```

`go list -m all` reported only the main module plus:

- `filippo.io/edwards25519 v1.2.0`
- `github.com/gtank/ristretto255 v0.2.0`

`govulncheck -version` reported:

```text
Go: go1.26.4
Scanner: govulncheck@v1.3.0
DB: https://vuln.go.dev
DB updated: 2026-06-16 23:55:18 +0000 UTC
```

`govulncheck -test -show verbose ./...` scanned the module, the two dependency modules, and the Go 1.26.4 standard library. Result: no vulnerabilities found.

The pinned gosec command reported zero issues:

```text
Summary:
  Gosec  : dev
  Files  : 13
  Lines  : 1478
  Nosec  : 0
  Issues : 0
```

The `Gosec : dev` summary value is the string emitted by the upstream gosec binary for this pinned `go run …gosec@v2.26.1` invocation. The file and line counts increased after the accepted-ADR implementation sequence split the package internals into additional files.

This refresh follows the accepted-ADR implementation sequence through the exact candidate commit, including ADR-0003 peer-share errors, ADR-0001 core extraction and zero-value hardening, ADR-0002 suite API cleanup, ADR-0009 caller input, issue #80 responder decoded-share reuse, and PR #199's Go fix modernization. No dependency versions changed from the previous review.

## Residual Risk

Repeat this review against the exact release tag if any dependency, toolchain,
or parser/security-relevant code changes before release.
