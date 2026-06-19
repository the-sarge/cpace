# Capslock Report

Date: 2026-06-19

Target module: `github.com/the-sarge/cpace`

Package-code baseline: `f7efa6a963a954952b1ecad3f46530f13799fe89`

Status: external-review evidence. Capslock is experimental static capability
analysis; this report is review signal, not a release gate.

This report refreshes the Capslock evidence after the accepted-ADR implementation sequence and PR #199's Go fix modernization. The refresh was run from a clean worktree at the exact package-code baseline. The capability classes remain the same as the previous report, while reference counts increased with the internal core extraction and caller-input refactor.

Transcript: `docs/evidence/f7efa6a-20260619/local-analysis.log`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned Capslock report is fresh for the latest release candidate.

## Tool

```sh
go run github.com/google/capslock/cmd/capslock@v0.3.2 -version
```

Result:

```text
capslock version v0.3.2
compiled with Go version go1.26.4
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

ARBITRARY_EXECUTION: 11 references
UNANALYZED: 13 references
```

Verbose output preserved the same capability classes and example call paths:

```text
ARBITRARY_EXECUTION: 11 references (11 direct, 0 transitive)
Example callpath:
  github.com/the-sarge/cpace.Respond
  api.go:76:26:github.com/the-sarge/cpace.respondWithRandom
  api.go:92:41:github.com/the-sarge/cpace.newResponderCore
  core.go:103:37:(github.com/the-sarge/cpace.irTranscript).responderConfirmationTag
  transcript.go:73:24:github.com/the-sarge/cpace.confirmationTag
  crypto.go:131:15:crypto/hmac.New
  hmac.go:48:25:crypto/internal/fips140only.Enforced
  fips140only.go:20:25:crypto/fips140.Enforced
  enforcement.go:37:31:crypto/fips140.isBypassed

UNANALYZED: 13 references (13 direct, 0 transitive)
Example callpath:
  github.com/the-sarge/cpace.Respond
  api.go:76:26:github.com/the-sarge/cpace.respondWithRandom
  api.go:92:41:github.com/the-sarge/cpace.newResponderCore
  core.go:90:24:github.com/the-sarge/cpace.sampleScalar
  crypto.go:59:27:io.ReadFull
```

Line numbers and reference counts shifted with the accepted-ADR refactor sequence; the broad capability classes are unchanged.

The package does not directly expose filesystem, network, subprocess, dynamic
loading, environment mutation, or other broad operating-system capabilities in
the default Capslock summary.

## Finding Triage

| Capability | Count | Paths | Triage |
| --- | ---: | --- | --- |
| `ARBITRARY_EXECUTION` | 11 | Public exchange and export paths through `crypto/hmac.New` or `crypto/hkdf.Key` into Go's `crypto/fips140.isBypassed` path. | Tool classification from Go standard-library FIPS enforcement internals, not an application subprocess or dynamic-code execution path in this module. Keep under review when Go toolchains change. |
| `UNANALYZED` | 13 | Public exchange paths and internal core constructors through `sampleScalar` and `io.ReadFull`. | Expected for scalar-randomness reads. Public `Start` and `Respond` use `crypto/rand.Reader`; tests and fuzzing use package-internal deterministic readers. |

## Residual Risk

Repeat this report when dependencies, Go toolchain, randomness handling,
HKDF/HMAC usage, or package imports change. Treat new filesystem, network,
process, plugin, environment, or unsafe capability classes as external-review
findings before a release-readiness claim.
