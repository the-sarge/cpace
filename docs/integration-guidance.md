# Integration Guidance

This package implements a CPace draft-21 initiator-responder exchange. It does
not implement an outer application protocol. Applications remain responsible
for negotiating whether CPace is used, which CPace package profile is used, and
what authenticated application channel or keys are built from the confirmed
session.

## Downgrade Protection

If an application negotiates PAKE usage, protocol versions, ciphersuites, application modes, or fallback authentication mechanisms outside this package, that negotiation needs its own downgrade protection. This package authenticates only the inputs it receives through `Input` and the CPace messages it parses. It cannot detect a stripped offer, a changed fallback choice, or an outer protocol mode it was never told about.

A typical integration should bind the agreed outer protocol state into CPace:

- include stable party identities in role-local `SelfID` and `PeerID`;
- include a versioned application/domain label and the negotiated CPace package
  profile in `Context`;
- include transcript hashes, negotiation hashes, channel identifiers, or other outer-protocol commitments in `LocalAssociatedData`, and verify the peer's view after the exchange (below);
- use a fresh, parties-agree-on `SessionID` for each exchange;
- reject the connection unless both `Finish` calls complete successfully and any expected `Session.PeerAssociatedData` check passes.

These bindings have different failure semantics. `Password`, `Context`, and `SessionID` are shared inputs that fail the exchange automatically on divergence: a `SessionID` mismatch is rejected by `Respond` as a session-id mismatch before responder key derivation or confirmation, while `Password` or `Context` divergence changes the derived keys and fails confirmation. `LocalAssociatedData` is role-local and transmitted in the clear: confirmation proves both sides saw the same transmitted `ADa`/`ADb` values, not that either value matches what the other side expected, so an exchange in which the two sides bound different outer commitments still completes. Commitments placed in `LocalAssociatedData` therefore protect the outer negotiation only if the application compares `Session.PeerAssociatedData` against the value the peer was expected to bind and rejects the session on mismatch; shared values that must fail the exchange automatically belong in `Context` or `SessionID`.

The exact binding format is application-owned. It should be deterministic,
versioned, and identical on both sides for the same session. Large artifacts
should normally be represented by a digest, Merkle root, exporter, or other
fixed-size commitment rather than placed directly in associated data.

## Role-Local Identities

Each side supplies identities from its own point of view. The initiator calls `Start` with `SelfID=initiator` and `PeerID=responder`; the responder calls `Respond` with `SelfID=responder` and `PeerID=initiator`. If one side swaps those values, the CI values differ and confirmation fails.

Do not use only global role labels such as `"client"` and `"server"` as identities across all users or deployments. Use stable, application-meaningful party identities. `Session.PeerID` returns the caller-configured peer identity that the confirmed exchange proved both sides agreed on; it is not parsed from wire data.

## Single-Use State Lifecycle

`Start` and `Respond` return single-use state that holds local persistent secret material until a terminal operation consumes it. Call `Initiator.Close` or `Responder.Close` if the exchange can be abandoned before `Finish`, and prefer `defer state.Close()` immediately after successful construction. `Close` after `Finish` returns nil, including after a `Finish` call that consumed the state and failed, so the deferred cleanup pattern is safe for success, failure, and cancellation paths.

Copies of constructed `Initiator` and `Responder` values share terminal state. A terminal operation on one copy spends the state for all copies: `Finish` after `Close` returns `ErrStateUsed`, while `Close` after `Finish` returns nil. Returned `Session` values own independent key material and still need their own `Close` when done.

## Session Outputs

`Respond` success is not authentication. Treat message B as unauthenticated
until `Initiator.Finish` and `Responder.Finish` both succeed.

`Session.TranscriptID` is the draft `CPaceSidOutput` for the confirmed CPace transcript. It is useful as a CPace transcript identifier, but it is not a complete channel binding for outer negotiation unless the outer negotiation was already bound into `Input.Context` or `Input.LocalAssociatedData`.

Use `Session.Export` with domain-specific labels and contexts for application
keys. Exported bytes are deterministic key material from the confirmed ISK, not
fresh randomness.

## Error Triage

Peer public-share rejections from `Respond` and `Initiator.Finish` always satisfy `errors.Is(err, ErrAbort)`, and two exported sentinels refine the cause for local triage: `ErrPeerShareEncoding` reports 32 bytes that are not a canonical Ristretto255 encoding (a buggy or malicious peer), and `ErrPeerShareIdentity` reports an encoded identity element (almost certainly an active attacker probing for a forced neutral-element shared secret). The error message keeps the role context — `invalid initiator share` from `Respond`, `invalid responder share` from `Initiator.Finish` — so logs distinguish which share was rejected.

Malformed wire lengths never surface as peer-share sentinels. Framing decodes public-share fields with an exact 32-byte limit and rejects any other length with `ErrMessage` before a share reaches point decoding.

The peer-share sentinels are local observability signals for logs and metrics. Do not reflect the detailed rejection cause to the remote peer before confirmation; keep remote-facing failure responses generic so the error channel does not become a probing surface.
