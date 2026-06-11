---
status: accepted
date: 2026-06-10
review-runs:
  - 20260609T222806-02c76a2acebe3f891af174a3 # ras consider — accept with revisions; revisions applied via ras fix --decisions, ras verify clean
---

# Peer-share error semantics for `scalarMultVFY`

## Status

**Accepted (2026-06-10).** This ADR captures a v1.0.0 error-API decision surfaced by external code review (item H2). Gated per the project's ADR policy: the `ras consider` run above returned accept-with-revisions; the revisions were applied via a maintainer-decided resolution pass (`ras fix --decisions`) and re-gated, with `ras verify` returning clean (unresolved: []). Evidence trail: PR #66 comments and DEV-JOURNAL cpace.S15. The error-sentinel surface and the internal return shape of `scalarMultVFY` are settled; future reviews should not re-litigate them. One implementation-time clarification was tracked as issue #70 (how call sites rewrap the sentinels) and is resolved by the *Call-site sentinel mapping* subsection under Decision (2026-06-10); it refines the implementation outline and does not reopen the decision.

## Context

`crypto.go:99-109` defines:

```go
func scalarMultVFY(s *ristretto255.Scalar, encoded []byte) ([]byte, bool) {
    p, ok := decodePublicShare(encoded)
    if !ok {
        return clone(identityEncoding), false
    }
    out := ristretto255.NewIdentityElement().ScalarMult(s, p).Bytes()
    if hmac.Equal(out, identityEncoding) {
        return clone(identityEncoding), false
    }
    return out, true
}
```

with `decodePublicShare(encoded)` at `crypto.go:111-124` returning `(*ristretto255.Element, bool)` and rejecting for any of three reasons: wrong length, non-canonical Ristretto255 encoding, or identity-element encoding.

The function collapses **three distinct rejection causes** into a single boolean:

1. **Wrong length** — `len(encoded) != pointSize`. The wire decoder already enforces 32 bytes (`framing.go` `readExactField(pointSize, ...)`), so this branch only fires for an internal-caller bug — a programmer error in framing, not an active attacker. Today it is dead but defensive.
2. **Non-canonical encoding** — `SetCanonicalBytes` rejection. The peer sent 32 bytes that do not decode as a valid Ristretto255 point. This is either a buggy or malicious peer.
3. **Identity element on input** — `p.Bytes() == identityEncoding`. The peer sent the encoded identity. This is almost certainly an active attacker trying to force `K = identity`.

A fourth branch — the post-multiply identity check at `crypto.go:104-107` — is dead in prime-order Ristretto255 (for any non-identity `p` and non-zero scalar `s`, `s·p` is non-identity) but kept as defense-in-depth.

There is one direct `decodePublicShare` caller (`api.go:166-168`) and two `scalarMultVFY` callers (`api.go:180-184`, `api.go:222-226`). `respondWithRandom` deliberately validates the initiator's own share twice: first as an early public-share check, then again inside `scalarMultVFY` when computing `K`. Today those callers wrap the bool as a single `fmt.Errorf("%w: invalid initiator/responder share", ErrAbort)`. An operator seeing `cpace: protocol abort: invalid responder share` in production cannot distinguish a malformed Ristretto255 encoding from an identity-element attack. Both are concerning, but the responses differ.

There is also a separate concern at the function's return value: on every failure path `scalarMultVFY` returns `clone(identityEncoding)` (32 zero bytes) as the first return value, with `ok=false`. Every current caller checks `ok` before using the first return, but the silent-fallback shape is the canonical "safe default that masks failure" anti-pattern — a future `k, _ := scalarMultVFY(...)` would key with an all-zeros shared secret.

The current identity-encoding return deliberately mirrors draft-21's `G.scalar_mult_vfy` neutral-element convention; the draft invalid-vector case named "Invalid Y2" is the all-zero identity encoding. The proposed divergence is internal-only: protocol abort behavior is unchanged.

The error sentinel set in `errors.go:5-34` already exposes `ErrAbort` ("draft abort condition such as an invalid point or neutral-element Diffie-Hellman result"). The proposal does not remove or repurpose `ErrAbort`; it adds finer-grained plain sentinels whose returned errors also wrap `ErrAbort` so existing `errors.Is(err, ErrAbort)` checks remain correct.

