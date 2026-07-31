# Capslock Report

Date: 2026-07-31

Target module: `github.com/the-sarge/cpace`

Package-code baseline: `1f19e278112fa037890848ed6c086addeffdca4e`

Status: exact-candidate external-review evidence. Capslock is experimental static capability analysis; this report is review signal, not a release gate.

This report refreshes Capslock against the clean frozen `v0.1.3` candidate under Go 1.26.5. Compared with the historical Go 1.26.4 report at `f7efa6a963a954952b1ecad3f46530f13799fe89`, capability classes and reference counts are unchanged; only the reported `crypto.go` line number shifts after the scalar-sampling comment correction.

Transcript: `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`

Checksums: `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned Capslock report is fresh for the latest release candidate.

## Tool

```sh
go install github.com/google/capslock/cmd/capslock@v0.3.2
capslock -version
```

Result:

```text
capslock version v0.3.2
compiled with Go version go1.26.5
includes Go tools version v0.43.0
```

The transcript also preserves `go version -m` output for the built binary, including `github.com/google/capslock v0.3.2` and its module checksum.

## Commands

```sh
capslock -packages ./...
capslock -packages ./... -output=verbose
```

## Summary

```text
Analyzed packages:
  filippo.io/edwards25519 v1.2.0
  github.com/gtank/ristretto255 v0.2.0

ARBITRARY_EXECUTION: 11 references
UNANALYZED: 13 references
```

Verbose output preserves the same capability classes, counts, and example call paths:

```text
ARBITRARY_EXECUTION: 11 references (11 direct, 0 transitive)
Example callpath:
  github.com/the-sarge/cpace.Respond
  api.go:76:26:github.com/the-sarge/cpace.respondWithRandom
  api.go:92:41:github.com/the-sarge/cpace.newResponderCore
  core.go:103:37:(github.com/the-sarge/cpace.irTranscript).responderConfirmationTag
  transcript.go:73:24:github.com/the-sarge/cpace.confirmationTag
  crypto.go:132:15:crypto/hmac.New
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

The package does not directly expose filesystem, network, subprocess, dynamic-loading, environment-mutation, plugin, or unsafe capabilities in the default Capslock summary.

## Finding Triage

| Capability | Count | Change from historical baseline | Paths | Triage |
| --- | ---: | --- | --- | --- |
| `ARBITRARY_EXECUTION` | 11 direct | None | Public exchange and export paths through `crypto/hmac.New` or `crypto/hkdf.Key` into Go's `crypto/fips140.isBypassed` path. | Tool classification from Go standard-library FIPS enforcement internals, not an application subprocess or dynamic-code execution path in this module. Keep under review when Go toolchains change. |
| `UNANALYZED` | 13 direct | None | Public exchange paths and internal core constructors through `sampleScalar` and `io.ReadFull`. | Expected for scalar-randomness reads. Public `Start` and `Respond` use `crypto/rand.Reader`; tests and fuzzing use package-internal deterministic readers. |

No changed or new capability required a release finding. The Go 1.26.5 toolchain change itself was security-relevant, so both existing classes were re-examined rather than carried forward from the historical report without rerunning the tool.

## Residual Risk

Repeat this report when dependencies, Go toolchain, randomness handling, HKDF/HMAC usage, or package imports change. Treat new filesystem, network, process, plugin, environment, or unsafe capability classes as external-review findings before a release-readiness claim.
