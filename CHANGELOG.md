# Changelog

## Unreleased

- Pre-v1 API cleanup (breaking relative to `v0.1.2`): remove exported `Suite` and `SuiteCPaceRistretto255SHA512`; callers should drop those references and treat the package as single-suite with opaque framing. No wire/protocol behavior change.
- Internal hardening: `Initiator.Finish` and `Responder.Finish` now reject caller-fabricated zero-value protocol states with `ErrInvalidInput` before consuming state; this closes the pre-v1 zero-value responder forged-tag path where `Responder.Finish` could accept `encodeMessageC(confirmationTag(nil, nil, nil, nil))` and return a Session keyed from nil ISK. Nil-receiver error text now says "uninitialized initiator/responder"; error identity remains `ErrInvalidInput`.
- Pre-v1 error surface: add exported `ErrPeerShareEncoding` and `ErrPeerShareIdentity` sentinels distinguishing non-canonical peer public shares from identity-element submissions (ADR-0003). Peer-share rejections from `Respond`/`Initiator.Finish` continue to wrap `ErrAbort` and retain initiator/responder role context; the internal `scalarMultVFY`/`decodePublicShare` helpers now return a typed error, and `scalarMultVFY` returns nil instead of the draft-shaped all-zero fallback. Malformed wire lengths still surface as `ErrMessage` from framing. No wire-format or protocol-visible change.
- Bump the pinned toolchain directive to Go 1.26.4 after the 2026-06-02 Go
  security release (`crypto/x509`, `mime`, and `net/textproto` fixes, plus
  `crypto/fips140`, compiler, and runtime bug fixes). CI already runs 1.26.4
  via setup-go with `GOTOOLCHAIN=local`; this aligns the `go.mod` directive
  and local builds with it. No source change required.
- Add a reusable evidence-bundle policy and cross-toolchain vector-stability
  checklist for future exact-candidate evidence refreshes.
- Update external-review handoff and reviewer outreach notes to point at the
  published `v0.1.2` prerelease.