The public API is frozen for v1.0.0 unless this review reopens it. Adding new exported error sentinels is API surface expansion and must be done before the freeze.

## Decision

Make two changes to `scalarMultVFY` and its callers, neither of which alters wire format or protocol semantics:

1. **Return `nil` on failure instead of `clone(identityEncoding)`.** No current caller reads the first return when `ok=false`, so this is internally non-breaking. The change removes the spec-shaped, valid-looking 32-byte fallback and makes misuse distinguishable. The loud-failure property comes from the typed error return, not from `nil`: `lvCat` treats `nil` as a zero-length field, `deriveISK` passes `k` only through `lvCat`/`copy`, and `clearBytes` is `nil`-safe. As optional hardening, `deriveISK` may defensively reject `len(k) != pointSize`, but this ADR does not require that second layer.

2. **Add two exported sentinels in `errors.go`** as plain `errors.New` values:
   - `ErrPeerShareEncoding` — wire decoded but `SetCanonicalBytes` rejected (non-canonical Ristretto255).
   - `ErrPeerShareIdentity` — peer share decoded to the identity element.

   Each sentinel's doc comment follows the package precedent set by `ErrEmptySessionID`: the sentinel value itself is plain, while the comment states "The returned error also wraps ErrAbort." Uniform sentinel semantics across the package's exported error set outweighs the structural guarantee of pre-wrapped sentinels, and the forgotten-wrap risk is covered by a mandated public-API regression test.

Change the signature of `scalarMultVFY` (and `decodePublicShare`) to return `([]byte, error)` / `(*ristretto255.Element, error)` with all four failure paths specified:

- Wrong length returns an internal defensive error, `fmt.Errorf("%w: invalid peer share length", ErrAbort)`, with no exported peer-share sentinel. Malformed wire lengths surface as `ErrMessage` at framing because the wire decoder enforces `pointSize` before any share reaches `decodePublicShare`.
- Non-canonical encoding returns `nil, fmt.Errorf("%w: %w", ErrAbort, ErrPeerShareEncoding)`.
- Identity-element input returns `nil, fmt.Errorf("%w: %w", ErrAbort, ErrPeerShareIdentity)`.
- The post-multiply identity check at `crypto.go:104-107` stays in place as defense-in-depth and returns `nil, fmt.Errorf("%w: neutral-element shared secret", ErrAbort)` with no new exported sentinel. The branch is unreachable for prime-order Ristretto255, so exported surface is not warranted.

`ErrAbort` is retained. The returned errors use exactly one `ErrAbort` wrap plus the plain sentinel where one exists, in the same style as the existing `fmt.Errorf("%w: %w", ErrInvalidInput, ErrEmptySessionID)` layering in `api.go`. API call sites must retain role-context wrapping, for example `fmt.Errorf("%w: invalid initiator share: %w", ErrAbort, ErrPeerShareIdentity)` and `fmt.Errorf("%w: invalid responder share: %w", ErrAbort, ErrPeerShareEncoding)`, so the finer sentinels do not regress initiator-vs-responder triage. The resulting public error-string shape is `cpace: protocol abort: invalid initiator share: cpace: peer share identity` (or the responder/encoding variant), with no duplicated `cpace: protocol abort` prefix.

### Call-site sentinel mapping (clarification, 2026-06-10 — issue #70)

The call-site examples above are produced by **rewrapping the plain sentinel, never the helper's returned error**. The helper's error already carries its one `ErrAbort` wrap, so a call site that `%w`-wraps it while adding `ErrAbort` role context would produce the duplicated `cpace: protocol abort` prefix this Decision forbids. Concretely, a call site dispatches on the returned error and handles exactly three cases:

