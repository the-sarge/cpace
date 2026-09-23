# Key-confirmation reflection review (#308)

Date: 2026-09-23. Baseline: `87528342d68d26db4a5a69386ae5647baf7d05ed`. Scope: this package's ordered Ristretto255/SHA-512 A/B/C exchange. This is a bounded implementation review and regression investigation, not independent cryptographic review or a production-readiness claim.

## Conclusion and disposition

The baseline has a conditional false-confirmation path: with equal public shares and equal associated data, the responder accepts its own B tag as C before the initiator has called Finish. Deterministic internal fixtures reproduce this with distinct configured identities and a nonempty SessionID. No feasible network attack producing that condition with fresh production randomness was established. The upstream single-run initiator attack does not transfer directly because this package withholds the initiator's tag until responder verification succeeds.

Nevertheless, confirmation must distinguish a peer's tag from the local tag. The concrete equal-message acceptance finding justifies the narrow security-fix exception to the repository's behavior freeze: both Finish paths now reject an otherwise correct remote tag that equals the local confirmation tag, using `hmac.Equal` and the existing `ErrConfirmationFailed`. This closes the demonstrated conditional path rather than relying on share inequality as an implicit precondition. No public API, MAC role label, dependency, randomness interface, or wire encoding is added. This is a finding-driven correction within the existing confirmation boundary, not a broader architectural or package-policy reopen; no new ADR is introduced.

## Upstream status

