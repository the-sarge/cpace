# Security Assessment

Status: initial self-assessment for an auditable draft implementation.

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

Scalar randomness always comes from Go's `crypto/rand.Reader`; callers cannot
inject a custom random reader through the public API. Scalar sampling masks bits
above group size 252 and rejects zero. This creates a secret-dependent loop only
for the all-zero masked scalar case. That event has negligible probability with
the system random reader. Sampling failure wraps `ErrRandomness`.

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

## Dependencies

- `github.com/gtank/ristretto255 v0.2.0`
- `filippo.io/edwards25519 v1.2.0` as an indirect dependency

Dependency review is not complete. Run `govulncheck -test ./...` before any
release. Initial notes are recorded in `docs/dependency-review.md`.

## Release Bar

Do not mark a release production-ready until:

- official draft-21 Ristretto255/SHA-512 vectors pass
- `go test ./...` and `go test -race ./...` pass
- parser and protocol fuzz targets have completed a meaningful run
- every target in `.github/fuzz-targets.json` has run for more than five
  minutes on release hardware or the manual long-fuzz workflow
- `govulncheck -test ./...` and `staticcheck ./...` pass
- this assessment and `docs/spec-matrix.md` are reviewed
- no critical or high independent review findings remain

If draft-21 is superseded, freeze this package as a draft-21 implementation and
plan a separate compatibility update.