- `errors.Is(err, ErrPeerShareEncoding)` → mint a fresh error wrapping the **plain sentinel** with role context — `fmt.Errorf("%w: invalid initiator share: %w", ErrAbort, ErrPeerShareEncoding)` (or the responder variant) — and discard the helper's error value.
- `errors.Is(err, ErrPeerShareIdentity)` → same, with `ErrPeerShareIdentity`.
- **Default (non-sentinel branches)** — the wrong-length defensive error and the post-multiply neutral-element error — propagate the helper's error **unchanged**. It is already `ErrAbort`-wrapped, both branches are defensive and unreachable from the wire (framing enforces `pointSize`; prime-order Ristretto255 makes the post-multiply identity impossible), and deliberately adding no role context keeps the single-wrap rule trivially true while preserving the precise defensive diagnostic.

This dispatch is the only sanctioned shape: detect-and-rewrap for the two sentinels, pass-through for everything else.

## Acceptance criteria

Multi-agent review concurrence on this ADR moves it proposed -> accepted (the decision is ratified at review time). The acceptance criteria below are implementation-verification gates: they bind the implementing change and must all be satisfied before v1.0.0 is tagged - not before this ADR is accepted.

- **New sentinels exist** at `errors.go` as plain `errors.New` values for `ErrPeerShareEncoding` and `ErrPeerShareIdentity`, with doc comments stating "The returned error also wraps ErrAbort."
- **`scalarMultVFY` returns nil-on-failure** with a typed error; no caller path relies on the first return value when the error is non-nil.
- **`TestPeerShareErrorsWrapErrAbort` exercises every production return path through the public API** (`Respond`/`Finish`), not constructed errors, and asserts `errors.Is(err, ErrAbort)`, the appropriate peer-share sentinel where one exists, and the preserved `invalid initiator share` / `invalid responder share` role context.
- **Each exported sentinel is reachable from at least one test** — `TestPeerShareEncodingRejection` and `TestPeerShareIdentityRejection`.
- **The internal length defense is tested without pretending it is wire-reachable.** Use a direct internal-helper call for the wrong-length branch and assert an `ErrAbort`-wrapped internal error with no peer-share sentinel; malformed wire lengths continue to surface as `ErrMessage` from framing.
- **The post-multiply identity branch has an `ErrAbort` guarantee.** A narrow internal test hook, or documented unreachability if no hook is acceptable, verifies that the branch's error satisfies `errors.Is(err, ErrAbort)` and does not introduce a fourth exported sentinel.
- **No protocol-visible change.** Black-box protocol-level tests, including full exchanges and wire-level rejection, pass unchanged. `FuzzScalarMultVFY`, `TestScalarMultVFYDraftInvalidVectors`, and the `TestDraftVectors` call sites are explicitly updated to the new `([]byte, error)` signature while preserving their invariants: success requires `len(out) == pointSize`, non-identity output, and fixture match; failure requires `out == nil`, the correct sentinel where applicable, and `errors.Is(err, ErrAbort)`.
- **Wire format unchanged.** No new bytes on the wire, no new acceptance criteria for incoming messages.

## Considered options

- **A — Add sentinels, return nil on failure (recommended).** Resolves both the silent-fallback shape and the operator-triage information loss. Costs two new exported sentinels and a small refactor of `scalarMultVFY`'s return type. Backward-compatible at the `errors.Is(err, ErrAbort)` level.

- **B — Keep the `bool`, just change the first return to nil.** Removes the silent-fallback shape without expanding the error sentinel surface. Cheaper, but does not resolve the operator-triage problem and leaves identity-element submissions indistinguishable from peer encoding bugs in logs.

- **C — Add sentinels but keep the `bool` and the identity-encoding fallback.** The error sentinels exist for the call-site `fmt.Errorf` wrap, but the internal helper stays unchanged. Marginal improvement over status quo. Does not fix the silent-fallback risk for future callers.

- **D — Status quo.** Ship v1.0.0 with the current `(bytes, bool)` shape. `ErrAbort` continues to be the only signal callers get. Documented as a known limitation in `docs/integration-guidance.md`.

## Consequences

