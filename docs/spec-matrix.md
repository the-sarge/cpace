# Spec Matrix

Target: `draft-irtf-cfrg-cpace-21`, published April 23, 2026.

Audit: the original matrix was reviewed against implementation commit `2e09774f171dde8c62763d6e35a258b0fef88801` on 2026-05-08. The current evidence status is indexed in `docs/evidence-baseline.md`; `docs/security-spec-audit.md` records the latest completed audit at frozen `v0.1.3` source commit `1f19e278112fa037890848ed6c086addeffdca4e` under Go 1.26.5, including the Go 1.26.4 to Go 1.26.5 vector comparison. The `scalar_mult_vfy` and invalid/weak-point rows include ADR-0003 and issue #80's responder-side decoded-share reuse.

| Draft requirement | Implementation | Tests |
| --- | --- | --- |
| Use `transcript_ir(Ya,ADa,Yb,ADb)` in initiator-responder mode | `newIRTranscript` / `irTranscript` in `transcript.go`; symmetric/OC transcript helpers are test-vector-only via `testTranscriptOC` in `strings_test.go` | `TestStringUtilitiesDraftVectors`, `TestIRTranscriptDraftVectorFlow`, `TestRistrettoDraft21Vectors` |
| Use LEB128 length-value encoding for `prepend_len` and `lv_cat` | `length_value.go`, `framing.go` | `TestStringUtilitiesDraftVectors`, `TestLengthValueEncodingBoundaries`, malformed parser tests |
| `generator_string(DSI,PRS,CI,sid,s_in_bytes)` with SHA-512 block size 128 | `generatorString` | `TestRistrettoDraft21Vectors` |
| `G_Ristretto255.DSI = "CPaceRistretto255"` | `dsiRistretto255` | `TestRistrettoDraft21Vectors` |
| Hash generator string to 64 bytes and use Ristretto element derivation | `calculateGenerator` | `TestRistrettoDraft21Vectors` |
| Sample scalars by masking bits above group size 252 | `sampleScalar` implements draft §8.3 bit-masking and adds a defense-in-depth retry loop. The only reachable retry is the all-zero masked scalar (`~2^-252` per attempt); the canonical-decode branch is unreachable because masking bounds every sample below `2^252 < L`, and it is kept as hardening only. Zero rejection is conservative hardening beyond the draft's Ristretto255 recommendation; exhausting `maxScalarTries=128` retries under `crypto/rand.Reader` is negligible. | `TestScalarSamplingMasksDraftRistrettoBits`, `TestScalarSamplingRejectsRepeatedZero`, public exchange tests |
| `scalar_mult_vfy` aborts on decode failure or neutral output | `scalarMultVFY` keeps the draft-shaped encoded-byte helper for vector/spec traceability; `scalarMultVFYElement` applies the same multiplication and neutral-output defense after responder prevalidation has already decoded `Ya`. Both paths return nil plus an `ErrAbort`-wrapped error instead of the draft's function-level neutral-element return — an intentional internal-only divergence with identical abort behavior (ADR-0003). | `TestScalarMultVFYDraftInvalidVectors`, `TestScalarMultVFYElementMatchesEncodedPeerShare`, `TestProtocolAbortsOnInvalidRistrettoEncoding`, `TestScalarMultVFYPostMultiplyIdentityDefense`, embedded B.3.11 JSON fixture |
| Compute ISK from `lv_cat(DSI_ISK,sid,K)||transcript_ir(...)` | `deriveISK` | `TestRistrettoDraft21Vectors`; embedded B.3.9 JSON fixture pinned to the draft-decoded SHA-256 |
| Add explicit key confirmation with MAC key derived from ISK | `confirmationTag`, `Initiator.Finish`, `Responder.Finish`; tags remain draft-compatible with no package-added role labels | confirmed exchange and mismatch tests |
| Integrate initiator and responder identifiers into CI with role binding | `buildCI` | mismatch tests; CI format documented as package-owned |
| Abort on invalid/weak points | `Respond` prevalidates message A shares before responder scalar sampling as implementation hardening, then reuses that decoded non-identity element for responder Diffie-Hellman; `Initiator.Finish` keeps the encoded-byte `scalarMultVFY` check for responder shares. Rejections wrap `ErrAbort` plus `ErrPeerShareEncoding` or `ErrPeerShareIdentity` with role context. | invalid Ristretto tests, `TestPeerShareErrorsWrapErrAbort`, `TestPeerShareRoleDecodeSharedSecretAddsRoleContext`, `TestPeerShareEncodingRejection`, `TestPeerShareIdentityRejection` |

