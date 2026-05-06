# Spec Matrix

Target: `draft-irtf-cfrg-cpace-21`, published April 23, 2026.

Audit: reviewed against implementation commit
`4a8f629e59f0cc5c8f9351abacfa511fe6e4f441` on 2026-05-06; see
`docs/security-spec-audit.md`.

| Draft requirement | Implementation | Tests |
| --- | --- | --- |
| Use `transcript_ir(Ya,ADa,Yb,ADb)` in initiator-responder mode | `transcriptIR` in `strings.go`; symmetric transcript kept internal for vectors | `TestStringUtilitiesDraftVectors`, `TestRistrettoDraft21Vectors` |
| Use LEB128 length-value encoding for `prepend_len` and `lv_cat` | `strings.go`, `framing.go` | `TestStringUtilitiesDraftVectors`, malformed parser tests |
| `generator_string(DSI,PRS,CI,sid,s_in_bytes)` with SHA-512 block size 128 | `generatorString` | `TestRistrettoDraft21Vectors` |
| `G_Ristretto255.DSI = "CPaceRistretto255"` | `dsiRistretto255` | `TestRistrettoDraft21Vectors` |
| Hash generator string to 64 bytes and use Ristretto element derivation | `calculateGenerator` | `TestRistrettoDraft21Vectors` |
| Sample scalars by masking bits above group size 252 | `sampleScalar`; this keeps the draft-21 Ristretto255 recommendation rather than the draft's allowed uniform-sampling alternative with zero rejection/retry | `TestScalarSamplingMasksDraftRistrettoBits`, public exchange tests; release needs statistical review |
| `scalar_mult_vfy` aborts on decode failure or neutral output | `scalarMultVFY`, protocol abort paths | `TestScalarMultVFYDraftInvalidVectors`, `TestProtocolAbortsOnInvalidRistrettoEncoding`, embedded B.3.11 JSON fixture |
| Compute ISK from `lv_cat(DSI_ISK,sid,K)||transcript_ir(...)` | `deriveISK` | `TestRistrettoDraft21Vectors`; embedded B.3.9 JSON fixture pinned to the draft-decoded SHA-256 |
| Add explicit key confirmation with MAC key derived from ISK | `confirmationTag`, `Initiator.Finish`, `Responder.Finish`; tags remain draft-compatible with no package-added role labels | confirmed exchange and mismatch tests |
| Integrate initiator and responder identifiers into CI with role binding | `buildCI` | mismatch tests; CI format documented as package-owned |
| Abort on invalid/weak points | `Respond` prevalidates message A shares before responder scalar sampling as implementation hardening; `scalarMultVFY` remains the final protocol check in `Respond` and `Initiator.Finish` | invalid Ristretto tests |

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
- documented application integration guidance for any outer PAKE/version
  negotiation and downgrade protection