- **Option A (recommended):**
  - Two new exported error sentinels become part of the v1.0.0 surface.
  - Internal refactor of `scalarMultVFY` signature (unexported, no external impact).
  - Operators triaging local production logs and metrics can distinguish non-canonical public-share encodings from identity-element submissions.
  - Returning `nil` removes the valid-looking all-zero fallback; future callers still get the loud signal from the typed error return, not from any panic property of `nil`.
  - The finer error granularity is not a secret-dependent oracle: all three `decodePublicShare` rejection causes are functions only of the encoded public wire bytes, and no secret is an input. The only secret-adjacent branch, the post-multiply identity check, is unreachable in the current prime-order Ristretto255 suite.
  - Detailed peer-share errors are local observability signals for logs and metrics. Integration guidance must say they are not reflected to the remote peer before confirmation; remote responses stay generic.
  - Doc updates in `docs/integration-guidance.md`, `docs/security-assessment.md`, and `docs/security-spec-audit.md` explain the sentinel taxonomy, the framing-level `ErrMessage` behavior for malformed wire lengths, the fact that malformed wire lengths never surface as peer-share sentinels, and the intentional internal divergence from draft-21's function-level neutral-element return.

- **Option B:**
  - One-line fix removes the silent-fallback footgun but does not improve operator-side triage.
  - No API surface expansion.
  - If operator-side triage becomes important later, the work is deferred.

- **Option C:**
  - Adds API surface without fixing the silent-fallback shape. Worst of both worlds.

- **Option D:**
  - Zero risk of getting the sentinel taxonomy wrong, but the silent-fallback shape and the triage gap persist for the lifetime of v1.x.

## Implementation outline (Option A)

1. Add `ErrPeerShareEncoding` and `ErrPeerShareIdentity` to `errors.go` as plain `errors.New` sentinels with doc comments that say the returned error also wraps `ErrAbort`.
2. Change `decodePublicShare` to return `(*ristretto255.Element, error)` using the two new sentinels for encoding and identity rejection, and an internal `ErrAbort`-wrapped error for the wrong-length defensive branch.
3. Change `scalarMultVFY` to return `([]byte, error)`, returning `nil` on every failure path; thread the appropriate error through, and return `fmt.Errorf("%w: neutral-element shared secret", ErrAbort)` for the unreachable post-multiply identity branch.
4. Update `api.go:166-168, 180-184, 222-226` so public API errors retain `invalid initiator share` / `invalid responder share` role context and still satisfy `errors.Is(err, ErrAbort)` plus `errors.Is(err, ErrPeerShareEncoding)` or `errors.Is(err, ErrPeerShareIdentity)` where applicable, using the *Call-site sentinel mapping* specified under Decision (rewrap the plain sentinel; pass non-sentinel errors through unchanged).
5. Add tests:
   - `TestPeerShareErrorsWrapErrAbort` — public-API tests over `Respond` and `Finish`, asserting both role-context strings and `errors.Is` behavior for `ErrAbort` plus the applicable peer-share sentinel.
   - `TestPeerShareEncodingRejection` — feed a non-canonical Ristretto255 encoding via wire and assert `ErrPeerShareEncoding`.
   - `TestPeerShareIdentityRejection` — feed the identity encoding via wire and assert `ErrPeerShareIdentity`.
   - A direct internal-helper length test — call the internal helper with a short slice and assert `errors.Is(err, ErrAbort)` with no peer-share sentinel; malformed wire lengths remain covered by existing framing tests that return `ErrMessage`.
   - A post-multiply identity branch check — use a narrow internal hook, or document the branch's mathematical unreachability, and ensure the specified error is `ErrAbort`-wrapped with no new sentinel.
6. Update `fuzz_test.go` and `vectors_test.go` for the new `scalarMultVFY` signature, including `FuzzScalarMultVFY`, `TestScalarMultVFYDraftInvalidVectors`, and the `TestDraftVectors` direct call sites.
7. Update `docs/integration-guidance.md`, `docs/security-assessment.md`, and `docs/security-spec-audit.md` to describe the sentinel taxonomy, local-only disclosure guidance, malformed wire length behavior as `ErrMessage` rather than any peer-share sentinel, and the intentional update to the "matches documented draft behavior" claim.
8. Add a `CHANGELOG.md` Unreleased entry under "Pre-v1 error surface".