Package-owned profile and extensions:

| Package behavior | Status | Tests |
| --- | --- | --- |
| Single `CPACE-RISTR255-SHA512` suite, exposed as `DraftVersion` plus opaque package behavior rather than a suite-selection API | The exported inert `Suite` markers are removed; package-owned framing keeps fixed suite byte `0x01`, and callers have no in-package negotiation surface | `TestBuildCIWireStability`, `TestWireFormatPrefixByte`, public API inventory in the exact-candidate transcript |
| Role-local `Input` maps `SelfID`, `PeerID`, and `LocalAssociatedData` into fixed initiator-responder ordering | Password, Context, and SessionID are shared session facts; caller input is copied, capped, and consumed per call, with empty SessionID requiring explicit compatibility opt-in | `TestInputValidation`, `TestInputFieldSizeLimits`, `TestProtocolAllowsEmptyLocalAssociatedData`, `TestRoleLocalIdentityReversalFailsConfirmation`, transcript-locking mismatch tests |
| `cpace-go` CI construction from draft version, suite, role labels, identities, and context | Package profile over draft-21 CI input, not a generic raw-CI interface | transcript-locking mismatch tests |
| Binary wire framing with format byte `0xc1`, suite byte, role byte, and draft LEB128 fields | Package-owned application framing | `TestWireFormatPrefixByte`, parser tests |
| Per-field size caps for package-owned input and wire fields | Password and IDs 4 KiB; context and session ID 1 KiB; local associated data 64 KiB; public shares and tags exact-size decoded | `TestInputFieldSizeLimits`, `TestMessageFramingCatalogueRejectsFieldLimits`, `TestMessageFramingCatalogueOwnsFieldLengthAcceptance` |
| `Initiator` and `Responder` are copied single-use values whose constructed copies share terminal state; `Finish` and `Close` are terminal | `Finish` consumes and clears persistent core secrets on success or consuming failure; nil-safe/idempotent `Close` clears abandoned state, later `Finish` returns `ErrStateUsed`, and caller-fabricated zero values remain `ErrInvalidInput` | `TestSingleUseStateCloseNilAndZeroValue`, `TestSingleUseStateCloseCleansAbandonedState`, `TestSingleUseStateCloseAfterFinish`, `TestSingleUseStateCopiesShareTerminalState`, `TestSingleUseStateConcurrentFinishClose`, cleanup tests |
| `Session.Export` using HKDF-SHA512 over confirmed ISK | Package extension following the draft recommendation to process ISK with a KDF | `TestConfirmedExchangeAndExport`, example |
| `Session.Close`, `PeerID`, and `PeerAssociatedData` lifecycle and metadata | `Close` is nil-safe and idempotent, constructed Session copies share close state, zero-value Sessions remain invalid, and copied peer metadata stays available after close | `TestSessionClose`, `TestSessionValueCopiesShareCloseState`, `TestNilReceiverMethods`, `TestSessionPeerMetadata` |
| `Session.TranscriptID` as draft `CPaceSidOutput` | Public accessor for draft optional session identifier output; not a complete channel binding for outer negotiation | vector and exchange tests |

Known gaps before a production release:

- independent cryptographic review
- external review of package-owned message framing, CI, and profile choices
