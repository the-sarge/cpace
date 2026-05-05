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

The package owns its binary wire framing. Applications should treat the message
bytes as opaque and versioned by this module.

```go
initiator, msgA, err := cpace.Start(initCfg)
responder, msgB, err := cpace.Respond(respCfg, msgA)
msgC, initSession, err := initiator.Finish(msgB)
respSession, err := responder.Finish(msgC)
key, err := initSession.Export([]byte("application key"), nil, 32)
```

Release policy: keep tags in the `v0.x` range until independent review is
complete and the release bar in `docs/security-assessment.md` is satisfied.
