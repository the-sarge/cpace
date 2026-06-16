# ADR-0009 Secret-Lifetime Audit

Date: 2026-06-16

Scope: manual audit for the ADR-0009 caller-input implementation. This is not a replacement for the pinned security/spec audit or long-fuzz evidence; those remain stale for this security-relevant package-code change until refreshed against the exact release candidate.

## Code Paths Reviewed

- `acceptInput`, `validateRequiredCallerInputFields`, and `validateCallerInputCapFields` in `input.go`: validate `Input`, enforce package-owned caps, and clone caller-owned byte slices into a short-lived `acceptedInput`; the validation helpers and `callerInputCappedField` values hold only transient read-only slice headers into caller input and do not clone, persist, wipe, or extend password byte residency.
- `acceptedInput.wipe` in `input.go`: clears accepted password, role-local identities, context, SessionID, and local associated data if normalization does not transfer ownership.
- `normalizeInput`, `normalizeStartInput`, `normalizeRespondInput`, and `callerInputRole.mapTranscriptIDs` in `input.go`: map role-local `SelfID`/`PeerID` into transcript-role `initiatorID`/`responderID`, build CI, clear accepted context immediately after CI construction, and transfer accepted byte slices into `normalizedInput`.
- `normalizedInput.wipe` in `input.go`: deferred by both `startWithRandom` and `respondWithRandom` to clear normalized password, mapped identities, CI, SessionID, and local associated data on every return path.
- `newInitiatorCore` and `newResponderCore` in `core.go`: derive the generator from the normalized password, then immediately clear the normalized password backing array before scalar sampling and transcript work continue.

## Findings

- No reusable validated-input object was introduced. `acceptedInput` and `normalizedInput` are unexported, by-value plumbing objects scoped to one `Start` or `Respond` call; successful construction returns only `Initiator` or `Responder` single-use state.
- No new persistent password field was introduced. `acceptedInput.password` and `normalizedInput.password` are transient; `initiatorCore`, `responderCore`, and `Session` do not store password bytes.
- On validation failure before cloning, caller-owned `Input.Password` is not modified. If validation or cap checks fail after earlier required-field checks, no accepted clone has been returned.
- On normalization failure or panic before ownership transfer, `normalizeInput` defers `accepted.wipe()` while `keep` is false.
- On successful normalization, ownership of the accepted password moves into `normalizedInput`; `startWithRandom` and `respondWithRandom` immediately defer `nc.wipe()`.
- On core-constructor success, `newInitiatorCore` and `newResponderCore` clear `nc.password` immediately after `calculateGenerator` returns, bounding password residency to generator derivation rather than the whole constructor. The outer deferred `nc.wipe()` remains a backstop and is idempotent.
- On core-constructor error, including randomness errors after generator derivation, the outer deferred `nc.wipe()` clears any remaining normalized slices. The constructors also clear the password before scalar sampling, so the password is cleared before randomness errors can return.
- On panic after `startWithRandom` or `respondWithRandom` installs `defer nc.wipe()`, the normalized password and other normalized slices are cleared during panic unwinding.
- `Respond` decodes and validates message A after `defer nc.wipe()` is installed, so malformed message A and session-ID mismatch paths clear normalized input.

## Residual Risks

- Go still does not guarantee secure zeroization, pinning, or avoidance of compiler/runtime copies. This audit is a best-effort lifetime review, not a claim of resistance to local memory disclosure.
- `calculateGenerator`, `lvCat`, `prependLen`, SHA-512, HMAC, and HKDF internals can create heap or runtime-owned intermediates outside package control; this risk is unchanged in kind from the existing security assessment.
- This branch changes security-relevant caller-input mapping and validation vocabulary. Dependency review, Capslock, security/spec audit, and paired long-fuzz evidence remain historical until refreshed at the exact release candidate.
