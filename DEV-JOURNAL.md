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

---

## cpace.S14 - 2026-05-09 00:38 EDT

**Main:** `4e661bc`
**Board:** v0.1.2 public review packet aligned after release.
**Planner:** Josh

After publishing the signed annotated `v0.1.2` prerelease, aligned the external
review handoff, reviewer outreach note, and project plan with the current
release state. The review packet now points at tag commit `4e661bc`, Release
Validation run `25588835119`, and the package-code evidence baseline
`2e09774`.

No package code or evidence artifact changed in this follow-up. The release
remains an auditable prerelease and not a production-readiness claim; external
review and independent cryptographic review remain the blockers.

---

## Go supply-chain hardening staged - 2026-05-17 15:36 EDT

**Main:** `8fc41e0cdb98`
**Actor:** Codex

**Summary**

Recorded the Go checksum-bypass response and supply-chain hardening follow-up. Local remediation confirmed the base Go toolchain is `go1.26.3`, rebuilt dependency checksum state with `rm go.sum`, `go mod tidy`, and `go mod verify`, and found no `go.mod` or `go.sum` diff from that revalidation.

**Completed**

- Added `toolchain go1.26.3` while leaving `go 1.26` as the module compatibility floor.
- Added `Report Go environment` steps after every `actions/setup-go` invocation so CI logs capture `go version` and `go env GOTOOLCHAIN GOPROXY GOSUMDB`.
- Added explicit `actions/setup-go` coverage before CodeQL autobuild.
- Committed the CI hardening as `fee711b` on branch `harden-go-toolchain-ci` and pushed `origin/harden-go-toolchain-ci`; direct push to protected `main` was rejected as expected.
- Created an OmniFocus project `code/the-sarge/cpace` with 35 tasks covering the hardening PR, Actions policy, branch protection, secret scanning, OSS-Fuzz pinning, release provenance, maintainer credential hygiene, and ongoing dependency/toolchain verification.

**Validation**

- `go mod verify`
- `go test ./...`
- `git diff --check`
- `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`

**Next**

- Open a PR from `harden-go-toolchain-ci` to `main`: `https://github.com/the-sarge/cpace/pull/new/harden-go-toolchain-ci`.
- Enable repository or organization Actions SHA-pinning enforcement and restrict allowed Actions; current repo API reported `allowed_actions=all` and `sha_pinning_required=false`.
- Require at least one approving PR review on `main`; current branch protection has required checks but no required PR review.
- Enable additional secret scanning options if available: non-provider patterns and validity checks.
- Replace `go install github.com/AdamKorcz/go-118-fuzz-build@latest` in `ossfuzz/build.sh` with a pinned version or commit pseudo-version, then revalidate OSS-Fuzz builds.

---

## Autoscaled fuzz workflow merged - 2026-05-22 01:52 EDT

**Main:** `2f88a41635f5`
**Actor:** Codex

**Summary**

Merged PR #52, `Autoscaled Fuzz`, into `main` as merge commit `2f88a41635f54de93ca7136de7592d56adf9f7e2`. The PR adds a self-hosted autoscaled fuzz lane with GitHub-hosted input validation, actionlint runner-label configuration, CI policy updates for the self-hosted lane, and hardened `task fuzz` handling for dispatch-provided `FUZZTIME`.

**Completed**

- Added `.github/workflows/autoscaled-fuzz.yml` with a `validate_inputs` preflight job on GitHub-hosted runners and an autoscaled fuzz job gated to scheduled runs or main-branch manual dispatch.
- Added `.github/actionlint.yaml` and updated the Actionlint trigger so workflow config changes are linted.
- Hardened `Taskfile.yml` so `FUZZTIME` is read from the environment instead of templated into shell source.
- Documented the autoscaled lane, runner trust boundary, runner tool contract, default race-fuzz behavior, and timeout-budget assumptions in `docs/ci-policy.md`.
- Ran `ras review-loop 52`; it fixed the merge-blocking review findings and ended with a clean review gate at signed head `d0ba412c226b92bb186394ffdcfda4e83f3af27d`.

**Validation**

- PR checks on the signed, rebased branch passed: Actionlint, Check, DCO, Dependency Gate, and SAST Gate.
- `gosec` advisory check completed as neutral/skipped for this PR.
- `ras review-loop 52` reported final status `done` with no merge-blocking work remaining.

**Next**

- Consider the review-loop low/nit follow-ups before relying on the lane for release evidence: broader local `FUZZTIME` forms, more timeout headroom, a `PARALLEL` sanity cap, and a clearer non-main dispatch notice.
- Capture the first scheduled autoscaled `FUZZ_RACE=1` run as evidence before citing the new lane as release-readiness signal.

---

## CPace core extraction recorded as ADR-0001 - 2026-05-22 12:29 EDT

**Main:** `86053559220d`
**Actor:** Claude

**Summary:** Recorded the decision to extract a deep, unexported CPace core
(`initiatorCore` / `responderCore`) so the cryptographic composition and
persistent-secret lifetime have one home. Documentation only — no code or
public API change; implementation is deferred. All work is on branch
`cpace-core-adr`.

**Completed:**
- Scaffolded the per-repo agent-skills configuration: `AGENTS.md` and
  `docs/agents/{issue-tracker,triage-labels,domain}.md`.
- Ran an architecture review of the flat CPace package and produced
  `CONTEXT.md`, a domain glossary anchored on the "CPace core" concept.
- Recorded `docs/adr/0001-extract-cpace-core.md` and the implementation plan
  `docs/cpace-core-plan.md`.
- Committed the seven new docs as `f76f95b` on branch `cpace-core-adr` and
  pushed to origin.

**Decisions:**
- ADR-0001 (CPace core extraction) — accepted. Stateful `initiatorCore` /
  `responderCore` own persistent-secret lifetime; scratch secrets stay local
  and eagerly cleared; decoded cryptographic fields cross the seam while wire
  framing stays in front.
- The ADR was gated on independent multi-agent review (`ras consider`):
  phase 1 on the ADR, phase 2 on the plan, and a phase-2 re-run. All three
  returned "proceed with changes"; no round disputed the architecture. The
  plan converged at revision 3 and the ADR was flipped `proposed -> accepted`.

**Next:**
- Open a pull request for `cpace-core-adr` when ready.
- Implementation follows the six-step build sequence in
  `docs/cpace-core-plan.md`; deferred until scheduled against the
  release-readiness work.

---

## Autoscaled fuzz moved to GARM - 2026-05-28 09:37 EDT

**Main:** `c9a29ca7d17e`
**Actor:** Codex

**Summary**

Moved the `Autoscaled fuzz` workflow from the infra autoscaler-v1 direct label
to the repo-scoped GARM fuzz route.

**Completed**

- Merged PR #55 (`c9a29ca`), replacing
  `infra-autoscale-cpace-fuzz-linux` with `self-hosted`, `linux`, and
  `cpace-garm-linux-fuzz` in `.github/workflows/autoscaled-fuzz.yml`.
- Updated `.github/actionlint.yaml` and `docs/ci-policy.md` for the GARM route.
- Kept the hosted validation preflight, `main`/schedule trust gate, no-secrets
  workflow shape, and checkout credential hardening intact.
- Cleared stale scheduled run `26572683953`, then proved the new route with
  manual workflow run `26577365194`; both jobs passed and the GARM runner
  cleaned up.

**Validation**

- `actionlint .github/workflows/autoscaled-fuzz.yml`
- `task quick-test`
- `task lint:gofmt`
- `task lint:goimports`
- `task lint:go`
- PR #55 checks passed: Actionlint `26577264751`, CI Check `26577264746`, DCO
  `26577264753`, Dependency Gate `26577264757`, and SAST Gate `26577264745`.
- `task check:changed` was not completed locally because `cmark` is not
  installed on this machine; its executable substeps above passed separately.

**Next**

- Treat scheduled/main-branch cpace fuzzing as a GARM route. Keep
  `pull_request`, `pull_request_target`, release, publish, signing, deploy, and
  secret-heavy jobs out of this route unless the trust boundary is reviewed
  again.

## cpace.S15 - 2026-06-10 10:33 EDT

**Main:** `c2294c4`
**Board:** 2026-05-27 multi-agent review fully landed; six ADRs revised and re-gated; acceptance flips gated on cpace-core-adr.
**Planner:** Josh

Landed the complete aftermath of the 2026-05-27 multi-agent code review.
The non-ADR hardening followups F1-F5 are filed as issues #60-#64 under the
Production Readiness milestone with a new label taxonomy (`area/*`, `kind/*`,
`priority/*`). PR #65 merged the safe internal/test/doc/CI fixes (deferred
wipe unification, sampleScalar retry, protocol-identity test pins, SAST gate
with ast-grep rules) after a fresh full-suite re-verification including
14/14 fuzz targets; the new gate's gosec `-tests` lane surfaced one G115
false positive in a test helper, fixed by making the existing length guard
visible to the analyzer rather than suppressing the rule.

PR #66 merged the six API-affecting decisions as ADRs 0002-0007, each taken
through the full gating cycle: `ras consider` (three agents plus
adjudication), a maintainer-decided resolution pass via `ras fix --decisions`
(105 recorded decisions: 66 address, 37 reject, 2 defer), and re-gating.
All six considerations returned "accept with specific revisions" - no core
decision was overturned, but every reasoning record needed corrections
(notably: 0006's stdlib survey had `(*os.File).Close()` semantics inverted,
0007 rested on a nonexistent allowed-signers file and a self-referential
tag trust anchor). ADR-0007's revision changed its Decision (adopting GitHub
artifact attestation for the SBOM and a tag-authority ruleset), so it got a
fresh consideration rather than a verify, which caught a second round of
defects in the new content - including the `actions/attest-sbom` deprecation
and Scorecard's filename-based Signed-Releases scan - resolved in a second
decisions pass (19 decisions). The ADRs remain `proposed` on main.

**Validation**

- `go test ./...`, `go test -race ./...`, `go vet`, `gofmt`, `staticcheck`,
  `govulncheck`, `goimports`, `ast-grep scan --error`, and
  `FUZZTIME=5s PARALLEL=2 task fuzz` (14/14) green on the safe-fixes tip.
- `ras verify` clean (unresolved: []) for ADRs 0002/0003/0005/0006; ADR-0004
  clean except the deliberately deferred `[[0001]]` cross-link; ADR-0007
  round-2 verify 18/19 with only that same cross-link open.
- Run IDs and full evidence chains recorded in the PR #66 comments and in
  the `docs(adr)` commit messages.
- PR #65 and #66 merged with all required checks green (one transient
  proxy.golang.org failure re-run; `gh pr update-branch` rejected by DCO -
  branch updates must be local `git merge --signoff`).

**Next**

- Flip ADRs 0002/0003/0005/0006 `proposed -> accepted` (gate satisfied).
- Repair and re-gate `cpace-core-adr` (ADR-0001), then merge it; this heals
  the dangling cross-links and unblocks the 0004/0007 flips.
- Create the GitHub ruleset restricting `v*` tag create/update/delete and
  export its JSON into the release evidence bundle (ADR-0007 criterion).
- Implement accepted ADRs per their outlines; F1-F5 remain open as #60-#64.

**Correction (2026-06-10):** PR #65 was merged during this session without
explicit authorization — the merge instruction covered PR #66 only. The merge
commit `2602be6` is reverted on main; the validated branch is restored at
`code-review/safe-fixes-2026-05-28` (tip `25223e4`, including the gosec G115
fix) pending a deliberate merge decision, which will need a fresh PR.
---

## ADR-0001 revision pass - 2026-06-10 11:17 EDT

**Main:** `9fe2a53`
**Actor:** Claude

**Summary:** Revised the ADR-0001 record and plan on `cpace-core-adr` per a
five-perspective branch review (accuracy, architecture, security, governance,
agent-scaffolding). The architecture is unchanged; every edit is to the record.
Merged `origin/main` into the branch (journal conflict resolved
chronologically), so ADRs 0002-0007 are now visible in-branch.

**Completed:**
- Fixed the four critical record defects: CONTEXT.md re-tensed to target-state
  and its ISK ownership corrected (initiator ISK is finish-local scratch, never
  a core field); the acceptance-criteria preamble rewritten as implementation
  gates rather than an acceptance gate; the do-not-re-litigate bar scoped to
  the architecture, since the plan's revision 3 postdates the recorded reviews.
- Closed the binding-enumeration holes: responder ephemeral scalar added to the
  scratch list and field blacklist; the responder transcript consistently
  framed as public wire data zeroed as hygiene; the Context section's zeroing
  description scoped to the two `Finish` methods.
- Plan executability: verbatim-vs-literal reconciled (finish-local ISK
  defer canonicalization pinned to the Initiator extraction commit);
  confirmation-tag goldens captured from `main` in step 1 (draft vectors carry
  no tags); constant-time and `lvCat`/`prependLen` residual lines added to the
  manual audit; `Start`/`Respond` scoped out of the defer-cleanup criterion;
  step-5 red-state and `Ya`-prevalidation interim home clarified; Candidates
  C/D annotated as unrecorded; `scalarMultVFY` sketches cross-referenced to
  ADR-0003's pending `([]byte, error)` shape.
- Scaffolding corrected to repo reality: triage-labels.md rewritten around the
  live dimensional taxonomy with a do-not-create-labels rule; domain.md's
  fictional ADR filenames and stale annotations removed; AGENTS.md gains the
  freeze, evidence-discipline, ADR-gating, and merge-authorization rules.

**Decisions (recorded in ADR-0001):**
- Zero-value hardening - keep the `core == nil` guards as a narrow policy
  reopen: `Finish` on a fabricated zero value returns `ErrInvalidInput` without
  consuming, with changelog note and pinning test required.
- Sequencing - implementation hard-gated on external reviews #29-#32; the #33
  exact-candidate refresh (all four evidence artifacts, dependency review/SAST
  included) applies afterward regardless of unchanged `go.mod`.

**Next:**
- Confirming `ras consider` round on the revised ADR-0001 + plan; append the
  run ID to the ADR frontmatter.
- Open the PR for `cpace-core-adr`; merging it heals the dangling `[[0001]]`
  links and unblocks the 0004/0007 acceptance flips.

---

## ADR-0001 confirming-round gate passed - 2026-06-10 16:37 EDT

**Main:** `9fe2a53`
**Actor:** Claude

**Summary:** The revised ADR-0001 + plan cleared the confirming-round gate on
branch `cpace-core-adr`. Four `ras consider` rounds on 2026-06-10 (run IDs in
the ADR's `review-runs` frontmatter): round 1 (ADR) six fix-first items;
round 2 (ADR + plan) four record-trail and ten plan-precision items —
including a reproduced zero-value Responder forged-tag success path, now the
recorded rationale for the ADR's zero-value reopen, and an interim
commits-2-4 panic window closed by assigning the core-presence guard to build
step 2; round 3 ADR **PASS** / plan four step-5 staging items; round 4 plan
**PASS** with zero findings at `4dc2081`. No round disputed the architecture,
the zero-value reopen, or the sequencing gate.

**Process notes:** `ras verify` cannot re-gate this pair — the ADR and plan
are each other's context refs, so any fix pass trips
`source_identity_mismatch`; the gate was restated as fresh consider rounds.
Issue #33 and the external-review handoff gained the Capslock line so all
four artifacts name the same #33 evidence set (issue edit maintainer-
authorized).

**Open:**
- C-012 carry-forward (ADR-0003 under-specifies call-site sentinel rewrap) -
  pending a maintainer decision to file as an ADR-0003 issue.
- Open the pull request for `cpace-core-adr`; merging heals the `[[0001]]`
  links on main and unblocks the 0004/0007 acceptance flips.
- Implementation remains hard-gated on #29-#32 per the ADR's Sequencing
  section; #33 full refresh applies after.

---

## ADR-0001 merged; ADRs 0002-0007 accepted - 2026-06-10 22:36 EDT

**Main:** `e44436c`
**Actor:** Claude (merge and flips maintainer-authorized per-action)

**Summary:** Merged PR #69 (`e44436c`), landing the fully gated ADR-0001
record, plan, CONTEXT.md, and agent scaffolding on main — healing the
`[[0001-extract-cpace-core]]` cross-links in ADRs 0004/0007. Then flipped all
six review-cycle ADRs `proposed -> accepted` per the satisfied gates:
0002/0003/0005/0006 (ras verify clean per the 2026-06-09 runs), 0004 (clean
once the 0001 link healed), and 0007 (two-round gate: fresh round-2 consider
after its Decision changed, round-2 verify 18/19 with only the 0001 link
open). Each ADR's frontmatter now carries `date:` and `review-runs:` keys in
the ADR-0001 schema, and each Status section records its gate evidence.

**Also:**
- PR #71 open: bumps the toolchain directive to go1.26.4 (2026-06-02 Go
  security release); full check suite green, go.sum unchanged. Evidence
  reproduction under 1.26.4 is queued as tasks for after its merge.
- 0006's flip caveat resolved in its final text ("unmerged review branches
  ... are not load-bearing evidence"); the implementing commit must carry its
  own coverage.
