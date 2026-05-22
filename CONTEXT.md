# cpace

A Go implementation of the draft-21 CPace password-authenticated key exchange,
restricted to the `CPACE-RISTR255-SHA512` suite. This glossary covers the
protocol vocabulary and the internal module language used when reasoning about
the implementation.

## Language

**CPace core**:
The deep, unexported module that owns one role's CPace computation — generator
derivation, scalar sampling, Diffie-Hellman, transcript assembly, ISK
derivation, confirmation tags — and the lifetime of every secret it touches.
Exists as `initiatorCore` and `responderCore`; the public `Initiator` and
`Responder` are thin shells that delegate to it.
_Avoid_: crypto layer, engine, helper.

**Single-use state**:
An `Initiator` or `Responder` value, which the protocol permits to be consumed
exactly once. Reuse is rejected, and the state is spent even when a step fails.
_Avoid_: handle, context (Go's `context.Context` is unrelated).

**ISK**:
The Intermediate Session Key — the shared secret CPace derives from the
Diffie-Hellman result and the transcript. It has two independent copies with
separate owners: the **CPace core** holds the working copy; a confirmed
**Session** holds an independent clone. Each owner clears its own copy.
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

**Maintainer:** Two owners, one copy each. The initiator's **CPace core**
derives the ISK as a working secret and clears it when the core is cleared. On
success it also builds a **Session**, which gets an independent clone. After
that, the core's copy and the Session's copy never alias.

**Dev:** So if I copy the Session value, I get a third ISK copy?

**Maintainer:** No — a Session copy shares the *same* underlying ISK and close
state. Closing one copy closes them all. The "two copies" rule is core-vs-Session,
not Session-vs-Session.

**Dev:** And if `Finish` fails?

**Maintainer:** The **single-use state** is still spent, the **CPace core**
still clears its ISK, and no Session is built. A failed **confirmation tag**
check is just one such failure path.
