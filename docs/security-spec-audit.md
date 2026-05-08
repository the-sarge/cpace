# Security/Spec Audit

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Implementation baseline: `737bc56ffba81e2df5e9caa0df1ff180bfdb594b`

Documentation/evidence baseline: PR #43 head or merge commit containing this
file.

Toolchain: Go 1.26.3

Evidence transcript: `docs/evidence/go1263-20260508/local-analysis.log`

Draft source: `draft-irtf-cfrg-cpace-21`
(`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This audit checked the PR #43 versions of `docs/security-assessment.md` and
`docs/spec-matrix.md` against the implementation baseline, tests, refreshed Go
1.26.3 release evidence, and the draft-21 text. It is a documentation and
conformance audit, not an independent cryptographic review. This is a
self-audit by the project maintainer, distinct from independent cryptographic
review or external review.

The audit covered:

- implemented suite and protocol mode;
- CPace transcript, generator, ISK, scalar sampling, and confirmation behavior;
- package-owned CI construction, wire framing, and per-field caps;
- session lifecycle, export, peer metadata, and memory-handling claims;
- invalid-share handling and parser rejection behavior;
- test/vector/fuzz/dependency/Capslock evidence referenced by the
  release-readiness docs;
- Go 1.26.3 toolchain impact after the 2026-05-07 security release.

## Result

No security/spec drift was found at the implementation baseline.

`task check` passes under Go 1.26.3 for PR #43. The clean-worktree evidence
transcript records dependency, gosec, and Capslock commands at the
implementation baseline.

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
it does transitively use Go crypto internals, so dependency, fuzz, and Capslock
evidence was refreshed under Go 1.26.3. No package source change was required.
A Go 1.26.2-to-1.26.3 draft/RFC vector bit-identity comparison was not
performed for this refresh because the prior Go 1.26.2 toolchains were no
longer retained on the maintainer machines after upgrade. This refresh relies
on the Go 1.26.3 `task check` run, which reran the existing draft/RFC vector
assertions under the refreshed toolchain.

## Residual Risk

External review of package-owned CI/framing/profile choices remains open.
Independent cryptographic review remains required before any production-ready
claim.

Repeat this audit if protocol code, parser/framing code, package-profile docs,
dependencies, toolchain, or the targeted CPace draft revision changes before a
release tag.
