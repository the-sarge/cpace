# Code Conventions

These conventions record the repository's established implementation dialect so maintenance sweeps and new work apply the same rules. They do not reopen the frozen public API, package profile, wire format, observable behavior, or error strings.

## Go Construction And Panics

- Name an unexported constructor `newX`, where `X` identifies the type or concept it constructs; use a different verb when a function performs an operation rather than construction.
- Prefix every production panic message with `cpace: ` so an invariant failure is attributable to the package.
- Prefix panics raised by test fixtures, catalogues, and helpers with `cpace test: ` so test-infrastructure failures are distinguishable from production invariant failures.
- Keep the conventional plain `panic(err)` idiom in godoc `Example` functions, which demonstrate caller behavior rather than test-infrastructure invariants.

## Tests

- Write expected-value failures in the comma-free form `<subject> got X want Y`.
- Use `t.Fatalf` as the standard assertion dialect so a failed prerequisite or comparison stops the test immediately; use a non-fatal assertion only when the test intentionally collects independent discrepancies and later checks remain meaningful.
- Name a helper `mustX` when it loads or constructs a required fixture and terminates the test on failure.
- Name a helper `assertX` when it checks a condition or comparison for the caller.
- Name a helper `snapshotX` when it captures hygiene-relevant state that a later assertion will inspect after a terminal operation.

## Shell

- Write files under `scripts/` that are run as commands as strict POSIX shell: start with `#!/bin/sh`, immediately enable `set -eu`, and use POSIX syntax rather than Bash-only features.
- Treat files under `scripts/` that are sourced for definitions as sourced libraries regardless of executable mode: start them with `#!/bin/sh`, keep them side-effect-free, and namespace their definitions for the caller; they inherit strict mode from the command script instead of changing the caller's shell options.
- Start every Taskfile-embedded Bash block with `set -euo pipefail` before performing work.

## Validation Tool Packages

Give each Go module under `tools/` a package comment that names its module using the exact terminology from `CONTEXT.md` and states its validation-only boundary.

- `tools/releasepolicy` names the **Release policy checker** and explains that it validates accepted release-pipeline policy without generating workflows or querying live GitHub state.
- `tools/evidencebaseline` names the **Evidence baseline** and explains that it validates committed evidence references and freshness metadata without refreshing evidence or making a production-readiness claim; maintaining its derived summary-doc manifest remains within that boundary.
