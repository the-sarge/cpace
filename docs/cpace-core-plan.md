# CPace core — deepening plan

Implementation plan for the refactor recorded in
[ADR-0001](adr/0001-extract-cpace-core.md): extract a deep, unexported **CPace
core** so the cryptographic composition and persistent-secret lifetime have one
home. Internal-only — the public API and profile policy stay frozen.

This plan incorporates three multi-agent reviews (`ras consider`): phase 1 on
ADR-0001 (run `20260522T081906-6c67083f870be5ac1f971508`), phase 2 on this plan
(run `20260522T150534-dc141248a1c30a4d025c1c1f`), and a phase-2 re-run
confirming the revisions (run `20260522T152641-7d2ca5ac0cd36d5a1062254b`). All
three returned *proceed with changes*; the disposition tables below record how
each round's findings were resolved.

Revision 4 (2026-06-10) folds in a five-perspective review of this branch: gating-language and enumeration fixes, two recorded decisions (zero-value hardening; sequencing against the external reviews), confirmation-tag golden capture, and audit-checklist additions. Pending the confirming `ras consider` round recorded in ADR-0001 — implementation must not begin before it passes.

## Goal

The CPace cryptographic composition — generator derivation, scalar sampling,
Diffie-Hellman, transcript assembly, ISK derivation, confirmation tags — has no
module of its own. It is smeared across `crypto.go` primitives, `strings.go`
transcript builders, and four orchestration functions in `api.go`. The
persistent-secret invariant is enforced by hand-placed deferred closures in the
two `Finish` functions; the constructors' `defer`s cover scratch secrets only.

Extract a deep `initiatorCore` / `responderCore` — collectively the **CPace
core** (see `CONTEXT.md`). The public `Initiator` / `Responder` become thin
shells. The core owns one role's cryptographic computation and the lifetime of
its persistent secrets, so the persistent-secret audit concentrates in two
`clear()` methods instead of a trace across four functions.

## Design — locked decisions

| Decision | Resolution |
|---|---|
| Candidate | A — extract a deep CPace core |
| Secret ownership | Stateful core objects own **persistent**-secret lifetime (initiator scalar; responder ISK — plus the responder transcript, public wire data zeroed as hygiene). Scratch secrets stay local, cleared eagerly. |
| Seam content | Decoded cryptographic fields cross; wire framing stays in front. Responder core validates `Ya` before sampling. |
| Randomness | Core constructor takes `io.Reader` (new core-test seam). `startWithRandom` / `respondWithRandom` **retained**, unexported, as the full-pipeline seam. |
| `buildCI` | **In front** — `buildCI` runs inside `normalizeConfig`; `normalizedConfig` keeps its `ci` field unchanged. (Revises the earlier "behind the seam" call — lower churn, and it removes the seam contradiction phase-2 flagged.) |
| Session | Core constructs it; only the Session's independent ISK clone persists past `clear()`. |
| Naming | `initiatorCore` / `responderCore` — concept **CPace core** (`CONTEXT.md`). |
| Tests | Primitive-level vector tests retained as internal-seam tests; add core-level vector tests + an ISK-isolation test. |

## Target shape

The sketches below are **literal** — every defensive `defer` and guard the
current code relies on is shown, because an implementer follows the sketch.

