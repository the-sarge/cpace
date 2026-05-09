# Evidence Artifacts

This directory stores raw evidence for release candidates, toolchain-security
refreshes, and other review packets that need more than a handwritten summary.
Evidence summaries in `docs/*.md` should point to these artifacts or to
immutable GitHub workflow runs.

## Evidence Modes

Use one or both of these modes for exact-candidate evidence:

- Immutable GitHub Actions or manual workflow artifacts tied to the exact
  commit under review.
- Committed transcript bundles under `docs/evidence/<candidate-or-date>/` with
  `SHA256SUMS`.

For maintainer-machine evidence, prefer committed raw transcripts plus
checksums. Add detached signatures for `SHA256SUMS` when practical.

## Workflow-Artifact Evidence

When using workflow artifacts or workflow runs as evidence, summary docs should
record:

- workflow file path;
- workflow run ID and URL;
- exact target commit SHA;
- triggering event and ref;
- expected jobs or checks;
- artifact retention period or permanence limitation;
- summary document that records the workflow link;
- fallback evidence if the workflow run or artifact is later unavailable.

If the workflow run itself is the durable evidence, state which logs, check
conclusions, or uploaded artifacts are being cited.

## Transcript Bundle Shape

Each committed bundle should include:

- `README.md`, describing the candidate commit, purpose, files, verification
  command, and residual limitations.
- Raw logs or transcripts for the commands used as evidence.
- `SHA256SUMS`, covering every raw transcript in the bundle.
- `SHA256SUMS.sig`, when detached signing is practical.

When detached signatures are used, prefer the signer identity documented in
`docs/release-verification.md`. Name the signing tool, signer identity, and key
fingerprint in the bundle README, or document the exception.

Raw logs should preserve:

- exact commit SHA and clean or dirty worktree status;
- host, operating system, architecture, Go version, and relevant tool versions;
- exact commands;
- UTC start and end timestamps for long-running commands;
- raw command output rather than only a handwritten summary;
- return code for wrapped commands.

## Verification

Verify committed transcript bundles before citing them from summary docs.

On macOS:

```sh
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
sha256sum -c SHA256SUMS
```

If a detached signature is present, verify it with the signing tool named in
the bundle README before tagging or publishing a release.

## Toolchain Vector Stability

When a Go toolchain update triggers an evidence refresh, run the draft/RFC
vector lane under both the old and new toolchains when both are available:

```sh
go test -v -run 'Test(StringUtilitiesDraftVectors|EmbeddedDraft.*|RistrettoDraft21Vectors|ScalarMultVFYDraftInvalidVectors)$' -count=1 ./...
```

The `-v` flag is intentional so the transcript enumerates each matched vector
test by name. Preserve raw logs for both runs and record whether the results are
bit-identical across toolchains. If the previous toolchain is unavailable, say
that explicitly in the workflow run summary or bundle README, depending on the
evidence mode, and under `Toolchain Vector Stability` in
`docs/security-spec-audit.md`; do not imply cross-toolchain identity from
current-toolchain validation alone.
