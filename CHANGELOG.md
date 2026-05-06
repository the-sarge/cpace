# Changelog

## Unreleased

- Clarify that `Respond` success is not authentication; only successful
  `Initiator.Finish` and `Responder.Finish` calls return authenticated
  sessions.
- Document role-ID orientation, `SessionID` freshness, outer downgrade
  responsibility, `TranscriptID` scope, and `Export` domain separation.
- Add draft-21 Ristretto255 generator-vector coverage, assert
  `sid_output_oc`, and add reflected-share regression coverage.
- Expand fuzz coverage with protocol-entry, scalar verification, and message
  round-trip fuzz targets.
- Add security tooling updates: `govulncheck -test`, weekly vulnerability
  scanning, Dependabot, and advisory `gosec` SARIF artifacts.
- Add best-effort cleanup for owned protocol temporaries, consumed scalar
  state, derived generator elements, and consumed responder state.
- Prevalidate responder-side message A public shares before responder scalar
  sampling while retaining the final `scalarMultVFY` protocol check.
- Add `docs/interview-results-triage.md` and start `DEV-JOURNAL.md` to record
  review triage and landing notes.
- Define the package-owned wire framing as format v1 with prefix byte `0xc1`.
  No released versions used the earlier draft-revision byte.
- Add `ErrRandomness` for random-source read failures and unusable scalar
  samples.
- Document and test that `Finish` consumes protocol state even when parsing or
  confirmation fails.
- Lighten GitHub-hosted CI before release, add docs-only PR validation, and add
  local `quick`, `check:changed`, docs, formatting, and ast-grep validation
  lanes.
- Remove public `Config.Rand`; `Start` and `Respond` now always draw scalar
  randomness from `crypto/rand.Reader`, with deterministic readers confined to
  package-internal tests and fuzzing.
- Reject empty `SessionID` by default, add `ErrEmptySessionID`, and add
  `AllowEmptySessionID` as an explicit draft-21 compatibility opt-in.
- Add `Session.Close`, `ErrSessionClosed`, `PeerAssociatedData`, and `PeerID`;
  closing a session clears session-owned key material best-effort and prevents
  future `Export` calls.
- Keep confirmation tags draft-compatible and replace the former 1 MiB generic
  field cap with non-configurable per-field caps: 4 KiB for passwords and IDs,
  1 KiB for context and session IDs, 64 KiB for associated data, and exact
  public-share/tag lengths on the wire.
- Keep draft-21 Ristretto255 scalar sampling via masked canonical 32-byte
  values and document why the `SetUniformBytes` alternative is not used.
- Close the policy/API decision phase in the project docs and shift release
  tracking to dependency review, long fuzzing, security/spec audit, and
  external review readiness.
