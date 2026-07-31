# Security/Spec Audit

Date: 2026-07-31

Target module: `github.com/the-sarge/cpace`

Implementation baseline: `1f19e278112fa037890848ed6c086addeffdca4e`

Documentation/evidence baseline: the merge commit containing this file.

Toolchain: Go 1.26.5, with vector comparison against Go 1.26.4

Evidence transcript: `docs/evidence/1f19e27-20260731-protocol/protocol-audit.log`

Evidence checksums: `docs/evidence/1f19e27-20260731-protocol/SHA256SUMS`

Baseline status: `docs/evidence-baseline.md` is the current source of truth for whether this pinned audit is fresh for the latest release candidate.

Draft source: `draft-irtf-cfrg-cpace-21` (`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This project-side audit checked `docs/security-assessment.md` and `docs/spec-matrix.md` against the exact frozen candidate, its exported documentation, focused tests, accepted ADRs, the historical Go 1.26.4 audit, and the draft-21 text. It is a documentation and conformance self-audit, not external review or independent cryptographic review.

The audit covered:

- the complete exported API: role-local `Input`, `Start`, `Respond`, `Initiator` and `Responder` lifecycle methods, confirmed `Session` methods, `DraftVersion`, and exported error identities;
- the single `CPACE-RISTR255-SHA512` suite and initiator-responder protocol mode;
- CPace transcript, generator, ISK, scalar sampling, confirmation-tag, and `CPaceSidOutput` behavior;
- package-owned CI construction, Message framing, fixed suite/format/role bytes, LEB128 encoding, and non-configurable per-field caps;
- nil, zero-value, copy, concurrency, terminal-operation, abandoned-state cleanup, confirmation, and Session close/export lifecycle behavior;
- peer-share prevalidation, validation-before-randomness ordering, invalid/identity classification, role-context errors, and the internal nil-plus-error divergence from the draft helper convention;
- caller-input role mapping, copy ownership, empty-session compatibility, local associated data, peer metadata, and outer-negotiation limits;
- the production and evidence-relevant delta from historical audit baseline `f7efa6a963a954952b1ecad3f46530f13799fe89`, including issue #219's internal close-protocol concentration, the scalar-sampling explanation correction, workflow/dependency updates, and the Go 1.26.5 pin;
- exact-candidate dependency evidence, Capslock capability evidence, cross-toolchain vectors, and the freshness limits of the still-historical long-fuzz and release-control lanes.

## Result

No implementation, protocol, Message framing, public API, dependency, or package-profile drift was found at the exact candidate. The audit updated stale historical-baseline wording and expanded the assessment/matrix mapping for the already-shipped Caller input and lifecycle surfaces; these are documentation corrections, not behavior changes.

The raw transcript records a clean detached candidate before and after capture, the full `go doc -all .` public API, the historical-to-candidate commit and source-change inventory, both Go toolchains, Capslock, and focused tests. The public API remains the frozen pre-v1 surface: one role-local `Input`; `Start` and `Respond`; `Initiator.Finish`/`Close`; `Responder.Finish`/`Close`; confirmed `Session.Export`/`Close`/`TranscriptID`/`PeerAssociatedData`/`PeerID`; `DraftVersion`; and the documented error sentinels. No exported `Suite`, caller-supplied randomness, suite negotiation, raw-CI interface, or package-profile configuration was introduced.

The focused Go 1.26.5 audit lane ran 42 top-level tests plus their subtests. It covered successful confirmed exchange and export, peer metadata, Session and single-use-state lifecycle, nil and zero-value behavior, copied-state terminal sharing, concurrent terminal operations, persistent-secret cleanup, role-local input validation and mapping, peer-share rejection/error wrapping, Message framing and cap boundaries, fixed CI/suite/wire bytes, and draft invalid-share behavior. Every selected test passed.

The security assessment and spec matrix now explicitly and accurately describe:

- only `CPACE-RISTR255-SHA512` from draft-21 is implemented, with no suite-selection API and wire suite byte `0x01`;
- only initiator-responder mode is exposed, `Respond` success is not authentication, and a `Session` is returned only after explicit key confirmation;
- role-local `Input` is mapped to fixed initiator-responder CI ordering while password, Context, and SessionID remain shared session facts;
- `Finish` and `Close` are terminal for `Initiator` and `Responder`, constructed value copies share terminal state, abandoned-state `Close` clears persistent core secrets, and successful or consuming-failed `Finish` clears those secrets on every return path;
- `Session.Close` is nil-safe and idempotent, zero-value Sessions remain invalid, confirmed Session copies share close state, and peer metadata remains copied and available after close;
- `transcript_ir`, generator derivation, scalar sampling, Diffie-Hellman, ISK derivation, confirmation tags, and `CPaceSidOutput` match the documented draft-21 profile and fixed test vectors;
- invalid peer shares preserve `ErrAbort`, add the documented encoding/identity sentinels with role context, and are rejected after Message framing size checks, with responder validation before generator derivation or randomness;
- package-owned CI construction, Message framing, non-configurable caps, `Session.Export`, and peer metadata remain package-profile behavior rather than generic draft negotiation surfaces.

### Go 1.26.5 Impact

Go 1.26.5 is the 2026-07-07 security release with fixes in `crypto/tls` and `os` plus compiler, runtime, `go` command, `net`, `os`, and `syscall` bug fixes. The cpace production package does not import `crypto/tls` or `os`; the compiler/runtime/toolchain changes still apply to the built implementation, and reachable HMAC/HKDF paths still enter Go FIPS enforcement internals according to Capslock. Rerunning the vectors, capability analysis, and focused protocol lane under Go 1.26.5 found no changed behavior or capability class.

### Toolchain Vector Stability

The exact candidate ran the same required vector/profile lane under Go 1.26.4 and Go 1.26.5. The lane covers `TestStringUtilitiesDraftVectors`, `TestIRTranscriptDraftVectorFlow`, all four `TestEmbeddedDraft*` fixture checks, `TestRistrettoDraft21Vectors`, `TestCoreDraft21Vectors`, `TestScalarMultVFYDraftInvalidVectors`, `TestBuildCIWireStability`, and `TestWireFormatPrefixByte`.

Both toolchains passed all required tests. After removing only JSON event timestamps, elapsed fields, and textual duration values, each normalized event stream contains 112 lines with SHA-256 `05e966e9b4cc8ea6883d5f8a750cc2d91098a3e59aee365ed852e717067efbd9`; `cmp` reports the normalized outputs bit-identical. Because these tests assert implementation results against pinned draft literals, package-local confirmation-tag goldens, exact CI bytes, and fixed framing bytes, the result records bit identity for the exercised draft and package-owned vector surface rather than merely two uncorrelated green test commands.

### Capability Analysis

Capslock v0.3.2, built with Go 1.26.5 and Go tools v0.43.0, reports the same classes and counts as the historical Go 1.26.4 baseline: 11 direct `ARBITRARY_EXECUTION` references through Go FIPS enforcement internals and 13 direct `UNANALYZED` references through `io.ReadFull`. Both are triaged in `docs/capslock-report.md`; no new or changed broad capability was found.

## Current Implementation Notes

ADR-0001 remains satisfied: `Initiator` and `Responder` are thin shells over the CPace core, caller-fabricated zero values return `ErrInvalidInput`, persistent secrets have local core ownership, and secret-derived comparisons use `hmac.Equal`. The issue #219 refactor moved the duplicated claim/idempotency/clear sequence into `singleUseState.closeCore`; the focused lifecycle and secret-snapshot tests confirm the public behavior and cleanup result are unchanged.

ADR-0002 remains satisfied: exported inert suite markers are absent, the package remains single-suite, and package-owned framing still emits suite byte `0x01`.

ADR-0003 remains satisfied: exported peer-share sentinels classify non-canonical and identity inputs while preserving `ErrAbort`, role context, validation-before-randomness, and protocol abort behavior. The internal nil-plus-error convention remains an intentional internal-only divergence from the draft helper return shape.

ADR-0006 and ADR-0008 remain satisfied: nil `Close` calls are successful no-ops, zero-value state is strict, `Finish` and `Close` are shared terminal operations, and Close provides deterministic cleanup for abandoned single-use state.

ADR-0009 remains satisfied: role-local `Input` maps `SelfID` and `PeerID` per role, `LocalAssociatedData` names caller-local associated data, package caps and copy ownership are preserved, and no reusable password-owning validated-input object exists. The named manual secret-lifetime audit remains `docs/adr-0009-secret-lifetime-audit.md`.

The scalar-sampling correction is integrated: masking bounds samples below `2^252 < L`, making canonical-decode rejection unreachable defense-in-depth; the only reachable retry is the all-zero masked sample at approximately `2^-252` per attempt. The implementation behavior did not change.

## Residual Risk

This audit and its evidence are self-review. Passing fixed vectors, focused tests, and static capability analysis does not establish cryptographic correctness, side-channel resistance, or real-world integration safety.

External review of package-owned CI, Message framing, and suite-profile choices remains open. Independent cryptographic review remains required before any production-ready claim. The paired exact-candidate long-fuzz, GitHub alert, tag-ruleset, Scorecard, and signed Release Validation lanes remain separate work and are not implied by this audit.

Repeat this audit if protocol code, parser/framing code, dependencies, toolchain, package-profile behavior or documentation, evidence-sensitive tooling, or the targeted CPace draft revision changes before the release tag.
