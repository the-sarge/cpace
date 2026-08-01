# External Review Handoff

Date: 2026-07-31

Target module: `github.com/the-sarge/cpace`

Last released tag: `v0.1.2`

Last released commit: `4e661bc1f925ebedf1f270668129d85bab73e468`

Current evidence status: see `docs/evidence-baseline.md`. The dependency, vulnerability, and SAST/gosec lane is current for the frozen `v0.1.3` candidate in `docs/evidence/1f19e27-20260731/`; Capslock, security/spec audit, and Go 1.26.4 to Go 1.26.5 vector stability are current in `docs/evidence/1f19e27-20260731-protocol/`; paired long fuzz is current in `docs/evidence/1f19e27-20260731-fuzz/` with a recorded Intel all-target deadline non-pass and same-host targeted recovery. The original `v0.1.2` packet and `docs/evidence/f7efa6a-20260619/` remain historical for the still-pending release-control lanes.

Latest complete multi-lane evidence packet: `docs/evidence/f7efa6a-20260619/`, pinned to package-code candidate `f7efa6a963a954952b1ecad3f46530f13799fe89`. It is unaudited historical prerelease evidence, not current-candidate evidence or a production-readiness claim.

Frozen `v0.1.3` candidate source: `1f19e278112fa037890848ed6c086addeffdca4e`. Review package code against that exact commit. The dependency, vulnerability, SAST/gosec, Capslock, security/spec, cross-toolchain vector, and paired long-fuzz lanes are fresh for this candidate under Go 1.26.5; release-control evidence remains historical and must still be refreshed before `v0.1.3`.

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
- `docs/security-spec-audit.md` for the latest completed exact-candidate internal security/spec audit and cross-toolchain vector result.
- `docs/threat-model.md` for assets, attackers, non-goals, and security
  boundaries.
- `docs/integration-guidance.md` for outer negotiation, downgrade protection, role-local identity input, and session-output guidance.
- `docs/dependency-review.md` for dependency and vulnerability scan evidence.
- `docs/fuzz-evidence.md` for local smoke and long-fuzz campaign evidence.
- `docs/capslock-report.md` for static capability-analysis evidence.
- `docs/evidence-baseline.md` for the current evidence-status and stale-trigger index.
- `docs/evidence/1f19e27-20260731/` for exact-candidate dependency/SAST raw evidence and SHA-256 digests.
- `docs/evidence/1f19e27-20260731-protocol/` for exact-candidate Capslock, security/spec, focused protocol, and cross-toolchain vector raw evidence and SHA-256 digests.
- `docs/evidence/f7efa6a-20260619/` for the historical multi-lane raw evidence bundle and SHA-256 digests.
- `docs/evidence/v012-candidate-20260508/` for raw v0.1.2 candidate transcript
  files and SHA-256 digests.
- `docs/evidence/v012-soak-20260509/` for raw v0.1.2 supplemental fuzz soak
  transcripts and SHA-256 digests, including the iMacPro all-target non-pass
  and clean same-host targeted `FuzzProtocolConsistency` rerun.
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

- `cpace-go` CI construction from draft version, suite, role labels, role-local party identities, and caller context.
- Binary wire framing with format byte `0xc1`, suite and role bytes, and
  draft LEB128 length-value fields.
- Non-configurable per-field size caps: passwords and party IDs at 4 KiB,
  context and session IDs at 1 KiB, associated data at 64 KiB, and exact-sized
  public shares and confirmation tags.
- Default rejection of empty `SessionID`, with `AllowEmptySessionID` kept only
  for draft-21 compatibility or deliberately compatible profiles.
- Draft-compatible confirmation tag inputs, with no package-added role labels
  in the confirmation MACs.
- Scalar sampling profile: masked canonical 32-byte sampling with a bounded
  defensive retry loop whose only reachable retry is an all-zero masked sample,
  following the draft-21 Ristretto255 recommendation rather than the allowed
  64-byte uniform-sampling alternative.
- `Session.Export` as HKDF-SHA512 over the confirmed ISK, and
  `Session.TranscriptID` as the draft `CPaceSidOutput` rather than a complete
  channel binding for outer negotiation.
- Best-effort session key cleanup through `Session.Close`, with no claim of
  resistance to local memory disclosure under the Go runtime.

## Evidence Snapshot

`v0.1.2` is an SSH-signed annotated prerelease tag at commit
`4e661bc1f925ebedf1f270668129d85bab73e468`. The tag-triggered Release
Validation workflow passed `Check`, `Race`, `Govulncheck`, and `Gosec`; the
Gosec job uploaded SARIF to GitHub Code Scanning:

`https://github.com/the-sarge/cpace/actions/runs/25588835119`

The `v0.1.2` prerelease contains the external-review packet, Go 1.26
modernization, and refreshed evidence. It has no intended Go API,
wire/protocol, dependency, or vector behavior change.

The current evidence status and freshness caveats are indexed in `docs/evidence-baseline.md`. The Go 1.26.5 dependency/SAST, protocol/capability, and paired long-fuzz bundles cover frozen candidate source `1f19e278112fa037890848ed6c086addeffdca4e`, including bit-identical normalized vector outputs under Go 1.26.4 and Go 1.26.5 and a same-host targeted fuzz recovery for the recorded Intel deadline miss. The historical `f7efa6a963a954952b1ecad3f46530f13799fe89` bundle covers release-control lanes only for that older Go 1.26.4 commit; repeat those lanes against the frozen candidate before a complete current-candidate packet.

Capslock capability-analysis evidence is recorded in `docs/capslock-report.md`; the exact-candidate rerun found unchanged classes and counts, and its pinned baseline and freshness caveat are indexed in `docs/evidence-baseline.md`.

OSS-Fuzz onboarding is not currently available. Upstream PR `google/oss-fuzz#15480` had passed the PR helper build, header check, and Google CLA check, but it was closed on 2026-05-11 after OSS-Fuzz maintainers declined the project for current project-size/user-base reasons and suggested ClusterFuzzLite instead.

Current compensating fuzz signal remains maintainer-controlled long fuzzing, scheduled hosted fuzzing, and autoscaled trusted-runner fuzzing; possible ClusterFuzzLite integration would be future additional CI signal rather than a replacement release claim.

## Review Questions

- Is the package-owned CI construction appropriate for a Go package profile over draft-21, and are the role-local identity-input requirements clear enough for real integrations?
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
- Refresh exact-release dependency review, long fuzz evidence, Capslock
  capability evidence, and security/spec audit after review-driven or
  security-relevant changes before any production-readiness claim.
- Resolve any critical or high review findings before moving beyond the `v0.x`
  prerelease line.
