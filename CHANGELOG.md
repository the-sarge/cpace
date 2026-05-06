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