```go
// api.go — public shells; frozen surface unchanged
type Initiator struct {
    mu   sync.Mutex
    used bool
    core *initiatorCore
}

func Start(cfg Config) (*Initiator, []byte, error) {
    return startWithRandom(cfg, rand.Reader)
}

// startWithRandom stays unexported — the deterministic full-pipeline seam for
// api_test.go / fuzz_test.go / bench_test.go. It owns the password backstop.
func startWithRandom(cfg Config, random io.Reader) (*Initiator, []byte, error) {
    nc, err := normalizeConfig(cfg)
    if err != nil {
        return nil, nil, err
    }
    defer clearBytes(nc.password)            // backstop — fires on core-ctor error/panic too
    core, ya, err := newInitiatorCore(nc, random)
    if err != nil {
        return nil, nil, err
    }
    return &Initiator{core: core}, encodeMessageA(nc.sid, ya, nc.ad), nil
}

func (i *Initiator) Finish(messageB []byte) ([]byte, *Session, error) {
    if i == nil || i.core == nil {
        return nil, nil, fmt.Errorf("%w: uninitialized initiator", ErrInvalidInput)
    }
    if err := i.consume(); err != nil {
        return nil, nil, err
    }
    defer i.core.clear()                     // every path after consume
    b, err := decodeMessageB(messageB)
    if err != nil {
        return nil, nil, err
    }
    tagA, sess, err := i.core.finish(b.yb, b.adb, b.tag)
    if err != nil {
        return nil, nil, err
    }
    return encodeMessageC(tagA), sess, nil
}

func Respond(cfg Config, messageA []byte) (*Responder, []byte, error) {
    return respondWithRandom(cfg, messageA, rand.Reader)
}

func respondWithRandom(cfg Config, messageA []byte, random io.Reader) (*Responder, []byte, error) {
    nc, err := normalizeConfig(cfg)
    if err != nil {
        return nil, nil, err
    }
    defer clearBytes(nc.password)            // backstop — fires on core-ctor error/panic too
    a, err := decodeMessageA(messageA)
    if err != nil {
        return nil, nil, err
    }
    if !bytes.Equal(a.sid, nc.sid) {
        return nil, nil, fmt.Errorf("%w: session id mismatch", ErrMessage)
    }
    core, yb, tagB, err := newResponderCore(nc, a.ya, a.ada, random)
    if err != nil {
        return nil, nil, err
    }
    return &Responder{core: core}, encodeMessageB(yb, nc.ad, tagB), nil
}

func (r *Responder) Finish(messageC []byte) (*Session, error) {
    if r == nil || r.core == nil {
        return nil, fmt.Errorf("%w: uninitialized responder", ErrInvalidInput)
    }
    if err := r.consume(); err != nil {
        return nil, err
    }
    defer r.core.clear()                     // every path after consume
    c, err := decodeMessageC(messageC)
    if err != nil {
        return nil, err
    }
    return r.core.finish(c.tag)
}
```

The `i.core == nil` / `r.core == nil` guards are deliberate hardening, recorded in ADR-0001 as a **narrow policy reopen** (decided 2026-06-10). `Initiator` and `Responder` are exported structs, so a caller *can* fabricate a zero value. Today `Finish` on one consumes the single-use state and then: the **initiator** returns `ErrMessage` (malformed message B) or `ErrAbort` (invalid or identity share), panicking on its nil scalar only when the share is valid; the **responder** returns `ErrMessage`/`ErrConfirmationFailed` — or **succeeds** against a crafted message C, because a zero-value responder's expected tag is computed from all-nil inputs with no secret material (a public constant), handing the caller a real `*Session` keyed from a nil ISK whose `Export` output is attacker-predictable. With the guards, `Finish` returns `ErrInvalidInput` **without** consuming — pinned by `TestFinishZeroValueHardening` (build step 5) and a changelog note that must state the forged-tag success path, per the ADR's acceptance criteria. The error text says "uninitialized", not "nil", because the merged guard also fires for a non-nil shell whose `core` is nil; this changes the nil-receiver message text from "nil initiator"/"nil responder" too (error identity is unchanged).

Two load-bearing invariants the sketches rely on, stated so no implementer relaxes them: **the shells never reassign or nil `.core` after construction** — cleanup nils the core's *fields*, never the pointer — so the unsynchronized `i.core == nil` read is race-free and a second `Finish` still reaches `consume()` and returns `ErrStateUsed`, not `ErrInvalidInput`; and **core methods are never called after `clear()`** — the shell single-use guard is the enforcement boundary, and post-`clear()` core behavior is out of contract (the initiator path would nil-deref; the responder path would compute a tag over a nil ISK).

