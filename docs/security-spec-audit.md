# Security/Spec Audit

Date: 2026-06-19

Target module: `github.com/the-sarge/cpace`

Implementation baseline: `f7efa6a963a954952b1ecad3f46530f13799fe89`

Documentation/evidence baseline: the merge commit containing this file.

Toolchain: Go 1.26.4

Evidence transcript: `docs/evidence/f7efa6a-20260619/local-analysis.log`

Vector-stability transcript: `docs/evidence/f7efa6a-20260619/vector-stability.log`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned audit is fresh for the latest release candidate.

Draft source: `draft-irtf-cfrg-cpace-21`
(`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This audit checked `docs/security-assessment.md` and `docs/spec-matrix.md` against the implementation baseline, tests, refreshed Go 1.26.4 release evidence, and the draft-21 text. It is a documentation and conformance audit, not an independent cryptographic review. This is a project-side self-audit, distinct from independent cryptographic review or external review.

The audit covered:

- implemented suite and protocol mode;
- CPace transcript, generator, ISK, scalar sampling, and confirmation behavior;
- package-owned CI construction, wire framing, and per-field caps;
- session lifecycle, export, peer metadata, and memory-handling claims;
- invalid-share handling and parser rejection behavior;
- test/vector/fuzz/dependency/Capslock evidence referenced by the
  release-readiness docs;
- the go1.26.4 toolchain security release (2026-06-02) impact;
- the accepted-ADR implementation sequence through ADR-0003, ADR-0001, ADR-0002, ADR-0009, issue #80, and PR #199's Go fix modernization.

## Result

No security/spec drift was found at the implementation baseline.

`task check` passes under Go 1.26.4 at the implementation baseline (transcript records exit 0 with both test lanes green). The clean-worktree evidence transcript records dependency, gosec, Capslock, pinned Staticcheck, an explicit non-cached race test, and `task check`. Separate bundle artifacts record candidate GitHub status (`github-status-capture.log` and `github-runs-for-candidate.json`), fresh tag-ruleset capture (`tagruleset-capture.log`, `rulesets-list.json`, `ruleset-16048307.json`, and `ruleset-16048307-verify.json`), and the fresh Scorecard run (`github-scorecard-20260619-run.json`).

The security assessment and spec matrix accurately describe the current
implementation:

- only `CPACE-RISTR255-SHA512` from draft-21 is implemented;
- only initiator-responder mode is exposed;
- `Respond` success is not authentication; sessions are returned only after
  explicit key confirmation;
- `transcript_ir`, generator derivation, ISK derivation, confirmation tags, scalar sampling, and `scalar_mult_vfy` handling match the documented draft-21 profile, including the intentional internal-only ADR-0003 nil-plus-error convention for invalid peer-share handling;
- package-owned CI construction, binary framing, non-configurable field caps,
  `Session.Export`, `Session.Close`, `PeerAssociatedData`, and `PeerID` are
  correctly documented as package-profile behavior;
- dependency, Capslock, and paired long-fuzz evidence references point to the refreshed Go 1.26.4 pinned evidence baseline indexed in `docs/evidence-baseline.md`.

The go1.26.4 release (2026-06-02) is a security release: fixes to
`crypto/x509`, `mime`, and `net/textproto`, plus bug fixes to
`crypto/fips140`, the compiler, and the runtime. CPace does not import the
three patched packages; it does transitively use Go crypto internals including
`crypto/fips140`, so evidence is re-recorded under Go 1.26.4. The accepted-ADR implementation sequence changed package internals and public caller input, but the current docs correctly record the intended public API, wire format, package profile, error surface, and residual memory-handling limits. At this baseline `task check` reran the draft/RFC vector assertions. No unintended wire/protocol, dependency, or vector behavior change was found.

### Toolchain Vector Stability

For this refresh, the draft/RFC vector assertions were run under both toolchains and recorded in `docs/evidence/f7efa6a-20260619/vector-stability.log`: the default go1.26.4 toolchain, and `GOTOOLCHAIN=go1.26.3` as the previous-toolchain comparison. The selected vector tests pass under both (`TestStringUtilitiesDraftVectors`, `TestEmbeddedDraftVectorJSON`, `TestEmbeddedDraftGeneratorJSON`, `TestEmbeddedDraftConfirmationTagGoldens`, `TestEmbeddedDraftInvalidVectorJSON`, `TestRistrettoDraft21Vectors`, and `TestScalarMultVFYDraftInvalidVectors`). Future toolchain-triggered refreshes should continue the old/new comparison pattern.

## Current Implementation Notes

ADR-0003 is included in this baseline: exported peer-share sentinels classify public-wire decode and identity-element failures while preserving `ErrAbort` wrapping and protocol abort behavior. Internally, `scalarMultVFY` returns nil plus an error on failure instead of draft-21's function-level neutral-element return convention; that divergence is internal-only and documented in `docs/security-assessment.md` and `docs/spec-matrix.md`.

ADR-0001 is included in this baseline: `Initiator` and `Responder` are thin shells over core state, caller-fabricated zero-value shells return `ErrInvalidInput` without consuming state, and the manual secret-lifetime findings remain the same in kind as documented. Secret-derived comparisons remain `hmac.Equal`; no `bytes.Equal` or `reflect.DeepEqual` over secret-derived values was introduced.

ADR-0002 is included in this baseline: the exported inert suite markers are removed before v1.0.0, while the package remains single-suite and the wire suite byte remains `0x01` through internal constants.

ADR-0009 is included in this baseline: role-local `Input` maps `SelfID` and `PeerID` per role before building CI, `LocalAssociatedData` names caller-local associated data, and the wire format is preserved. The named manual secret-lifetime audit for this implementation is `docs/adr-0009-secret-lifetime-audit.md`.

Issue #80 is included in this baseline: responder-side decoded-share reuse avoids reparsing `Ya` after role-aware peer-share prevalidation while preserving validate-before-randomness, ADR-0003 peer-share sentinels, post-multiply neutral-element defense, public API, and wire behavior.

## Residual Risk

External review of package-owned CI/framing/profile choices remains open.
Independent cryptographic review remains required before any production-ready
claim.

Repeat this audit if protocol code, parser/framing code, package-profile docs,
dependencies, toolchain, or the targeted CPace draft revision changes before a
release tag.

## Post-baseline correction note - 2026-07-02

A documentation-accuracy review found that the scalar-sampling analysis in
`docs/security-assessment.md`, `docs/spec-matrix.md`, and the `crypto.go`
sampling comment described a reachable `~2^-125` canonical-decode rejection
window `[L, 2^252)`. That interval is empty: masking bounds every sample below
`2^252 < L`, so `SetCanonicalBytes` cannot reject a masked sample, the
canonical-decode retry branch is unreachable defense-in-depth, and the only
reachable retry is the all-zero masked sample at `~2^-252` per attempt. The
audited code behavior is unchanged and remains correct; this baseline's
no-drift conclusion stands for behavior, but its endorsement of the erroneous
probability description is corrected as of this note. No package behavior,
public API, or dependency changed.
