# Issue #308 reflection validation

Source candidate: `d2db6eed5923395bf0dfb57e73275ea0a1506c44` (`fix: reject reflected confirmation tags in both roles (#308)`). Base: `87528342d68d26db4a5a69386ae5647baf7d05ed`. Validation ran from a clean dedicated worktree on 2026-09-23 under Go 1.27.1 on Darwin/arm64. This bundle and the link from the analysis are a documentation-only descendant of that source candidate; the candidate contains all implementation, tests, and security-conclusion changes.

## Results

| Command | Result |
| --- | --- |
| `go test -run '^(TestReflection\|TestRistrettoDraft21Vectors\|TestIRTranscriptDraftVectorFlow)' -count=1 -v .` | PASS; 22 reflection subtests, IR transcript flow, and draft-21 vector test with its six generator sid mutation subtests |
| `task --silent check` | PASS; full normal/race suites, Markdown/whitespace, release helper/policy checks, CI classifier/tool/fuzz-input checks, evidence checks, formatting/imports, pinned GolangCI-Lint, ast-grep, and govulncheck |
| `go mod verify` | PASS; all modules verified |

`validation.log` preserves the exact commands, candidate SHA, worktree state, UTC command-start timestamps, environment, raw output, and return codes. `--silent` suppresses Task command echo only; check output is preserved. Go's build/test cache was not globally disabled; the focused regression/vector command and the evidence checker explicitly use `-count=1`, and the root normal/race results in this transcript are uncached (no cached marker). The definition of every constituent `task check` command is pinned by the candidate's `Taskfile.yml`.

Verify the raw transcript from this directory with:

```sh
shasum -a 256 -c SHA256SUMS
```

No detached signature is attached to this local implementation-validation packet; it is not a signed release packet.

## Review

The code-review skill ran independent read-only Standards and Spec agents against the pinned diff `87528342d68d26db4a5a69386ae5647baf7d05ed...d2db6eed5923395bf0dfb57e73275ea0a1506c44`.

- Standards: no material actionable findings. The checks preserve the CPace core boundary and terminal cleanup, use `hmac.Equal`, and add no persistent secret fields. Tests follow helper, assertion, and snapshot conventions; documentation records the narrow finding-driven exception and evidence limits. No actionable baseline smells were reported.
- Spec: no findings against issue #308 and its triage brief. Both roles, the AD matrix, correctly framed authentic-tag reflection, forced equality, honest controls, replay, nil outputs, consumption, and cleanup were covered. The review found no scope creep or incorrect implementation. Its pending validation-capture caveat is satisfied by this bundle.

## Limitations

The [analysis](../../key-confirmation-reflection.md) bounds the security conclusion: the forced equal-message false confirmation was reproduced, but no feasible way for a network attacker to force that condition with fresh production randomness was established. Randomized regression observations do not estimate collision probability, and replay tests do not exhaust concurrent schedules.

These checks and automated reviews are not independent cryptographic review or a refreshed release-evidence packet. Historical dependency/SAST, Capslock, security/spec/vector, and paired long-fuzz evidence remains pinned to its older candidates. Refresh the affected lanes against the exact selected release candidate before any stronger release claim, as recorded in the [evidence baseline](../../evidence-baseline.md). No hosted CI, cross-platform campaign, paired long-fuzz run, or new Capslock capture is claimed here.
