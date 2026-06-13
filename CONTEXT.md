# cpace

A Go implementation of the draft-21 CPace password-authenticated key exchange,
restricted to the `CPACE-RISTR255-SHA512` suite. This glossary covers the
protocol vocabulary and the internal module language used when reasoning about
the implementation.

## Language

**CPace core**:
The deep, unexported module that owns one role's CPace computation — generator derivation, scalar sampling, Diffie-Hellman, transcript assembly, ISK derivation, confirmation tags — and the lifetime of its *persistent* secrets; scratch secrets are cleared in place inside its methods, never stored on it. Implemented per ADR-0001 as `initiatorCore` and `responderCore`, with public `Initiator` and `Responder` as thin single-use shells that hold a named `core` field and delegate cryptographic work to it.
_Avoid_: crypto layer, engine, helper.

**Single-use state**:
An `Initiator` or `Responder` value, which the protocol permits to be consumed
exactly once. Reuse is rejected, and the state is spent even when a step fails.
_Avoid_: handle, context (Go's `context.Context` is unrelated).

**Message framing**:
The package-owned binary envelope for CPace messages A, B, and C: format byte, suite byte, role byte, and length-value encoded fields with package-owned caps. Message framing sits in front of the **CPace core**; decoded cryptographic fields cross that seam, but parsing, role checks, size caps, and wire bytes stay here.
_Avoid_: wire protocol, transport, serialization helper.

**Package-owned cap policy**:
The internal policy that names and caps caller-provided `Config` fields and package-owned **Message framing** fields before they can drive allocation, parsing, or CPace computation. It is not a public profile knob: changing a cap is an observable behavior change and must be treated as release-policy work.
_Avoid_: limit constants, validation helpers, size settings.

**Release policy checker**:
The tooling module that validates accepted release-pipeline policy, especially ADR-0007, against the Release Validation workflow and local release helper files. It is validation-only: it parses workflow YAML, checks tag-only execution, signed-tag verification, SBOM and attestation publication, action pinning, least permissions, release-note extraction, and no in-place release replacement, but it does not generate release workflow YAML and does not query live GitHub ruleset state.
_Avoid_: release generator, CI abstraction, policy engine.

**ISK**:
The Intermediate Session Key — the shared secret CPace derives by hashing the sid, the Diffie-Hellman result, and the transcript. Ownership is role-asymmetric. The responder derives its ISK at construction and holds a working copy in `responderCore` until cleanup by `clear()`. The initiator's ISK exists only as a local inside the core's `finish`, cleared before `Finish` returns — it is never stored on the initiator or its core. A confirmed **Session** holds its own independent clone. Each owner clears its own copy.
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
