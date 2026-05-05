# Spec Matrix

Target: `draft-irtf-cfrg-cpace-21`, published April 23, 2026.

| Draft requirement | Implementation | Tests |
| --- | --- | --- |
| Use `transcript_ir(Ya,ADa,Yb,ADb)` in initiator-responder mode | `transcriptIR` in `strings.go`; symmetric transcript kept internal for vectors | `TestStringUtilitiesDraftVectors`, `TestRistrettoDraft21Vectors` |
| Use LEB128 length-value encoding for `prepend_len` and `lv_cat` | `strings.go`, `framing.go` | `TestStringUtilitiesDraftVectors`, malformed parser tests |
| `generator_string(DSI,PRS,CI,sid,s_in_bytes)` with SHA-512 block size 128 | `generatorString` | `TestRistrettoDraft21Vectors` |
| `G_Ristretto255.DSI = "CPaceRistretto255"` | `dsiRistretto255` | `TestRistrettoDraft21Vectors` |
| Hash generator string to 64 bytes and use Ristretto element derivation | `calculateGenerator` | `TestRistrettoDraft21Vectors` |
| Sample scalars by masking bits above group size 252 | `sampleScalar` | public exchange tests; release needs statistical review |
| `scalar_mult_vfy` aborts on decode failure or neutral output | `scalarMultVFY`, protocol abort paths | `TestScalarMultVFYDraftInvalidVectors`, `TestProtocolAbortsOnInvalidRistrettoEncoding` |
| Compute ISK from `lv_cat(DSI_ISK,sid,K)||transcript_ir(...)` | `deriveISK` | `TestRistrettoDraft21Vectors`; embedded JSON vector loader |
| Add explicit key confirmation with MAC key derived from ISK | `confirmationTag`, `Initiator.Finish`, `Responder.Finish` | confirmed exchange and mismatch tests |
| Integrate initiator and responder identifiers into CI with role binding | `buildCI` | mismatch tests; CI format documented as package-owned |
| Abort on invalid/weak points | `Respond`, `Initiator.Finish` | invalid Ristretto tests |

Known gaps before a production release:

- independent cryptographic review
- broader fuzzing campaign beyond smoke targets
- dependency review sign-off
- external review of package-owned message framing
