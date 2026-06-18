# cpace

A Go implementation of the draft-21 CPace password-authenticated key exchange,
restricted to the `CPACE-RISTR255-SHA512` suite. This glossary covers the
protocol vocabulary and the internal module language used when reasoning about
the implementation.

## Language

**CPace core**:
The deep, unexported module that owns one role's CPace computation — generator derivation, scalar sampling, Diffie-Hellman, transcript assembly, ISK derivation, confirmation tags — and the lifetime of its *persistent* secrets; scratch secrets are cleared in place inside its methods, never stored on it. Implemented per ADR-0001 as `initiatorCore` and `responderCore`, with public `Initiator` and `Responder` as thin single-use shells that hold shared terminal state and delegate cryptographic work to the core.
_Avoid_: crypto layer, engine, helper.

**Single-use state**:
An `Initiator` or `Responder` value, which the protocol permits to be consumed exactly once. `Finish` and `Close` are terminal operations: reuse is rejected, the state is spent even when `Finish` fails, `Close` releases local persistent secrets when an exchange is abandoned, and constructed value copies share the same terminal state.
_Avoid_: handle, context (Go's `context.Context` is unrelated).

**Message framing**:
The package-owned binary envelope for CPace messages A, B, and C: format byte, suite byte, role byte, and length-value encoded fields with package-owned caps. Message framing sits in front of the **CPace core**; decoded cryptographic fields cross that seam, but parsing, role checks, size caps, and wire bytes stay here.
_Avoid_: wire protocol, transport, serialization helper.

**Peer-share rejection**:
The internal module that classifies peer public share bytes, applies ADR-0003 role-context errors, and computes the Diffie-Hellman result only after validation. It sits inside the **CPace core** seam: **Message framing** enforces field size first, and the responder path calls this module before generator derivation or scalar sampling.
_Avoid_: point validation helper, scalar multiplication wrapper, peer key parser.

**Package-owned cap policy**:
The internal policy that names and caps caller-provided **Caller input** fields and package-owned **Message framing** fields before they can drive allocation, parsing, or CPace computation. It is not a public profile knob: changing a cap value is an observable behavior change and must be treated as release-policy work; caller-input diagnostic names may change only through an accepted public-input ADR.
_Avoid_: limit constants, validation helpers, size settings.

**Caller input**:
The public-facing module for application facts supplied before `Start` or `Respond` can construct **single-use state**: password, role-local identities, context, SessionID, local associated data, and compatibility flags after package policy accepts them. ADR-0009 names `Input` as the role-local adapter, with `SelfID`, `PeerID`, and `LocalAssociatedData` mapped by the package while `Password`, `Context`, and `SessionID` remain shared session values. **Caller input** is the seam where validation, role mapping, **Package-owned cap policy**, copy ownership, accepted/normalized input plumbing, and package-profile commitments meet.
_Avoid_: config validation, options, input helper.

**Release policy checker**:
The tooling module that validates accepted release-pipeline policy, especially ADR-0007, against the Release Validation workflow and local release helper files. It is validation-only: it parses workflow YAML, checks tag-only execution, signed-tag verification, SBOM and attestation publication, action pinning, least permissions, release-note extraction, and no in-place release replacement, but it does not generate release workflow YAML and does not query live GitHub ruleset state.
_Avoid_: release generator, CI abstraction, policy engine.

**Accepted release policy**:
The internal catalogue of ADR-0007 release-pipeline facts that the **Release policy checker** applies to repository files: workflow shape, jobs, step order, `needs`, permissions, exact protected shell snippets, required helper scripts, and maintainer signing keys. It is a validation input, not a workflow template; changing it changes what release drift the checker accepts.
_Avoid_: workflow template, release spec, generated CI.

**Evidence baseline**:
The documentation module that indexes pinned release-evidence claims: exact evidence commit or workflow, toolchain, raw artifact paths, summary docs, stale reasons, and refresh triggers. It does not refresh evidence and does not make a production-readiness claim; it makes current evidence freshness and gaps visible from one place.
_Avoid_: evidence summary, release claim, checklist.

**ISK**:
The Intermediate Session Key — the shared secret CPace derives by hashing the sid, the Diffie-Hellman result, and the transcript. Ownership is role-asymmetric. The responder derives its ISK at construction and holds a working copy in `responderCore` until cleanup by `clear()`. The initiator's ISK exists only as a local inside the core's `finish`, cleared before `Finish` returns — it is never stored on the initiator or its core. A confirmed **Session** holds its own independent clone. Each owner clears its own copy.
_Avoid_: session key, shared secret, master key.

**Transcript**:
The internal module that owns the injective initiator-responder ordering of both parties' public shares and associated data and derives from it the ISK, the two role confirmation tags, and the CPaceSidOutput transcript id. It holds only public wire data: the ISK it derives is a return value owned and cleared by the **CPace core**, never stored on the transcript. The initiator builds a finish-local transcript; the responder builds the transcript at construction and carries that same value through to Finish, so both roles compute their finish-time confirmation tag through one transcript interface rather than recomputing it from decomposed fields. This suite uses the IR ordering only.
_Avoid_: log, history, decomposed ya/ada responder fields.

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
