# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

This repo is **single-context**: one `CONTEXT.md` plus `docs/adr/` at the repo root cover the whole project. The repository contains the root cpace library module and two nested validation-tool modules, but they share one domain vocabulary and one decision-record set.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root — the project's domain glossary.
- **`docs/adr/`** — read ADRs that touch the area you're about to work in.

Both exist. (If a future split ever removes one, **proceed silently** — don't flag the absence; the producer skill `/grill-with-docs` recreates them lazily when terms or decisions actually get resolved.)

## File structure

Single-context layout (this repo):

```
/
├── CONTEXT.md                      ← shared domain glossary
├── docs/adr/                       ← shared architecture decision records; read the live directory
├── go.mod                          ← root cpace library module
└── tools/
    ├── evidencebaseline/go.mod     ← evidence validation-tool module
    └── releasepolicy/go.mod        ← release-policy validation-tool module
```

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in `CONTEXT.md`. Don't drift to synonyms the glossary explicitly avoids.

If the concept you need isn't in the glossary yet, that's a signal — either you're inventing language the project doesn't use (reconsider) or there's a real gap (note it for `/grill-with-docs`).

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-0001 (extract a deep CPace core) — but worth reopening because…_
