# Capslock Report

Date: 2026-05-07

Target module: `github.com/the-sarge/cpace`

Package-code baseline: `39ccb58f827d88f6742628c1fadf9375539fb017`

Status: external-review evidence. Capslock is experimental static capability
analysis; this report is review signal, not a release gate.

PR #40 added benchmarks, examples, documentation, and OSS-Fuzz staging files
without changing the package implementation call graph. This report remains
useful capability-review signal for the current external-review packet. See
Residual Risk for rerun triggers.

## Tool

```sh
go run github.com/google/capslock/cmd/capslock@v0.3.2 -version
```

Result:

```text
capslock version v0.3.2
compiled with Go version go1.26.2
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
