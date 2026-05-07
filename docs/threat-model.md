# Threat Model

Status: review input for an auditable draft implementation. This document
describes the security boundaries this package is trying to maintain. It is not
an independent cryptographic review.

Reference implementation scope: `CPACE-RISTR255-SHA512` from
`draft-irtf-cfrg-cpace-21`, initiator-responder mode only, with mandatory
explicit key confirmation.

## Assets

- Password material supplied by callers.
- Ephemeral CPace scalar values and public shares.
- The CPace shared secret `K` and intermediate session key material.
- Confirmed session key material used by `Session.Export`.
- Exported application key material.
- Transcript/session metadata: identities, context, session ID, associated
  data, transcript ID, and peer metadata.
- Release evidence used by downstream consumers to judge prerelease quality.

## Trusted Computing Base

- Go runtime and compiler behavior.
- Go standard-library cryptographic primitives used by this package.
- `github.com/gtank/ristretto255` and its indirect dependency
  `filippo.io/edwards25519`.
- The package's parser/framing code, context-info construction, confirmation
  logic, exporter, and session lifecycle.
- Maintainer release process: signed annotated tags, protected `main`, protected
  `v*` tags, dependency review, fuzz evidence, and CI workflows.

## In-Scope Attackers

- Network attackers that can observe, replay, drop, reorder, or modify CPace
  messages.
- Malicious peers that send malformed, oversized, mismatched, reflected, or
  invalid-point protocol messages.
- Callers that accidentally misconfigure party identities, context, session ID,
  or associated data.
- Public contributors submitting untrusted pull-request code to hosted CI.
- Dependency or toolchain drift that changes release evidence or security
  assumptions.

## Out-Of-Scope Attackers And Non-Goals

- Local memory disclosure adversaries. The package does best-effort cleanup of
  owned key material, but the Go runtime does not provide secure zeroization,
  pinning, or copy avoidance guarantees.
- Compromised caller process, compromised operating system, or compromised
  random source.
- Authentication of outer application negotiation. If an application negotiates
  PAKE version, ciphersuite, protocol mode, or whether CPace is used at all,
  that negotiation needs its own downgrade protection.
- Multi-suite CPace support. This package intentionally implements only the
  Ristretto255/SHA-512 draft-21 suite.
- Production-readiness claims before external review, independent cryptographic
  review, and exact-candidate evidence refresh are complete.

## Security Boundaries

### Protocol Inputs

The package authenticates only the inputs it is given: password, party
identities, context, session ID, associated data, and wire messages. Callers are
responsible for making outer negotiation results part of those inputs or
protecting negotiation separately.

### Identity Orientation

Both parties must agree that `InitiatorID` names the party running `Start`, and
`ResponderID` names the party running `Respond`. If each side puts itself first,
the context info differs and confirmation fails. Global role labels such as
`client` and `server` are not enough as stable party identities for all users or
deployments.

### Wire Framing

The CPace draft leaves wire encoding to applications. This package owns a
binary framing layer with a format byte, suite byte, role byte, LEB128
length-value fields, exact-sized public shares and tags, and per-field size
caps. Decoders reject malformed, non-canonical, oversized, trailing, wrong-role,
wrong-suite, and invalid-length inputs.

### Session Outputs

`Session.TranscriptID` is the draft `CPaceSidOutput` for the confirmed CPace
exchange. It is not a complete channel binding for outer negotiation. Exported
keys come from `Session.Export`, which uses HKDF-SHA512 over the confirmed ISK
with caller-provided labels and context.

### Release Process

Required PR CI is intentionally narrow because public pull requests run
untrusted code on hosted runners. Security, fuzzing, static-analysis, and
release-validation workflows provide background and release evidence. Long
fuzzing remains maintainer-controlled evidence rather than a required PR gate.

## Review Focus

External reviewers should evaluate whether the package-owned context-info
construction, identity orientation, binary framing, size caps, empty-session-ID
policy, scalar sampling profile, invalid-point handling, exporter semantics,
session lifecycle, and release evidence match the claims in the public docs.

Independent cryptographic review remains required before any production-ready
claim.
