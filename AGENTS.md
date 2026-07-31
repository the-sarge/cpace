# Agent instructions — cpace

## Repo constraints every agent must know

- **Release-readiness freeze**: the public API and package-profile policy are frozen except accepted ADR exceptions. Do not propose API or observable-behavior changes unless a review finding forces one or an accepted ADR explicitly authorizes it; outside accepted exceptions, such a change is a *policy reopen* and needs an explicit maintainer decision first. ADR-0009 currently authorizes only the follow-up caller-input `Input` implementation and leaves unrelated public API and package-profile choices frozen.
- **Rapid deferred**: `pgregory.net/rapid` was evaluated for property-based testing and intentionally not adopted during the release-readiness freeze. Its MPL-2.0 license is compatible with test-only use in principle, but it is outside the Dependency Gate's current license allowlist, and adding any dependency would require refreshed exact-candidate dependency, Capslock, security/spec, and paired long-fuzz evidence. Do not add Rapid unless a concrete review finding justifies it and the maintainer explicitly approves both the dependency-policy change and the evidence-refresh cost.
- **Evidence discipline**: any security-relevant change invalidates the pinned dependency-review / fuzz / security-audit evidence (each pinned to a commit). Stronger release claims require refreshing that evidence against the exact candidate commit.
- **ADR gating**: ADRs start `status: proposed` and flip to `accepted` only after an independent multi-agent review (`ras consider`) concurs. Decisions live in `docs/adr/`.
- **Merges are the maintainer's**: never merge or close a PR, or push to `main`, without an explicit per-action instruction.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for `the-sarge/cpace`, managed via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The live taxonomy is dimensional (`priority/*`, `kind/*`, `area/*`, plus `release blocker`, `external review`, `security`, `wontfix`), not workflow-state. See `docs/agents/triage-labels.md` for how the skills' canonical triage roles map onto it — do not create new labels.

### Domain docs

Single-context: one `CONTEXT.md` (domain glossary) + `docs/adr/` (decision records) at the repo root. See `docs/agents/domain.md`.
