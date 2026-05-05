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

This module is a package-specific `cpace-go` profile over draft-21. It builds
CI internally from the draft version, suite, roles, initiator ID, responder ID,
and caller context. It also owns its binary wire framing; applications should
treat message bytes as opaque and versioned by this module. The current wire
format prefix byte is `0xc1`.

Provide a fresh, non-secret `SessionID` agreed by both parties for every
session. Empty session IDs remain accepted because draft-21 only recommends
uniqueness, but they weaken replay and transcript separation properties.

`Initiator.Finish` and `Responder.Finish` are single-use calls. Passing a
malformed message or a message that fails confirmation consumes the state and
requires restarting the exchange.

`Config.Rand`, when set, must be a CSPRNG that provides fresh entropy for every
exchange. Deterministic readers are only appropriate in tests.

## Validation

This repository uses `Taskfile.yml` as the local and CI validation facade:

```sh
task check
FUZZTIME=30s PARALLEL=2 task fuzz
```

Fuzz targets live in `.github/fuzz-targets.json`. Pull-request CI runs the full
check gate plus a short fuzz smoke pass; `.github/workflows/nightly-fuzz.yml`
provides a manual long-fuzz matrix that uploads newly captured corpus files on
failure.

```go
initiator, msgA, err := cpace.Start(initCfg)
responder, msgB, err := cpace.Respond(respCfg, msgA)
msgC, initSession, err := initiator.Finish(msgB)
respSession, err := responder.Finish(msgC)
key, err := initSession.Export([]byte("application key"), nil, 32)
```

Release policy: keep tags in the `v0.x` range until independent review is
complete and the release bar in `docs/security-assessment.md` is satisfied.
