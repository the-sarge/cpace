# Security Assessment

Status: self-assessment for an auditable draft implementation. Reviewed
against commit `2e09774f171dde8c62763d6e35a258b0fef88801` on 2026-05-08;
see `docs/security-spec-audit.md`.

## Cryptographic Scope

Only `CPACE-RISTR255-SHA512` from `draft-irtf-cfrg-cpace-21` is implemented.
The public API exposes initiator-responder mode only and requires explicit key
confirmation before returning a `Session`.

## Secret-Dependent Behavior

The implementation delegates Ristretto255 group operations and scalar field
operations to `github.com/gtank/ristretto255 v0.2.0`, which documents constant
time operation except for explicitly variable-time APIs that this module does
not call.

Password handling follows the draft generator string padding rule. For short
passwords the first SHA-512 input block length is independent of password
length, but Go slice allocation and caller-side password handling are not
constant time.

The draft recommends, but does not require, a unique session identifier. This
package rejects an empty `SessionID` by default; callers must provide a fresh,
non-secret, parties-agree-on sid for every session. `AllowEmptySessionID`
preserves draft-21 compatibility for tests or deliberate compatibility profiles,
but empty sids weaken replay and transcript separation properties. Default
empty-sid failures wrap both `ErrInvalidInput` and `ErrEmptySessionID`.

Any outer application negotiation of PAKE version, ciphersuite, protocol mode,
or whether CPace is used needs downgrade protection outside this package. The
package authenticates only the inputs it is given and has no negotiation API.

Both parties must use a consistent role-ID orientation: `InitiatorID` names the
party running `Start`, and `ResponderID` names the party running `Respond`. If
each side puts itself first, the CI values differ and confirmation fails. Role
labels such as `"client"` and `"server"` are not enough as global identities for
all users or deployments; callers should bind stable party identities.

Scalar randomness always comes from Go's `crypto/rand.Reader`; callers cannot inject a custom random reader through the public API. Scalar sampling masks the top four bits of byte 31 (clearing bits above group size 252), parses the result as a canonical Ristretto255 scalar, and rejects the zero scalar. The Ristretto255 scalar order `L = 2^252 + 27742...` exceeds `2^252` by approximately `2^125`, so a uniformly random masked value has probability approximately `2^-125` of falling in `[L, 2^252)` and being rejected by `SetCanonicalBytes`. That outcome is statistically negligible but reachable in principle; the sampling loop treats it as an unusable sample and retries rather than aborting. The zero check creates a secret-dependent loop only for the all-zero masked scalar case, which has negligible probability with the system random reader. Sampling failure after `maxScalarTries` retries wraps `ErrRandomness`. The package implements draft §8.3 bit-masking with defense-in-depth retries for unusable samples; using the ristretto255 library's `SetUniformBytes` plus zero rejection/retry would be an allowed draft alternative, but it consumes 64 random bytes and reduces modulo the scalar order, changing deterministic behavior and defining a different package profile.

## Memory Handling

All mutable public inputs and received message fields are copied. The
implementation clears selected owned byte-slice temporaries, consumed scalar
state, derived generator elements, consumed responder state, and session key
material on a best-effort basis. `Session.Close` clears the session-owned ISK
and makes future `Export` calls fail with `ErrSessionClosed`. The Go runtime
does not guarantee secure zeroization, pinning, or avoidance of copies made by
the compiler or garbage collector. This package does not claim resistance to a
local memory disclosure adversary.

## Key Access

Raw `K`, scalar values, and ISK are not exposed through the public API. Exported
application material is derived from the confirmed ISK using HKDF-SHA512 and is
deterministic for a given label and context; it is not fresh randomness or a
randomness pool. A session is returned only after key confirmation succeeds.
`Respond` success alone does not authenticate the peer. `Session.PeerID` and
`Session.PeerAssociatedData` return metadata that each side configured locally
and that the confirmed exchange proves both sides agreed on. `PeerID` is copied
from `Config`, not parsed from peer-controlled wire data.

## Framing

