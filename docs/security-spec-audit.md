# Security/Spec Audit

Date: 2026-06-11

Target module: `github.com/the-sarge/cpace`

Implementation baseline: `933ece246e6170b11e838395bf36f852cba0cd02`

ADR-0001 interim addendum: 2026-06-12 at `7aa79e4a40304a14610df36d0bd906fd6c7e3a24`

Documentation/evidence baseline: the merge commit containing this file.

Toolchain: Go 1.26.4

Evidence transcript: `docs/evidence/go1264-20260611/local-analysis.log`

Draft source: `draft-irtf-cfrg-cpace-21`
(`https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21`)

## Scope

This audit checked `docs/security-assessment.md` and `docs/spec-matrix.md` (as
updated by PR #73, which shipped its own documentation updates alongside its
code changes) against the implementation baseline, tests, refreshed Go 1.26.4
release evidence, and the draft-21 text. It is a documentation and conformance
audit, not an independent cryptographic review. This is a project-side
self-audit, distinct from independent cryptographic review or external review.

The audit covered:

- implemented suite and protocol mode;
- CPace transcript, generator, ISK, scalar sampling, and confirmation behavior;
- package-owned CI construction, wire framing, and per-field caps;
- session lifecycle, export, peer metadata, and memory-handling claims;
- invalid-share handling and parser rejection behavior;
- test/vector/fuzz/dependency/Capslock evidence referenced by the
  release-readiness docs;
- the go1.26.4 toolchain security release (2026-06-02) impact;
- the PR #73 package-code changes (the safe fixes from the 2026-05-27
  multi-agent review: deferred wipe unification, `sampleScalar` retry,
  protocol-identity test pins, SAST gate workflow).

## Result

No security/spec drift was found at the implementation baseline.

`task check` passes under Go 1.26.4 at the implementation baseline (transcript
records exit 0 with both test lanes green). The clean-worktree evidence
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
- dependency and Capslock evidence references point to the refreshed Go 1.26.4
  evidence documents; the long-fuzz evidence refresh under 1.26.4 is pending
  the paired maintainer-machine campaigns and is tracked separately.

The go1.26.4 release (2026-06-02) is a security release: fixes to
`crypto/x509`, `mime`, and `net/textproto`, plus bug fixes to
`crypto/fips140`, the compiler, and the runtime. CPace does not import the
three patched packages; it does transitively use Go crypto internals including
`crypto/fips140`, so evidence is re-recorded under Go 1.26.4. Separately,
PR #73 merged the safe fixes from the 2026-05-27 multi-agent review into
`api.go`, `crypto.go`, and `session.go` (deferred wipe unification,
`sampleScalar` retry, protocol-identity test pins); those changes were
multi-agent-reviewed, shipped with their own security-assessment and
spec-matrix updates, and pin protocol identity with dedicated tests. At this
baseline `task check` reran the draft/RFC vector assertions. No Go API,
wire/protocol, dependency, or vector behavior change was found.

### Toolchain Vector Stability

For this refresh, the draft/RFC vector assertions were run under **both**
toolchains and recorded in the evidence transcript: `go test -count=1 -run
'Vector' -v ./...` under go1.26.4, and the same command under
`GOTOOLCHAIN=go1.26.3`. All vector tests pass identically under both
(`TestStringUtilitiesDraftVectors`, `TestEmbeddedDraftVectorJSON`,
`TestEmbeddedDraftInvalidVectorJSON`, `TestRistrettoDraft21Vectors`,
`TestScalarMultVFYDraftInvalidVectors`, and the vector-loader fuzz seeds).
This is the first refresh to record the old/new pair the evidence policy asks
for; future toolchain-triggered refreshes should continue the practice.

## Post-Baseline Changes

