# ADR-0009 Secret-Lifetime Audit

Date: 2026-06-16

Scope: manual audit for the ADR-0009 caller-input implementation. This is not a replacement for the pinned security/spec audit or long-fuzz evidence; those remain stale for this security-relevant package-code change until refreshed against the exact release candidate.

## Code Paths Reviewed

- `acceptInput`, `validateRequiredCallerInputFields`, and `validateCallerInputCapFields` in `input.go`: validate `Input`, enforce package-owned caps, and clone caller-owned byte slices into a short-lived `callerInput`; the validation helpers and `callerInputCappedField` values hold only transient read-only slice headers into caller input and do not clone, persist, wipe, or extend password byte residency.
- `callerInput.wipe` in `input.go`: clears any package-owned caller input slices still referenced by `callerInput` if normalization panics before handoff completes, and is harmless after a successful handoff because transferred slice headers have been nilled.
- `callerInput.handoff` in `input.go`: maps role-local `SelfID`/`PeerID` into transcript-role `initiatorID`/`responderID`, builds CI, clears and drops residual context storage, transfers package-owned password, identity, SessionID, and local associated data slice headers into `normalizedInput`, and nils the transferred headers in `callerInput`.
- `normalizeInput`, `normalizeStartInput`, `normalizeRespondInput`, and `callerInputRole.mapTranscriptIDs` in `input.go`: install `defer caller.wipe()` immediately after accepting caller input, then return the handed-off `normalizedInput` for the role-specific start/respond path.
- `normalizedInput.wipe` in `input.go`: deferred by both `startWithRandom` and `respondWithRandom` to clear normalized password, mapped identities, CI, SessionID, and local associated data on every return path.
- `newInitiatorCore` and `newResponderCore` in `core.go`: derive the generator from the normalized password, then immediately clear the normalized password backing array before scalar sampling and transcript work continue.

## Findings

- No reusable validated-input object was introduced. `callerInput` and `normalizedInput` are unexported, by-value plumbing objects scoped to one `Start` or `Respond` call; successful construction returns only `Initiator` or `Responder` single-use state.
- No new persistent password field was introduced. `callerInput.password` and `normalizedInput.password` are transient; `initiatorCore`, `responderCore`, and `Session` do not store password bytes.
- On validation failure before cloning, caller-owned `Input.Password` is not modified. If SessionID validation or cap checks fail after earlier required-field checks, no package-owned clone has been created or returned.
- On `normalizeInput` error, the error comes from `acceptInput` before any package-owned clone is returned.
- On panic before ownership transfer after `acceptInput` succeeds, `normalizeInput`'s deferred `caller.wipe()` clears package-owned clones still referenced by `callerInput` during panic unwind.
- On successful normalization, `callerInput.handoff` transfers package-owned slices into `normalizedInput`, clears residual context storage after CI construction, and nils the transferred slice headers in `callerInput`; the deferred `caller.wipe()` then runs without retaining or clearing the slices now owned by `normalizedInput`.
- After successful normalization, `startWithRandom` and `respondWithRandom` immediately defer `nc.wipe()` so `normalizedInput` owns the transferred slices only until the role-specific constructor path returns or unwinds.
- On core-constructor success, `newInitiatorCore` and `newResponderCore` clear `nc.password` immediately after `calculateGenerator` returns, bounding password residency to generator derivation rather than the whole constructor. The outer deferred `nc.wipe()` remains a backstop and is idempotent.
- On core-constructor error, including randomness errors after generator derivation, the outer deferred `nc.wipe()` clears any remaining normalized slices. The constructors also clear the password before scalar sampling, so the password is cleared before randomness errors can return.
- On panic after `startWithRandom` or `respondWithRandom` installs `defer nc.wipe()`, the normalized password and other normalized slices are cleared during panic unwinding.
- `Respond` decodes and validates message A after `defer nc.wipe()` is installed, so malformed message A and session-ID mismatch paths clear normalized input.

## Residual Risks

- Go still does not guarantee secure zeroization, pinning, or avoidance of compiler/runtime copies. This audit is a best-effort lifetime review, not a claim of resistance to local memory disclosure.
- `calculateGenerator`, `lvCat`, `prependLen`, SHA-512, HMAC, and HKDF internals can create heap or runtime-owned intermediates outside package control; this risk is unchanged in kind from the existing security assessment.
- This branch changes security-relevant caller-input mapping and validation vocabulary. Dependency review, Capslock, security/spec audit, and paired long-fuzz evidence remain historical until refreshed at the exact release candidate.
