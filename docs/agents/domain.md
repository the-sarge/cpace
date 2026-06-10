# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

This repo is **single-context**: one `CONTEXT.md` plus `docs/adr/` at the repo root cover the whole project. cpace is a single flat Go package implementing one CPace suite, so there is one domain vocabulary.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root — the project's domain glossary.
- **`docs/adr/`** — read ADRs that touch the area you're about to work in.

Both exist. (If a future split ever removes one, **proceed silently** — don't flag the absence; the producer skill `/grill-with-docs` recreates them lazily when terms or decisions actually get resolved.)

## File structure

Single-context layout (this repo):

```
/
├── CONTEXT.md            ← domain glossary
├── docs/adr/             ← architecture decision records — read the live
│                            directory; do not rely on a hardcoded list
└── *.go                  ← flat package at the repo root
```

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in `CONTEXT.md`. Don't drift to synonyms the glossary explicitly avoids.

If the concept you need isn't in the glossary yet, that's a signal — either you're inventing language the project doesn't use (reconsider) or there's a real gap (note it for `/grill-with-docs`).

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-0001 (extract a deep CPace core) — but worth reopening because…_
