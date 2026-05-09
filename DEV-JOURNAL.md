# Development Journal

**Append-only. New entries go at the END of this file.**

## cpace.S1 - 2026-05-05 20:34 EDT

**Main:** `f5ae7e2`
**Board:** CPace hardening workstreams merged; review triage documentation pending.
**Planner:** Josh

Recorded the post-review landing state for the CPace draft-21 hardening pass.
PRs #2 through #8 merged the documentation clarifications, draft-vector
coverage, fuzz expansion, security tooling, responder public-share
prevalidation, and best-effort memory hygiene workstreams.

This session adds the previously local interview-results triage document to the
repository and updates the changelog so the merged hardening work and triage
artifact are visible from the release notes.

---

## cpace.S2 - 2026-05-06 02:42 EDT

**Main:** `3f38b43`
**Board:** Policy/API decisions closed; release-readiness tracking begins.
**Planner:** Josh

Recorded the close of the CPace policy/API decision phase after PRs #13 through
#17 landed. Those PRs removed public randomness injection, required explicit
empty-session-ID compatibility opt-in, added session lifecycle and peer metadata
accessors, tightened framing/profile caps while keeping draft-compatible
confirmation tags, and documented the draft-21 scalar-sampling policy.

This session shifts the active project plan to release readiness. The next work
items are dependency review refresh, long fuzzing evidence, security/spec
documentation audit, and external review handoff before any production-readiness
claim.

---

## cpace.S3 - 2026-05-06 09:39 EDT

**Main:** `e1d0c6d`
**Board:** Dependency review and fuzz evidence for release readiness.
**Planner:** Josh

Combined the dependency-review refresh and long-fuzz evidence work into one
release-readiness branch. The advisory gosec lane flagged LEB128 integer
conversions, so this session records the parser cleanup and the clean rerun
alongside `govulncheck -test -show verbose ./...` evidence.

This session also starts recording fuzz evidence as a first-class artifact:
local smoke coverage plus long runs for all registered targets on ARM and Intel
hardware. The remaining release-readiness items are the security/spec audit and
external review handoff.

---

## cpace.S4 - 2026-05-06 12:20 EDT

**Main:** `4a8f629`
**Board:** Security/spec audit for release readiness.
**Planner:** Josh

Reviewed the security assessment and spec matrix against the merged
release-readiness implementation, the draft-21 source, and the recorded
dependency/fuzz evidence. No security/spec drift was found.

This session records the audit as `docs/security-spec-audit.md`, removes the
security/spec audit from the remaining release-readiness gaps, and leaves
external review handoff plus independent cryptographic review as the remaining
production-readiness blockers.

---

## cpace.S5 - 2026-05-06 14:49 EDT

**Main:** `5c97ae6`
**Board:** v0.1.0 draft-snapshot release preparation.
**Planner:** Josh

Prepared the first `v0.x` release snapshot after the project-owned
release-readiness gaps were reduced to integration guidance and release
hygiene. This session adds `docs/integration-guidance.md`, applies the
BSD-3-Clause license, keeps the release positioning explicit as unaudited and
not production-ready, and moves the changelog from `Unreleased` to `v0.1.0`.

The remaining production-readiness blockers are external review of
package-owned framing/CI/profile choices and independent cryptographic review.

---

## cpace.S6 - 2026-05-06 18:31 EDT

**Main:** `74b82cb`
**Board:** `v0.1.1` CI hardening snapshot shipped.
**Planner:** Josh

Closed the first CI-hardening pass after the draft-snapshot release. PRs #22
through #27 added the public hosted CI posture around required checks, CodeQL,
OpenSSF Scorecard, Staticcheck Advisory, Actionlint, cross-platform smoke,
scheduled vulnerability scanning, scheduled gosec, scheduled fuzz regression,
and release validation.

This session also published the signed annotated `v0.1.1` prerelease tag as a
CI and security-process hardening snapshot. The tag remained explicitly scoped
to release hygiene and evidence, not a production-readiness claim.

---

## cpace.S7 - 2026-05-07 02:27 EDT

**Main:** `39ccb58`
**Board:** External-review handoff, governance, and required gates prepared.
**Planner:** Josh

Moved from CI hardening into reviewer-readiness and project governance. PR #28
added the external-review handoff, and the follow-up public-hygiene and
release-planning work cleaned up reviewer-facing docs, release checklist
language, public contact handling, and OpenSSF badge posture.

PRs #37 through #39 added DCO, coordinated vulnerability disclosure, and
branch-protection-ready Dependency Gate and SAST Gate workflows. At this point
the project had the public process scaffolding for external review, but the
remaining release blockers stayed unchanged: independent cryptographic review,
external review of package-owned choices, and exact-candidate evidence refresh
after review-driven changes.

---

## cpace.S8 - 2026-05-07 22:25 EDT

**Main:** `955855b`
**Board:** Review-readiness tooling and evidence merged.
**Planner:** Josh

Closed the review-readiness tooling batch in PR #40. The project gained
allocation-reporting benchmarks for the full round trip, individual protocol
phases, exporters, and parser/message paths; additional godoc examples for
export, transcript IDs, close behavior, and confirmation-failure handling; a
Capslock capability-analysis report; and an `ossfuzz/` staging directory for
all fourteen native Go fuzz targets.

