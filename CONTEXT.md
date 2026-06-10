# cpace

A Go implementation of the draft-21 CPace password-authenticated key exchange,
restricted to the `CPACE-RISTR255-SHA512` suite. This glossary covers the
protocol vocabulary and the internal module language used when reasoning about
the implementation.

## Language

**CPace core**:
The deep, unexported module that owns one role's CPace computation — generator
derivation, scalar sampling, Diffie-Hellman, transcript assembly, ISK
derivation, confirmation tags — and the lifetime of its *persistent* secrets;
scratch secrets are cleared in place inside its methods, never stored on it.
**Target architecture, not yet implemented**: per ADR-0001 (accepted;
implementation gated on the external reviews) it will exist as `initiatorCore`
and `responderCore`, with the public `Initiator` and `Responder` becoming thin
shells that delegate to it. Until that lands, `Initiator` / `Responder` hold
the cryptographic state directly in `api.go`.
_Avoid_: crypto layer, engine, helper.

**Single-use state**:
An `Initiator` or `Responder` value, which the protocol permits to be consumed
exactly once. Reuse is rejected, and the state is spent even when a step fails.
_Avoid_: handle, context (Go's `context.Context` is unrelated).

**ISK**:
The Intermediate Session Key — the shared secret CPace derives by hashing the
sid, the Diffie-Hellman result, and the transcript. Ownership is
role-asymmetric. The responder derives its ISK at construction and holds a
working copy until cleanup (today a `Responder` field; under ADR-0001 a
`responderCore` field cleared by `clear()`). The initiator's ISK exists only as
a local inside `Finish`, cleared before `Finish` returns — it is never stored
on the initiator or its core. A confirmed **Session** holds its own
independent clone. Each owner clears its own copy.
_Avoid_: session key, shared secret, master key.

**Transcript**:
The injective initiator-responder ordering of both parties' public shares and
associated data, fed into ISK derivation. This suite uses the IR ordering only.
_Avoid_: log, history.

**Confirmation tag**:
The explicit key-confirmation MAC each party sends and verifies, proving both
sides derived the same ISK before a **Session** is authenticated.
_Avoid_: checksum, signature, HMAC (too generic).

**Session**:
The public type returned by a successful `Finish` — an explicitly confirmed
CPace result that exports application key material and holds its own ISK clone.
Copies of a Session share close state.
_Avoid_: connection, channel.

## Example dialogue

**Dev:** When `Initiator.Finish` succeeds, who owns the ISK?

**Maintainer:** After `Finish` returns, exactly one owner: the **Session**,
which got an independent clone. The initiator's own ISK is a `Finish`-local
scratch secret — derived, used for the tag exchange, and cleared before
`Finish` returns; it is never stored on the initiator (nor, under ADR-0001, on
its core). The responder is the asymmetric case: it derives its ISK at
construction and holds a working copy until cleanup. The working copy and the
Session's clone never alias.

**Dev:** So if I copy the Session value, I get a third ISK copy?

**Maintainer:** No — a Session copy shares the *same* underlying ISK and close
state. Closing one copy closes them all. The "two copies" rule is core-vs-Session,
not Session-vs-Session.

**Dev:** And if `Finish` fails?

**Maintainer:** The **single-use state** is still spent and no Session is
built. On the initiator side, if the failure happens after derivation — a
failed **confirmation tag** check is the canonical case — the `Finish`-local
ISK is cleared on that path too; on a parse failure no initiator ISK ever
existed. On the responder side the working copy is cleared by cleanup the same
as on success.
