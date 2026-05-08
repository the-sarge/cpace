# External Review Handoff

Date: 2026-05-08

Target module: `github.com/the-sarge/cpace`

Last released tag: `v0.1.1`

Last released commit: `74b82cbc65a1ea6186f2732749c9c5e5b03eecc3`

Current recorded evidence baseline:
`737bc56ffba81e2df5e9caa0df1ff180bfdb594b`

Evidence status: historical after the Go 1.26 `go fix` modernization until
dependency, fuzz, Capslock, and security/spec evidence is refreshed against the
post-merge package-code commit.

Status: auditable draft implementation. This package has not had independent
cryptographic review and is not production-ready.

## Review Goal

The immediate review goal is external review of the package-owned choices around
CPace draft-21 integration, wire framing, caller-facing profile policy, and
release evidence. This is separate from, and does not replace, independent
cryptographic review before any production-ready claim.

## Primary References

- `README.md` for the public API contract, integration warnings, and validation
  commands.
- `docs/security-assessment.md` for the current security self-assessment.
- `docs/spec-matrix.md` for the draft-21 requirement mapping.
- `docs/security-spec-audit.md` for the latest internal security/spec audit.
- `docs/threat-model.md` for assets, attackers, non-goals, and security
  boundaries.
- `docs/integration-guidance.md` for outer negotiation, downgrade protection,
  identity orientation, and session-output guidance.
- `docs/dependency-review.md` for dependency and vulnerability scan evidence.
- `docs/fuzz-evidence.md` for local smoke and long-fuzz campaign evidence.
- `docs/capslock-report.md` for static capability-analysis evidence.
- `docs/evidence/go1263-20260508/` for raw Go 1.26.3 transcript files and
  SHA-256 digests.
- `docs/performance.md` for local benchmark and allocation-measurement guidance.
- `docs/ci-policy.md` for hosted-runner policy, advisory lanes, long-fuzz
  evidence, and signed release tags.
- `docs/release-checklist.md` for exact-candidate release evidence steps.

## Implemented Scope

The package implements only `CPACE-RISTR255-SHA512` from
`draft-irtf-cfrg-cpace-21`.

The public API exposes initiator-responder mode only. A session is returned
only after explicit key confirmation succeeds. `Respond` returning success is
not authentication.

The package is intentionally not a generic CPace framework. It does not expose
other CPace suites, X25519/X448/NIST curves, symmetric mode, a raw-CI API, or
application negotiation. Applications must provide downgrade protection for any
outer negotiation that happens before CPace inputs are fixed.

## Package-Owned Choices To Review

- `cpace-go` CI construction from draft version, suite, role labels, initiator
  ID, responder ID, and caller context.
- Binary wire framing with format byte `0xc1`, suite and role bytes, and
  draft LEB128 length-value fields.
- Non-configurable per-field size caps: passwords and party IDs at 4 KiB,
  context and session IDs at 1 KiB, associated data at 64 KiB, and exact-sized
  public shares and confirmation tags.
- Default rejection of empty `SessionID`, with `AllowEmptySessionID` kept only
  for draft-21 compatibility or deliberately compatible profiles.
- Draft-compatible confirmation tag inputs, with no package-added role labels
  in the confirmation MACs.
- Scalar sampling profile: masked canonical 32-byte sampling with zero retry,
  following the draft-21 Ristretto255 recommendation rather than the allowed
  64-byte uniform-sampling alternative.
- `Session.Export` as HKDF-SHA512 over the confirmed ISK, and
  `Session.TranscriptID` as the draft `CPaceSidOutput` rather than a complete
  channel binding for outer negotiation.
- Best-effort session key cleanup through `Session.Close`, with no claim of
  resistance to local memory disclosure under the Go runtime.

## Evidence Snapshot

`v0.1.1` is an SSH-signed annotated prerelease tag at commit
`74b82cbc65a1ea6186f2732749c9c5e5b03eecc3`. The tag-triggered Release
Validation workflow passed `Check`, `Race`, `Govulncheck`, and `Gosec`; the
Gosec job uploaded SARIF to GitHub Code Scanning:

`https://github.com/the-sarge/cpace/actions/runs/25465518681`

The `v0.1.1` prerelease contains CI and documentation hardening only. It does
not change the Go API, protocol behavior, or dependencies from the earlier
draft snapshot.

Go 1.26.3 dependency, gosec, long-fuzz, Capslock, and security/spec evidence is
recorded for package-code commit
`737bc56ffba81e2df5e9caa0df1ff180bfdb594b`. The paired long-fuzz refresh ran
all 14 registered targets for `FUZZTIME=1h` on local ARM and Intel maintainer
machines. The Go 1.26 `go fix` modernization touches `crypto.go` and
`framing.go`, so that evidence becomes historical after the modernization
merges. Repeat dependency review, long fuzzing, Capslock, and security/spec
audit against the exact release candidate before any production-readiness
claim, or sooner if protocol, parser/framing, fuzz harness, dependency,
toolchain, or package-profile docs change.

Capslock capability-analysis evidence is recorded in
`docs/capslock-report.md` for
`737bc56ffba81e2df5e9caa0df1ff180bfdb594b`.

OSS-Fuzz onboarding is open upstream in `google/oss-fuzz#15480`. The upstream
PR helper build, header check, and Google CLA check passed; merge is waiting on
upstream review.

## Review Questions

- Is the package-owned CI construction appropriate for a Go package profile over
  draft-21, and are the role and identity-orientation requirements clear enough
  for real integrations?
- Is the binary wire framing unambiguous, injective for the represented fields,
  and sufficiently future-versioned?
- Are the per-field size caps reasonable for a library API, and are the
  associated-data warnings sufficient to keep callers from treating AD as a
  large payload channel?
- Is default rejection of empty session IDs the right package posture while
  preserving explicit draft compatibility?
- Are the scalar sampling, invalid-point handling, confirmation, exporter, and
  session lifecycle claims in the docs accurate and complete?
- Are the CI, dependency, fuzz, and release-tag controls sufficient evidence for
  an auditable prerelease, assuming independent cryptographic review remains
  required?

## Remaining Release Blockers

- Complete external review of package-owned framing, CI construction, and
  profile choices.
- Obtain independent cryptographic review before any production-ready claim.
- Refresh exact-release dependency review, long fuzz evidence, and
  security/spec audit after review-driven changes and before any
  production-readiness candidate.
- Resolve any critical or high review findings before moving beyond the `v0.x`
  prerelease line.