```go
// core.go — unexported, deep
type initiatorCore struct {
    scalar       *ristretto255.Scalar  // persistent secret — owned by clear()
    sid, ya, ada []byte
    peerID       []byte
}
type responderCore struct {
    isk          []byte  // persistent secret — owned by clear()
    transcript   []byte
    sid, ya, ada []byte
    peerID       []byte
}

func newInitiatorCore(nc normalizedConfig, random io.Reader) (*initiatorCore, []byte, error) {
    if random == nil {
        random = rand.Reader                 // nil-randomness guard lives here, the seam
    }
    g := calculateGenerator(nc.password, nc.ci, nc.sid)
    defer clearElement(g)                    // scratch
    clearBytes(nc.password)                  // scratch — eager, narrowest scope
    nc.password = nil
    y, err := sampleScalar(random)
    if err != nil {
        return nil, nil, err                 // wrapper's defer backstops nc.password
    }
    ya := scalarMult(y, g)
    return &initiatorCore{
        scalar: y, sid: clone(nc.sid), ya: clone(ya),
        ada: clone(nc.ad), peerID: clone(nc.responderID),
    }, ya, nil
}

func (c *initiatorCore) finish(peerYb, peerAdb, peerTag []byte) ([]byte, *Session, error) {
    k, ok := scalarMultVFY(c.scalar, peerYb)
    defer clearBytes(k)                       // scratch
    if !ok {
        return nil, nil, fmt.Errorf("%w: invalid responder share", ErrAbort)
    }
    tr := transcriptIR(c.ya, c.ada, peerYb, peerAdb)
    isk := deriveISK(c.sid, k, tr)
    defer clearBytes(isk)                     // scratch — initiator finish-local ISK
    if !hmac.Equal(confirmationTag(isk, c.sid, peerYb, peerAdb), peerTag) {
        return nil, nil, ErrConfirmationFailed
    }
    tagA := confirmationTag(isk, c.sid, c.ya, c.ada)
    return tagA, newSession(isk, tr, peerAdb, c.peerID), nil  // newSession clones isk
}

func newResponderCore(nc normalizedConfig, peerYa, peerAda []byte, random io.Reader) (*responderCore, []byte, []byte, error) {
    if random == nil {
        random = rand.Reader
    }
    if _, ok := decodePublicShare(peerYa); !ok {   // validate Ya FIRST — before generator / sampling
        return nil, nil, nil, fmt.Errorf("%w: invalid initiator share", ErrAbort)
    }
    g := calculateGenerator(nc.password, nc.ci, nc.sid)
    defer clearElement(g)                    // scratch
    clearBytes(nc.password)                  // scratch — eager
    nc.password = nil
    y, err := sampleScalar(random)
    if err != nil {
        return nil, nil, nil, err
    }
    defer clearScalar(y)                     // scratch — responder scalar is NOT persistent
    yb := scalarMult(y, g)
    k, ok := scalarMultVFY(y, peerYa)
    defer clearBytes(k)                      // scratch
    if !ok {
        return nil, nil, nil, fmt.Errorf("%w: invalid initiator share", ErrAbort)
    }
    tr := transcriptIR(peerYa, peerAda, yb, nc.ad)
    isk := deriveISK(nc.sid, k, tr)          // PERSISTENT — stored on the core
    tagB := confirmationTag(isk, nc.sid, yb, nc.ad)
    return &responderCore{
        isk: isk, transcript: tr, sid: clone(nc.sid),
        ya: clone(peerYa), ada: clone(peerAda), peerID: clone(nc.initiatorID),
    }, yb, tagB, nil
}

func (c *responderCore) finish(peerTagC []byte) (*Session, error) {
    if !hmac.Equal(confirmationTag(c.isk, c.sid, c.ya, c.ada), peerTagC) {
        return nil, ErrConfirmationFailed
    }
    return newSession(c.isk, c.transcript, c.ada, c.peerID), nil  // newSession clones isk
}

// clear() zeroes THEN nils each persistent-secret field; a second call finds
// nil and is a safe no-op.
func (c *initiatorCore) clear() {
    if c == nil {
        return
    }
    clearScalar(c.scalar)
    c.scalar = nil
}
func (c *responderCore) clear() {
    if c == nil {
        return
    }
    clearBytes(c.isk)
    clearBytes(c.transcript)
    c.isk = nil
    c.transcript = nil
}
```

