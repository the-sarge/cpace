# Spec Matrix

Target: `draft-irtf-cfrg-cpace-21`, published April 23, 2026.

Audit: the original matrix was reviewed against implementation commit `2e09774f171dde8c62763d6e35a258b0fef88801` on 2026-05-08. The current security/spec audit evidence baseline is indexed in `docs/evidence-baseline.md`; `docs/security-spec-audit.md` records the `933ece246e6170b11e838395bf36f852cba0cd02` audit and post-baseline caveats. The `scalar_mult_vfy` and invalid/weak-point rows were amended 2026-06-11 for ADR-0003 (post-audit-baseline; see *Post-Baseline Changes* in `docs/security-spec-audit.md`).

| Draft requirement | Implementation | Tests |
| --- | --- | --- |
| Use `transcript_ir(Ya,ADa,Yb,ADb)` in initiator-responder mode | `transcriptIR` in `strings.go`; symmetric transcript kept internal for vectors | `TestStringUtilitiesDraftVectors`, `TestRistrettoDraft21Vectors` |
| Use LEB128 length-value encoding for `prepend_len` and `lv_cat` | `strings.go`, `framing.go` | `TestStringUtilitiesDraftVectors`, malformed parser tests |
| `generator_string(DSI,PRS,CI,sid,s_in_bytes)` with SHA-512 block size 128 | `generatorString` | `TestRistrettoDraft21Vectors` |
| `G_Ristretto255.DSI = "CPaceRistretto255"` | `dsiRistretto255` | `TestRistrettoDraft21Vectors` |
| Hash generator string to 64 bytes and use Ristretto element derivation | `calculateGenerator` | `TestRistrettoDraft21Vectors` |
| Sample scalars by masking bits above group size 252 | `sampleScalar` implements draft §8.3 bit-masking and adds defense-in-depth retries for the zero scalar and the (~2^-125) canonical-decode rejection window where the masked value falls in `[L, 2^252)`. Zero-rejection and canonical-rejection retries are conservative hardening beyond the draft's Ristretto255 recommendation; probability of exhausting `maxScalarTries=128` retries under `crypto/rand.Reader` is negligible. | `TestScalarSamplingMasksDraftRistrettoBits`, `TestScalarSamplingRejectsRepeatedZero`, public exchange tests |
| `scalar_mult_vfy` aborts on decode failure or neutral output | `scalarMultVFY`, protocol abort paths; internally returns nil plus an `ErrAbort`-wrapped error instead of the draft's function-level neutral-element return — an intentional internal-only divergence with identical abort behavior (ADR-0003) | `TestScalarMultVFYDraftInvalidVectors`, `TestProtocolAbortsOnInvalidRistrettoEncoding`, `TestScalarMultVFYPostMultiplyIdentityDefense`, embedded B.3.11 JSON fixture |
| Compute ISK from `lv_cat(DSI_ISK,sid,K)||transcript_ir(...)` | `deriveISK` | `TestRistrettoDraft21Vectors`; embedded B.3.9 JSON fixture pinned to the draft-decoded SHA-256 |
| Add explicit key confirmation with MAC key derived from ISK | `confirmationTag`, `Initiator.Finish`, `Responder.Finish`; tags remain draft-compatible with no package-added role labels | confirmed exchange and mismatch tests |
| Integrate initiator and responder identifiers into CI with role binding | `buildCI` | mismatch tests; CI format documented as package-owned |
| Abort on invalid/weak points | `Respond` prevalidates message A shares before responder scalar sampling as implementation hardening; `scalarMultVFY` remains the final protocol check in `Respond` and `Initiator.Finish`; rejections wrap `ErrAbort` plus `ErrPeerShareEncoding` or `ErrPeerShareIdentity` with role context | invalid Ristretto tests, `TestPeerShareErrorsWrapErrAbort`, `TestPeerShareEncodingRejection`, `TestPeerShareIdentityRejection` |

Package-owned profile and extensions:

| Package behavior | Status | Tests |
| --- | --- | --- |
| `cpace-go` CI construction from draft version, suite, role labels, identities, and context | Package profile over draft-21 CI input, not a generic raw-CI interface | transcript-locking mismatch tests |
| Binary wire framing with format byte `0xc1`, suite byte, role byte, and draft LEB128 fields | Package-owned application framing | `TestWireFormatPrefixByte`, parser tests |
| Per-field size caps for package-owned config and wire fields | Password and IDs 4 KiB; context and session ID 1 KiB; associated data 64 KiB; public shares and tags exact-size decoded | `TestConfigFieldSizeLimits`, `TestMessageParserFieldSizeLimits` |
| `Session.Export` using HKDF-SHA512 over confirmed ISK | Package extension following the draft recommendation to process ISK with a KDF | `TestConfirmedExchangeAndExport`, example |
| `Session.TranscriptID` as draft `CPaceSidOutput` | Public accessor for draft optional session identifier output; not a complete channel binding for outer negotiation | vector and exchange tests |

Known gaps before a production release:

- independent cryptographic review
- external review of package-owned message framing, CI, and profile choices