- Record `v0.1.2` supplemental 4-hour ARM/Intel fuzz soak evidence with the
  Intel deadline-miss recovery in `docs/fuzz-evidence.md` (#49).
- Pin the Go toolchain via `toolchain go1.26.3` in `go.mod` and add a
  per-workflow "Report Go environment" step so CI consistently records the
  toolchain version used for each lane (#51).
- Add a self-hosted `Autoscaled Fuzz` workflow on the
  `infra-autoscale-cpace-fuzz-linux` runner label, gated to scheduled and
  trusted manual dispatches, with a GitHub-hosted input-validation preflight;
  add `.github/actionlint.yaml` to register the self-hosted runner label and
  harden the `task fuzz` recipe to validate `FUZZTIME` and `PARALLEL` env
  inputs (#52).
- Reaffirm draft-21 scalar sampling behavior in `docs/security-assessment.md`
  and `docs/spec-matrix.md`: bit-masking is the draft §8.3 recommendation and
  the package adds defense-in-depth retries for the (~2^-125) canonical-decode
  rejection window and the zero scalar; no protocol-visible change.
- Internal hardening: unify the deferred wipe of normalized config fields so
  every cloned `Config` byte slice is zeroized on every Start/Respond exit
  path, and mirror `Responder.Finish`'s deferred ISK wipe in `Initiator.Finish`
  so future early-returns cannot leak the session key. No public-API or
  wire-format change.
- Pin protocol-identity strings (`DraftVersion`, `suiteName`, `currentSuite`) and the byte output of `buildCI` via
  `TestBuildCIWireStability`. Add password-mismatch, nil-receiver,
  Export-prefix-free, and `ErrConfirmationFailed`/`ErrAbort` state-consumption
  tests. No protocol-visible change.

## v0.1.2 - 2026-05-08

- Add external review handoff notes, public contribution guidance, issue
  templates, a pull-request template, and reviewer outreach notes.
- Add threat-model and release-checklist docs for external review and future
  exact-candidate evidence refresh.
- Add the OpenSSF Best Practices Baseline badge to the README.
- Add Developer Certificate of Origin signoff policy and PR validation.
- Add coordinated vulnerability disclosure response timeframes to
  `SECURITY.md`.
- Add a project secrets and credentials policy covering storage, access, and
  rotation.
- Add release verification instructions for signed Git tags and future release
  assets.
- Document the expected release signer identity for release verification.
- Document support scope and duration for `v0.x` prereleases.
- Add governance, CI, test-update, SCA/SAST threshold, VEX, SBOM, multi-repo,
  and attack-surface policy docs for OpenSSF review.
- Clarify SSH allowed-signers setup for release tag verification.
- Add branch-protection-ready dependency and SAST gate workflows for OpenSSF
  review, covering GitHub Dependency Review, module verification,
  `govulncheck`, blocking `gosec -tests`, and same-repository gosec SARIF
  upload.
- Add benchmark coverage and a `task bench` facade for hot protocol, exporter,
  and parser paths with `-benchmem`.
- Add godoc examples for exporter domain separation, transcript IDs,
  confirmation-failure handling, and session close behavior.
- Add Capslock capability-analysis evidence for the external-review packet.
- Stage OSS-Fuzz project files for the existing native Go fuzz targets and
  validate them with the local OSS-Fuzz helper flow.
- Open the upstream OSS-Fuzz onboarding PR and refresh local ARM64 long-fuzz
  evidence for the merged review-readiness commit.
- Refresh dependency, gosec, Capslock, security/spec, and paired ARM/Intel
  one-hour fuzz evidence under Go 1.26.3 after the Go toolchain security
  release, with raw transcripts and SHA-256 digests.
- Document the calibrated evidence-artifact policy for release candidates,
  toolchain-security refreshes, and lighter external-review refreshes.
- Switch the README badge from the OpenSSF Baseline endpoint to the OpenSSF
  Best Practices `passing` endpoint.
- Move private local planning/interview/review artifacts into the private
  `the-sarge/meta` archive and remove stale public-repo ignore rules for those
  root files.
- Apply Go 1.26 `go fix` modernizations to internal crypto/framing loops and
  concurrent tests, with no intended Go API, wire/protocol, dependency, or
  vector behavior change.
- Refresh exact-current v0.1.2 candidate dependency, gosec, Capslock,
  security/spec, and paired ARM/Intel one-hour fuzz evidence.

## v0.1.1 - 2026-05-06

- Publish a CI/security-process hardening prerelease with no Go API, protocol,
  or dependency changes from `v0.1.0`.
- Add tag-triggered release validation covering tests, race tests,
  `govulncheck`, and `gosec`.
- Add CodeQL, OpenSSF Scorecard, Staticcheck Advisory, Actionlint, and
  cross-platform smoke workflows.
- Upload gosec SARIF to GitHub Code Scanning and keep Code Scanning open
  alerts at zero after false-positive/noise triage.
- Tighten workflow permissions, keep third-party actions SHA-pinned, and keep
  release tags as signed annotated tags.
- Clarify public fuzz evidence and ignore private local scratch/planning files.

## v0.1.0 - 2026-05-06

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
- Refresh dependency review with `govulncheck -test -show verbose ./...` and
  advisory `gosec v2.26.1`; clean up the LEB128 parser to avoid integer
  conversions flagged by gosec.
- Make `task fuzz` disable the outer Go test timeout and allow explicit
  `FUZZ_RACE=0` long campaigns after race-instrumented tests pass.
- Record fuzz evidence for all 14 registered fuzz targets, including a local
  smoke run and long ARM/Intel runs.
- Audit the security assessment and spec matrix against the merged
  release-readiness implementation and record the result in
  `docs/security-spec-audit.md`.
- Add application integration guidance for outer PAKE/version negotiation,
  downgrade protection, identity orientation, and CPace session outputs.
- License the project under BSD-3-Clause.
- Publish the first `v0.x` draft-21 snapshot. This release is not
  production-ready.
