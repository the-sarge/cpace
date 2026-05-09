# Security/Spec Audit

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Implementation baseline: `2e09774f171dde8c62763d6e35a258b0fef88801`

Documentation/evidence baseline: v0.1.2 evidence PR head or merge commit
containing this file.

Toolchain: Go 1.26.3

Evidence transcript: `docs/evidence/v012-candidate-20260508/local-analysis.log`

Draft source: `draft-irtf-cfrg-cpace-21`
(`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This audit checked the v0.1.2 evidence PR versions of
`docs/security-assessment.md` and `docs/spec-matrix.md` against the
implementation baseline, tests, refreshed Go 1.26.3 release evidence, and the
draft-21 text. It is a documentation and conformance audit, not an independent
cryptographic review. This is a self-audit by the project maintainer, distinct
from independent cryptographic review or external review.

The audit covered:

- implemented suite and protocol mode;
- CPace transcript, generator, ISK, scalar sampling, and confirmation behavior;
- package-owned CI construction, wire framing, and per-field caps;
- session lifecycle, export, peer metadata, and memory-handling claims;
- invalid-share handling and parser rejection behavior;
- test/vector/fuzz/dependency/Capslock evidence referenced by the
  release-readiness docs;
- Go 1.26 `go fix` modernization impact after PR #45.

## Result

No security/spec drift was found at the implementation baseline.

`task check` passes under Go 1.26.3 for the v0.1.2 candidate. The
clean-worktree evidence transcript records dependency, gosec, and Capslock
commands at the implementation baseline.

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
- dependency, fuzz, and Capslock evidence references point to the refreshed Go
  1.26.3 evidence documents.

The Go 1.26.3 release note included security fixes in the `go` command, the
`pack` tool, and several standard-library packages, plus bug fixes including
`crypto/fips140`. CPace does not import the named web/template/mail packages;
it does transitively use Go crypto internals, so evidence remains recorded
under Go 1.26.3. PR #45 then applied mechanical Go 1.26 `go fix`
modernizations in `crypto.go`, `framing.go`, and concurrent tests. The
generator-string clamp, scalar-sampling loop bound, and LEB128 loop bound are
unchanged in meaning, and `task check` reran the existing draft/RFC vector
assertions at the implementation baseline. No Go API, wire/protocol,
dependency, or vector behavior change was found.

## Residual Risk

External review of package-owned CI/framing/profile choices remains open.
Independent cryptographic review remains required before any production-ready
claim.

Repeat this audit if protocol code, parser/framing code, package-profile docs,
dependencies, toolchain, or the targeted CPace draft revision changes before a
release tag.