The CPace IR asymmetry is intentional: `newResponderCore` performs the DH and
ISK derivation; for the initiator that work lands in `finish`. The responder
holds no scalar field — its scalar is a scratch secret cleared inside the
constructor. The `responderCore` carries no `adb` field — the responder's own
associated data is already baked into the stored `transcript` and `tagB` — and
no `yb` field either: like `adb`, the responder's own share is read nowhere
after construction. (Today's `Responder` stores both; both are dead weight the
sketch deliberately drops.)

The sketches show `scalarMultVFY`'s current `([]byte, bool)` shape. ADR-0003 (peer-share error semantics, `proposed`, review gate satisfied) changes it to `([]byte, error)` with exported sentinels and role-context wrapping at call sites. If 0003 is accepted before this plan executes, implement step 2 against the 0003 shape (`k, err := scalarMultVFY(...)`, wrapping per 0003's call-site examples); the seam placement and the clearing structure here are unaffected either way.

## The seam

**Behind the seam (the CPace core):** generator derivation, scalar sampling,
Diffie-Hellman and peer-share validation, transcript assembly, ISK derivation,
confirmation tag build and verify, `*Session` construction, and the clearing of
both persistent and scratch secrets.

**In front (`api.go` shell):** `Config` normalization — *including* `buildCI`,
which runs inside `normalizeConfig` so `normalizedConfig.ci` reaches the core
prebuilt — validation, public error wrapping; wire framing
(`encode`/`decodeMessage*`); the single-use guard; the `a.sid == nc.sid`
message-vs-config check; and the unexported `startWithRandom` /
`respondWithRandom` randomness wrappers with their password backstop `defer`.

Error ownership at the seam: the shells mint config and framing errors (`ErrInvalidInput`, `ErrMessage`); the core returns protocol errors ready-made (`ErrAbort` wraps, `ErrConfirmationFailed`, and `ErrRandomness` propagated from `sampleScalar`), and the shells pass them through without re-wrapping — exactly as the sketches show. An implementer must not introduce internal sentinels that the shell re-wraps; that shape risks double-wrapping and changed error identity.

## The `clear()` contract

`clear()` is the single audit surface for persistent secrets, so its trigger
contract is pinned:

- **Idempotent & nil-safe** — `clear()` on a nil core returns without
  panicking. `clear()` zeroes **then nils** each persistent-secret field, so a
  second call finds nil and is a safe no-op.
- **Deferred on every path** — each `Finish` runs `defer core.clear()`
  immediately after `consume()` succeeds, so the persistent secret is zeroed on
  parse failure, confirmation failure, and success alike.
