# CPace for Go

This repository implements `draft-irtf-cfrg-cpace-21` for the
`CPACE-RISTR255-SHA512` suite only.

Status: auditable draft implementation. This code has not had independent
cryptographic review and is not production-ready.

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
session. Empty session IDs remain accepted because draft-21 only recommends
uniqueness, but they weaken replay and transcript separation properties. If an
outer protocol negotiates PAKE versions, ciphersuites, or whether CPace is used
at all, that negotiation needs its own downgrade protection; this package does
not authenticate negotiation it never sees.

`Initiator.Finish` and `Responder.Finish` are single-use calls. Passing a
malformed message or a message that fails confirmation consumes the state and
requires restarting the exchange.

`Session.TranscriptID` is the draft `CPaceSidOutput` for the confirmed CPace
transcript, not a complete channel binding for outer protocol negotiation.
`Session.Export` derives deterministic, domain-separated application key
material from the confirmed ISK. It is not a source of fresh randomness; use
specific labels and contexts for each exported key purpose.

`Config.Rand`, when set, must be a CSPRNG that provides fresh entropy for every
exchange. Deterministic readers are only appropriate in tests.

## Validation

This repository uses `Taskfile.yml` as the local validation facade:

```sh
task quick
task check
task check:changed
task docs:check
FUZZTIME=30s PARALLEL=2 task fuzz
```

Fuzz targets live in `.github/fuzz-targets.json`. Until self-hosted runners are
available, GitHub-hosted pull-request CI is intentionally light: code changes
run `go test ./...`, and Markdown-only PRs run docs validation without setting
up Go. Run the full local gate, fuzzing, vulnerability scan, and advisory
`gosec` scan locally before release-oriented changes.

```go
initiator, msgA, err := cpace.Start(initCfg)
responder, msgB, err := cpace.Respond(respCfg, msgA)
msgC, initSession, err := initiator.Finish(msgB)
respSession, err := responder.Finish(msgC)
key, err := initSession.Export([]byte("application key"), nil, 32)
```

Release policy: keep tags in the `v0.x` range until independent review is
complete and the release bar in `docs/security-assessment.md` is satisfied.
