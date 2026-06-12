---
status: accepted
date: 2026-06-10
review-runs:
  - 20260609T221534-bb273e71d960efe03b4b0fd0 # ras consider — accept with revisions; revisions applied via ras fix --decisions, ras verify clean
---

# Disposition of the `Suite` type before v1.0.0 freeze

## Status

**Accepted (2026-06-10).** This ADR captures a v1.0.0 freeze decision surfaced by external code review (item B1). Gated per the project's ADR policy: the `ras consider` run above returned accept-with-revisions; the revisions were applied via a maintainer-decided resolution pass (`ras fix --decisions`) and re-gated, with `ras verify` returning clean (unresolved: []). Evidence trail: PR #66 comments and DEV-JOURNAL cpace.S15. The decision is ratified; the implementation-verification gates below bind the implementing change before v1.0.0 is tagged. Status may later record `implemented at <sha>` once implementation lands.

## Context

`api.go:22-28` exports `Suite byte` together with the single constant `SuiteCPaceRistretto255SHA512 Suite = 0x01`. No exported function takes a `Suite` parameter, no exported method returns a `Suite`, and no `Config` field has type `Suite`. The constant is used internally only at `framing.go:7` as `wireSuite byte = byte(SuiteCPaceRistretto255SHA512)`, where the type is immediately discarded.

The shape was added in anticipation of a future multi-suite version of cpace. Today the package supports exactly one suite (CPACE-RISTR255-SHA512) and the threat model, docs, and tests all reflect that. The public API and profile policy are frozen for v1.0.0 unless this review reopens them, and any post-freeze removal or repurposing of this exported type is breaking.

The deferred problem: shipping `Suite` exported but unused commits the project to carry observable API surface. No zero-value reservation is needed now: `0x00` is unassigned and no exported API consumes `Suite`. If multi-suite support appears later, adding `Config.Suite Suite` with a documented zero value meaning "default = current suite" is non-breaking by Go convention, with the caveat that adding any field to exported `Config` can break external unkeyed composite literals because `Config` is an exported struct with all-exported fields; that usage is excluded from the Go 1 compatibility promise but is not compile-proof. A purely additive constructor such as `StartSuite` is the strictly compile-safe future surface. If multi-suite support never materializes, unexporting `Suite` later is breaking.

## Decision

Unexport `Suite` and `SuiteCPaceRistretto255SHA512` before v1.0.0. Delete the internal suite type entirely and declare `const currentSuite byte = 0x01`; an internal single-value defined type carries no information, and the type can be reintroduced in the same future commit that ever adds a second suite. The single use site at `framing.go:7` simplifies to `wireSuite byte = currentSuite`.

When (if) a future version of cpace supports multiple suites, the project re-introduces an exported `Suite` enum at that time with a deliberately-chosen multi-suite negotiation surface. `Config.Suite Suite` with zero-value-as-default is non-breaking by Go convention subject to the unkeyed-`Config` literal caveat above; a purely additive constructor such as `StartSuite` is strictly compile-safe. Re-exporting a previously-unexported type is non-breaking; removing or repurposing a previously-exported type is breaking. Unexporting now avoids carrying dead exported surface through v1.x while preserving both future paths.

This ADR rejects the option of threading `Suite` through `Config`/`Start`/`Respond` *now* on the grounds that the multi-suite negotiation design (zero-value semantics, suite-mismatch error path, downgrade-attack story for outer negotiation) is not load-bearing for v1.0.0 and committing to it before there are multiple suites means committing without evidence.

## Acceptance criteria

Multi-agent review concurrence on this ADR moves it proposed -> accepted (the decision is ratified at review time). The acceptance criteria below are implementation-verification gates: they bind the implementing change and must all be satisfied before v1.0.0 is tagged - not before this ADR is accepted.

- **API diff is exact for the ADR-0002 delta:** the isolated implementation delta from current `main` reports exactly two removals (`type Suite`, `const SuiteCPaceRistretto255SHA512`) and no other API change; the incompatible-only `v0.1.2` -> implementation view reports the same two removals. Because ADR-0003 landed first, the full `v0.1.2` -> implementation API diff also reports the compatible additions `ErrPeerShareEncoding` and `ErrPeerShareIdentity`; those additions are out of scope for this ADR-0002 gate. No exported identifier whose name contains `Suite` and no exported identifier of the now-internal suite type remains in `api.go`, `doc.go`, or any other file under the repo root.
- **`framing.go` still constructs the wire byte** from the internal constant; the wire format is unchanged (still `0xc1 || 0x01 || role || ...`). `TestWireFormatPrefixByte` is extended, or a sibling assertion is added, to pin `wireSuite == 0x01` as a literal, mirroring the existing `0xc1` format-byte pin, so a constant-value typo during the rename cannot pass the suite while breaking wire interop with released v0.1.x peers. This test pins currently-shipped behavior and may land immediately, independent of this ADR's acceptance.
- **`docs/spec-matrix.md`, `docs/security-assessment.md`, and `README.md`** are verified to contain no references to the exported `Suite` or `SuiteCPaceRistretto255SHA512` identifiers (none exist today; update only if found).
- **`CHANGELOG.md` Unreleased** records the unexport as breaking relative to v0.1.2 and flags it as pre-v1 cleanup, naming the removed symbols (`Suite`, `SuiteCPaceRistretto255SHA512`) and the migration path: drop the references and treat the package as single-suite with opaque framing.