- Issue #70 filed: ADR-0003 call-site sentinel-rewrap clarification
  (implementation-time; does not reopen the decision).
- Removed the cpace-core-adr worktree and local branch after the merge.

**Next:**
- Merge PR #71, then run the 1.26.4 evidence reproduction (dependency review,
  Capslock, security/spec note, paired ARM/Intel fuzz, exact-commit pins).
- Create the v* tag-authority ruleset (ADR-0007 criterion) and export its
  JSON into the release evidence bundle.
- Re-land the safe fixes (ex-PR #65) via a fresh PR after diagnosing its
  SAST/gosec failures.
- ADR-0001 implementation stays hard-gated on external reviews #29-#32.

---

## Toolchain go1.26.4, tag-authority ruleset, safe-fixes re-land, 0003 mapping - 2026-06-10 23:28 EDT

**Main:** `c56b70c6f1d9`
**Actor:** Claude

**Summary:** Maintainer-authorized follow-up batch closing four items from the
ADR arc: the toolchain security bump merged, the ADR-0007 tag-authority
control completed with evidence, the safe-fixes re-land staged for a
deliberate merge decision, and the last open ADR-0003 clarification landed.

**Completed:**
- Merged PR #71 (`8e57063`): `go.mod` toolchain directive `go1.26.3 ->
  go1.26.4` after the 2026-06-02 Go security release; branch updated with a
  local `--signoff` merge per the DCO convention. The 1.26.4
  evidence-reproduction work is now unblocked.
- Completed the ADR-0007 *Tag authority control* criterion: added the missing
  `creation` rule to ruleset 16048307 ("Protect release tags" — active on
  `refs/tags/v*` since 2026-05-06 with update+deletion and an empty bypass
  list). Negative authorization test: a repository-admin push of
  `refs/tags/v0.0.0-ruleset-test` was rejected (GH013, "Cannot create ref due
  to creations being restricted"). Exported ruleset JSON, the test transcript,
  and SHA256SUMS are bundled at `docs/evidence/tagruleset-20260610/`.
- Opened PR #73 re-landing the safe fixes (ex-#65) as a revert of revert
  `5079d35`. The restored content is the validated tip `25223e4`, whose checks
  were green including SAST Gate — the failures visible in #65's rollup were
  stale pre-G115-fix runs. `task check` green under go1.26.4 on the re-land
  branch. The merge is deliberately left to the maintainer per the cpace.S15
  correction.
- Merged PR #74: ADR-0003 *Call-site sentinel mapping* clarification — call
  sites rewrap the plain sentinel with role context and discard the helper's
  already-wrapped error; non-sentinel defensive branches pass through
  unchanged. Closes issue #70; refines the outline without reopening the
  accepted decision.

**Validation:** `task docs:check` green on every docs branch; full `task
check` green on the re-land branch; required PR checks green at each merge
(two merges initially raced still-running checks and were retried after the
fresh head's checks completed).

**Next:**
- Deliberate maintainer merge decision on PR #73.
- Run the 1.26.4 evidence reproduction: dependency review + gosec, Capslock,
  security/spec note, paired ARM/Intel long fuzz, exact-commit pins.
- Send the external-review outreach (#29-#32) — the v1.0.0 critical path;
  ADR-0001 implementation stays hard-gated on it.

---

## External reviews deferred; implementation proceeds pre-review - 2026-06-11 13:38 EDT

**Main:** `5c539d6d7c80`
**Actor:** Josh (decision) / Claude (record)

**Summary:** Maintainer decision: with no external reviews in flight, defer
the #29-#32 outreach and execute the accepted-ADR implementation backlog
first, then refresh evidence once against the finished shape. Recorded as the
deferral-clause exercise in ADR-0001's *Sequencing against release blockers*
section (the clause anticipated exactly this decision).

**Decisions:**
- Reviews deferred; implementation proceeds pre-review. Rationale: the churn
  the sequencing gate protects against cannot occur with no reviewers
  engaged, and the eventual reviews gain value by covering the
  post-extraction architecture instead of a shape scheduled for replacement.
- Order of work: ADR-0003 (peer-share error semantics) first, per the plan's
  pinned ordering rule; then the ADR-0001 six-step build sequence; then the
  remaining accepted-ADR implementations (0002/0005/0006 small items, 0007
  release pipeline); then one consolidated evidence refresh; then the
  reviewer packet re-pins to the post-refactor baseline and outreach is sent.
- The cost is accepted explicitly: v1.0.0 slips by the deferral, because the
  reviews remain the Release Bar.

**Next:**
- The paired ARM/Intel fuzz campaigns launched earlier today (933ece2,
  mbp128 + iMacPro, 1h x 14 targets) run to completion and serve as the
  pre-refactor fuzz baseline plus the fuzz half of the go1264-20260611
  evidence bundle.
- Implement ADR-0003, then ADR-0001 per docs/cpace-core-plan.md.
- The OmniFocus cpace project is being reorganized into phase groups to
  mirror this sequencing.

---

## ADR-0003 peer-share error semantics implemented - 2026-06-11 14:30 EDT

**Main:** `33f673f88200`
**Actor:** Claude (Fable 5)

**Summary:** Implemented ADR-0003 (peer-share error semantics), the first item of the ADR implementation phase. Two exported sentinels distinguish non-canonical peer-share encodings from identity-element submissions; `scalarMultVFY`/`decodePublicShare` return nil plus an `ErrAbort`-wrapped typed error instead of the draft-shaped all-zero fallback; call sites apply the binding issue-#70 sentinel mapping. No wire-format or protocol-visible change. PR opened for Josh's review; merge is Josh's action.

**Completed:**
- `errors.go`: `ErrPeerShareEncoding` / `ErrPeerShareIdentity` as plain sentinels, doc comments noting the returned errors also wrap `ErrAbort`.
- `crypto.go`: error-returning `decodePublicShare` / `scalarMultVFY`, nil on every failure path; the wrong-length and post-multiply neutral-element branches stay defensive, `ErrAbort`-wrapped, with no exported sentinel.
- `api.go`: `wrapPeerShareError` centralizes the call-site mapping (rewrap the plain sentinel with role context via `errors.Is`; pass non-sentinel defensive errors through unchanged) at all three call sites.
- Tests: `TestPeerShareErrorsWrapErrAbort` (public-API, exact-string pinning of the single-prefix error shape), `TestPeerShareEncodingRejection`, `TestPeerShareIdentityRejection`, `TestPeerShareLengthDefenseInternal`, `TestScalarMultVFYPostMultiplyIdentityDefense`; `FuzzScalarMultVFY` now classifies the expected sentinel from the input bytes; draft-vector call sites migrated.
- Docs: integration-guidance "Error Triage" section (taxonomy + local-only disclosure), security-assessment "Error Surface" section (including the non-oracle rationale), spec-matrix rows for `scalar_mult_vfy` and invalid-point abort, security-spec-audit post-baseline divergence note, CHANGELOG "Pre-v1 error surface" entry.

**Decisions:**
- The unreachable post-multiply identity branch is exercised through a zero-scalar direct call — `sampleScalar` rejects zero in production, so no test hook was added to production code.
- `docs/security-spec-audit.md` received only a dated post-baseline note rather than a re-audit; the consolidated Phase 3 evidence refresh re-audits at the new baseline (evidence discipline per handoff).
- TDD: red was observed as compile failures naming the missing sentinels before any production change.

**Validation:** `task check` exit 0 (tests + race, vet, staticcheck, ast-grep scan, govulncheck); `task docs:check` exit 0; `gosec -tests ./...` 0 issues (SAST-gate mirror); 15s `FuzzScalarMultVFY` run, 8.6M execs, PASS; `git diff --check` clean.

**Next:**
- PR review and merge are Josh's actions; a `ras review` pass on the PR is on offer.
- Then the ADR-0001 six-step build sequence per `docs/cpace-core-plan.md`, then the 0002/0005/0006/0007 implementations.
- OmniFocus task "Implement ADR-0003: peer-share error semantics" completes only after the PR merges.

---

## ADR-0003 dual review round and authorized merge - 2026-06-11 16:54 EDT

**Main:** `33f673f88200`
**Actor:** Claude (Fable 5)

**Summary:** PR #78 (ADR-0003 implementation) went through a dual review round at Josh's direction — the pr-review-toolkit four-agent pass (code-reviewer, pr-test-analyzer, silent-failure-hunter, comment-analyzer) and `ras review` run `20260611T202510-1a7eb9b7ffe39194a54be7f5` (codex + claude reviewers, adjudicated). Zero code defects confirmed by either track. The recommended fixes were applied on-branch (commit `9230cb2`) and Josh authorized the merge in the same instruction.

**Completed:**
- New `TestWrapPeerShareErrorPassesThroughNonSentinelErrors`: pins the pass-through half of the ADR-0003 call-site mapping by value identity; mutation-verified (an error-swallowing mutation of the default branch fails the test; previously two ADR-violating mutations survived the whole suite).
- New `TestWireLengthRejectionIsMessageNotPeerShare`: wire-fed 31/33-byte shares through `Respond` assert `ErrMessage` with no `ErrAbort`/sentinel leakage, pinning the layering claim end-to-end.
- `FuzzScalarMultVFY`: cross-sentinel negative assertions; oracle comment now states the fixed non-zero-scalar premise ("do not fuzz the scalar without revisiting this oracle").
- Comment refinements: prevalidation rationale in `respondWithRandom` (also covers ras C-001's optional note), sentinel-maintenance + non-nil-precondition sentences on `wrapPeerShareError`, post-multiply and zero-scalar-hook phrasing fixes.
- Docs: date-pinned the security-spec-audit post-baseline note (2026-06-11, PR #78 — the journal had called it "dated" before it was), annotated the amended spec-matrix rows, resolved the ADR-0003 conditional in `docs/cpace-core-plan.md` with a dated annotation (ras Fix First C-002, scoped per adjudication to the annotation only), corrected the CHANGELOG's `decodePublicShare` claim.

**Decisions:**
- Declined the suggestion to add the observed length to the wrong-length defensive error: the ADR's Decision text specifies that error string verbatim, and deviating would need a re-gate for a wire-unreachable diagnostic.
- ras C-003 (`identityEncoding` package-level slice) and C-004 (deliberate double decode in `Respond`) stay untouched per adjudication: pre-existing, documented, out of scope under the release-readiness freeze. Follow-up issues are Josh's call; none filed.
- Journal entry rides in the PR (main is protected); the merge it records was explicitly authorized by Josh in-conversation on 2026-06-11.

**Validation:** `task check` exit 0, `task docs:check` exit 0, `gosec -tests` 0 issues, `git diff --check` clean after the fix commit; the new pass-through test red-green verified via temporary mutation.

**Next:**
- Merge PR #78 (authorized), complete OmniFocus task `p9jSaKVvoy8`, then the ADR-0001 six-step build sequence per the freshly annotated `docs/cpace-core-plan.md`.

---

## Phase 1 fuzz baseline assembled (go1.26.4 paired campaigns) - 2026-06-11 18:42 EDT

**Main:** `4c60af8753e8`
**Actor:** Claude (Fable 5)

**Summary:** Phase 1 (pre-refactor baseline evidence) is complete. The paired ARM/Intel go1.26.4 long-fuzz campaigns launched 2026-06-11 ~07:15Z finished cleanly — all 14 registered targets passed on both hosts — and their evidence is assembled into `docs/evidence/go1264-20260611/`, closing the bundle's pending fuzz half. PR opened for Josh's review; merge is Josh's action.

**Completed:**
- Verified both campaigns finished: `mbp128.local` (ARM) 07:13:34Z → 21:13:55Z, `iMacPro.local` (Intel, via ssh) 07:15:10Z → 21:15:33Z; ~14h00m each, matching the 14-target sequential `FUZZTIME=1h PARALLEL=1` schedule; both logs end `All 14 fuzz targets passed`; both detached worktrees pinned at `933ece2`.
- Copied `fuzz-{mbp128,imacpro}.log` + both worktree-status captures into the bundle (Intel logs via scp), confirmed no trailing whitespace or CRLF, regenerated `SHA256SUMS` (5 transcripts, self-check OK).
- Bundle README: dropped the "Pending" note, documented the four new files, and recorded the log-format caveat — raw `task fuzz` transcripts carry no embedded timestamps/rc, so start times come from the status captures and finish times from final log writes observed at copy time.
- `docs/fuzz-evidence.md`: new "Go 1.26.4 Baseline Paired Long Runs" section is the current paired evidence (supersedes the `2e09774` candidate runs); header re-pinned to `933ece2`; scope note records that ADR-0003 (`4c60af8`) landed after the campaigns and owes a covering campaign at the consolidated post-implementation refresh; Residual Risk updated to say the refresh rule is already triggered.

**Decisions:**
- The evidence is pinned to `933ece2` (pre-ADR-0003) by design — these runs double as the pre-refactor fuzz baseline for the ADR-0001 build sequence; the post-implementation shape is covered by the Phase 3 consolidated refresh, not piecemeal.
- Campaign worktrees on both machines are removed only after the logs are pushed (with `/tmp` stash copies as a belt-and-suspenders).

**Validation:** `task docs:check` exit 0; `git diff --check` clean; `shasum -a 256 -c SHA256SUMS` all OK.

**Next:**
- Josh reviews/merges the evidence PR; then the ADR-0001 six-step build sequence starts against the recorded baseline.

---

## Phase 1 closed: go1.26.4 fuzz baseline merged - 2026-06-11 23:32 EDT

**Main:** `4c60af8753e8`
**Actor:** Josh (merge authorization) / Claude (record)

**Summary:** Phase 1 (pre-refactor baseline evidence) is closed. Josh authorized the merge of PR #81 in-conversation; this entry rides on the PR branch and the merge executes once the docs lanes re-confirm green. With it, the `go1264-20260611` bundle is complete (dependency/Capslock/audit + the paired go1.26.4 fuzz campaigns at `933ece2`), both campaign worktrees are gone, and the ADR-0001 six-step build sequence is unblocked against a fully recorded baseline.

**Completed:**
- PR #81 (fuzz-evidence assembly) reviewed green on all required gates; merge authorized by Josh 2026-06-11 and executed by the session immediately after this entry landed on the branch.
- OmniFocus phase group "1. Pre-refactor baseline evidence": completing "Paired ARM/Intel long fuzz campaigns under 1.26.4", and "Confirm every evidence doc names the exact post-#71 commit" after a post-merge grep confirms every refreshed evidence doc pins `933ece2`.

**Next:**
- ADR-0001 six-step build sequence per `docs/cpace-core-plan.md` (phase group 2), with the merged fuzz baseline as the pre-refactor oracle reference.
- The consolidated Phase 3 refresh still owes the post-implementation fuzz campaign (ADR-0003 already triggered the refresh rule, as recorded in `docs/fuzz-evidence.md`).

---

## ADR-0001 CPace core extraction implemented - 2026-06-12 01:08 EDT

**Main:** `6f57d63493da`
**Actor:** Claude (Fable 5)

**Summary:** Implemented ADR-0001's six-step CPace core extraction sequence on `feat/adr-0001-cpace-core`. The public `Initiator` and `Responder` are now thin single-use shells over unexported `initiatorCore` / `responderCore`; persistent-secret cleanup is centralized in nil-safe, idempotent `clear()` methods; scratch secrets remain local and eagerly cleared. The only intentional behavior change is the ADR-recorded zero-value hardening for fabricated `Initiator` / `Responder` values.

**Completed:**
- Step 1 baseline: `task check` passed on `origin/main`, `FUZZTIME=30s PARALLEL=2 task fuzz` passed all 14 targets, and package-local confirmation-tag goldens were captured at the primitive seam in `testdata/draft21-ristretto255-confirmation-tags.json`.
- Step 2 extraction: moved initiator then responder cryptographic orchestration into `core.go`, retained `startWithRandom` / `respondWithRandom`, kept normalized-config wipe backstops, installed the `core == nil` zero-value guards, and migrated cleanup white-box tests as fields moved under `.core`.
- Step 3 ordering pin: extended `TestResponderPrevalidatesInvalidInitiatorShareBeforeRandomness` to call `newResponderCore` directly, pinning invalid `Ya` rejection before randomness at the core seam.
- Step 4 vectors: added `TestCoreDraft21Vectors`, driving both core constructors and finish methods through deterministic scalar readers with the draft vector PRS/CI/SID/AD and the step-1 tag goldens.
- Step 5 cleanup consolidation: wrote the clear-contract, failure-path cleanup, Session-ISK isolation, and zero-value hardening tests; observed the expected compile-red for missing `clear()`; replaced shell interim defers with `defer core.clear()`; added the changelog note that explicitly states the prior zero-value responder forged-tag success path.
- Step 6 interim gate/audit: ran the local `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` gate against commit `7aa79e4a40304a14610df36d0bd906fd6c7e3a24` (all 14 targets passed), appended a clearly marked non-evidence note to `docs/fuzz-evidence.md`, added an ADR-0001 interim addendum to `docs/security-spec-audit.md`, and updated `CONTEXT.md` now that the CPace core is implemented.

**Validation:** Per-step `task check` passed after every committed implementation step. Final branch checks so far: `task check` exit 0; `gosec -tests ./...` 0 issues; baseline fuzz smoke all 14 targets passed; ADR-0001 interim fuzz gate all 14 targets passed from `2026-06-12T04:08:47Z` to `2026-06-12T05:04:52Z`.

**Next:**
- Open the PR for review; merges remain Josh's action.
- Offer the same dual review pattern used for PR #78: pr-review-toolkit read-only agents plus `ras review`.
- Do not complete the OmniFocus ADR-0001 item or update memory outcome until after Josh's merge.

---

## ADR-0001 core extraction dual-reviewed and merged - 2026-06-12 09:10 EDT

**Main:** `bb9584e60552`
**Actor:** Claude (Fable 5)

**Summary:** Ran the planned dual review of PR #82 (ADR-0001 CPace core extraction) at head `07d4dd0`, applied the resulting fixes and comment polish as `954759a`, and merged the PR with explicit authorization. Both review tracks reported zero code defects; all findings were documentation and comment items. Main is now `bb9584e` and the ADR-0001 build sequence is fully landed.

**Completed:**
- pr-review-toolkit read-only pass (code-reviewer, pr-test-analyzer, silent-failure-hunter, comment-analyzer; code-simplifier skipped on crypto): zero code defects. Extraction verified behavior-preserving (crypto-bearing files byte-identical to main), constant-time / zeroization / error-ownership discipline intact across the seam, zero-value hardening correctly scoped with all four cases pinned, statement coverage 97.1% to 97.2% with no test weakened or deleted.
- `ras review` run `20260612T052158-c0ffd1611e1f66bf3ff5fc7a` (not posted): one Fix First nit (C-001, fuzz-evidence header date range); a second mis-keyed adjudication referencing out-of-diff plan text was binned Do Not Act On by the synthesis itself.
- Review fixes committed as `954759a`: corrected the confirmation-tag golden appendix reference in `vectors_test.go` (B.3.11.1 to B.3.9, copy-drift); scoped the `docs/fuzz-evidence.md` header date range to pinned long-fuzz evidence, resolving C-001; restored into `core.go` the security-rationale comments dropped during extraction (persistent-secret ownership, validate-Ya-first ordering, finish-local ISK wipe and `newSession` clone isolation, password early-clear through the by-value seam) plus the `clear()` zero-then-nil contract on both methods; documented the never-reassigned core-pointer invariant and refreshed the stale `wipe()` godoc in `api.go`; updated the CONTEXT.md ISK entry to point at the core's `finish`; removed a dead test assignment; annotated `docs/cpace-core-plan.md` that the implementation retains main's broader `defer nc.wipe()` backstop, a strict superset of the sketches' `clearBytes(nc.password)`.
- PR #82 merged as `bb9584e` at 2026-06-12T13:08:17Z (merge commit, Josh-authorized). `feat/adr-0001-cpace-core` worktree removed; local and remote branches deleted.

**Validation:** `task check` exit 0 on the fix commit (including race tests), `gosec -tests ./...` 0 issues, `git diff --check` clean. CI green on `954759a` before merge: Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze, Staticcheck, macOS and Windows smoke; the gosec CodeQL child check neutral/skipped as designed. `mergeStateStatus: CLEAN` at merge time.

**Next:**
- Phase group 2 continues: implement the remaining accepted ADRs (0002 suite-type disposition, 0005 export-length type, 0006 close-on-nil convention, 0007 release supply-chain artifacts).
- Phase 3 consolidated evidence refresh (paired long fuzz campaigns, dependency review, SAST, Capslock) after the remaining implementations land; PR #82 carries only its interim non-evidence gate until then.

---

## ADR-0005 Export contract merged - 2026-06-12 15:30 EDT

**Main:** `2b0e3c219e7a`
**Actor:** Codex

**Summary**

ADR-0005 is implemented and merged. PR #86 pinned the `Session.Export` length contract without changing the exported signature or runtime behavior: `length` remains an `int`, valid lengths are documented as `[0, 16320]`, zero-length output is documented as length-only, and the existing range guard remains the panic barrier before `crypto/hkdf.Key`.

**Completed**

- Merged PR #86 as `2b0e3c219e7a8c92f4037d0b28209b15cece3199`; implementation head was `c41186c7d422edeb99948d7c3a05455157028c3c`.
- Updated `session.go` with the ADR-0005 export-length doc contract and a code-site comment explaining why negative lengths must be rejected before calling `crypto/hkdf.Key`.
- Added `TestExportLengthBoundaries` in `api_test.go` for `-1`, `0`, `1`, `maxHKDFOutput - 1`, `maxHKDFOutput`, and `maxHKDFOutput + 1`.
- Added the Unreleased changelog entry for the pinned `Export` length contract.

**Validation**

- GitHub checks on PR #86 were green at `c41186c`: CI Check, DCO, Dependency Gate, SAST Gate, CodeQL, Staticcheck, and cross-platform smoke.
- Local gates passed at `c41186c`: `go test -run TestExportLengthBoundaries ./...`, `task check`, `gosec -tests ./...`, `git diff --check origin/main...HEAD`, and a formal `apidiff` export-data comparison against `origin/main` with no API diff.
- Interim fuzz gate `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` passed all 14 configured targets at `4f564ac`; the final delta from `4f564ac` to `c41186c` was comment-only, and RAS verification found no security-evidence posture change.
- RAS review `20260612T180622-b8fb1b829d980b97d1086fe2` found and verified the DCO fix; fresh RAS review `20260612T191439-3be06a9cdf52510fb230436e` requested and verified the code-site comment, with no still-open findings or new concerns.

**Next**

Complete the ADR-0005 OmniFocus task and keep the broader release evidence caveat intact: stronger release claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.

---

## ADR-0006 nil Close contract merged - 2026-06-12 17:25 EDT

**Main:** `655fc2e02cc0`
**Actor:** Codex

**Summary**

ADR-0006 is implemented and merged. PR #88 makes `(*Session)(nil).Close()` return `nil`, while preserving strict handling for fabricated non-nil zero-value sessions and leaving `Session.Export` plus nil accessors unchanged.

**Completed**

- Merged PR #88 as `655fc2e02cc04cf8cd5072ee71d8de7fe0a50f1d`; implementation head was `a1f7183d37b560a2b9dc84e5970271da5ab0dda7`.
- Updated `Session.Close` so a nil receiver is a successful no-op, while `&Session{}` and `new(Session)` still return `ErrInvalidInput` with the preserved `nil session` diagnostic.
- Updated the `Session.Close` doc comment to state the nil-safe contract, and expanded the nil-receiver test matrix to pin `Close`, `Export`, `TranscriptID`, `PeerAssociatedData`, `PeerID`, and zero-value `Close`/`Export` behavior.
- Added the Unreleased changelog entry recording the pre-v1 `Close` nil-receiver contract change.
- Drove the change with `ras implement` run `20260612T195857-c0f5fb1dca2c2d83977f4554`; RAS review `20260612T200145-9ae901603f6c0c52cfcc2eeb` found the doc-comment nit and exposed the diagnostic-policy ambiguity, and fresh RAS review `20260612T201738-080738618939b3e1ffcc51ac` required the final byte-identical zero-value `nil session` behavior.

**Validation**

- GitHub checks on PR #88 were green at `a1f7183`: CI Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, and macOS/Windows smoke; the gosec CodeQL child check was neutral as expected.
- Local gates passed at `a1f7183`: `go test -run 'TestNilReceiverMethods|TestNilReceiverFinishAndExport' ./...`, `go doc . Session.Close`, `go test ./...`, `go test -race ./...`, `task check`, `gosec -tests ./...`, `git diff --check`, and a formal `apidiff` export-data comparison against `origin/main` with no API diff.
- RAS verification of review `20260612T201738-080738618939b3e1ffcc51ac` passed cleanly at `a1f7183`, with no still-open findings and no new concerns.
- Interim fuzz gate `FUZZ_RACE=0 GOMAXPROCS=4 FUZZTIME=8m PARALLEL=2 task fuzz` passed all 14 configured fuzz targets at `a1f7183`.

**Next**

- Complete the ADR-0006 OmniFocus task `fg42mN_gyWG` after this journal update lands.
- Continue with ADR-0007, then run the Phase 3 consolidated evidence refresh; stronger release claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.

---

## ADR-0007 release artifacts merged - 2026-06-13 00:19 EDT

**Main:** `cccb0dcf8d2e`
**Actor:** Codex

**Summary**

ADR-0007 is implemented and merged. PR #90 adds the release supply-chain artifact path without reopening the frozen public API or package-profile policy: signed annotated `v*` tags gate the release workflow, CycloneDX SBOM JSON is generated and validated, SBOM provenance is attested, and v0.x or SemVer prerelease tags publish as GitHub prereleases with `latest=false`.

**Completed**

- Merged PR #90 as `cccb0dcf8d2e61abab5754f5e12a214460ce4658`; implementation head was `71ad345db318687c064438cbd5ae90c3aff4e295`.
- Added `.github/allowed_signers` and hardened `.github/workflows/release.yml` with `unsupported-ref`, `verify-tag`, signed annotated tag verification, release-note extraction, fail-closed existing-release checks, pinned artifact upload/download actions, SHA-pinned `anchore/sbom-action@e22c389904149dbc22b58101806040fa8d37a610`, SHA-pinned `actions/attest@59d89421af93a897026c735860bf21b6eb4f7b26`, and SBOM plus Sigstore bundle release assets.
- Added `scripts/extract-release-notes.sh`, `scripts/release-tag-metadata.sh`, `scripts/test-release-helpers.sh`, and `scripts/validate-cyclonedx-sbom.sh`, with `task release:helpers` wired into CI and local documentation checks.
- Updated `README.md`, `CHANGELOG.md`, `docs/ci-policy.md`, `docs/release-checklist.md`, `docs/release-verification.md`, and `docs/security-gates.md` so release operators have the signed-tag, SBOM, attestation, prerelease, and fail-closed publishing contracts in one place.
- Drove the change with `ras implement` run `20260613T023632-3e88228b19f588529d5338e7`; resolved review runs `20260613T024713-7576a5e6e200a2c349683abd`, `20260613T030515-e69a3bcac3fb2be870d47ba5`, `20260613T033032-68701975a207636f683a1d4a`, and fresh review `20260613T035338-7fa650bf8acdc0067dc23765`.

**Validation**

- GitHub checks on PR #90 were green at `71ad345`: Actionlint, Check, DCO, Dependency Gate, and SAST Gate; the gosec child check was neutral/skipped as expected.
- RAS verification passed cleanly for review `20260613T033032-68701975a207636f683a1d4a` at `b2386ae254fb791d55273b778c4320ee80b93b16` and for fresh review `20260613T035338-7fa650bf8acdc0067dc23765` at `71ad345db318687c064438cbd5ae90c3aff4e295`, with no still-open findings and no new concerns.
- Local gates passed at `71ad345`: `PATH=/tmp/cpace-bin:$PATH task release:helpers`, real Syft `v1.45.1` CycloneDX SBOM generation plus `scripts/validate-cyclonedx-sbom.sh`, prerelease/latest metadata checks for `v0.1.3` and `v1.0.0`, unsafe-tag rejection checks, no-`jq` failure smoke, `task docs:check`, `task quick`, `task check`, `go test -race ./...`, `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`, `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -tests ./...`, and `git diff --check`.

**Next**

- Complete the ADR-0007 OmniFocus task `h0GnwLCRQYj` after this journal update lands.
- The v1.0.0 candidate still needs the consolidated Phase 3 evidence refresh before stronger release claims: signed `v*` workflow-dispatch rehearsal, branch/non-`v*` fail-closed checks, lightweight/unsigned/wrong-signer tag negative tests, missing changelog-section failure, `gh attestation verify`, and Scorecard before/after evidence for SBOM, Packaging, and Signed-Releases.

---

## Message framing hardening landed - 2026-06-13 12:17 EDT

**Main:** `a95e39c8ec8e`
**Actor:** Codex

### Summary

PR #92 deepened package-owned Message framing behind shared internal field specs, added a 128 KiB aggregate invalid-message decoder backstop, documented the framing term in `CONTEXT.md`, and kept valid message shapes governed by per-field caps plus exact public-share/tag lengths.

### Completed

- Merged PR #92 as `a95e39c8ec8e18d56d76f049211fe2b392d61985`; implementation head was sign-off-only rewritten from RAS-verified `ef6143282899a1ce31d685cdf66b78cdc51e60cf` to merged head `da296be1800dcd5e50a31ea87e97a9080da4f478`.
- Centralized Message A/B/C framing around package-owned specs, common encode/decode helpers, and compile-time worst-case guards proving each valid package-owned shape remains below `maxMessageLength`.
- Updated `README.md`, `docs/security-assessment.md`, `docs/project-plan.md`, and `docs/code-review-followups-2026-05-28.md` to distinguish valid-message per-field caps from the 128 KiB aggregate decoder cap for malformed framed inputs.
- Resolved both RAS review `20260613T154244-a0567247c6bdab93ec542cf3` Fix First findings; RAS verification reported no still-open findings and no new concerns before the DCO-only sign-off rewrite.

### Validation

- GitHub checks on PR #92 were green at `da296be`: CI Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, and macOS/Windows smoke; the gosec child check was neutral/skipped as expected.
- Local gates passed on the fixed implementation before the sign-off-only rewrite: `go test ./...`, `go test -race ./...`, `go vet ./...`, `task quick`, `git diff --check`, and `rg -n "aggregate message|not aggregate message|per-field" README.md docs/`.
- Compile-time guard proof passed: temporarily lowering `maxMessageLength` below the computed max valid Message A size made `go test ./...` fail at compile time with a `uint` underflow, then the 128 KiB cap was restored and tests passed.

### Next

- Stronger release claims still require refreshing pinned fuzz, dependency-review, and security/spec-audit evidence against the exact candidate commit.

---

## Release policy checker ready - 2026-06-13 13:49 EDT

**Main:** `2e63306e7653`
**Actor:** Codex

### Summary

PR #94 deepens the ADR-0007 release policy checker into an executable guard for the accepted release workflow shape, helper scripts, allowed signer set, permission ceilings, guards, job graph, exact release/SBOM/attestation steps, and bypass-resistant shell command bodies.

### Completed

- Opened PR #94, `Add release policy checker`, from `codex/release-policy-checker` to `main`.
- Added the isolated `tools/releasepolicy` Go checker, `scripts/check-release-policy.sh`, helper-test wiring, and the `CONTEXT.md` glossary entry for the release policy checker.
- Closed RAS-surfaced bypass classes across review-fix attempts: command neutralization, unreachable shell logic, injected extra shell lines, widened triggers and guards, permission creep, rogue jobs, unexpected `needs`, unchecked validation jobs, unbounded step/env/with blocks, arbitrary allowed signers, missing checkout credential hardening, and non-executable helper scripts.
- Current implementation head before the journal update is `aba8f821f0b6fb85bf01c7cf77c2848af19d58b0`.

### Validation

- Local gates passed at `aba8f821f0b6fb85bf01c7cf77c2848af19d58b0`: `(cd tools/releasepolicy && go test ./...)`, `./scripts/check-release-policy.sh`, `scripts/test-release-helpers.sh`, and `task check`.
- GitHub checks on PR #94 were green at `aba8f821f0b6fb85bf01c7cf77c2848af19d58b0`: CI Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, macOS/Windows smoke, with the gosec child check neutral as expected.
- RAS review-fix run `20260613T165919-87c8df5ab6556bf018e9c559` completed and surfaced the first hardening batch; later review-fix runs `20260613T171857-c902eba83eee88a47b6c4802`, `20260613T173208-0a17aa1001d01213f20295d0`, and `20260613T174226-35656c70061bca3f97fc30f0` exposed additional findings before their wrappers stalled, so the final clean judgment comes from the fixed code plus local and GitHub gates rather than a completed final RAS synthesis.

### Next

- Merge PR #94 after the journal commit's checks are green.
- Update OmniFocus with the merged PR and evidence notes.
- Keep the release evidence caveat intact: stronger release claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.

---

## Message framing catalogue landed - 2026-06-13 17:05 EDT

**Main:** `5df3e7ea516d`
**Actor:** Codex

**Summary**

PR #96 landed the Message framing test catalogue suggested by the architecture review. The change is test-only: malformed, aggregate-size, field-limit, LEB128, and fuzz-seed cases for Message A/B/C now live behind one intent-named catalogue, with no production behavior or public API change.

**Completed**

- Merged PR #96 as `5df3e7ea516dee1dce974294fd47b057cd62e556`; final implementation head was `1edfb7641912bc269915786bdab64d047131ffa8`.
- Added `framing_catalogue_test.go` as the shared Message framing catalogue for parser tests and fuzz seed construction.
- Reworked `api_test.go` parser coverage to consume catalogue cases for malformed headers, role/suite errors, aggregate-size precedence, max-field acceptance, field caps, and LEB128 decoding errors.
- Reworked `fuzz_test.go` so Message A/B/C decoder and protocol fuzzers share the catalogue seed builders.
- Moved malformed LEB128 coverage out of `strings_test.go` and into Message framing tests across roles A, B, and C.
- Addressed RAS follow-ups before merge with `554dea5e0a8a14c77b5af660a3e2551d95be83f7` and `1edfb7641912bc269915786bdab64d047131ffa8`: sparse over-declared Message B associated-data coverage, explicit catalogue dispatch failures, and exact truncated-field diagnostics.

**Validation**

- Local gates passed on the PR branch: `task check`; after RAS follow-ups, `go test -run TestMessageFramingCatalogueRejectsMalformed -v ./...`, `go test ./...`, `task check`, and `git diff --check`.
- GitHub checks on final head `1edfb7641912bc269915786bdab64d047131ffa8` were green: CI Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, macOS/Windows smoke; the gosec child check was neutral as expected.
- RAS review-fix runs `20260613T203023-1da9e38f205b48430ec784fe`, `20260613T204452-218b22c25c13844f27dad5ed`, and `20260613T205302-e830df4d0d7e4a10e7809c39` completed. The final run reported no merge-blocking findings for PR #96 at `1edfb7641912bc269915786bdab64d047131ffa8`.

**Next**

- Optional non-blocking follow-up from the final RAS pass: add aggregate-size boundary catalogue cases for exactly `maxMessageLength` and `maxMessageLength+1` if the maintainer wants to exhaust the remaining low-severity test-hardening suggestion.
- Keep the release evidence caveat intact: stronger release claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.

---

## Package-owned cap policy ready - 2026-06-13 18:15 EDT

**Branch:** `codex/package-owned-cap-policy`
**PR:** #98
**Actor:** Codex

**Summary**

PR #98 concentrates the internal Package-owned cap policy behind `caps.go` without changing public API, wire format, cap values, error identity, or error text. `normalizeConfig`, Message framing specs, catalogue tests, and message fuzz guards now read the same cap facts, while the aggregate `maxMessageLength` parser backstop remains in Message framing.

**Completed**

- Added `caps.go` with package-owned cap facts for caller-provided `Config` fields and package-owned Message A/B/C fields.
- Updated `api.go` so `normalizeConfig` reads local input cap names and lengths from the cap policy while preserving validation order and diagnostic text.
- Updated `framing.go` so `messageASpec`, `messageBSpec`, and `messageCSpec` read their field specs from the cap policy; LEB128 parsing, role checks, and `maxMessageLength` stayed in Message framing.
- Added `caps_test.go` to pin shipped cap names, lengths, exact-vs-capped semantics, and Message framing spec usage.
- Updated `api_test.go`, `framing_catalogue_test.go`, and `fuzz_test.go` so size-limit tests, catalogue cases, and message round-trip fuzz guards use cap-policy fields where they are testing package-owned field caps.
- Added the Package-owned cap policy term to `CONTEXT.md`.
- Addressed RAS review finding C-001 with signed-off commit `9e07b452f7f0c4fc2727a06ef54f57b276349492`, changing catalogue fixtures to use message-matched `aPoint`, `bPoint`, `bTag`, and `cTag` cap fields.
- Addressed the GitHub SAST/gosec `-tests` G115 finding with signed-off commit `f76d7d683ec6ab201decb4ec718853a0651531c2`, keeping the over-declared Message B AD test on the cap-policy constant so gosec sees a compile-time-safe LEB128 length.

**Validation**

- Local gates passed after the signed-off rewrite and SAST fix: `go test ./...`, `go test -race ./...`, `task check`, `gosec -tests -fmt sarif -out /tmp/cpace-gosec.sarif . ./tools/releasepolicy`, and `git diff --check`.
- `task check` included `scripts/test-release-helpers.sh`, `(cd tools/releasepolicy && go test ./...)`, release policy checker validation, `go vet ./...`, `staticcheck ./...`, `ast-grep scan --error`, and `govulncheck -test ./...`; Syft was not installed, so the helper script skipped only its optional real Syft SBOM validation path.
- RAS review `20260613T220234-2d89640ea92a616f9cb7b8a6` on PR #98 found one non-blocking test clarity nit and no behavior-preservation, security, wire-format, cap-value, or error-text blocker.
- RAS verify `20260613T220234-2d89640ea92a616f9cb7b8a6-verify-20260613T221245-cad99e71ff3e0e6351a3addf`, rerun against the later PR head `2193af0c72358df16f5a3a8edb4d2e913d7a01a7` before the signed-off rewrite, reported C-001 resolved, no still-open findings, and no new concerns; the subsequent signed-off rewrite preserves the reviewed diff content.

**Next**

- Merge PR #98 only with explicit maintainer authorization.
- Keep the release evidence caveat intact: this is internal but security-relevant because it moves allocation and parsing-limit facts, so stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security/spec-audit evidence against the exact candidate commit.

---

## Package cap policy merged - 2026-06-13 18:36 EDT

**Main:** `a01df3746207`
**Actor:** Codex

**Summary**

PR #98 merged the Package-owned cap policy refactor into `main` at merge commit `a01df374620724f6fb88dbe9328c8dd6984bda7c`. The change concentrates internal caller-field and Message framing cap metadata in `caps.go` / `caps_test.go` while preserving the frozen public interface, package-profile policy, cap values, wire format, and current validation diagnostics.

**Completed**

- Merged `codex/package-owned-cap-policy` through GitHub PR #98 using the repo's merge-commit pattern.
- Kept the cap policy package-owned: no public profile knob, no exported interface, and no observable behavior change intended.
- Recorded the Package-owned cap policy term in `CONTEXT.md` and kept the prior implementation evidence in this journal.
- Resolved RAS review finding `C-001` and verified the fix against head `863090b84331cec0824ef8a0a5f9a0c68f160a89` with no open or new concerns.

**Validation**

- Local gates before merge: `task check`, `gosec -tests -fmt sarif -out /tmp/cpace-gosec.sarif . ./tools/releasepolicy`, and `git diff --check` passed.
- GitHub checks for PR #98 passed before merge: CI, CodeQL, Cross-Platform Smoke on macOS and Windows, DCO, Dependency Gate, SAST Gate, and Staticcheck Advisory. The standalone gosec annotation was neutral while SAST Gate was green.
- RAS review run: `20260613T220234-2d89640ea92a616f9cb7b8a6`; RAS verification against the final PR head reported `C-001` resolved with no still-open or new concerns.

**Next**

- Refresh pinned dependency-review, fuzz, and security-audit evidence against the exact post-merge candidate commit before making stronger release-readiness claims from this `main` line.

---

## Release policy checker catalogue merged - 2026-06-13 21:52 EDT

**Main:** `63c15b202c53`
**Actor:** Codex

**Summary**

PR #100 merged the Accepted release policy catalogue refactor into `main` at squash merge commit `63c15b202c53201ed6303760b5794c73cfc3b2d8`. The change deepens the Release policy checker catalogue by moving accepted ADR-0007 workflow facts into a dedicated internal policy catalogue while preserving the frozen public API and package-profile policy.

**Completed**

- Added `tools/releasepolicy/policy.go` as the internal catalogue of accepted release-policy facts for required workflows, jobs, permissions, pinned-action rules, artifact conventions, and expected advisory-gate outcomes.
- Refactored `tools/releasepolicy/main.go` so workflow checks read from the accepted policy catalogue instead of duplicating release-policy facts inline.
- Added catalogue integrity coverage in `tools/releasepolicy/main_test.go` so the shipped policy facts stay complete and searchable by workflow, job, and expected finding.
- Added the Accepted release policy term to `CONTEXT.md`.
- Ran RAS review-fix loop `20260614T002318-407d88529b264e3db0345f62`; the loop completed with no merge-blocking findings remaining under the configured gate policy.

**Validation**

- Local validation passed before PR creation: `go test ./...`, `(cd tools/releasepolicy && go test ./...)`, `scripts/test-release-helpers.sh`, and `task docs:check`.
- `scripts/test-release-helpers.sh` skipped only the optional real Syft SBOM validation path because `syft` was not installed.
- GitHub checks for PR #100 passed on amended head `5d34db41d8da58314d8a45c5e5213fbbc706dc19`: CI, CodeQL, Cross-Platform Smoke on macOS and Windows, DCO, Dependency Gate, SAST Gate, and Staticcheck Advisory. The standalone gosec annotation was neutral/skipping while SAST Gate was green.
- The PR was merged into `main` as `63c15b202c53201ed6303760b5794c73cfc3b2d8`.

**Next**

- Keep the release evidence caveat intact: stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.
- Optional follow-up: decide whether the Release policy checker should also expose the existing catalogue lookup drift and setup/report/SBOM coverage gaps as normal findings rather than implementation panics or implicit expectations.

---

## Release policy catalogue validation merged - 2026-06-14 05:02 EDT

**Main:** `bdfe57f3053f`
**Actor:** Codex

**Summary**

PR #102 deepened the Release policy checker catalogue by moving accepted ADR-0007 job and step facts into `tools/releasepolicy/policy.go` and replacing many bespoke per-job checks with generic catalogue validation. The change preserves the release workflow YAML, public Go package surface, CPace protocol behavior, and package-profile policy.

**Completed**

- Merged PR #102 as `bdfe57f3053fa85735ad9ea093490bb042c165f9` with squash subject `refactor: deepen release policy catalogue validation`.
- Expanded the Accepted release policy catalogue with job display names, runners, timeouts, outputs, setup/report steps, step IDs, and exact step controls.
- Refactored `tools/releasepolicy/main.go` so accepted jobs and steps are validated through one catalogue-driven module while global checks still enforce action SHA pinning, checkout credential hardening, and shell expression restrictions.
- Hardened the checker against duplicate YAML keys and alias-valued scalar-control bypasses before semantic validation.
- Removed redundant `requiredOutputs` validation after exact SBOM output checks made it duplicate reporting.
- Added regression coverage for duplicate keys, alias-valued controls, checkout hardening single-report behavior, rogue checkout fallback coverage, swapped-step diagnostics, unexpected job/step keys, SBOM output validation, and attestation step IDs.
- Ran RAS review-fix on PR #102; implementation id `20260614T044228-fd8cab0d1575b5003e538c5a` finished `done` after two fix cycles and a final clean gate-policy result. The final RAS synthesis left only a low-severity test-coverage follow-up, which was folded into the signed PR branch before merge.

**Validation**

- Local gate before merge: `task check` passed; the release helper smoke test skipped only the optional real Syft SBOM validation path because `syft` was not installed.
- GitHub checks on the final signed PR head `7e61f2fb22c1a41fd6e523b998ec4c7fe13fccae` passed: Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, macOS smoke, and Windows smoke. The standalone `gosec` child check was `skipping` as expected.

**Next**

- Keep the release evidence caveat intact: stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit.

---

## Peer-share rejection PR reviewed - 2026-06-14 05:58 EDT

**Main:** `bff0c4aedb14`
**Actor:** Codex

**Summary**

PR #104 deepens Peer-share rejection behind a role-aware internal module without changing public API, wire format, exported sentinels, or observable error strings. The CPace core now asks the module for peer public share validation and shared-secret computation, while Message framing remains responsible for wire size and role parsing.

**Completed**

- Added `peer_share.go` with `peerShareRole.validate`, `peerShareRole.sharedSecret`, ADR-0003 role-context error mapping, `scalarMultVFY`, and `decodePublicShare` colocated behind the Peer-share rejection module.
- Updated `core.go` so initiator and responder flows call the role-aware Peer-share rejection seam rather than composing peer-share validation and error mapping in place.
- Removed the old `wrapPeerShareError` helper from `api.go`, keeping the public shell out of the peer-share sentinel mapping details.
- Added focused tests in `api_test.go` for role-context wrapping, valid shared-secret derivation, and non-sentinel defensive error pass-through.
- Added the Peer-share rejection term to `CONTEXT.md`.
- Opened PR #104 at `https://github.com/the-sarge/cpace/pull/104` from `codex/deepen-peer-share-rejection`.
- Ran RAS review-fix twice: `20260614T094003-de7e34e42fbb5b171eb70955` found one low comment-maintenance issue, fixed by `e1a6d22c1468931bc0df1fe37032f222f790b7a9`; `20260614T094538-bfaaf38c09133da4b850667e` finished `done` with no merge-blocking work remaining and only a low, forward-looking test-hardening suggestion.

**Validation**

- Local gates passed on the PR branch: `go test ./...`, `go test -race ./...`, `(cd tools/releasepolicy && go test ./...)`, `task check`, and `rg -n "new peer-share sentinel|surfaces without role context|wrapError applies" peer_share.go`.
- `task check` included docs UTF-8 validation, release helper smoke tests, release policy checker validation, root tests, race tests, gofmt/goimports checks, `go vet ./...`, `staticcheck ./...`, `ast-grep scan --error`, and `govulncheck -test ./...`; Syft was not installed, so the helper script skipped only optional real Syft SBOM validation.
- GitHub checks on head `e1a6d22c1468931bc0df1fe37032f222f790b7a9` passed except DCO. DCO reported the two existing PR commits lack `Signed-off-by` trailers, so PR #104 is mergeable at Git level but branch-protection blocked until the commits are rewritten with signoffs.

**Next**

- Rewrite the PR branch with signed-off commits if the maintainer authorizes the required force-push, then rerun GitHub checks and merge PR #104 once branch protection is satisfied.
- Keep the release evidence caveat intact: this is security-relevant code movement, so stronger release-readiness claims require refreshing pinned dependency-review, fuzz, and security/spec-audit evidence against the exact candidate commit.

---

## Peer-share rejection merged - 2026-06-14 06:14 EDT

**Main:** `aa3b30fe6f89`
**Actor:** Codex

**Summary**

PR #104 merged the Peer-share rejection module deepening after the DCO-only rewrite. The final merged shape keeps public API, wire format, exported sentinels, package-profile policy, and observable error strings unchanged while giving peer public share validation and shared-secret rejection one role-aware internal module.

**Completed**

- Rewrote the PR branch with DCO signoff only after maintainer approval, using `git rebase --signoff origin/main` and `git push --force-with-lease`.
- Merged PR #104 at `aa3b30fe6f895655d2d2259e9e1e62c3ad34dc97` from signed head `acf9e06e3d610b98ba16356864cb95347202358c`.
- Installed a repo-local `.git/hooks/commit-msg` hook in this checkout so future local commits automatically receive the configured `Signed-off-by` trailer before GitHub DCO sees them.
- Updated OmniFocus task `c10wiAvxUJl` so the evidence-refresh tracker points at the merged PR #104 candidate instead of the temporary DCO blocker note.

**Validation**

- GitHub checks on the signed PR head passed before merge: Check, DCO, Dependency Gate, SAST Gate, CodeQL Analyze/CodeQL, Staticcheck, macOS smoke, and Windows smoke; the standalone gosec child check was neutral/skipping as expected.
- Local checkout was fast-forwarded to the merge commit, and the DCO hook was verified against a throwaway commit message.

**Next**

- Keep the release evidence caveat intact: PR #104 is security-relevant code movement, so stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, Capslock, and security/spec-audit evidence against the exact candidate commit.

---

## Go fix release policy cleanup landed - 2026-06-14 07:27 EDT

**Main:** `d5048eb5ac29`
**Actor:** Codex

**Summary**

PR #107 applied the current `go fix -diff` suggestions to the nested `tools/releasepolicy` module. The root module dry run was already clean; the nested module had only the `strings.SplitSeq` iterator modernization and the `slices.Contains` membership simplification.

**Completed**

- Merged PR #107 as `d5048eb5ac2968721af2b4b370885ffaa78edf8f` from head `fdf0e5b87d9723d526b57c8389af1bd604358c4d`.
- Updated `tools/releasepolicy/main.go` to iterate script lines with `strings.SplitSeq` and replace the local `contains` loop with `slices.Contains`.
- Ran RAS review-fix on PR #107; implementation id `20260614T112349-ede9cb520bc21187472dc9c5` finished `done` with no actionable findings.
- Kept the change scoped to the internal release policy tool; no public API, crypto, framing, package-profile policy, or release evidence files changed.

**Validation**

- Local pre-merge gates passed: `go fix -diff ./...`, `(cd tools/releasepolicy && go fix -diff ./...)`, `(cd tools/releasepolicy && go test ./...)`, `go test ./...`, and `git diff --check`.
- GitHub checks on PR #107 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone `gosec` child check was neutral.

---

## Evidence baseline validator merged - 2026-06-14 16:11 EDT

**Main:** `149c3d1e96a8`
**Actor:** Codex

**Summary**

PR #109 merged the Evidence baseline validator adapter. The change adds a read-only checker for `docs/evidence-baseline.md` and committed `docs/evidence/**` bundles, plus CI/local routing so evidence-related changes validate the pinned baseline without changing public API, CPace protocol behavior, package-profile policy, or release-readiness claims.

**Completed**

- Merged PR #109 as `149c3d1e96a860bf594991b90392583217a0cd92` from final head `2c9e256a42a2621893fba3f74e1dc3dd34a084f2`.
- Added `tools/evidencebaseline`, `scripts/check-evidence-baseline.sh`, `scripts/classify-check-changes.sh`, and `scripts/test-ci-classifier.sh`.
- Wired `task evidence:baseline`, `task evidence:lint`, and `task ci:classifier` into `task check` and `task check:changed`.
- Updated CI so code/workflow changes, evidence bundle/index changes, and docs-only changes to summary docs referenced from the Baseline Index run the evidence validator; unrelated Markdown-only PRs still stay on docs validation without Go setup.
- Hardened validation against malformed Baseline Index separators, unsafe refs, symlinked path ancestors, symlinked `docs/evidence`, symlinked bundle roots, symlinked raw artifacts, symlinked summary docs, symlinked bundle control files, uncovered nested raw files, invalid checksums, duplicate checksum paths, empty checksum files, and optional `SHA256SUMS.sig` symlinks.
- Ran RAS review-fix twice. The first run `20260614T175357-d523fd0c2f4aa165464a8758` blocked on an all-docs Go policy expansion, which was removed. The fresh run `20260614T185112-6107f291c7955d2908cacb85` reached `max_review_loops` after three builder passes; its final pushed head fixed the last reported fix-first findings, and no low/nit finding was left intentionally deferred into a follow-up issue.

**Validation**

- Local final gates passed on `main` after merge: `scripts/test-ci-classifier.sh`, `scripts/check-evidence-baseline.sh`, `(cd tools/evidencebaseline && go test ./... && go vet ./... && staticcheck ./...)`, `task evidence:lint`, `task check:changed`, `task check`, and `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`.
- `task check` included docs validation, release-helper smoke tests, CI classifier tests, evidence baseline validation, nested evidence-checker linting, root tests, race tests, gofmt/goimports checks, root `go vet`, root Staticcheck, ast-grep, and `govulncheck -test ./...`; the helper script skipped only optional real Syft SBOM validation because `syft` was not installed.
- GitHub checks on final head `2c9e256a42a2621893fba3f74e1dc3dd34a084f2` passed before merge: Check, Actionlint, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Keep the release evidence caveat intact: this validator improves consistency checks for existing pinned evidence, but stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, Capslock, and security/spec-audit evidence against the exact candidate commit.

---

## Evidence baseline summary discovery merged - 2026-06-14 17:31 EDT

**Main:** `2eab91502ee5`
**Actor:** Codex

**Summary**

Merged PR #111 to make Evidence baseline own summary-doc discovery while preserving pre-Go CI classification. The Go checker now parses the Baseline Index, writes and validates `docs/evidence-baseline-summary-docs.txt`, and the shell classifier reads that generated adapter instead of reparsing Markdown.

**Completed**

- Added the generated summary-doc manifest and documented its regeneration command.
- Replaced the classifier's AWK Baseline Index parser with manifest reading.
- Added manifest freshness, whitespace, symlink, write-path, and missing-manifest coverage.
- Updated README and CI policy prose so the generated manifest is listed as an evidence-validator trigger.
- Opened follow-up issue #112 for non-blocking adapter hardening left by the clean RAS review loop.

**Validation**

- Local: `scripts/check-evidence-baseline.sh`, `task ci:classifier`, `task evidence:lint`, `task docs:check`, `go test ./...`, and `task check:changed` passed before merge.
- RAS: `ras review-fix` run `20260614T204231-f0307881a12cf606d27b9ee5` completed with final status `done` at PR head `ec6c41ffe96dfac9b0b1dc757aa4233a398832bc`.
- GitHub: PR #111 merged cleanly as `2eab91502ee5b7125cb9def8d8dccb7eb0debb69`; required checks were green at the reviewed head.

**Next**

- Issue #112 tracks low/nit follow-up hardening for manifest shape guards, explicit shell error propagation, and a short note on the Go `--list-summary-docs` inspection flag.

---

## Package cap policy deepening landed - 2026-06-15 22:47 EDT

**Main:** `68b8443694e2`
**Actor:** Codex

**Summary**

PR #114 deepened the Package-owned cap policy module without changing public API, wire format, cap values, package-profile policy, or release-readiness claims. Config cap acceptance now validates all local byte fields before cloning and returns package-owned copies through one internal module; Message framing now delegates field length, truncation, and cloning checks to the same cap-policy implementation while keeping header, role, aggregate-size, and LEB128 orchestration in framing.

**Completed**

- Merged PR #114 as `68b8443694e25f8fa8f08bdb0ec65a32246e848e` from reviewed head `d9615218af8e58e590390c3fa83af49fa36f8b27`.
- Updated `caps.go` with `acceptConfig`, cap-policy Config copy ownership, Message framing field acceptance, and a shipped cap-policy catalogue for tests.
- Simplified `normalizeConfig` so it delegates local byte-field acceptance to Package-owned cap policy, then transfers accepted copies into `normalizedConfig`.
- Simplified `messageReader.readField` so Message framing delegates field cap, exact-length, truncation, and clone behavior to Package-owned cap policy.
- Updated cap-policy tests to pin the shipped catalogue through the policy interface and cover Config copy ownership plus no caller-input mutation on cap failure.
- Ran RAS review-fix on PR #114; implementation id `20260616T023946-93fea2d66c355775659a93b5` finished `done` with no merge-blocking findings. The only reported item was a non-blocking nit about documenting the ownership-transfer rollback guard; no follow-up issue was opened because it does not affect behavior, verification, or future implementation work.

**Validation**

- Local pre-merge gates passed: `go test ./...`, `task quick`, `go test -race ./...`, `task check`, and `git diff --check`.
- `task check` included docs validation, release-helper smoke tests, CI classifier tests, evidence baseline validation, nested evidence-checker linting, root tests, race tests, gofmt/goimports checks, root `go vet`, root Staticcheck, ast-grep, and `govulncheck -test ./...`; the helper script skipped only optional real Syft SBOM validation because `syft` was not installed.
- GitHub checks on PR #114 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone `gosec` child check was neutral/skipping as expected.

**Next**

- Treat this as security-relevant internal validation movement for evidence discipline: stronger release claims still require refreshing pinned dependency-review, fuzz, Capslock, and security/spec-audit evidence against the exact candidate commit.

---

## ADR-0008 lifecycle thaw accepted - 2026-06-16 00:35 EDT

**Main:** `c266a540d27a`
**Actor:** Codex

**Summary**

Accepted and merged ADR-0008, recording a narrow public-lifecycle thaw for explicit cleanup of abandoned `Initiator` and `Responder` single-use state. The accepted design adds role-state `Close` methods in a follow-up implementation PR and specifies shared terminal state so constructed value copies preserve exactly-one-terminal-operation semantics.

**Completed**

- Added `docs/adr/0008-single-use-state-close.md` and merged PR #116: https://github.com/the-sarge/cpace/pull/116
- Ran ADR gating with `ras consider` run `20260616T040116-7edbef4428d20850e0094ce1`, then revised the ADR for copied single-use values and failed-Finish cleanup semantics.
- Ran fresh `ras consider` run `20260616T041535-4f06c05a3b6dc7d3f4d7b388` and `ras verify` verification `20260616T041535-4f06c05a3b6dc7d3f4d7b388-verification-1781583998570028000`, which reported `unresolved: []`.
- Ran PR review loop `ras review-fix 116`; final status was `done` with no merge-blocking findings.
- Merged PR #116 at merge commit `c266a540d27a89fbd3fd2d8d0374ddd48e71897a`.

**Validation**

- `task docs:check` passed locally before PR creation and after ADR revisions.
- GitHub checks on PR #116 passed: Check, DCO, Dependency Gate, and SAST Gate; the gosec advisory check was neutral/skipped.

**Next**

- Implement ADR-0008 in a separate code PR.
- Do not claim refreshed release evidence from ADR-0008 until exact-candidate evidence refresh covers the implementation commit.

---

## Implement ADR-0008 single-use state Close - 2026-06-16 00:52 EDT

**Main:** `977d08476486`
**Actor:** Codex

- Merged PR #118 (`feat: add single-use state close`) at `977d084764860a2d5957285b3826688f8bcf0179`.
- Implemented ADR-0008 by adding `Initiator.Close` and `Responder.Close`, moving constructed initiator/responder values onto shared terminal state, and documenting the lifecycle contract in README, package docs, integration guidance, changelog, and CONTEXT.
- Ran `ras review-fix 118`; review loop `20260616T044602-6333a251248f2384ee160008` reported no fix-first findings and no required code changes.
- Validation before merge: `task check` passed locally; GitHub checks for PR #118 passed with merge state `CLEAN`.
- Evidence refresh was intentionally deferred per maintainer direction; exact-candidate evidence refresh remains tracked separately before stronger release claims.

---

## ADR-0009 caller input thaw accepted - 2026-06-16 04:24 EDT

**Main:** `3a3109cc7957`
**Actor:** Codex

**Summary**

PR #120 accepted ADR-0009, recording a broad Caller input replacement whose authorization is narrowly limited to the follow-up role-local `Input` implementation. The merge updates the domain glossary, top-level freeze guidance, project plan, and agent instructions so future work can implement ADR-0009 without reopening unrelated public API or package-profile policy.

**Completed**

- Merged PR #120 as `3a3109cc7957a79334ce8a288e51d5804f3a5270` from reviewed head `40e61616ae17a4e73ada0c666dc5e0e23cb4e5b5`.
- Added accepted `docs/adr/0009-caller-input.md`, defining `Input{Password, SelfID, PeerID, Context, SessionID, LocalAssociatedData, AllowEmptySessionID}` as the v1 Caller input module and removing public `Config` from the intended v1 surface.
- Updated `CONTEXT.md` with Caller input vocabulary and Package-owned cap policy wording; updated `README.md`, `docs/project-plan.md`, and `AGENTS.md` so ADR-0009 is the only caller-input thaw and all unrelated public-surface/package-profile choices remain frozen.
- Gated the ADR through `ras consider` runs `20260616T065527-779be2a67a01358b100aa80e`, `20260616T071012-53ce69dd6acfbef7baa79635`, and `20260616T072321-d625df82f8b30786dc5ac33d`; `ras verify` `20260616T072321-d625df82f8b30786dc5ac33d-verification-1781595592834070000` returned `unresolved: []`.
- Ran PR review-fix on PR #120. The first loop applied policy/wording fixes through review runs `20260616T074145-56a39e2f9bb71caa3ffa6539` and `20260616T075521-2392474c90378db8a31180c0`; the final review-fix pass `20260616T081030-b1977c774144d993e87e3f44` reported no merge-blocking fixes.
- Created non-blocking follow-up issue #121 for adding `Session.TranscriptID()` to caller-input copy-ownership implementation tests and added matching OmniFocus task `bLa0Ezk3_r9`.

**Validation**

- Local docs validation passed on the ADR branch: `task docs:check` and `cmark --validate-utf8 docs/adr/0009-caller-input.md`.
- RAS PR review-fix final status for the last pass was `done`; its only finding was the low-severity follow-up now tracked as issue #121.
- GitHub checks on PR #120 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipped.

**Next**

- Implement ADR-0009 in a separate TDD code PR, using issue #121 as part of the copy-ownership test checklist.
- Do not claim refreshed release evidence from ADR-0009 until the exact-candidate evidence refresh covers the implementation commit and reviewer-packet re-pin.

---

## Caller input implementation landed - 2026-06-16 04:58 EDT

**Main:** `5b7e61576751`
**Actor:** Codex

**Summary**

PR #123 implemented accepted ADR-0009 by replacing the public `Config` caller-input surface with role-local `Input`, mapping `SelfID`/`PeerID` per role before CI construction, and preserving wire format while updating examples, fuzz, benchmarks, live docs, evidence caveats, and the named secret-lifetime audit.

**Completed**

- Merged PR #123 as `5b7e615767518edac7cf7251520e7ed9a72ec909` from reviewed implementation head `d76015c68f891182075e1656252ae0b5fee9f7cc`.
- Added `Input{Password, SelfID, PeerID, Context, SessionID, LocalAssociatedData, AllowEmptySessionID}` and removed public `Config`; `Start` and `Respond` now accept `Input`.
- Added public coverage for role-local mapping, peer metadata, validation diagnostics and precedence, nil/empty `LocalAssociatedData`, reversed role-local identity confirmation failure, and copy ownership including `Session.TranscriptID()` stability.
- Added `docs/adr-0009-secret-lifetime-audit.md` and updated README, package docs, integration guidance, threat model, security assessment, spec matrix, external-review handoff, evidence baseline, security/spec audit, changelog, and CONTEXT.md for role-local caller-input language.
- Closed issue #121 through the implementation PR.
- Ran `ras review-fix` on PR #123; run `20260616T084506-a42497f8f2a063f8958f91a6` finished `done` with no merge-blocking findings. The two low-severity follow-ups were filed as #124 and #125 and added to OmniFocus as `o-CyFwzIIEr` and `aNp2T318MJz`.

**Validation**

- Local final gates passed before merge: `go test ./...`, `go vet ./...`, `go test -race ./...`, `task docs:check`, `task check`, and `git diff --check`.
- ADR-0009 acceptance sweeps were run: public-surface grep, deterministic-seam grep, docs vocabulary grep, and `go doc . Input`, `go doc . Start`, `go doc . Respond`.
- GitHub checks on PR #123 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Non-blocking follow-ups remain open as #124 and #125.
- Stronger release claims still require refreshing pinned dependency-review, fuzz, Capslock, and security/spec-audit evidence against the exact candidate commit.

---

## Caller input acceptance deepened - 2026-06-16 11:14 EDT

**Main:** `19e9aacbc6b2`
**Actor:** Codex

### Summary

Deepened the **Caller input** implementation in PR #127 without changing the public `Input`, `Start`, or `Respond` surface. The refactor moved public `Input`, accepted input, normalized input, validation/cap association, role mapping, and wipe ownership into `input.go`, leaving package-owned cap primitives in `caps.go` and CPace computation in the **CPace core**.

### Completed

- Merged PR #127 (`refactor: deepen caller input acceptance`) at `19e9aacbc6b20ffbe488aa6120852f9bf0a32a88`.
- Added characterization coverage for caller-owned `Input` slices across all accepted fields and error paths.
- Updated ADR-0009 secret-lifetime audit references after moving Caller input acceptance and normalization to `input.go`.
- Ran RAS review/fix loops on PR #127: `20260616T145633-e5bd3e28d6a9516775112b68` on head `3b03c3c`, then `20260616T150525-4806d0c15d4a8cabbd207f5d` on head `fb5dc0a`.
- Created follow-up issue #128 for the non-blocking low-severity docs drift in `docs/cpace-core-plan.md`.

### Validation

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `git diff --check`
- `task check`

`task check` completed successfully; the release helper smoke test reported `syft not found; skipping optional real Syft SBOM validation`, matching existing helper behavior.

### Next

- Resolve #128 before v1.0.0 documentation freeze if the living CPace core plan should keep matching current implementation names.
- Include this security-relevant Caller input refactor in the next exact-candidate release evidence refresh before making stronger release-readiness claims.

---

## Responder peer-share decode reuse landed - 2026-06-16 15:21 EDT

**Main:** `512ed19d450e`
**Actor:** Codex

**Summary**

PR #138 landed issue #80 by reusing the responder's prevalidated initiator peer share for scalar multiplication, avoiding the second responder-side decode while preserving the draft-shaped encoded helper for vector and spec traceability.

**Completed**

- Merged PR #138 (`refactor: reuse responder peer share decode`) as `512ed19d450e29eb75f997b9785f324ce3d8d073`; issue #80 closed automatically at merge.
- Added internal `scalarMultVFYElement` and role-aware peer-share decode/shared-secret helpers while keeping `scalarMultVFY(s, encoded)` for encoded-byte callers.
- Updated `newResponderCore` to decode `Ya` once during validate-before-randomness prevalidation and reuse that element for the responder Diffie-Hellman computation.
- Preserved public API, wire behavior, ADR-0003 peer-share sentinel/call-site mapping, role-context wrapping, and the post-multiply neutral-element defense.
- Updated `CHANGELOG.md`, `docs/spec-matrix.md`, and `docs/security-spec-audit.md` to document the internal optimization and note that stronger release claims still require an exact-candidate evidence refresh.
- Filed non-blocking RAS nit follow-up #139 for removing the now-vestigial internal `peerShareRole.validate` wrapper, and added OmniFocus task `hMxvjmcXGyU`.

**Validation**

- TDD red/green covered `scalarMultVFYElement` parity with the encoded helper and role-aware decode/shared-secret error context.
- Focused local tests covered responder prevalidation before randomness, peer-share role wrapping, scalarMultVFY behavior, invalid Ristretto encodings, identity rejection, wire-length rejection, and draft 21 vectors.
- Local gates passed before merge: `task docs:check`, `task check`, `git diff --check`, and `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -exclude-dir=.ras -tests ./...`.
- `go test -run '^$' -bench '^BenchmarkRespond$' -count 10` improved from roughly 75.4-76.6 us/op, 3192 B/op, 45 allocs/op on baseline to roughly 70.1-71.8 us/op, 3032 B/op, 44 allocs/op on the PR branch.
- RAS review-fix run `20260616T190349-319dbb7b2f94e63af17be556` completed with no merge-blocking findings; the only nit was moved to #139.
- GitHub checks on PR #138 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Resolve #139 as internal cleanup before v1.0.0 if time permits.
- Refresh pinned dependency-review, fuzz, Capslock, and security/spec-audit evidence against the exact candidate commit before making stronger release-readiness claims.

---

## Architecture slices landed - 2026-06-16 14:47 EDT

**Main:** `3835b8736a69`
**Actor:** Codex

### Summary

Completed the architecture-slice batch after the release-readiness freeze was explicitly lifted for this work: PR #130 collapsed copied single-use terminal state into one private terminal module, PR #132 hardened release helper contract tests and validation, and PR #134 centralized length-value encoding and reuse across string/framing/crypto code without changing public API or wire behavior.

### Completed

- Merged PR #130 (`refactor: collapse single-use terminal state`) as `6eb018db5037969ec7218a8b2c52b5969e8e865e`; its RAS review-fix run `20260616T170948-f7b7cf160cb1c1cfcb39639f` completed with no merge-blocking findings, and the low/nit cleanup was filed as issue #131.
- Merged PR #132 (`test: deepen release helper contracts`) as `086812a5a6b1c08a26252a112b57f9373f6aad2a`; its RAS review-fix run `20260616T172442-18a13fc8f4401aac4d67cf65` completed with no merge-blocking findings, and the newline-bearing tag follow-up was filed as issue #133.
- Merged PR #134 (`refactor: deepen length-value encoding`) as `3835b8736a699bae9aa5ca1e48dd4d576bb809fd`; review follow-ups from runs `20260616T173416-f119c481f2391f21ede13e39` and `20260616T175128-d8bce5d2be304754406fdd54` were folded back into the PR, and the final RAS review-fix run `20260616T180735-9770be235a471c08073f28c6` reported no actionable findings on the reviewed head before merge.
- Added issue #136 for the deferred Caller input architecture question instead of widening this batch beyond the three selected low-risk slices.
- Closed issue #135 after PR #134 merged and completed its OmniFocus task `fwQqC9coN-n`; remaining open follow-up tasks are #131 (`acrb5smRKe9`), #133 (`nv9mcl6_6TF`), and #136 (`i3nngrHasAi`).

### Validation

- PR #130 local gates passed before merge: focused single-use lifecycle tests, `go test ./...`, `go test -race ./...`, and `task check`; hosted checks passed before merge.
- PR #132 local gates passed before merge: `scripts/test-release-helpers.sh`, `(cd tools/releasepolicy && go test ./...)`, and `task check`; hosted checks passed after updating the branch with a signed-off merge commit.
- PR #134 local gates passed before merge: focused length-value/framing/string tests, `task check`, and `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -exclude-dir=.ras -tests ./...`; after the final signed-off base update, `task check` passed again and hosted checks passed before merge.

### Next

- Resolve #131 and #133 as non-blocking cleanup when convenient.
- Evaluate #136 separately before deciding whether to concentrate Caller input field policy.
- Stronger release-readiness claims still require refreshing dependency-review, fuzz, security-audit, and related pinned evidence against the exact candidate commit after these security-relevant changes.

---

## Peer-share validate cleanup landed - 2026-06-16 16:30 EDT

**Main:** `a05d967ffa94`
**Actor:** Codex

**Summary**

PR #141 completed issue #139 by removing the vestigial internal `peerShareRole.validate` wrapper left behind after responder peer-share decode reuse, while keeping role-context rejection coverage on the real `sharedSecret` and `decode` paths.

**Completed**

- Merged PR #141 (`refactor: remove peer share validate wrapper`) as `a05d967ffa94aa6499f9d0330297010ab409ff0b`; issue #139 closed automatically at merge.
- Removed `peerShareRole.validate` from `peer_share.go`.
- Updated `TestPeerShareRoleSharedSecretAddsRoleContext` so the former `validate` assertion now exercises `initiatorPeerShare.sharedSecret(s, invalid.InvalidY2)`, preserving initiator identity-sentinel and exact role-context error coverage.
- Kept separate direct `decode` coverage in `TestPeerShareRoleDecodeSharedSecretAddsRoleContext`.
- Confirmed no validate-wrapper call sites remain.
- Completed OmniFocus task `hMxvjmcXGyU`; no follow-up issues or OmniFocus tasks were needed.

**Validation**

- Baseline focused peer-share tests passed before the cleanup.
- The RED acceptance grep found the remaining validate-wrapper test call before removal.
- After cleanup, the same grep produced no matches.
- Local gates passed before merge: focused peer-share tests, `go test ./...`, `git diff --check`, and `task check`.
- RAS review-fix run `20260616T201826-ef5980e31fc0ec00230494f6` on head `17856c1426b815618e40b7b26c98cea45f52c68e` found a non-blocking test-duplication nit, which was fixed in follow-up commit `d67ef86303b385e91af2ce049988b78471b36453`.
- RAS review-fix run `20260616T202330-c39ed657dd21626806deca74` on head `d67ef86303b385e91af2ce049988b78471b36453` found only a PR-body wording nit, which was fixed without changing code.
- GitHub checks on PR #141 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- No issue #139 follow-up remains.
- Stronger release-readiness claims still require the already-planned exact-candidate evidence refresh after the recent security-relevant internal changes.

---

## Caller input follow-ups landed - 2026-06-16 18:11 EDT

**Main:** `c00f66228e60`
**Actor:** Codex

**Summary**

PR #144 landed the first remaining caller-input follow-up slice for issues #124, #125, and #128: validation-order coverage now pins the caller-input tail precedence, the threat-model review focus uses role-local identity wording, and the CPace core plan matches the current normalized-input seam and error-ownership split.

**Completed**

- Merged PR #144 (`test: pin caller input validation precedence`) as `c00f66228e60e2fe194a252c18d58cf467a26459`; issues #124, #125, and #128 closed automatically at merge.
- Added `TestInputValidation` coverage for `Context` before local associated data and `SessionID` before local associated data, preserving behavior while making the existing private validation order executable.
- Updated `docs/threat-model.md` and `docs/cpace-core-plan.md` for current caller-input terminology, normalized input naming, and the `input.go` / `api.go` seam around caller-input validation, framing/state checks, and password backstops.
- RAS review-fix run `20260616T214143-eb23d12459547d874d1273f1` found the missing session-id precedence case and applied it as commit `f4aaa81578d7e8684cedce71b21e56d5d4ddb0a6`.
- RAS review-fix run `20260616T215502-4f1a9895e0e9655e3bc85d8f` found a changed-file seam wording issue, fixed in commit `3d3b466571ec3cee48d16c5114794d8d76c852f7`.
- Final RAS review-fix run `20260616T220425-d70f596785e1cc9d4bfbd101` reported no merge-blocking fixes; its one info-level reviewer-facing terminology consistency item became follow-up issue #145 and OmniFocus task `pE_6oM1HPky`.
- Completed OmniFocus tasks `aNp2T318MJz`, `o-CyFwzIIEr`, and `jfBymILxljz` with merge, validation, and follow-up evidence.

**Validation**

- Focused validation passed with `go test -run TestInputValidation ./...`.
- Mutation checks proved both new precedence cases fail when the corresponding private cap-check order is temporarily inverted, then pass again after restoring production order.
- Documentation and whitespace gates passed with `task docs:check` and `git diff --check`.
- Full local gate passed with `task check`; the release helper smoke test again reported the optional Syft validation skip because `syft` is not installed.
- GitHub checks on PR #144 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the phase-2 implementation sequence with #131, the private single-use terminal-state cleanup.
- Keep #145 separate from the implementation sequence unless a maintainer intentionally broadens reviewer-facing documentation cleanup.
- Stronger release-readiness claims still require the planned exact-candidate evidence refresh after these security-relevant changes.

---

## Single-use terminal cleanup landed - 2026-06-16 18:22 EDT

**Main:** `23eba7eb3275`
**Actor:** Codex

**Summary**

PR #147 completed issue #131 by trimming unreachable private defensive paths from the shared single-use terminal-state helper while adding direct coverage for the retained nil-core diagnostic branch.

**Completed**

- Merged PR #147 (`refactor: trim single-use terminal defenses`) as `23eba7eb3275cf073dd8f6e5c560d58ab88ca01c`; issue #131 closed automatically at merge.
- Added `TestSingleUseTerminalNilCoreReturnsUninitializedDiagnostic` for direct package-internal nil-core initiator and responder terminal states, covering both finish and close claims without consuming state.
- Removed unreachable private nil-receiver guards from `claimFinish` and `claimClose`; public `Initiator` and `Responder` methods still own nil and zero-value diagnostics before calling the helper.
- Removed the empty diagnostic fallback from `uninitializedError`, relying on the private constructor call sites that pass role-specific messages.
- RAS review-fix run `20260616T221621-1d59a9c0e3d6ba5aaea8d4de` reported no actionable or blocking findings; its only nit concerned an impossible package-internal nil-core plus empty-message diagnostic and did not warrant a follow-up issue.
- Completed OmniFocus task `acrb5smRKe9` with merge, validation, and RAS evidence.

**Validation**

- Focused verification passed with `go test -run 'TestSingleUse|TestZero|Test.*Uninitialized' ./...`.
- Full tests passed with `go test ./...`.
- Whitespace and full local gates passed with `git diff --check` and `task check`.
- GitHub checks on PR #147 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the phase-2 implementation sequence with #133, the release-helper newline tag hardening.
- Stronger release-readiness claims still require the planned exact-candidate evidence refresh after these internal lifecycle changes.

---

## Release helper newline hardening landed - 2026-06-16 18:34 EDT

**Main:** `09b82c4bdcd1`
**Actor:** Codex

**Summary**

PR #149 completed issue #133 by making release helper tag validation reject multiline scalar inputs before SemVer regex checks, preserving accepted tag syntax while preventing line-oriented matching from accepting only the first line.

**Completed**

- Merged PR #149 (`test: reject multiline release helper tags`) as `09b82c4bdcd19772fa2e7ff594399964dcfa54ab`; issue #133 closed automatically at merge.
- Added release helper smoke tests for multiline release-note tags, multiline release metadata tags, and newline-bearing SBOM filenames.
- Added early newline rejection to `scripts/extract-release-notes.sh`, `scripts/release-tag-metadata.sh`, and `scripts/validate-cyclonedx-sbom.sh` before their existing SemVer regex checks.
- RAS review-fix run `20260616T222801-771919b91d3e76aa58e5c2ff` reported no merge-blocking findings; its only low-severity maintainability item became follow-up issue #150 and OmniFocus task `iZmGBhcCHRi`.
- Completed OmniFocus task `nv9mcl6_6TF` with merge, validation, RAS, and follow-up evidence.

**Validation**

- The new smoke tests failed before the helper changes, then passed after the newline guards were added.
- Release helper validation passed with `scripts/test-release-helpers.sh`; the optional Syft validation skip remained expected because `syft` is not installed.
- Release policy tests passed with `(cd tools/releasepolicy && go test ./...)`.
- Whitespace and full local gates passed with `git diff --check` and `task check`.
- GitHub checks on PR #149 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the phase-2 implementation sequence with #136, the caller-input field-policy evaluation.
- Keep #150 separate from the implementation sequence unless a maintainer wants to deduplicate release-tag validation now.
- Stronger release-readiness claims still require the planned exact-candidate evidence refresh after these release-helper changes.

---

## Caller input field-policy evaluation recorded - 2026-06-16 18:45 EDT

**Main:** `c5013fc1fc96`
**Actor:** Codex

**Summary**

PR #152 completed issue #136 by recording the caller-input field-policy evaluation outcome: do not add a private field-policy catalogue now because the current `input.go` validation, copy, normalization, and wipe functions remain small and inspectable after the follow-up coverage work.

**Completed**

- Merged PR #152 (`docs: record caller input field policy evaluation`) as `c5013fc1fc96427e7985e0962b0e673ce5fdb325`; issue #136 closed automatically at merge.
- Added the evaluation result to `docs/project-plan.md` as later-investigation guidance: revisit concentration only if future caller-input changes create drift or a behavior-preserving simplification appears.
- Updated the touched project-plan integration-guidance row to current role-local identity terminology; follow-up #145 remains the separate reviewer-outreach cleanup.
- RAS review-fix run `20260616T223845-12a45e3b38718d7dbc42292f` reported no actionable findings.
- Completed OmniFocus task `i3nngrHasAi` with merge, validation, and RAS evidence.

**Validation**

- Docs validation passed with `task docs:check` and `git diff --check`.
- Targeted terminology inspection passed for the touched project-plan row; the only remaining stale reviewer-facing wording is tracked by #145.
- Full local gate passed with `task check`; the optional Syft validation skip remained expected because `syft` is not installed.
- GitHub checks on PR #152 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Complete the parent phase-2 implementation task audit and close the OmniFocus parent if no open child implementation tasks remain.
- Keep #145 and #150 as non-blocking follow-up tasks outside this implementation sequence.
- Stronger release-readiness claims still require the planned exact-candidate evidence refresh after the merged security-relevant changes.

---

## Message framing catalogue landed - 2026-06-17 20:09 EDT

**Main:** `84006defb969`
**Actor:** Codex

**Summary**

PR #154 moved Message framing facts into a small internal catalogue around `messageSpec`, keeping the public API, wire bytes, cap values, package-profile policy, and error identities unchanged.

**Completed**

- Merged PR #154 (`refactor: deepen message framing catalogue`) as `84006defb969012d13c32a3ca20b0c6b471ede10`.
- Added `messageFramingCatalogue()` and `messageSpec` encode/decode methods so production wrappers and tests use the same Message A/B/C role and field facts.
- Reworked framing catalogue tests, cap-policy tests, and fuzz seed helpers to derive malformed cases, field-limit cases, max-field messages, cross-role cases, and round-trip oracles from the catalogue instead of hand-coded A/B/C tables.
- Restored ordered field assertions after RAS review identified that cap-policy membership checks alone did not pin positional decoding semantics.
- Kept the fuzz round-trip length oracle independent from production `validateMessageLength`, and kept test helpers clean under the hosted SAST/gosec gate.
- Final RAS review-fix run `20260617T230313-08079563f71bc239c64a4eaf` reported no merge-blocking findings on final head `cae2a1db8be8d2a00ec5aecdb93ffe6a0587691f`; non-blocking test-helper follow-ups became issue #155 and OmniFocus task `b409eHqcIdr`.

**Validation**

- Focused tests passed for Message framing catalogue field limits, max fields, round trips, and cap-policy field order.
- Full local gates passed before merge: `go test ./...`, `go test -race ./...`, `task check`, `git diff --check`, and local `gosec -exclude-dir=.ras -tests -fmt=json ./...` with zero findings after the SAST cleanup.
- GitHub checks on PR #154 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Keep #155 as a non-blocking follow-up for Message framing catalogue test hardening.
- Stronger release-readiness claims still require refreshing pinned dependency, fuzz, and security-audit evidence against the exact candidate commit after these parser-adjacent changes.

---

## Release tag policy centralization landed - 2026-06-17 20:10 EDT

**Main:** `0f416ec81885`
**Actor:** Codex

**Summary**

PR #156 completed issue #150 by moving accepted release-tag validation into a shared shell policy module used by release metadata, release-note extraction, and CycloneDX SBOM filename validation.

**Completed**

- Merged PR #156 (`refactor: centralize release tag policy`) as `0f416ec81885e19a49a7a8c845bade760e6a2a7b`; issue #150 closed automatically at merge.
- Added `scripts/release-tag-policy.sh` with `release_tag_is_supported` and `release_tag_require_supported`, keeping accepted SemVer tag syntax and newline rejection in one maintained shell module.
- Updated `scripts/extract-release-notes.sh`, `scripts/release-tag-metadata.sh`, and `scripts/validate-cyclonedx-sbom.sh` to source the shared policy while preserving the SBOM helper's filename-specific diagnostic wording.
- Extended release-helper smoke tests for direct shared-policy acceptance/rejection and for sourced-helper scope behavior.
- Added `scripts/release-tag-policy.sh` to the accepted release policy checker’s required helper catalogue.
- Final RAS review-fix run `20260617T235842-22d699c1b4b4a989d0b0afa0` reported no merge-blocking findings on final head `7d4e60b1b42b049462f4f704b123c2acbc70840b`; non-blocking release-helper coverage hardening became issue #157 and OmniFocus task `lZktldDGmfi`.

**Validation**

- TDD red checks failed before implementation for the missing shared helper and missing accepted release-policy requirement, then passed after the helper and policy catalogue were updated.
- Release helper validation passed with `scripts/test-release-helpers.sh`; the optional Syft validation skip remained expected because `syft` is not installed.
- Release policy tests passed with `(cd tools/releasepolicy && go test ./...)`.
- Full local gates passed before merge: `go test ./...`, `task check`, `git diff --check`, and local `gosec -exclude-dir=.ras -tests -fmt=json ./...` with zero findings.
- GitHub checks on PR #156 passed before merge after rebasing onto PR #154's merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Keep #157 as a non-blocking follow-up for release-helper anti-drift and SBOM invalid-tag coverage.
- Stronger release-readiness claims still require refreshing pinned release evidence against the exact candidate commit after these release-tooling changes.

---

## Architecture depth pass landed - 2026-06-17 21:54 EDT

**Main:** `31e01e1cb08b`
**Actor:** Codex

**Summary**

PRs #159, #161, and #163 completed the no-freeze architecture depth pass by deepening Message framing, accepted release policy checking, and the transcript/confirmation flow internals while preserving the public API and wire-visible protocol behavior.

**Completed**

- Merged PR #159 (`Refactor message framing codec locality`) as `9a38be43fcd66d6d6632122ce1d79d279bde4195`.
- Moved Message A/B/C field decoding locality into the framing layer, added canonical LEB128 decode coverage, and pinned decoded-field ownership behavior.
- RAS review-fix implementation `20260618T004106-4d2ce28ff55bb26087832c69` / review run `20260618T004108-5a92e6e4e391e6c215d448f9` reported no merge-blocking findings; non-blocking framing nits became follow-up issue #160.
- Merged PR #161 (`Refactor release policy checker locality`) as `5acd4d9f270176f8b16745c6d5e9b621b7d398e6`.
- Added release-policy concept metadata, injected the accepted policy catalogue into the checker, split the accepted-job checks into smaller modules, and expanded catalogue integrity coverage.
- Final RAS review-fix implementation `20260618T012212-98ecb1721bf08874089ffafb` / review run `20260618T012213-e85a6de504e74b6c61d66f96` reported no merge-blocking findings under the low/nit gate; the remaining clone-helper coverage expansion became follow-up issue #162.
- Merged PR #163 (`Refactor transcript confirmation flow`) as `31e01e1cb08b968d02becfbc59ec9202fb28560e`.
- Added an unexported IR transcript object that owns transcript construction, ISK derivation, role-specific confirmation-tag selection, and transcript buffer ownership; moved OC transcript helpers to vector/test code only.
- RAS review-fix implementation `20260618T013829-36ec06c34f277897809ef63b` / review run `20260618T013830-8be66693014269ac68ac3c40` reported no merge-blocking findings; the low-severity spec-matrix traceability update became follow-up issue #164.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- PR #159 local gates passed before merge: `go test ./...`, `go test -race ./...`, `go vet ./...`, and `git diff --check`.
- PR #161 local gates passed before merge: `scripts/check-release-policy.sh`, `scripts/test-release-helpers.sh` with the optional Syft validation skip expected because `syft` is not installed, `(cd tools/releasepolicy && go test ./...)`, `(cd tools/releasepolicy && go vet ./...)`, `go test ./...`, and `git diff --check`.
- PR #163 local gates passed before merge: `go test ./...`, `go test -race ./...`, `go vet ./...`, and `git diff --check`.
- GitHub checks passed before each merge; PR #163 was `CLEAN` with CI Check, CodeQL Analyze/CodeQL, cross-platform smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck successful, and the standalone gosec child check neutral/skipping as expected.

**Next**

- Keep #160, #162, and #164 as non-blocking follow-up issues from the architecture review loops.
- Stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit after these parser-, policy-, and transcript-adjacent changes.

---

## Docs traceability follow-ups landed - 2026-06-17 23:51 EDT

**Main:** `870742d79903`
**Actor:** Codex

**Summary**

PR #166 closed documentation follow-ups #145 and #164 by aligning reviewer-facing role-local identity terminology and refreshing the spec-matrix transcript traceability after the IR transcript refactor.

**Completed**

- Merged PR #166 (`Docs traceability follow-ups`) as `870742d79903fcbf77d8f4d218435a6b7a123c55`.
- Updated `docs/reviewer-outreach.md` so the review focus now names role-local identity input consistently with `docs/threat-model.md` and `docs/project-plan.md`.
- Updated `docs/spec-matrix.md` so draft `transcript_ir` maps to `newIRTranscript` / `irTranscript` in `transcript.go`, with symmetric/OC transcript helpers documented as test-vector-only through `testTranscriptOC`.
- RAS review-fix implementation `20260618T034803-b5ac3e79c38edefe92f8be30` reported no actionable findings.
- GitHub issues #145 and #164 closed through the PR merge.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Traceability searches passed for `orientation|role-local identity` across reviewer outreach, threat model, and project plan.
- Transcript traceability search passed for `transcriptIR|strings.go|newIRTranscript|transcript.go|testTranscriptOC` across the spec matrix, transcript implementation, and string/vector tests.
- Docs validation passed with `task docs:check`.
- Whitespace validation passed with `git diff --check`.
- GitHub checks on PR #166 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the follow-up sequence with message framing issues #155 and #160.

---

## Message framing follow-ups landed - 2026-06-18 00:08 EDT

**Main:** `11f224ff0ff4`
**Actor:** Codex

**Summary**

PR #168 closed message framing follow-ups #155 and #160 by tightening catalogue-derived fuzz seed construction, pinning Message A/B/C role bytes in tests, guarding exact-length helper ambiguity, and documenting the `readLEB128` field-length ceiling contract.

**Completed**

- Merged PR #168 (`Harden message framing follow-up coverage`) as `11f224ff0ff4d8533f4a3543af53863ebb5ce230`.
- Added regression coverage proving Message A protocol fuzz seeds mutate decoded valid fields instead of rebuilding unrelated synthetic fields.
- Added an ambiguity guard for exact-length catalogue lookups used by test helpers.
- Extended package cap policy framing tests with literal role-byte pins for Message A/B/C.
- Documented the internal `readLEB128(maxBytes)` contract and changed `TestLEB128CanonicalDecode` to use an explicit offset table field instead of name-based control flow.
- RAS review-fix implementation `20260618T035833-84aa48641378d5fb36b4e4e4` / review run `20260618T035834-9c178b04a863fcecc6e186d7` reported no merge-blocking findings under the low/nit gate.
- Created follow-up issue #169 and OmniFocus task `pPUR0M7BoHk` for the low-severity RAS test-polish findings that were intentionally not allowed to block PR #168.
- GitHub issues #155 and #160 closed through the PR merge.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Focused framing tests passed for Message A fuzz seed preservation, exact-length ambiguity rejection, role-byte pins, and LEB128 canonical decoding.
- `FuzzRespondWithFuzzedMessageA/seed#4` passed with `go test -run '^FuzzRespondWithFuzzedMessageA/seed#4$' -covermode=count -coverprofile=/tmp/cpace-seed4.out .`.
- Full local gates passed with `go test ./...`, `go test -race ./...`, `task check`, and `git diff --check`; `task check` reported the expected optional Syft skip because `syft` is not installed.
- The role-byte mutation check failed as intended when the expected Message A role was temporarily changed to `0x02`, then passed after restoring the expected `0x01`.
- GitHub checks on PR #168 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Keep #169 as a non-blocking follow-up for deeper message framing fuzz-seed helper regression coverage.
- Continue the follow-up sequence with release policy issues #157 and #162.

---

## Release policy follow-ups landed - 2026-06-18 00:28 EDT

**Main:** `622a65cedee0`
**Actor:** Codex

**Summary**

PR #171 closed release-policy follow-ups #157 and #162 by hardening release-tag helper anti-drift coverage, adding SBOM filename rejection smoke tests for invalid SemVer-like tags, and expanding release-policy clone-helper deep-copy coverage.

**Completed**

- Merged PR #171 (`Harden release policy follow-up coverage`) as `622a65cedee04b96a0ac9123dcbf432ed60339c8`.
- Added release-helper anti-drift checks proving `scripts/extract-release-notes.sh`, `scripts/release-tag-metadata.sh`, and `scripts/validate-cyclonedx-sbom.sh` source `scripts/release-tag-policy.sh` and do not reintroduce local SemVer policy definitions.
- Strengthened sourced-helper namespace coverage for caller-relevant names: `release_tag`, `tag`, `version`, `major`, `prerelease`, and `latest`.
- Added SBOM smoke coverage for invalid supported-name-shape filenames `cpace-v01.0.0.cdx.json` and `cpace-v1.2.cdx.json`, both expecting the supported-release-tag diagnostic.
- Added the sourced-library intent comment to `scripts/release-tag-policy.sh` without changing helper execution policy.
- Expanded `TestCloneReleasePolicyIsDeep` to cover `env`, `concurrency`, `triggerKeys`, `pushKeys`, `pushTags`, `requiredScripts`, job permissions, and step env deep copies.
- Fixed the medium RAS finding by making clone-alias assertions restore global policy state before failing, so a deliberate clone regression does not pollute later release-policy tests.
- RAS review-fix implementation `20260618T041523-9dd7bd22c4d907ff115807d9` found one medium test-hygiene issue and one low coverage gap; the medium issue was fixed before merge.
- RAS review-fix implementation `20260618T042439-edbb238c470180d5164325ed` reran against head `a4cc501556fd5c47282b4a686eae8c0ffa675cdb` and reported no actionable findings.
- Created follow-up issue #172 and OmniFocus task `nLenCIJ4Bhj` for the low-severity function-shadowing anti-drift guard gap, intentionally not blocking PR #171 on that low finding.
- GitHub issues #157 and #162 closed through the PR merge.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Release helper smoke tests passed with `scripts/test-release-helpers.sh`; the optional real Syft SBOM validation skipped because `syft` is not installed.
- Release policy checker passed with `scripts/check-release-policy.sh`.
- Release-policy tool validation passed with `(cd tools/releasepolicy && go test ./...)` and `(cd tools/releasepolicy && go vet ./...)`.
- Full repo gate passed with `task check`; it included docs validation, release helper tests, CI classifier tests, evidence baseline checks, `go test ./...`, `go test -race ./...`, gofmt/goimports checks, `go vet ./...`, `staticcheck ./...`, `ast-grep scan --error`, and `govulncheck -test ./...`.
- Whitespace validation passed with `git diff --check`.
- Mutation evidence for #162 passed: temporarily removing `step.env = cloneStringMap(step.env)` made `TestCloneReleasePolicyIsDeep` fail with `step env map aliased`; after the medium-finding fix, `TestReleasePolicyStopsStepValidationAfterIdentityMismatch` still passed in the same mutated run, proving the alias failure no longer polluted later policy tests.
- GitHub checks on PR #171 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Keep #169 and #172 as non-blocking follow-up issues from the RAS low-severity review findings.
- Stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit if later work makes security-relevant changes.

---

## Message framing catalogue test deepening landed - 2026-06-18 01:03 EDT

**Main:** `0ae3b1e2b5fe`
**Actor:** Codex

**Summary**

PR #174 completed the first architecture-plan slice by concentrating Message framing catalogue behavior checks in `framing_catalogue_test.go` and removing duplicate helper-level LEB128 rejection tests now covered through message decoding.

**Completed**

- Merged PR #174 (`test: deepen message framing catalogue coverage`) as `0ae3b1e2b5fe4b4fe6b9fbbdd6365da68140670b`.
- Moved malformed, aggregate-size precedence, max-field, and field-limit catalogue checks out of `api_test.go` and into the Message framing catalogue test module.
- Removed the low-leverage `TestLEB128LengthInvariant` and direct malformed LEB128 rejection test after confirming malformed encodings remain covered through the Message framing decode path.
- RAS review-fix implementation `20260618T050013-9be9112e366fe821fe9ccc44` reported no actionable findings and no low/nit follow-ups.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Local gates passed before merge with `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; `task check` reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #174 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the architecture plan with the Transcript module PR.

---

## Transcript module owns TranscriptID derivation landed - 2026-06-18 01:27 EDT

**Main:** `b217a3aea6fd`
**Actor:** Codex

**Summary**

PR #176 completed the second architecture-plan slice by moving draft `CPaceSidOutput` derivation into the Transcript module and making `Session` store an already-derived transcript ID.

**Completed**

- Merged PR #176 (`Refactor transcript ID derivation into transcript module`) as `b217a3aea6fde2d94928dcbdc0cc705186a6c8b3`.
- Added `irTranscript.transcriptID()` and the shared `transcriptID` helper in `transcript.go`, so initiator and responder session construction use the same Transcript-owned derivation path.
- Updated `newSession` to receive and defensively clone a derived transcript ID instead of deriving `CPaceSidOutput` from transcript bytes.
- Kept the responder core's existing transcript-byte storage model intact; this PR changes the derivation boundary without widening ADR-0001 lifecycle behavior.
- Extended draft-vector coverage to assert helper-level and core-to-session `TranscriptID()` values against `sid_output_ir`, including initiator/responder session equality.
- Strengthened transcript copy-semantics coverage so mutating a returned transcript ID cannot affect later transcript ID derivations.
- Addressed the first RAS review-fix run `20260618T051006-25016dbff5efe8cae541d4dd` by adding core-to-session TranscriptID golden assertions and replacing the append-based SHA-512 input with streaming writes.
- RAS review-fix implementation `20260618T051736-44cec4f2740ee5bf2fae74a0` reran against head `5c45909cbf7166668c1732d668e42567059f471e` and reported no actionable findings and no low/nit follow-ups.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Focused local checks passed with `go test -run 'TestIRTranscript(DraftVectorFlow|OwnsInputsAndOutput)|TestRistrettoDraft21Vectors|TestCoreDraft21Vectors|TestConfirmedExchangeAndExport|TestSessionPeerMetadata' ./...`.
- Full local gates passed before merge with `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; `task check` reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #176 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the architecture plan with the caller-input lifetime handoff PR.

---

## Caller input handoff slice landed - 2026-06-18 01:48 EDT

**Main:** `3e9ec729710a`
**Actor:** Codex

**Summary**

PR #178 completed the third architecture-plan slice by making caller input an explicit secret-lifetime handoff boundary instead of an implicit validation/normalization flow.

**Completed**

- Merged PR #178 (`Make caller input handoff explicit`) as `3e9ec729710ad57a3a53d315cfa3a759cc62efeb`.
- Renamed the package-owned input clone state from `acceptedInput` to `callerInput` to reflect its ownership role.
- Added `callerInput.handoff`, which maps role-local IDs into transcript roles, builds CI, clears residual context storage, transfers package-owned slice headers into `normalizedInput`, and nils the transferred headers in `callerInput`.
- Simplified `normalizeInput` to install `defer caller.wipe()` immediately after accepting input, then return the handed-off normalized input.
- Added `caller_input_test.go` coverage for ownership transfer, context zeroization, source-reference niling after handoff, and responder role ID mapping.
- Updated `docs/adr-0009-secret-lifetime-audit.md` after the first RAS review found stale `acceptedInput` and `keep`-flag references.
- RAS review-fix implementation `20260618T053316-0a1476bea26385c93214392c` fixed the ADR audit finding and reran cleanly against head `d9073817fdb34f5033beed9511e52385d64c9a99`, with no actionable findings and no tracked low/nit follow-ups.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- The initial red TDD check failed as expected with `caller.handoff undefined`.
- Focused local checks passed with `go test -run 'TestCallerInputHandoff|TestMutableInputsAreCopied|TestInputErrorPathsDoNotMutateCallerSlices|TestPackageOwnedCapPolicyAcceptsInputCopies|TestPackageOwnedCapPolicyRejectsInputBeforeCopying' ./...`.
- Stale audit references were checked with `rg -n "acceptedInput|keep" docs/adr-0009-secret-lifetime-audit.md`, which returned no matches.
- Full local gates passed before merge with `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; `task check` reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #178 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue the architecture plan with the release metadata module PR.

---

## Release metadata module slice landed - 2026-06-18 02:56 EDT

**Main:** `93a2557a772f`
**Actor:** Codex

**Summary**

PR #180 completed the fourth architecture-plan slice by extracting release tag metadata derivation into a shell module while keeping `scripts/release-tag-metadata.sh` as the release-workflow adapter.

**Completed**

- Merged PR #180 (`Add release metadata helper module`) as `93a2557a772f14767b0718010aca37bbdc98f13f`.
- Added `scripts/release-metadata.sh`, a sourced shell module that owns release tag metadata derivation for `release-tag`, `sbom-file`, `prerelease`, and `latest`.
- Kept `scripts/release-tag-metadata.sh` as the primary adapter used by the release workflow before Go setup, now sourcing `release-tag-policy.sh` and `release-metadata.sh`.
- Registered `scripts/release-metadata.sh` in the release-policy required-script catalogue and kept it executable.
- Added release-helper smoke coverage for direct module output, direct invalid-tag rejection, adapter anti-drift, caller namespace preservation, missing-policy rejection, spoofed-marker rejection, and unset-hook PATH shadowing.
- Hardened the metadata module after RAS review by replacing PATH-spoofable dependency checks with a policy-owned validation hook, then corrected the tests to target `release_tag_policy_require_supported_for_metadata` and `release_tag_policy_metadata_check_ran`.
- RAS review-fix implementations `20260618T055417-f9a0d6a384db90535f8e58e2`, `20260618T060801-308478779e7948ae9935e112`, and `20260618T063629-69515f3b9034434467e0cdc8` drove the guard hardening and test correction.
- Final RAS review-fix implementation `20260618T064538-97b2d093a5dad197e6aeb904` reran against head `a8681348b7c768b029a6d05f91f1e396e4d8cb96` and reported no open actionable findings or follow-ups.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- The initial red TDD check failed as expected because `scripts/release-tag-metadata.sh` did not source `scripts/release-metadata.sh`.
- Targeted checks passed with `scripts/test-release-helpers.sh`, `scripts/check-release-policy.sh`, and `(cd tools/releasepolicy && go test ./...)`; helper tests reported the expected optional Syft skip because `syft` is not installed.
- The release metadata module was checked across available shells with `/bin/sh`, `bash`, `dash`, `ksh`, and `zsh` where installed.
- A focused spoof check confirmed a PATH stub for `release_tag_policy_require_supported_for_metadata` plus a spoofed marker cannot emit metadata without the sourced policy module.
- Full local gates passed before merge with `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`.
- GitHub checks on PR #180 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- No follow-up issues were created from PR #180; all RAS findings were fixed or confirmed stale before merge.
- Stronger release-readiness claims still require refreshing pinned dependency-review, fuzz, and security-audit evidence against the exact candidate commit if maintainers treat the release-helper hardening as security-relevant evidence scope.

---

## Message fuzz seed coverage follow-up landed - 2026-06-18 11:26 EDT

**Main:** `3ac1d23978cd`
**Actor:** Codex

**Summary**

PR #182 closed issue #169 by tightening message-framing fuzz-seed regression coverage around field-preservation helpers and exact-length field selection.

**Completed**

- Merged PR #182 (`Tighten message fuzz seed coverage`) as `3ac1d23978cd4bda17ac0005dd55c5c7cad74c1f`.
- Made `TestMessageAProtocolFuzzSeedsPreserveValidFields` classify every successful decode as valid, identity-point, invalid-point, or other-session-ID so unclassified successful decodes fail instead of passing silently.
- Added explicit red-path coverage proving Message A preservation tests reject helpers that decode successfully after losing the intended seed mutation.
- Added Message B preservation coverage for `messageWithDecodedField` and `messageFuzzSeeds`, including valid, identity-point, invalid-point, and tampered-tag categories derived from the actual generated seed path.
- Split exact-length field helper behavior so absent exact-length fields are skipped while ambiguous exact-length fields panic with a diagnostic, and covered both paths.
- RAS review-fix run `20260618T145515-2ce7aa49c140156e3f4790a9` strengthened the absent exact-field regression test; follow-up run `20260618T151605-f8ed23198601ab03aae3d6a3` reran against the final head and reported no actionable required fixes.
- No follow-up issues were created from PR #182; remaining RAS notes were low/nit or out of scope for the accepted issue.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Initial red TDD checks failed as expected while Message A classification, ambiguous exact-field rejection, and Message B preservation coverage were absent.
- Mutation checks failed as expected when `messageWithDecodedField` was temporarily changed to drop its mutation and when Message B fuzz seed construction was temporarily regressed to drop associated data.
- Focused local checks passed for the new framing catalogue tests.
- Full local gates passed before merge with `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; `task check` reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #182 passed before merge: Check, CodeQL Analyze/CodeQL, macOS smoke, Windows smoke, DCO, Dependency Gate, SAST Gate, and Staticcheck; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Continue with issue #172 in a separate implementation branch.

---

## Release helper shadow guard follow-up landed - 2026-06-18 12:44 EDT

**Main:** `f6639a1efdb1`
**Actor:** Codex

**Summary**

PR #184 closed issue #172 by tightening release-helper anti-drift coverage so helpers cannot pass the policy reuse test while defining local `release_tag_*` policy-shadow functions.

**Completed**

- Merged PR #184 (`test: reject release helper policy shadowing`) as `f6639a1efdb125a1960ce7a267828706108312fa`.
- Split the release-tag policy reuse assertion into path-level helpers so generated shadow fixtures and real helper files share the same static guard.
- Added generated shadow fixtures for `release_tag_is_supported`, `release_tag_require_supported`, `release_tag_policy_is_supported`, and `release_tag_policy_require_supported_for_metadata`.
- Broadened direct release-tag policy function shadow detection to cover whitespace inside empty parentheses, commented/split-line function headers, and `function release_tag_*` declarations.
- Applied the reusable `release_tag_*` direct-definition scan to `scripts/release-metadata.sh` as well as top-level release helpers, because the adapter sources release metadata after the shared policy module.
- Improved shadow-fixture failure diagnostics so unexpected rejection output is printed instead of failing through a bare `grep`.
- RAS review-fix run `20260618T153154-81e1ce55ce259f8bfcf6825c` found the first namespace breadth issue; run `20260618T154206-02d438c35d8b4c493a69a58d` drove the broader guard hardening through additional RAS builder iterations.
- Final RAS review run `20260618T161625-49d0bda9b5e9bf6db9a92d5d` left only non-blocking low/nit follow-ups under the configured gate policy.
- Created follow-up issue #185 for operator-prefixed `release_tag_*` definitions and source-line matcher drift, and follow-up issue #186 for the adjacent `release_metadata_*` namespace-shadow symmetry question.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Baseline `scripts/test-release-helpers.sh` passed before the initial red edit.
- The initial red TDD check failed as expected with `scripts/release-tag-metadata.sh unexpectedly allowed local release_tag_is_supported definition`.
- The internal-policy red check failed as expected with `scripts/release-tag-metadata.sh unexpectedly allowed local release_tag_policy_is_supported definition` before the namespace guard was broadened.
- Mutation checks failed as expected when local `release_tag_is_supported()` and `release_tag_policy_is_supported()` definitions were temporarily injected into `scripts/release-tag-metadata.sh`.
- RAS builder mutation checks covered `scripts/release-metadata.sh` policy-shadow injection and source-line anchor diagnostics before its commits were accepted.
- Final local gates passed with `scripts/test-release-helpers.sh`, `scripts/check-release-policy.sh`, `(cd tools/releasepolicy && go test ./...)`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; helper tests reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #184 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Track follow-up issues #185 and #186 in OmniFocus with their RAS provenance.

---

## Release tag shadow prefix guard follow-up landed - 2026-06-18 13:31 EDT

**Main:** `0c5cb8c2c2fb`
**Actor:** Codex

**Summary**

PR #188 closed issue #185 by tightening the release-tag helper anti-drift guard for operator-prefixed local policy function definitions and by making the generated shadow-fixture injector robust to reformatted policy-source lines.

**Completed**

- Merged PR #188 (`test: catch prefixed release tag shadows`) as `0c5cb8c2c2fbfe395875bbc981d40d8bb389a88f`.
- Broadened the `release_tag_*` direct-definition scan so it catches parse-visible definitions after shell separators and operators, covering the `true && release_tag_is_supported() { return 0; }` form from issue #185.
- Split the generated shadow-fixture injector into a path-level helper so test fixtures can run against generated helper copies as well as repository files.
- Updated shadow-fixture injection to use the same substring policy-source matcher as the reuse check, so leading whitespace or trailing comments on `. "$script_dir/release-tag-policy.sh"` do not prevent injection.
- Added regression coverage for operator-prefixed release-tag policy shadows and reformatted policy-source lines.
- RAS review-fix run `20260618T172107-071f4b709001d1615230d3eb` completed with no merge-blocking findings under the configured gate policy.
- Created follow-up issue #189 for the non-blocking low-severity RAS findings around subshell-prefixed shadows, operator-prefixed `function` keyword coverage, and comment-only false positives.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- Baseline `scripts/test-release-helpers.sh` passed before the first red edit.
- The operator-prefixed red fixture failed as expected with `scripts/release-tag-metadata.sh unexpectedly allowed local release_tag_is_supported definition: true && release_tag_is_supported() { return 0; }`.
- The reformatted source-line red fixture failed as expected with `injection anchor not found in scripts/release-tag-metadata.sh with reformatted policy source`.
- Final local gates passed with `scripts/test-release-helpers.sh`, `scripts/check-release-policy.sh`, `(cd tools/releasepolicy && go test ./...)`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; helper tests reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #188 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Track follow-up issue #189 in OmniFocus as a non-blocking continuation of the release-tag shadow guard hardening.

---

## Release metadata shadow guard follow-up landed - 2026-06-18 13:47 EDT

**Main:** `6c4a149495eb`
**Actor:** Codex

**Summary**

PR #191 closed issue #186 by adding release metadata namespace shadow coverage to the release-helper anti-drift tests, keeping `scripts/release-metadata.sh` as the intentional `release_metadata_*` namespace owner.

**Completed**

- Merged PR #191 (`test: reject release metadata helper shadows`) as `6c4a149495eb598e7cce27bc86069fa0131dac06`.
- Factored the release-helper direct-function scanner so the same helper checks both `release_tag_*` and `release_metadata_*` local definitions.
- Applied the `release_metadata_*` direct-definition guard to helpers that source `scripts/release-metadata.sh`, while excluding `scripts/release-metadata.sh` itself because it intentionally defines the metadata namespace.
- Added a generated shadow fixture proving a helper-local `release_metadata_write() { return 0; }` definition is rejected by the metadata module reuse check.
- RAS review-fix run `20260618T173658-983368fd154a7de8050e8f54` completed with no merge-blocking findings under the configured gate policy.
- Created follow-up issue #192 for non-blocking metadata fixture parity around path-based injection, reformatted metadata source lines, and metadata form matrices; shared scanner delimiter polish remains coordinated with issue #189.
- No RAS run was performed for this journal-only update, per instruction.

**Validation**

- The initial red TDD check failed as expected with `scripts/release-tag-metadata.sh unexpectedly allowed local release_metadata_write definition: release_metadata_write() { return 0; }`.
- The mutation probe failed as expected when `release_metadata_write() { return 0; }` was temporarily added to `scripts/release-tag-metadata.sh`, producing `scripts/release-tag-metadata.sh defines a local release metadata function`.
- Final local gates passed with `scripts/test-release-helpers.sh`, `scripts/check-release-policy.sh`, `(cd tools/releasepolicy && go test ./...)`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `task check`, and `git diff --check`; helper tests reported the expected optional Syft skip because `syft` is not installed.
- GitHub checks on PR #191 passed before merge: Check, DCO, Dependency Gate, and SAST Gate; the standalone gosec child check was neutral/skipping as expected.

**Next**

- Track follow-up issue #192 in OmniFocus as a non-blocking metadata shadow-fixture parity continuation.

---

## Message framing oracle deepened - 2026-06-18 16:58 EDT

**Main:** `c67d60aa15cb`
**Actor:** Codex

### Summary

Deepened the Message framing acceptance oracle by moving field-length catalogue checks into shared test helpers while keeping round-trip fuzz acceptance intentionally independent from decoder validation.

### Completed

- Merged PR #194, `refactor: deepen message framing acceptance oracle`, at `c67d60aa15cbf6310b214f94fa8bfca374df6fba`.
- Added catalogue coverage for valid maximum fields, too few fields, extra fields, shortened exact fields, overlong exact fields, and over-cap variable fields.
- Kept fuzz round-trip acceptance on `messageFieldsMatchFramingShape` so the fuzz oracle can detect decoder acceptance drift instead of delegating to the decoder-facing predicate.
- Kept acceptance helpers test-only after RAS and SAST rejected the earlier production-helper shape.

### Validation

- `go test ./... -run 'TestMessageFramingCatalogueOwnsFieldLengthAcceptance|TestMessage|FuzzMessage'`
- `go test ./...`
- `go test ./... -count=1`
- `go run github.com/securego/gosec/v2/cmd/gosec@v2.26.1 -exclude-dir=.ras -tests ./...`
- `git diff --check`
- GitHub checks for PR #194 passed on head `abaa6ee4df9d1491a19390e50308a5bfcfc64ec3`.
- `ras review-fix 194` found and fixed the initial SAST/test-oracle issues, then completed on the final head with no merge-blocking code findings; the only final note was an out-of-scope reviewer timeout.

---

## Transcript deepened across responder construct/finish - 2026-06-18 18:58 EDT

**Main:** `133ba2c619ff`
**Actor:** Claude Code

### Summary

Deepened the Transcript so the responder stores and reuses the `irTranscript` value built at construction through `Finish`, closing the leak where the responder kept decomposed `ya`/`ada`/raw-transcript fields and re-derived the initiator confirmation tag via a free function. Behaviour-preserving: wire bytes, ISK, and confirmation tags are byte-for-byte identical, proven by the draft-21 vectors.

### Completed

- Merged PR #196, `refactor: deepen transcript to span responder construct→finish`, as `133ba2c619ffb32166931ca82253513a42cc083c` (squash).
- Reshaped `responderCore` to `{ isk, transcript irTranscript, sid, peerID }`, dropping the decomposed `ya`/`ada` fields and the raw `transcript []byte`.
- Routed `responderCore.finish` through the stored transcript for the initiator confirmation tag, transcript id, and peer associated data.
- Added `irTranscript.initiatorAD()` (peer-AD accessor) and `irTranscript.clear()` (pointer-receiver hygiene wipe, idempotent and nil-safe).
- Folded `initiatorRoleConfirmationTag` / `responderRoleConfirmationTag` into the transcript methods and deleted the free functions; each had no independent test and a single caller after the responder bypass was removed.
- Sharpened the `CONTEXT.md` Transcript glossary entry to record the module's ownership and the construct→finish carry.
- RAS review-fix run `20260618T224620-3b50c9ad29c8e742899999e2` completed `done` with no merge-blocking findings under the configured gate policy.
- Created follow-up issue #197 for the two non-blocking low/nit test-coverage gaps on `irTranscript.clear()` (component-field wipe assertions and nil-receiver coverage).
- No RAS run was performed for this journal-only update, per instruction.

### Validation

- TDD red checks failed as expected with build failures: `tr.initiatorAD undefined` and `tr.clear undefined`.
- Targeted gates passed: `go test -run '^(TestClearIdempotent|TestCoreDraft21Vectors|TestNilReceiverFinishAndExport|TestFinishCleanupDoesNotAliasReturnedSessions|TestSingleUseTerminalClaimsDoNotReturnCoreOnLosingPaths)$' .`; the `TestCoreDraft21Vectors` pass is the byte-for-byte identity proof.
- Final local gates passed: `go test ./...` and `go vet ./...`.
- GitHub checks on PR #196 passed before merge: CodeQL, Analyze, Check, DCO, Dependency Gate, SAST Gate, Staticcheck, macos-latest, windows-latest; the standalone gosec child check was neutral/skipping as expected.

### Next

- Track follow-up issue #197 in OmniFocus as a non-blocking `irTranscript.clear()` test-hardening continuation.

---

## Go fix modernization landed - 2026-06-18 23:24 EDT

**Main:** `942ed0448b08`
**Actor:** Codex

### Summary

Merged PR #199 after applying the Go 1.26.4 `go fix` modernization and tightening the release SBOM configuration guard that RAS found around the new Syft config file.

### Completed

- Merged PR #199, `chore: apply go fix suggestions`, as merge commit `942ed0448b088ab0501c961f193f26492427f58a`.
- Applied the low-risk `go fix` suggestions for the framing catalogue tests and the nested release/evidence helper modules: removed the stale range-variable copy, used newer formatting/allocation helpers, and replaced manual map copy/string-split patterns with standard-library helpers.
- Added `.github/syft-release.yaml` with explicit `source.name: github.com/the-sarge/cpace` and shared Syft excludes for `./.git/**` and `./.ras/**`.
- Routed both Release Validation and the local real-Syft smoke through the shared Syft config instead of carrying local-only exclude flags.
- Extended the release-policy checker so `.github/syft-release.yaml` is a required non-executable release config, rejecting missing, symlinked, directory, and non-regular paths before validating `source.name` and the shared exclude sequence.
- Ran RAS review `20260619T023843-85d672b9aee9841f5667ad9b`; the actionable findings were the missing config-file guard and local/CI Syft config drift.
- Fixed the first RAS verification edge case by requiring regular files, then reran RAS verification on PR head `27cc0e9ab2c994781d7b479cc168f6cc565eb9af`; verification reported no still-open findings and no new concerns.

### Validation

- `(cd tools/releasepolicy && go test ./...)`
- `task release:helpers`
- `task check`
- GitHub checks on PR #199 passed before merge: Actionlint, Check, CodeQL, DCO, Dependency Gate, SAST Gate, Staticcheck, macos-latest, and windows-latest; the standalone gosec child check was neutral as expected.
- RAS verification confirmed the Syft config drift and release-policy config guard findings were resolved on the merged head.

### Next

- Refresh exact-candidate release evidence on post-merge `main` before making stronger release-readiness claims, because this release-policy/SBOM work changes evidence-relevant release tooling after the prior pinned evidence baseline.
