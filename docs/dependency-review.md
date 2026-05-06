# Dependency Review

Date: 2026-05-06

Target module: `github.com/the-sarge/cpace`

Review commit: `06f21c51645f54e2b7bde7c5b538479463be5d0e`

Dependencies:

| Module | Version | Role | Notes |
| --- | --- | --- | --- |
| `github.com/gtank/ristretto255` | `v0.2.0` | Direct | Ristretto255 group and scalar operations. The package documents constant-time operations except variable-time APIs; this module does not call the variable-time APIs. |
| `filippo.io/edwards25519` | `v1.2.0` | Indirect | Pulled by `github.com/gtank/ristretto255`; pinned above the `v1.1.0` release noted by some SCA tools. |

## Commands

- `go list -m all`
- `govulncheck -test -show verbose ./...`
- `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 ./...`

## Results

`go list -m all` reported only the main module plus:

- `filippo.io/edwards25519 v1.2.0`
- `github.com/gtank/ristretto255 v0.2.0`

`govulncheck -test -show verbose ./...` scanned the module, the two dependency
modules, and the Go 1.26.2 standard library. Result: no vulnerabilities found.

`gosec v2.26.1` initially reported G115 integer-conversion findings in the
LEB128 parser. The parser was changed to keep decoded LEB128 lengths as `int`
under the three-byte package ceiling rather than converting between `uint64`
and `int`. After that change, `gosec` reported zero issues.

## Residual Risk

Repeat this review against the exact release tag if any dependency, toolchain,
or parser/security-relevant code changes before release.
