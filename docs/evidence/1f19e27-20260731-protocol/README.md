# CPace v0.1.3 Toolchain, Vector, Capability, and Protocol Evidence

Candidate commit: `1f19e278112fa037890848ed6c086addeffdca4e`

Capture interval: 2026-07-31 08:15:24–08:15:32 UTC

Candidate state: clean detached worktree before and after capture

Host and platform: `mbp128.local`, macOS 26.6 (25G72), Darwin 25.6.0, `arm64`

Toolchains and analysis tools: Go 1.26.4, Go 1.26.5, `capslock@v0.3.2` built with Go 1.26.5 and Go tools v0.43.0, jq 1.7.1-apple, and Git 2.50.1.

This bundle records the Go 1.26.4 to Go 1.26.5 draft/package-vector comparison, exact-candidate Capslock capability analysis, public API inventory, post-historical-audit source-change inventory, and focused protocol/profile regression tests for the frozen `v0.1.3` candidate. It is project-side prerelease evidence, not an independent cryptographic review or a production-readiness claim.

## Contents

| File | Description |
| --- | --- |
| `capture.sh` | Reproduction script pinned to the candidate, historical audit baseline, toolchains, Capslock version, vector lane, and protocol audit lane; it refuses a mismatched or dirty candidate worktree and rejects empty test selections. |
| `protocol-audit.log` | Raw command transcript with candidate state, host/platform, tool versions, historical-to-candidate commit and source-change inventory, public `go doc` API, both vector runs, normalized comparison result, Capslock normal and verbose output, focused protocol/profile tests, command timestamps, and exit codes. |
| `SHA256SUMS` | SHA-256 digests for the reproduction script and raw transcript. |

## Results and Disposition

- Go 1.26.4 and Go 1.26.5 both ran the same 11 required top-level vector/profile tests, including draft string, transcript, generator, exchange-core, invalid-share, and package-local confirmation-tag fixtures plus package-owned CI and wire-prefix locks. Each lane passed.
- The capture removes only JSON event timestamps, elapsed fields, and textual duration values before comparison. Both normalized 112-line event streams have SHA-256 `05e966e9b4cc8ea6883d5f8a750cc2d91098a3e59aee365ed852e717067efbd9` and compare byte-for-byte equal; the recorded result is `vector_result=bit-identical`.
- Capslock reports the same two capability classes and counts as the historical Go 1.26.4 baseline: `ARBITRARY_EXECUTION` with 11 direct references through Go FIPS enforcement internals, and `UNANALYZED` with 13 direct references through `io.ReadFull` randomness reads. No capability class or count changed, and no filesystem, network, subprocess, plugin, environment-mutation, or unsafe capability appeared.
- The focused Go 1.26.5 protocol audit ran 42 top-level tests covering the exported exchange and Session API, role-local Caller input, lifecycle and persistent-secret cleanup, copied-state terminal behavior, peer metadata, peer-share rejection and error classification, Message framing and cap policy, CI/suite-profile locks, and draft invalid-share behavior. Every selected test and subtest passed.
- The historical-to-candidate production delta is limited to the issue #219 internal `Close` protocol concentration, the corrected scalar-sampling comment, and the Go 1.26.5 toolchain pin; the transcript preserves the exact commit and source-change inventory used by the audit. The security/spec audit found no public API, observable behavior, protocol, Message framing, dependency, or package-profile drift.

## Verification

On macOS:

```sh
cd docs/evidence/1f19e27-20260731-protocol
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/1f19e27-20260731-protocol
sha256sum -c SHA256SUMS
```

To perform an equivalent procedural recapture, run the following command from the repository root. The new output will contain live timestamps and temporary paths, so it is not expected to be byte-identical to the committed transcript and must not overwrite that hash-covered artifact.

```sh
docs/evidence/1f19e27-20260731-protocol/capture.sh /path/to/clean/candidate-worktree > /tmp/cpace-234-recapture.log 2>&1
```

## Residual Limitations

Capslock is experimental static analysis and its classifications require manual triage. Passing fixed vector/profile assertions and obtaining bit-identical normalized event streams provide strong cross-toolchain regression evidence, but do not prove the implementation cryptographically correct. The focused protocol lane does not replace the full, race, long-fuzz, GitHub alert, release-control, Scorecard, or signed Release Validation lanes tracked elsewhere.

External review of package-owned CI, Message framing, and suite-profile choices remains open, and independent cryptographic review remains required before any production-readiness claim. A later dependency, toolchain, protocol, parser/framing, security-relevant code, or package-profile change invalidates this lane and requires recapture against a newly selected candidate.

No `SHA256SUMS.sig` is included because no release-authorized signing key was used during branch implementation. The committed hashes provide tamper detection within repository history; the future signed release tag remains the release trust root.