On the review date, [upstream PR #21](https://github.com/cfrg/draft-irtf-cfrg-cpace/pull/21) remained open at `87b1aef16ee176511948f1b629424bca910732b1`, and [Datatracker](https://datatracker.ietf.org/doc/draft-irtf-cfrg-cpace/) still listed draft-21 as the latest revision. The PR patch rejects identical protocol messages; the [September 16 editor proposal](https://github.com/cfrg/draft-irtf-cfrg-cpace/pull/21#issuecomment-5700964252) instead describes rejecting equal reflected tags. The [September 22 update](https://github.com/cfrg/draft-irtf-cfrg-cpace/pull/21#issuecomment-5774999640) reports shepherd/chair approval to incorporate a clarification during RFC Editor review. There is not yet final published replacement wording to claim conformance with. Recheck it when updating the draft/RFC target.

[Draft-21 §10.4](https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21#section-10.4) offers optional confirmation guidance that also permits parallel tag emission. This package mandates confirmation with a fixed order. The remediation follows the editor's equal-tag condition, but its justification is the reproduced package behavior, not the upstream editorial classification.

## Attacker and exchange ordering

The attacker can observe, replace, reframe, replay, reorder, and relay wire messages, including their public shares, AD, and tags. It does not initially know the password, ephemeral scalars, shared secret K, ISK, or MAC key, and cannot control `crypto/rand.Reader` or read process memory. Password guessing, compromised randomness, and compromise of a password-bearing peer are not made safe by this check.

| Stage | Package action | Tag available to a network attacker |
| --- | --- | --- |
| `Start` | Samples the initiator scalar and sends A containing sid, Ya, ADa | None |
| `Respond` | Validates A, samples its own scalar, derives K/ISK, sends B containing Yb, ADb, Tb | Tb; no authenticated Session |
| `Initiator.Finish` | Derives K/ISK, verifies Tb, rejects equal Ta/Tb, then returns C containing Ta and a Session | Ta only after successful verification and reflection check |
| `Responder.Finish` | Verifies Ta and rejects equality with its own Tb before returning a Session | No new tag |

For the initiator attack, replacing Yb with Ya is easy: the point is valid. But the resulting K is the initiator's scalar multiplied by its own share, and no tag for that resulting session has been released by `Start`. A fabricated B tag fails cryptographic confirmation. Copying a real Tb from an ordinary responder also fails after replacing B's share/AD, because that Tb belongs to the responder's original K and ordered transcript. The initiator exposes no tag oracle on failure: C and Session are nil, its scalar is cleared, and a second Finish returns `ErrStateUsed`.

For the responder attack, Tb is available before C and can be copied into a correctly framed C without knowing a secret. At baseline this succeeds exactly when it is also the expected Ta. Equal shares and equal AD guarantee that equality: both tags use the same MAC key and identical length-value input. With distinct messages, equality would require a MAC collision; ordinary reflected Tb fails. The fix rejects equality even when the copied tag is otherwise correct.

## Why the bindings do and do not help

`confirmationTag` in `crypto.go` hashes `CPaceMac || sid || ISK` for its key and authenticates `lvCat(y, ad)`. `newIRTranscript` in `transcript.go` binds Ya, ADa, Yb, ADb in fixed order into ISK. There is no role-specific prefix in the tag input. Thus, within one session, equal `(y, ad)` pairs necessarily yield equal tags; distinct CI identities or a fresh nonempty sid cannot change that fact.

`buildCI` binds ordered, role-labelled identities, context, suite, and draft version into the common generator. This protects agreement on those configured facts, including role-local identity reversal. It does not turn identical within-session MAC inputs into different ones. Service-specific identities and appropriate application role assignment remain necessary for avoiding confused relays.

AD is authenticated as transmitted, not required to differ or to equal an application's expected remote value. Empty and equal nonempty AD are valid inputs. Differing AD distinguishes the tag inputs even for equal shares, and an honest exchange in that case remains supported. Applications must still validate expected peer AD as described in [integration guidance](integration-guidance.md).

Framing checks reject verbatim A-as-B or B-as-C wire bytes, but an attacker can copy fields into the proper role envelope. The tests deliberately use valid B/C framing and require `ErrConfirmationFailed`, rather than treating `ErrMessage` as reflection protection. The wire role byte is not a MAC role label.

SessionID is checked against A at the responder and bound into generator derivation, ISK, and confirmation-key derivation. It separates sessions when used freshly but does not distinguish the two tags inside one session. The package does not maintain a global replay cache or enforce cross-session sid uniqueness; the explicit empty-sid compatibility option does not change the equal-tag check.

## Forced equality and concurrent sessions

The equal-share fixture supplies the same nonzero scalar stream to both existing unexported randomness seams with matching password/CI/sid. It proves the conditional behavior, not that an attacker controls the responder scalar in production. A is fixed before the responder draws its scalar. For a nonidentity generator, multiplication by the scalar is injective, and this implementation samples uniformly from nonzero 252-bit values under ideal independent randomness. For a fixed A in that set, the per-attempt equal-share probability is `1/(2^252 - 1)`. This is a reasoning bound under the stated model, not a measurement from the production-randomness tests or a full security proof.

An attacker can collect authentic responder tags across concurrent sessions. Relaying Yb from one response as A to another responder does not make the next responder reuse Yb: it draws another scalar and binds the new ordered transcript. Likewise, copying B or C tags across independently sampled exchanges does not normally preserve K/ISK and their transcript binding. A regression replays authentic tags between two exchanges with identical caller inputs, deliberately including the same sid, and observes rejection. That bounded test does not authorize sid reuse or exhaust concurrent-session schedules.

Transparent relay of an actual exchange between password holders is different from manufacturing confirmation without a matching peer. This patch is not channel authentication or a general relay defense. [Draft-21 §10.1.2](https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21#section-10.1.2) discusses application measures for concurrent loopback; [§10.9](https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-cpace-21#section-10.9) requires fresh scalars and recommends unique sid values. Applications must preserve those assumptions and bind their intended roles and outer protocol.

## Compatibility and tests

The new observable result is `ErrConfirmationFailed` for otherwise valid equal-tag exchanges, including the negligible honest equal-share/equal-AD collision case. It is not `ErrAbort` (the share can be valid) or `ErrMessage` (the framing can be valid). The existing error identity and text are reused. `Respond` still emits B without claiming authentication; neither Finish returns a Session on rejection, and the initiator releases no C. Both states are consumed and persistent secrets cleared by the existing terminal lifecycle. MAC derivation, transcript bytes, draft vectors, public signatures, and the ordering of invalid-peer-share validation remain unchanged. The responder recomputes one local tag after successful peer-tag verification rather than storing another value in its core.

`reflection_test.go` covers:

- `TestReflectionResponderRejectsOwnTag`: authentic Tb copied into C before initiator Finish; empty/equal/different AD, production randomness and forced equal shares, nil Session, state consumption and cleanup.
- `TestReflectionInitiatorRejectsReflectedShareAndAuthenticTag`: A fields reframed as B with a genuine Tb; the forced equal-message case is also the genuine B, demonstrating that no ambiguous confirmation escapes.
- `TestReflectionInitiatorWithoutTagOracle`: correctly framed reflected share with a guessed mandatory tag; no C or Session escapes and state is spent.
- `TestReflectionHonestExchangeControls`: all three AD cases with production randomness, plus forced equal shares with differing AD; both sides export matching keys.
- `TestReflectionRejectsTagsFromAnotherExchange`: authentic B/C tags replayed across independent exchanges, with failure and cleanup assertions.

The responder and initiator tests were each run before their corresponding check was added. Both failed only in the empty/equal-nonempty AD cases with forced equal shares, returning a Session (and C on the initiator). They pass with the checks. These are durable rejection regressions, not assertions that baseline acceptance should remain supported. Existing honest-exchange and draft-vector tests remain part of validation.

## Evidence limits

This security-relevant change invalidates the applicability of the older commit-pinned dependency/SAST, Capslock, security/spec/vector, and paired long-fuzz release evidence to this candidate. [Evidence baseline](evidence-baseline.md) preserves the historical records and the refresh requirement. Local tests and automated code review are not an independent cryptographic assessment, a multi-host long-fuzz campaign, or a release-evidence refresh. No stronger release claim is made by this change. Exact source-candidate validation results are recorded separately from the source commit to avoid a self-referential SHA.

[The validation bundle](evidence/308-reflection-20260923/README.md) records source candidate `d2db6eed5923395bf0dfb57e73275ea0a1506c44`, successful focused regressions/vectors, full `task check`, module verification, and the Standards/Spec reviews. It includes raw command output and checksums; the follow-up evidence commit changes documentation only.
