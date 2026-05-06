# Integration Guidance

This package implements a CPace draft-21 initiator-responder exchange. It does
not implement an outer application protocol. Applications remain responsible
for negotiating whether CPace is used, which CPace package profile is used, and
what authenticated application channel or keys are built from the confirmed
session.

## Downgrade Protection

If an application negotiates PAKE usage, protocol versions, ciphersuites,
application modes, or fallback authentication mechanisms outside this package,
that negotiation needs its own downgrade protection. This package authenticates
only the inputs it receives through `Config` and the CPace messages it parses.
It cannot detect a stripped offer, a changed fallback choice, or an outer
protocol mode it was never told about.

A typical integration should bind the agreed outer protocol state into CPace:

- include stable party identities in `InitiatorID` and `ResponderID`;
- include a versioned application/domain label and the negotiated CPace package
  profile in `Context`;
- include transcript hashes, negotiation hashes, channel identifiers, or other
  outer-protocol commitments in `AssociatedData`;
- use a fresh, parties-agree-on `SessionID` for each exchange;
- reject the connection unless both `Finish` calls complete successfully.

The exact binding format is application-owned. It should be deterministic,
versioned, and identical on both sides for the same session. Large artifacts
should normally be represented by a digest, Merkle root, exporter, or other
fixed-size commitment rather than placed directly in associated data.

## Identity Orientation

Both sides must use the same role orientation. `InitiatorID` names the party
running `Start`, and `ResponderID` names the party running `Respond`, on both
machines. Do not have each side put itself first. That produces different CI
values and confirmation fails.

Do not use only global role labels such as `"client"` and `"server"` as
identities across all users or deployments. Use stable, application-meaningful
party identities. `Session.PeerID` returns the caller-configured peer identity
that the confirmed exchange proved both sides agreed on; it is not parsed from
wire data.

## Session Outputs

`Respond` success is not authentication. Treat message B as unauthenticated
until `Initiator.Finish` and `Responder.Finish` both succeed.

`Session.TranscriptID` is the draft `CPaceSidOutput` for the confirmed CPace
transcript. It is useful as a CPace transcript identifier, but it is not a
complete channel binding for outer negotiation unless the outer negotiation was
already bound into `Config.Context` or `Config.AssociatedData`.

Use `Session.Export` with domain-specific labels and contexts for application
keys. Exported bytes are deterministic key material from the confirmed ISK, not
fresh randomness.
