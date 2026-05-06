# Security/Spec Audit

Date: 2026-05-06

Target module: `github.com/the-sarge/cpace`

Audit commit: `4a8f629e59f0cc5c8f9351abacfa511fe6e4f441`

Draft source: `draft-irtf-cfrg-cpace-21`
(`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This audit checked `docs/security-assessment.md` and `docs/spec-matrix.md`
against the implementation, tests, release evidence, and the draft-21 text.
It is a documentation and conformance audit, not an independent cryptographic
review. This is a self-audit by the project maintainer, distinct from
independent cryptographic review or external review.

The audit covered:

- implemented suite and protocol mode;
- CPace transcript, generator, ISK, scalar sampling, and confirmation behavior;
- package-owned CI construction, wire framing, and per-field caps;
- session lifecycle, export, peer metadata, and memory-handling claims;
- invalid-share handling and parser rejection behavior;
- test/vector/fuzz/dependency evidence referenced by the release-readiness
  docs.

## Result

No security/spec drift was found at the audit commit.

`go test ./...` passes at the audit commit.

The security assessment and spec matrix accurately describe the current
implementation:

- only `CPACE-RISTR255-SHA512` from draft-21 is implemented;
- only initiator-responder mode is exposed;
- `Respond` success is not authentication; sessions are returned only after
  explicit key confirmation;
- `transcript_ir`, generator derivation, ISK derivation, confirmation tags,
  scalar sampling, and `scalar_mult_vfy` handling match the documented
  draft-21 profile;
- package-owned CI construction, binary framing, non-configurable field caps,
  `Session.Export`, `Session.Close`, `PeerAssociatedData`, and `PeerID` are
  correctly documented as package-profile behavior;
- dependency and fuzz evidence references point to the recorded release
  evidence documents.

## Residual Risk

External review of package-owned CI/framing/profile choices remains open.
Independent cryptographic review remains required before any production-ready
claim.

Repeat this audit if protocol code, parser/framing code, package-profile docs,
dependencies, toolchain, or the targeted CPace draft revision changes before a
release tag.