The CPace draft leaves wire encoding to applications. This implementation uses a
package-owned binary framing with explicit version, suite, and role bytes plus
draft LEB128 length-value fields. Decoders reject trailing data, wrong
version/suite/role, malformed or non-canonical LEB128, oversized fields, invalid
point lengths, and invalid tag lengths.

Size limits for valid package-owned message shapes are per-field: passwords and party IDs are capped at 4 KiB, context and session IDs at 1 KiB, and associated data at 64 KiB. Public-share and confirmation-tag fields are decoded with exact 32-byte and 64-byte limits. Malformed framed inputs also hit a 128 KiB aggregate decoder backstop before field parsing proceeds; this is an invalid-message throttle, not a replacement for the per-field valid-message caps. Associated data is intended to bind outer protocol context; large external artifacts should normally be represented by a digest, Merkle root, exporter, or other fixed-size commitment.

Confirmation tags intentionally remain draft-compatible. This package does not
add extra role-label inputs to the draft-21 confirmation MACs.

## Error Surface

Peer-share rejections wrap `ErrAbort`, and the public API adds one of two exported sentinels for local triage: `ErrPeerShareEncoding` for a non-canonical Ristretto255 encoding and `ErrPeerShareIdentity` for an identity-element submission. Errors from `Respond` and `Initiator.Finish` retain `invalid initiator share` / `invalid responder share` role context. Malformed wire lengths are rejected by framing as `ErrMessage` and never surface as peer-share sentinels; the internal wrong-length branch of share decoding is defensive, `ErrAbort`-wrapped, and unreachable from the wire.

The finer rejection granularity is not a secret-dependent oracle: every classification is a function only of the encoded public wire bytes, with no secret input. The only secret-adjacent branch, the post-multiply neutral-element check, is unreachable for prime-order Ristretto255 and is kept as defense-in-depth behind an `ErrAbort`-wrapped error.

Internally `scalarMultVFY` returns nil with a typed error on failure instead of draft-21's function-level neutral-element return convention. The divergence is intentional and internal-only: protocol abort behavior is unchanged and the draft's invalid-share vectors, including the neutral-element case, are still rejected. Detailed peer-share errors are for local logs and metrics; integrations must not reflect them to the remote peer before confirmation (see `docs/integration-guidance.md`).

## Dependencies

- `github.com/gtank/ristretto255 v0.2.0`
- `filippo.io/edwards25519 v1.2.0` as an indirect dependency

Dependency review was refreshed on 2026-05-08 at commit
`2e09774f171dde8c62763d6e35a258b0fef88801` under Go 1.26.3; see
`docs/dependency-review.md`. `govulncheck -test -show verbose ./...` found no
vulnerabilities, and the pinned `gosec@v2.26.1` command reported zero issues.
Repeat the review against the exact release tag if any dependency, toolchain, or
parser/security-relevant code changes before release.

## Fuzzing

Fuzz target evidence is recorded in `docs/fuzz-evidence.md`. The current paired
release-readiness run covers all 14 targets registered in
`.github/fuzz-targets.json` at commit
`2e09774f171dde8c62763d6e35a258b0fef88801`, with paired one-hour Go 1.26.3
long runs on ARM and Intel maintainer machines. Supplemental `v0.1.2` tag soak
evidence at commit `4e661bc1f925ebedf1f270668129d85bab73e468` ran
`FUZZTIME=4h` across all 14 targets on ARM and Intel hosts; ARM passed all
targets, while the Intel all-target run ended with a `FuzzProtocolConsistency`
deadline failure followed by a clean same-host 4-hour targeted rerun.

## Release Bar

Do not mark a release production-ready until:

- official draft-21 Ristretto255/SHA-512 vectors pass
- `go test ./...` and `go test -race ./...` pass
- parser and protocol fuzz targets have completed a meaningful run
- every target in `.github/fuzz-targets.json` has run for more than five
  minutes on release hardware or the manual long-fuzz workflow
- `govulncheck -test ./...`, advisory `gosec`, and `staticcheck ./...` pass
- this assessment and `docs/spec-matrix.md` are reviewed
- no critical or high independent review findings remain

If draft-21 is superseded, freeze this package as a draft-21 implementation and
plan a separate compatibility update.
