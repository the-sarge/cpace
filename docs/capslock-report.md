# Capslock Report

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Package-code baseline: `737bc56ffba81e2df5e9caa0df1ff180bfdb594b`

Status: external-review evidence. Capslock is experimental static capability
analysis; this report is review signal, not a release gate.

This report refreshes the earlier Go 1.26.2 Capslock evidence after the Go
1.26.3 security toolchain release. The refresh was run from a clean detached
worktree at the package-code baseline.

Transcript: `docs/evidence/go1263-20260508/local-analysis.log`

## Tool

```sh
go run github.com/google/capslock/cmd/capslock@v0.3.2 -version
```

Result:

```text
capslock version v0.3.2
compiled with Go version go1.26.3
includes Go tools version v0.43.0
```

## Command

```sh
go run github.com/google/capslock/cmd/capslock@v0.3.2 -packages ./...
```

## Summary

```text
Analyzed packages:
  filippo.io/edwards25519 v1.2.0
  github.com/gtank/ristretto255 v0.2.0

ARBITRARY_EXECUTION: 6 references
UNANALYZED: 5 references
```

Verbose output preserved the same capability classes and example call paths:

```text
ARBITRARY_EXECUTION: 6 references (6 direct, 0 transitive)
Example callpath:
  github.com/the-sarge/cpace.Respond
  api.go:147:26:github.com/the-sarge/cpace.respondWithRandom
  api.go:187:25:github.com/the-sarge/cpace.confirmationTag
  crypto.go:152:15:crypto/hmac.New
  hmac.go:48:25:crypto/internal/fips140only.Enforced
  fips140only.go:20:25:crypto/fips140.Enforced
  enforcement.go:37:31:crypto/fips140.isBypassed

UNANALYZED: 5 references (5 direct, 0 transitive)
Example callpath:
  github.com/the-sarge/cpace.Start
  api.go:108:24:github.com/the-sarge/cpace.startWithRandom
  api.go:125:24:github.com/the-sarge/cpace.sampleScalar
  crypto.go:61:27:io.ReadFull
```

The package does not directly expose filesystem, network, subprocess, dynamic
loading, environment mutation, or other broad operating-system capabilities in
the default Capslock summary.

## Finding Triage

| Capability | Count | Paths | Triage |
| --- | ---: | --- | --- |
| `ARBITRARY_EXECUTION` | 6 | `Respond`, `confirmationTag`, `respondWithRandom`, `Initiator.Finish`, `Responder.Finish`, `Session.Export` through `crypto/hmac.New` or `crypto/hkdf.Key` into Go's `crypto/fips140.isBypassed` path. | Tool classification from Go standard-library FIPS enforcement internals, not an application subprocess or dynamic-code execution path in this module. Keep under review when Go toolchains change. |
| `UNANALYZED` | 5 | `Start`, `Respond`, `startWithRandom`, `respondWithRandom`, and `sampleScalar` through `io.ReadFull`. | Expected for scalar-randomness reads. Public `Start` and `Respond` use `crypto/rand.Reader`; tests and fuzzing use package-internal deterministic readers. |

## Residual Risk

Repeat this report when dependencies, Go toolchain, randomness handling,
HKDF/HMAC usage, or package imports change. Treat new filesystem, network,
process, plugin, environment, or unsafe capability classes as external-review
findings before a release-readiness claim.
