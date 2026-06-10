# Agent instructions — cpace

## Repo constraints every agent must know

- **Release-readiness freeze**: the public API and package-profile policy are frozen. Do not propose API or observable-behavior changes unless a review finding forces one; such a change is a *policy reopen* and needs an explicit maintainer decision first.
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
