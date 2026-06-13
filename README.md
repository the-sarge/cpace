# CPace for Go

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/the-sarge/cpace/badge)](https://scorecard.dev/viewer/?uri=github.com/the-sarge/cpace)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12767/badge)](https://www.bestpractices.dev/projects/12767)

This repository implements `draft-irtf-cfrg-cpace-21` for the
`CPACE-RISTR255-SHA512` suite only.

Status: auditable draft implementation. This code has not had independent
cryptographic review and is not production-ready.

Current work is release readiness, not new policy design. The public API and
package-profile choices are frozen for review unless a new finding reopens a
decision. Before any production-readiness claim, the release bar in
`docs/security-assessment.md` must be satisfied.

The public API exposes only an initiator-responder flow with mandatory explicit
key confirmation:

1. `Start` returns initiator state and message A.
2. `Respond` consumes message A and returns responder state and message B.
3. `Initiator.Finish` verifies message B and returns message C plus `Session`.
4. `Responder.Finish` verifies message C and returns `Session`.

`Respond` returning success is not authentication. Only successful
`Initiator.Finish` and `Responder.Finish` calls return authenticated sessions.
Treat message B as an unauthenticated protocol message until the finish calls
complete.

This module is a package-specific `cpace-go` profile over draft-21. It builds
CI internally from the draft version, suite, roles, initiator ID, responder ID,
and caller context. It also owns its binary wire framing; applications should
treat message bytes as opaque and versioned by this module. The current wire
format prefix byte is `0xc1`.

Both sides must configure the same role orientation: `InitiatorID` is the party
that called `Start`, and `ResponderID` is the party that called `Respond`, on
both machines. A common integration bug is for each side to put its own
identity first; that makes the CI inputs differ and confirmation fails. Do not
use only globally hardcoded role labels such as `"client"` and `"server"` for
all users or deployments. Bind stable, application-meaningful party identities
into these fields.

Provide a fresh, non-secret `SessionID` agreed by both parties for every
session. Empty session IDs are rejected by default because they weaken replay
and transcript separation properties; that failure wraps both `ErrInvalidInput`
and `ErrEmptySessionID`. `AllowEmptySessionID` exists only for draft-21
compatibility tests or deliberately compatible profiles that accept the weaker
empty-sid behavior. If an outer protocol negotiates PAKE versions, ciphersuites,
or whether CPace is used at all, that negotiation needs its own downgrade
protection; this package does not authenticate negotiation it never sees. See
`docs/integration-guidance.md` for outer negotiation and downgrade-protection
guidance.

`Initiator.Finish` and `Responder.Finish` are single-use calls. Passing a
malformed message or a message that fails confirmation consumes the state and
requires restarting the exchange.

`Session.TranscriptID` is the draft `CPaceSidOutput` for the confirmed CPace
transcript, not a complete channel binding for outer protocol negotiation.
`Session.Export` derives deterministic, domain-separated application key
material from the confirmed ISK. It is not a source of fresh randomness; use
specific labels and contexts for each exported key purpose. `Session.Close`
performs best-effort cleanup of the session key material and makes future
`Export` calls fail with `ErrSessionClosed`; non-secret metadata accessors
remain available after close. `Session.PeerAssociatedData` returns the exact
peer AD bound into the confirmed exchange. `Session.PeerID` returns the
caller-configured peer identity that was bound into CI and confirmed by the
exchange; it is not parsed from peer-controlled wire data.

Scalar randomness is always drawn from Go's `crypto/rand.Reader`. Callers do
not provide randomness to `Start` or `Respond`.

Input and wire fields have package-owned per-field caps: passwords and party IDs are limited to 4 KiB, context and session IDs to 1 KiB, and associated data to 64 KiB. Valid package-owned message shapes are governed by those per-field caps and exact public-share/tag lengths; malformed framed inputs also hit a 128 KiB aggregate decoder backstop before field parsing proceeds. Associated data should bind protocol context, not carry large payloads; represent large external artifacts with a digest, Merkle root, exporter, or other fixed-size commitment.

## Validation

This repository uses `Taskfile.yml` as the local validation facade:

```sh
task quick
task check
task check:changed
task docs:check
task bench
FUZZTIME=30s PARALLEL=2 task fuzz
FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz
```

`task check` requires the normal Go/tooling prerequisites plus `jq`, because the full gate runs release-helper smoke tests that validate CycloneDX SBOM JSON.

Fuzz targets live in `.github/fuzz-targets.json`. GitHub-hosted pull-request CI
is intentionally light because it runs untrusted fork code: code changes run
`go test ./...`, and Markdown-only PRs run docs validation without setting up
Go. Scheduled hosted lanes run `govulncheck`, advisory `gosec`, and a
5-minute-per-target fuzz regression pass as background signal. Run the full
local gate, longer maintainer-machine fuzzing, vulnerability scan, and advisory
`gosec` scan locally before release-oriented changes. The default fuzz lane
keeps the race detector on for smoke runs; use `FUZZ_RACE=0` for longer
campaigns after `task check` has already covered race-instrumented tests. See
`docs/ci-policy.md` for the hosted and self-hosted runner threat model.

Release-readiness work should record exact evidence: commit SHA, command or
workflow, duration for fuzzing, target count, and residual risks. Dependency
review evidence lives in `docs/dependency-review.md`; fuzz campaign evidence
lives in `docs/fuzz-evidence.md`; security/spec audit evidence lives in
`docs/security-spec-audit.md`; Capslock capability-analysis evidence lives in
`docs/capslock-report.md`; local benchmark guidance lives in
`docs/performance.md`. External reviewer scope and review questions are
summarized in `docs/external-review-handoff.md`; security boundaries are
summarized in `docs/threat-model.md`. Downstream release verification
instructions live in `docs/release-verification.md`. Project governance,
security-gate, and VEX policies live in `docs/governance.md`,
`docs/security-gates.md`, and `docs/vex.md`.

```go
initiator, msgA, err := cpace.Start(initCfg)
responder, msgB, err := cpace.Respond(respCfg, msgA)
msgC, initSession, err := initiator.Finish(msgB)
defer initSession.Close()
respSession, err := responder.Finish(msgC)
defer respSession.Close()
key, err := initSession.Export([]byte("application key"), nil, 32)
```

Release policy: keep tags in the `v0.x` range until independent review is
complete and the release bar in `docs/security-assessment.md` is satisfied.
Use `docs/release-checklist.md` for future release candidates. See
`CONTRIBUTING.md` before opening public issues or pull requests; commits must
include DCO signoffs.

## License

BSD-3-Clause. See `LICENSE`.