## Considered options

- **A — Unexport now (recommended).** Removes the API surface tax, avoids carrying dead exported surface through v1.x, and preserves both future paths (re-export later or stay single-suite). Cost: ~5 LoC, one CHANGELOG entry.

- **B — Thread `Suite` through `Config` now.** Adds `Config.Suite Suite` defaulting to `SuiteCPaceRistretto255SHA512` (requires either remapping the constant to `0x00` so zero-value works, or treating `0x00` as "current" and leaving `0x01` as the explicit current-suite identifier). Locks in multi-suite negotiation semantics — zero-value default, suite-mismatch error code, outer-negotiation downgrade story — without a second suite to validate the design against.

- **C — Status quo.** Leaves `Suite` exported and unused for v1.0.0. Future multi-suite work can still add `Config.Suite` with zero-value-as-default (non-breaking by Go convention, subject to the unkeyed-`Config` literal caveat) or a strictly compile-safe additive `StartSuite` surface, but if multi-suite support never materializes the project carries dead exported surface for all of v1.x and any later unexport is breaking.

## Consequences

- **Option A (recommended):**
  - Breaking relative to v0.1.x for any consumer that references `cpace.Suite` or `cpace.SuiteCPaceRistretto255SHA512` (both exported in v0.1.0 through v0.1.2); permitted under pre-v1 Go module and semver conventions since no compatibility promise exists before v1.0.0.
  - No protocol-visible security change - the type is inert (consumed by no exported API, only the wire-constant derivation), the package remains single-suite with no in-package negotiation or downgrade path, and outer-negotiation downgrade protection remains the caller's responsibility (see integration guidance).
  - Re-exporting the type later costs only a doc note and a `CHANGELOG` entry.
  - `framing.go` and tests already do not depend on the exported shape.

- **Option B:**
  - Locks `Config.Suite` semantics for v1.0.0 with no second suite to validate against.
  - Forces a decision on zero-value defaulting *now* (remap `SuiteCPaceRistretto255SHA512` to `0x00`, or accept that zero is invalid and require callers to set it explicitly).
  - Any constant remap (for example, to `0x00`) requires first decoupling `wireSuite` from the enum value, because the wire suite byte `0x01` is frozen and shipped in v0.1.x releases; a remap without decoupling silently changes every message's second byte and breaks wire interop with all released peers.
  - Adds a documented suite-mismatch error path that has no testable wire-format outcome until a second suite exists.

- **Option C:**
  - Carries dead exported surface for all of v1.x plus a permanent doc footnote ("`Suite` is reserved for future use").
  - If multi-suite support never materializes, a later unexport is breaking after the type has propagated through a stable v1.x API.

## Implementation outline (Option A)

1. Delete exported `type Suite byte` and `SuiteCPaceRistretto255SHA512`; declare `const currentSuite byte = 0x01` in `api.go`, and either remove the obsolete exported-name doc comment or rewrite it to the final internal name (for example, `// currentSuite is the only suite implemented by v1 of this package.`).
2. Update `framing.go:7` to `wireSuite byte = currentSuite`; no `byte()` conversion is needed because `currentSuite` is already a byte constant.
3. Remove the public-doc reference in `doc.go` if any (none currently).
4. Verify `docs/spec-matrix.md`, `docs/security-assessment.md`, and `README.md` contain no references to the exported `Suite` or `SuiteCPaceRistretto255SHA512` identifiers (none exist today; update only if found).
5. Add a `CHANGELOG.md` Unreleased entry under "Pre-v1 API cleanup".
6. Extend `TestWireFormatPrefixByte` (or add a sibling assertion) to assert the literal `wireSuite == 0x01`, enforcing criterion 2 and pinning currently-shipped behavior; this may land immediately, independent of this ADR's acceptance.
7. Run `go test ./...`, `go test -race ./...`, `task check`. No protocol-visible or wire-format change is expected.