ADR-0003 (peer-share error semantics, implemented 2026-06-11, PR #78) landed after this audit's implementation baseline. It adds the exported `ErrPeerShareEncoding`/`ErrPeerShareIdentity` sentinels and changes the internal `scalarMultVFY` failure convention to nil plus an `ErrAbort`-wrapped error — an intentional internal-only divergence from draft-21's function-level neutral-element return with no protocol-visible change; the `scalar_mult_vfy` abort behavior audited above is unchanged. Read the "match the documented draft-21 profile" claim with that divergence in mind; the consolidated post-implementation evidence refresh will re-audit at the new baseline.

ADR-0001 (deep CPace core extraction) landed on this feature branch after the audit's implementation baseline. The public API and wire format are unchanged except for the ADR-recorded zero-value hardening: caller-fabricated zero-value `Initiator` and `Responder` shells now return `ErrInvalidInput` without consuming state, closing the previous zero-value responder forged-tag path where `Responder.Finish(encodeMessageC(confirmationTag(nil, nil, nil, nil)))` could return a Session keyed from nil ISK.

Interim verification for the ADR-0001 addendum: `task check` passed after the cleanup-consolidation commit; `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` ran locally on `darwin/arm64` with Go 1.26.4 and Task 3.51.1 from `2026-06-12T04:08:47Z` to `2026-06-12T05:04:52Z`, passing all 14 registered targets with `rc=0`. This is an interim gate only and does not replace the pinned paired long-fuzz evidence or the consolidated Phase 3 exact-candidate refresh.

Manual zeroization audit for the ADR-0001 addendum: `Initiator` and `Responder` are now thin shells with named `core` fields; `initiatorCore` holds the persistent initiator scalar only, and `responderCore` holds the persistent responder ISK plus the public transcript zeroed as hygiene. Neither core has a password, generator, DH point `k`, responder scalar, or initiator finish-local ISK field. `startWithRandom` and `respondWithRandom` retain the normalized-config `wipe()` backstop; core constructors eagerly clear the normalized password after generator derivation; `initiatorCore.finish` defers `clearBytes(isk)` immediately after `deriveISK`; `initiatorCore.clear()` and `responderCore.clear()` are nil-safe, idempotent, and zero-then-nil persistent fields; and both shell `Finish` methods defer `core.clear()` immediately after successful `consume()`, covering parse failure, confirmation failure, and success. Secret-derived comparisons remain `hmac.Equal`; no `bytes.Equal` or `reflect.DeepEqual` over secret-derived values was introduced.

Residual memory risks after the ADR-0001 addendum are unchanged in kind: a single-use state abandoned without `Finish` can retain its core-owned persistent secret until garbage collection; `lvCat`/`prependLen` still create heap intermediates that are not individually cleared, including password material inside `calculateGenerator` and K material inside `deriveISK`; and `hmac.New` retains internal key-pad copies outside package control. These are best-effort-zeroization limitations, not protocol-visible behavior changes.

ADR-0002 (suite API cleanup) landed on this feature branch after the audit's implementation baseline. It removes only the exported inert suite markers (`Suite` and `SuiteCPaceRistretto255SHA512`) before v1.0.0, keeps the package single-suite, and preserves the wire suite byte as `0x01` through internal `currentSuite`/`wireSuite` constants. Interim verification for the ADR-0002 implementation: `task check` passed; `go test ./...`, `go test -race ./...`, and `gosec -tests ./...` passed; and `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` ran locally on `darwin/arm64` with Go 1.26.4 and Task 3.51.1 from `2026-06-12T14:48:13Z` to `2026-06-12T15:44:29Z`, passing all 14 registered targets with `rc=0`. This is an interim gate only and does not replace the pinned paired long-fuzz evidence or the consolidated Phase 3 exact-candidate refresh.

## Residual Risk

External review of package-owned CI/framing/profile choices remains open.
Independent cryptographic review remains required before any production-ready
claim.

Repeat this audit if protocol code, parser/framing code, package-profile docs,
dependencies, toolchain, or the targeted CPace draft revision changes before a
release tag.
