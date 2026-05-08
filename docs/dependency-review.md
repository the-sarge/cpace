# Dependency Review

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Review commit: `737bc56ffba81e2df5e9caa0df1ff180bfdb594b`

Review worktree: clean detached worktree at the review commit.

Toolchain: Go 1.26.3 (`darwin/arm64`)

Transcript: `docs/evidence/go1263-20260508/local-analysis.log`

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
go version go1.26.3 darwin/arm64
```

`go list -m all` reported only the main module plus:

- `filippo.io/edwards25519 v1.2.0`
- `github.com/gtank/ristretto255 v0.2.0`

`govulncheck -version` reported:

```text
Go: go1.26.3
Scanner: govulncheck@v1.3.0
DB: https://vuln.go.dev
DB updated: 2026-05-07 19:21:40 +0000 UTC
```

`govulncheck -test -show verbose ./...` scanned the module, the two dependency
modules, and the Go 1.26.3 standard library. Result: no vulnerabilities found.

The pinned gosec command reported zero issues:

```text
Summary:
  Gosec  : dev
  Files  : 7
  Lines  : 940
  Nosec  : 0
  Issues : 0
```

The `Gosec : dev` summary value is the string emitted by the upstream gosec
binary for this invocation. The transcript also records `go version -m` module
metadata for an installed copy of the same pinned command, showing
`github.com/securego/gosec/v2 v2.26.1`.

The Go 1.26.3 security release was treated as a toolchain evidence trigger. No
dependency versions changed from the previous review.

## Residual Risk

Repeat this review against the exact release tag if any dependency, toolchain,
or parser/security-relevant code changes before release.
