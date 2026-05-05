# Dependency Review Notes

Date: 2026-05-05

Target module: `github.com/the-sarge/cpace`

Dependencies:

| Module | Version | Role | Notes |
| --- | --- | --- | --- |
| `github.com/gtank/ristretto255` | `v0.2.0` | Direct | Ristretto255 group and scalar operations. The package documents constant-time operations except variable-time APIs; this module does not call the variable-time APIs. |
| `filippo.io/edwards25519` | `v1.2.0` | Indirect | Pulled by `github.com/gtank/ristretto255`; pinned above the `v1.1.0` release noted by some SCA tools. |

Local checks run:

- `govulncheck ./...`: no vulnerabilities found.
- `go list -m all`: only the two modules above are required.

Before release, repeat this review with the exact release commit and archive the
full `govulncheck -show verbose ./...` output.