- **Partial-construction safe (forward-looking)** — no constructor error path
  currently returns *after* a persistent secret exists: the initiator scalar
  and the responder ISK are each created on the last fallible step or later, so
  this clause is **vacuous for the present constructor shapes**. It binds any
  future constructor that gains an earlier persistent-secret creation — such a
  constructor must clear that secret before returning an error. No test is
  mandated for it while it stays vacuous (see [Tests](#tests)).
- **Persistent scope only** — `clear()` owns the initiator scalar and the
  responder ISK + transcript (the transcript is public wire data, zeroed as
  hygiene alongside the ISK — ADR-0001 records the same framing). Scratch
  secrets — the normalized password, the generator element, the DH point `k`,
  the responder's own scalar, **and the initiator's `finish`-local ISK** — are
  never core fields; each is cleared eagerly at the narrowest scope inside the
  core method that creates it, by local `defer` or inline immediately after
  last use (the password is cleared inline, with the shell backstop `defer` in
  front of the seam). Derivation buffers are likewise scratch, cleared inside
  the `crypto.go` primitives the core calls (`deriveISK`, `confirmationTag`,
  `calculateGenerator`) — with the `lvCat`/`prependLen` intermediates excepted
  as recorded residual risk (see the audit checklist).

## Build sequence

Ordered so the regression net catches mistakes early and the dangerous step
lands with its tests already in place.

1. **Baseline.** `task check` and a fuzz smoke run pass on `main` — the
   behavioral oracle for everything that follows. From this unmodified
   baseline, capture confirmation-tag goldens for the draft-vector inputs and
   commit them to `testdata/` (see step 4 — the draft vectors carry no tags, so
   these goldens are the only direct bit-equivalence anchor for the tag path).
2. **Extract `core.go`, one role at a time** (Initiator first), with the
   `io.Reader` constructors **from the first extraction commit** —
   `newInitiatorCore` / `newResponderCore` take `io.Reader`; `startWithRandom` /
   `respondWithRandom` pass it through. Each role's extraction moves that role's
   *full* crypto — the constructor **and** the `finish` method — into `core.go`,
   leaving the shell `Finish` as wire framing plus the interim clearing
   `defer`s. Move logic verbatim, with two pinned exceptions that reconcile
   *verbatim* with the literal sketches: (a) `Initiator.Finish` today clears
   its finish-local ISK with two explicit per-path `clearBytes(isk)` calls; the
   Initiator extraction commit canonicalizes this to the sketch's single
   `defer clearBytes(isk)` immediately after `deriveISK` — identical coverage
   plus panic paths. (b) Today's scalar snapshot (`scalar := i.scalar`) before
   the clearing closure disappears with the move; the interim shell `defer`
   reads `i.core.scalar` at fire time, equivalent because nothing reassigns the
   field between `consume()` and return. The responder's `Ya` prevalidation
   moves into `newResponderCore` in the Responder extraction commit — it never
   sits in the shell (step 3 then *pins* the ordering with its test).
   Persistent-secret clearing is **preserved
   verbatim** as interim `defer`s in the shell `Finish` (e.g.
   `defer func(){ clearScalar(i.core.scalar); i.core.scalar = nil }()`),
   including the `= nil` assignment — zeroization never regresses across commits
   2→5. Migrate `TestFinishCleanupDoesNotAliasReturnedSessions` **in step with
   the extractions**: the Initiator commit migrates `initiator.scalar →
   initiator.core.scalar` only (the responder assertions still read
   `responder.isk` / `responder.transcript`); the Responder commit migrates
   `responder.isk` / `responder.transcript → responder.core.*`. The test
   compiles and stays green at every commit. Tests green after each role.
3. **Pin `Ya` validation order in the responder core.** `newResponderCore`
   decodes and validates `Ya` (canonical, non-identity) **before** generator
   derivation and scalar sampling.
   `TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness` stays green.
4. **Add core-level draft-vector tests.** Drive `newInitiatorCore` /
   `newResponderCore` and `finish` with draft vector inputs; assert `ya` / `yb`
   / `isk` against the draft vectors and the confirmation tags against the
   step-1 goldens (the draft defines no tag values — the `"CPaceMac"`
   construction is package-local, so main-captured goldens are the oracle).
   Note on entropy plumbing: the test reader feeds raw canonical scalar bytes
   through the core's `io.Reader` seam; this reproduces the vector scalars only
   because `sampleScalar`'s `b[31] &= 0x0f` mask is the identity for them. A
   future vector with byte 31 ≥ `0x10` cannot be injected through the sampler
   and would need direct core-field construction. Primitive-level vector tests
   stay.
5. **Consolidate persistent-secret clearing into `clear()` — the dangerous
   step, done test-first.** First write the `clear()`-contract tests (nil-safe,
   double-`clear()` idempotence, `Finish` parse-failure and confirmation-failure
   cleanup for both roles), `TestSessionISKSurvivesCoreClear`, and
   `TestFinishZeroValueHardening` — they fail
   (red). Then implement `core.clear()` per
   [the contract](#the-clear-contract) and replace the interim `defer`s with
   `defer core.clear()` (green). Step 5 only *consolidates* clearing that is
   already present — it introduces none. The "red" observation is local-only:
   the new tests reference `core.clear()`, which does not compile before the
   implementation half exists, so tests and implementation land in the same
   commit (red observed by stashing the implementation locally) and every
   *committed* state stays green per the per-commit gate.
   ⚠️ Tests prove behavior, not zeroization — the manual audit below is
   mandatory.
6. **Refresh evidence.** Re-run the fuzz corpus; update `docs/fuzz-evidence.md`
   and `docs/security-spec-audit.md`.

## Tests

- **Regression net — survive unchanged:** `bench_test.go` and
  `examples_test.go` drive the frozen public interface and the retained
  `startWithRandom` / `respondWithRandom` wrappers.
  `TestInternalRandomHelpersDefaultNilRandomness` also stays green unchanged —
  the `nil → rand.Reader` guard is preserved in the core constructors.
- **Migrated in build step 2:** `TestFinishCleanupDoesNotAliasReturnedSessions`
  white-box-reads the fields the refactor relocates; it is migrated to reach
  `.core.scalar` / `.core.isk` / `.core.transcript`, staged per role with the
  extraction commits (see step 2). `api_test.go` is therefore *not* "unchanged".
- **Retained seam:** `startWithRandom` / `respondWithRandom` stay unexported;
  `fuzz_test.go`'s `repeatingRand` injection is unchanged.
- **New — core-level vector tests:** drive `newInitiatorCore` /
  `newResponderCore` and `finish` with draft vector inputs. These construct
  `normalizedConfig` directly (it is package-private and test-constructible),
  including a raw `ci` where a draft vector specifies CI rather than IDs. The
  generator-from-CI primitive stays covered by the retained primitive-level
  `vectors_test.go` tests, which already feed raw draft CI to
  `calculateGenerator`.
- **New — ISK deep-clone isolation test (`TestSessionISKSurvivesCoreClear`):**
  this is a **responder** test — `responderCore` is the only role whose core
  ISK persists until `clear()`. Complete a handshake, build the responder's
  `Session`, call `responderCore.clear()`, then assert `Session.Export` still
  returns the correct non-zero bytes while `responderCore.isk` reads as zeroed.
  An initiator-side variant, if wanted, asserts a *different* property: after
  `initiatorCore.finish`, `clear()` zeroes the scalar, and the finish-local ISK
  never aliased the Session's cloned ISK. Written test-first in build step 5.
- **New — zero-value guard pinning test (`TestFinishZeroValueHardening`):**
  both roles. `Finish` on a caller-fabricated zero-value `Initiator` /
  `Responder` returns `ErrInvalidInput` **without** consuming the single-use
  state — covering a malformed message and, for the responder, the crafted
  message C whose forged tag *succeeds* today (the regression this guard
  exists to close). Written test-first in build step 5 alongside the
  `clear()`-contract tests.
- **Internal-seam tests — retained:** `vectors_test.go`'s primitive-level
  checks (`calculateGenerator`, `scalarMultVFY`, transcript builders) still
  pinpoint which primitive diverges from the spec.
- **Precedence — preserved:**
  `TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness`.
- **Not tested — partial-construction cleanup:** the `clear()`-contract
  partial-construction clause is vacuous for the current constructor shapes (no
  error path returns after a persistent secret exists), so no test is mandated.
  A failure-injection seam must **not** be added to the constructors to
  manufacture one.

## Phase-1 findings — disposition

| # | Finding | Sev | Addressed in this plan |
|---|---|---|---|
| 1 | Scratch secrets keep eager local zeroization; `clear()` owns persistent only | High | Design; Target shape; `clear()` contract; Build step 5 |
| 2 | Retain `startWithRandom` / `respondWithRandom` | High | Design (Randomness); Target shape; Build step 2; Tests |
| 3 | Specify the `clear()` trigger contract | Med | `clear()` contract; Build step 5 |
| 4 | Pin `Ya` validation order | Med | The seam; Build step 3; Tests (precedence preserved) |
| 5 | Scope the zeroization guarantee for abandoned states | Med | ADR-0001 Consequences; Out of scope |
| 6 | ISK deep-clone isolation regression test | Low | Tests (new); Build step 5 |

## Phase-2 findings — disposition

| # | Finding | Sev | Addressed in this revision |
|---|---|---|---|
| 1 | Password leak — restore `defer clearBytes(nc.password)` backstop | High | Target shape (shell sketches); Verification audit checklist |
| 2 | Initiator `finish`-local ISK is a scratch secret missing from enumerations | High | Target shape (`initiatorCore.finish`); `clear()` contract; ADR-0001; Verification |
| 3 | `api_test.go` not "unchanged"; white-box test compile-breaks at step 2 | High | Tests; Build step 2 (migration); `clear()` contract (zero-then-nil) |
| 4 | `io.Reader` seam must be created in step 2, not step 3 | High | Build step 2 (constructors take `io.Reader` from the first commit) |
| 5 | `buildCI` seam ownership contradiction | Med | Design table; The seam (`buildCI` in front); Tests (vector-test CI note) |
| 6 | Step 2 must preserve persistent-secret clearing across commits 2–5 | Med | Build step 2 (verbatim interim `defer`s, incl. `= nil`) |
| 7 | `clear()`-contract / ISK-isolation tests must precede step 5 | Med | Build step 5 (test-first: tests red, then implement) |
| 8 | nil-randomness guard dropped from the sketch | Med | Target shape (`newInitiatorCore` / `newResponderCore`) |
| 9 | Evidence step omits dependency/capability disposition | Low | Evidence & release-readiness section |
| 10 | `responderCore` unused `adb` field | Nit | Target shape (`responderCore` struct — `adb` dropped) |

## Phase-2 re-run findings — disposition

| # | Finding | Sev | Addressed in this revision |
|---|---|---|---|
| F1 | Audit checklist contradicted `responderCore.isk` ownership | Med | Verification — audit checklist reworded, roles distinguished |
| F2 | Mandated partial-construction test had no reachable code path | Med | `clear()` contract clause relabelled forward-looking/vacuous; test dropped from step 5 + `TestClear` gate; Tests "Not tested" note |
| F3 | Step-2 white-box test migration spanned two commits but specified atomically | Med | Build step 2 — migration staged per role with the extraction commits |
| F4 | `TestSessionISKSurvivesCoreClear` did not map to the initiator core | Low | Tests — scoped explicitly as a responder test; initiator variant noted |
| F5 | Responder shell had no literal sketch | Low | Target shape — literal `respondWithRandom` / `Responder.Finish` sketch added |
| F6 | Step 2 left `finish`-method extraction implicit | Low | Build step 2 — extraction of constructor **and** `finish` made explicit |

## Verification

```sh
# Per-commit regression net — must stay green at every step
task check        # go test + -race; covers concurrent consume()
task quick

# Compile gate — step 2 must build (phase-2 items 3, 4)
go build ./...
go test ./... -run TestFinishCleanupDoesNotAliasReturnedSessions   # migrated to .core.*, staged per role
go test ./... -run TestInternalRandomHelpersDefaultNilRandomness   # nil-guard

# Phase-1 #4 precedence — must stay green through the refactor
go test ./... -run TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness

# New tests — must exist before/with step 5, not after
go test ./... -run TestSessionISKSurvivesCoreClear   # responder-scoped
go test ./... -run 'TestClear'        # nil-safe, idempotent, failure-path
go test ./... -run TestFinishZeroValueHardening      # zero-value guard, both roles

# Fuzz — smoke before; the post-refactor run is an interim gate ONLY, not
# evidence: the recorded bar (docs/fuzz-evidence.md) re-runs at the #33 refresh
FUZZTIME=30s PARALLEL=2 task fuzz
FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz

# Evidence disposition
git diff --exit-code -- go.mod go.sum    # expect clean: no dependency change

# Mandatory manual zeroization audit — api_test.go cannot prove this:
#  - neither core has a password, generator, k, or responder-scalar field
#  - initiatorCore has no isk field; responderCore.isk is the only persistent
#    core secret (responderCore.transcript is public wire data zeroed alongside
#    it as hygiene), both zeroed-then-nilled by responderCore.clear()
#  - initiatorCore.finish: defer clearBytes(isk) immediately after deriveISK
#  - startWithRandom / respondWithRandom retain defer clearBytes(nc.password)
#    as a backstop covering core-constructor error and panic paths
#  - confirmation-tag and identity comparisons remain hmac.Equal; no
#    bytes.Equal / reflect.DeepEqual on secret-derived values
#  - trace both clear() methods and every shell defer site
#  - residual (pre-existing, unchanged by this refactor): lvCat/prependLen
#    build intermediate heap copies that are not cleared (the password inside
#    calculateGenerator; K inside deriveISK), and hmac.New retains key pads
#    internally — record in docs/security-spec-audit.md as residual risk
```

## Evidence & release-readiness

This is a security-relevant refactor. Per the project's evidence-discipline rule, **all four** pinned evidence artifacts require a refresh against the post-refactor code, sequenced by issue #33 (exact-candidate evidence refresh):

- **Fuzz** (`docs/fuzz-evidence.md`) — the step-6 `FUZZTIME=8m` campaign is an interim gate only; the recorded evidence bar (long per-target campaigns across hosts) re-runs at the #33 refresh. Do not replace the pinned evidence with the interim run.
- **Security/spec audit** (`docs/security-spec-audit.md`) — refreshed in step 6 (commit SHA, command, duration, target count, residual risks — including the abandoned-state risk and the `lvCat`/`prependLen` intermediate-copy residual from the audit checklist).
- **Dependency review / SAST** (`docs/dependency-review.md`) — `go.mod`/`go.sum` stay byte-identical, but the recorded evidence policy requires repeating dependency review when security-relevant code changes, and the pinned `gosec` source scan is sensitive to exactly this kind of relocation. Re-run at the #33 refresh; the unchanged-module check is necessary, not sufficient.
- **Capslock** (`docs/capslock-report.md`) — no capability-surface change is expected; confirm with a re-run at the #33 refresh rather than asserting it.

The behavioral net proves behavior is preserved but **cannot** prove zeroization — the manual audit pass is mandatory.

**Sequencing** — per ADR-0001 (*Sequencing against release blockers*, decided 2026-06-10): implementation does not begin until the external review (#29–#31) and the independent cryptographic review (#32) conclude (or a later explicit maintainer decision accepts the review churn), and the revised ADR/plan text has passed its confirming `ras consider` round.

## Out of scope

- **Candidate B** (`singleUse`) composes cleanly — after this, `Initiator` is
  `{mu, used, core}`, and B would consolidate the `{mu, used}` remainder.
  Independent; do later if wanted.
- **Candidates C and D** untouched. (The lettered candidates come from the
  pre-ADR design exploration; only A — chosen — and B are described in this
  repo. C concerned the wire-framing seam: keeping framing in front of the
  CPace-core seam deliberately leaves that seam free to move on its own. D is
  not otherwise recorded; the letters are kept only to match the exploration's
  numbering.)
- **Abandoned-state secret retention** (phase-1 finding #5) — a single-use
  state dropped without `Finish` leaves its secret to GC. Pre-existing under the
  current `api.go`; neither introduced nor worsened here. Resolvable only if the
  public API freeze is reopened to add a `Close`. Documented as residual risk
  in ADR-0001 Consequences; record it in `docs/security-spec-audit.md`. Do
  **not** add `runtime.SetFinalizer` — finalizers are not a zeroization
  guarantee.