This session kept the public API and runtime package implementation unchanged.
The change was deliberately evidence-oriented: improve what reviewers can run,
inspect, and cite without resetting the package behavior under review.

---

## cpace.S9 - 2026-05-08 03:40 EDT

**Main:** `737bc56`
**Board:** Fuzz evidence refreshed; OSS-Fuzz upstream review is open.
**Planner:** Josh

PR #41 refreshed the fuzz-evidence packet after PR #40 merged and after the
OSS-Fuzz submission was opened upstream as `google/oss-fuzz#15480`. The
candidate evidence now records the merged PR #40 code commit, preserves the
older paired ARM/Intel runs as historical evidence, and states the residual
single-architecture risk plainly instead of relabeling older runs as current.

The OSS-Fuzz handoff is now staged locally and open upstream with CLA, header,
and helper-build checks green. Today also started paired one-hour
maintainer-machine fuzz campaigns on `m4mini.local` and `iMacPro.local` against
`737bc56`, and corrected the README badge from the numeric OSPS Baseline
endpoint to the OpenSSF Best Practices `passing` endpoint.

---

## cpace.S10 - 2026-05-08 11:55 EDT

**Main:** `737bc56`
**Board:** Go 1.26.3 toolchain security release impact assessed.
**Planner:** Josh

Go 1.26.3 was released on 2026-05-07 with security fixes in the Go command,
the pack tool, several standard-library packages, and bug fixes including
`crypto/fips140`. We treated this as a release-evidence trigger because CPace
uses Go crypto internals that transitively include `crypto/fips140`, and the
current dependency, fuzz, and Capslock evidence was recorded under Go 1.26.2.

The code impact check found no source change required. Under Go 1.26.3,
`go list -deps ./...` did not show use of the web/template/mail packages named
in the release note, and `task check` passed, including tests, race tests,
`go vet`, Staticcheck, ast-grep, and `govulncheck -test ./...`.

The earlier one-hour fuzz attempts from today were discarded as Go 1.26.2 or
potentially mixed-toolchain evidence. Both maintainer machines now report Go
1.26.3, and clean paired one-hour campaigns were restarted against `737bc56`.
The follow-up plan is to refresh fuzz evidence, dependency/gosec evidence, and
Capslock under Go 1.26.3 before treating the evidence packet as current again.

---

## cpace.S11 - 2026-05-08 12:37 EDT

**Main:** `737bc56`
**Board:** Go 1.26.3 evidence refresh completed for current `main`.
**Planner:** Josh

Completed the Go 1.26.3 evidence refresh after the paired maintainer-machine
fuzz campaigns finished cleanly on `m4mini.local` and `iMacPro.local`. Both
hosts ran all 14 registered fuzz targets for `FUZZTIME=1h` with `PARALLEL=2`
and recorded `RC=0`.

The dependency review, pinned gosec command, Capslock report, and security/spec
self-audit were refreshed from a clean detached worktree at `737bc56` under Go
1.26.3. No source-code changes were needed. `go fix` modernization remains
tracked separately in issue #42 so it does not blur this evidence-only refresh.
The PR follow-up committed raw transcripts and SHA-256 digests under
`docs/evidence/go1263-20260508/` and documented the calibrated artifact policy
for release candidates, toolchain-security refreshes, and lighter review
updates.

---

## cpace.S12 - 2026-05-08 14:24 EDT

**Main:** `fa70f28`
**Board:** Go 1.26 `go fix` modernization evaluated.
**Planner:** Josh

Started the post-evidence `go fix` follow-up from clean `main` after PR #43
merged. The Go 1.26.3 `go fix` diff is mechanical: use the built-in `max` for
the generator-string zero-padding clamp, use integer `range` loops in scalar
sampling and LEB128 parsing, and modernize concurrent tests to
`sync.WaitGroup.Go`.

Because the diff touches `crypto.go` and `framing.go`, the prior Go 1.26.3
long-fuzz, dependency, Capslock, and security/spec evidence remains valuable
historical signal but should not be treated as exact-current release-candidate
evidence after this branch merges. Local validation for the modernization
branch passed `task check`, pinned `gosec@v2.26.1`, and a short all-target fuzz
smoke with `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=30s PARALLEL=2 task fuzz`.
After PR review, the touched concurrency tests also passed
`go test -race -count=200`.

---

## cpace.S13 - 2026-05-08 22:07 EDT

**Main:** `2e09774`
**Board:** v0.1.2 candidate evidence refreshed.
**Planner:** Josh

After PR #45 merged, treated merge commit `2e09774` as the v0.1.2 package-code
candidate and refreshed the evidence packet without additional package-code
changes. Local clean-worktree analysis under Go 1.26.3 passed `task check`,
`go mod verify`, verbose `govulncheck`, pinned `gosec@v2.26.1`, and Capslock
`v0.3.2`.

Paired maintainer-machine fuzz campaigns ran all 14 registered targets with
`FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=1h PARALLEL=2 task fuzz` on `m4mini.local`
and `iMacPro.local`. Both logs recorded `RC=0` and the full target PASS set.
Raw transcripts and SHA-256 digests are preserved under
`docs/evidence/v012-candidate-20260508/`.
