# CPace v0.1.3 Assembled Candidate Packet

Frozen source candidate: `1f19e278112fa037890848ed6c086addeffdca4e`

Frozen source tree: `a62d4279b23be8e6e47f7ff5869c8adceb85111f`

Packet assembly date: 2026-08-01 UTC

This packet assembles the current dependency/SAST, protocol/vector/capability, paired long-fuzz, and GitHub release-control lanes for the frozen `v0.1.3` source contract and records the final pre-tag gate. It is project-side evidence for an explicitly non-production-ready prerelease, not independent external or cryptographic review.

## Evidence Lanes

| Lane | Raw bundle | Capture and result |
| --- | --- | --- |
| Dependency, vulnerability, and SAST/gosec | `docs/evidence/1f19e27-20260731/` | Clean detached source candidate; 2026-07-31 06:58:34–06:58:39 UTC; module integrity passed, `govulncheck@v1.3.0` found no vulnerabilities, and `gosec@v2.26.1 -tests` reported zero issues. |
| Toolchain, vectors, Capslock, and protocol audit | `docs/evidence/1f19e27-20260731-protocol/` | Clean detached source candidate; 2026-07-31 08:15:24–08:15:32 UTC; Go 1.26.4 and Go 1.26.5 normalized vector event streams were bit-identical, Capslock classes/counts were unchanged, and all 42 focused top-level protocol/profile tests passed. |
| Paired long fuzzing | `docs/evidence/1f19e27-20260731-fuzz/` | Clean detached source candidate; 2026-07-31 08:29:08–23:50:10 UTC; ARM passed all 14 targets in 50,416 seconds, Intel passed 13 before a one-hour `FuzzMessageARoundTrip` deadline miss in 50,420 seconds, and the same-host 3,601-second targeted rerun passed with no new artifacts. |
| GitHub release controls and scheduled signal | `docs/evidence/1f19e27-20260801-github/` | Point-in-time capture 2026-08-01 02:54:40–02:55:00 UTC; active protected tag ruleset with no bypass, exact required `main` contexts, empty open alert sets, Scorecard 7.4, and recorded workflow failure/recovery and runner-availability qualifications. |
| Final pre-tag integration gate | `final-gate.log` | Clean documentation-only descendant `04816e1c71df0c2f4add52408e237c26e2ff1c5b` (tree `fb61242bbdaae84eba95a49c9bd0b95e7f87082f`); 2026-08-01 04:10:15–04:11:56 UTC; all release-facing local gates and pinned analysis reruns passed, with 97.2% statement coverage. |

Every source-sensitive lane names candidate `1f19e278112fa037890848ed6c086addeffdca4e`. The GitHub lane also names that frozen source contract while preserving that the source SHA is an intermediate PR #239 commit without direct check suites, the successful required checks belong to its documentation-only checked descendant, and several scheduled signals ran on other explicitly recorded SHAs.

## Packet Contents

| File | Description |
| --- | --- |
| `packet-manifest.txt` | Frozen source commit/tree plus the SHA-256 digest of each current lane's `SHA256SUMS`; it also marks `f7efa6a963a954952b1ecad3f46530f13799fe89` historical and keeps publication-only checks pending. |
| `release-notes.txt` | Exact output of `scripts/extract-release-notes.sh CHANGELOG.md v0.1.3`, preserving the breaking pre-v1 migration first, unchanged wire/suite behavior, non-production-ready posture, evidence scope, and remaining #29-#32 production-readiness blockers. |
| `final-gate.log` | Raw final pre-tag command transcript with repository identity, clean state, host/tool metadata, commands, timestamps, durations, outputs, and return codes. |
| `SHA256SUMS` | SHA-256 digests for the manifest, extracted release notes, and final-gate transcript. |

## Candidate-Gate Disposition

The dependency review, `govulncheck`, gosec, CodeQL alert, Dependabot alert, secret-scanning alert, and manual license results contain no violation under `docs/security-gates.md`; no `docs/vex.md` entry or suppression is required. The Scorecard deductions, Nightly Fuzz deadline miss with successful failed-job rerun, Intel long-fuzz deadline miss with same-host targeted recovery, and autoscaled arm64 queue limits are preserved as residual process or availability qualifications rather than silently promoted to clean source-candidate runs.

The clean integration gate ran `task docs:check`, `task quick`, `task check`, an explicit race rerun, coverage, pinned Staticcheck, pinned test-aware govulncheck, pinned test-inclusive gosec, normal and verbose Capslock, release-policy validation, release-note extraction, release metadata checks, every child-bundle checksum, and clean-state assertions. The full Task gate took 86 seconds; `govulncheck@v1.3.0` found no vulnerabilities, `gosec@v2.26.1 -tests` reported zero issues, Capslock repeated the expected 11 `ARBITRARY_EXECUTION` and 13 `UNANALYZED` direct references, and the final worktree remained clean.

The clean gate necessarily checked a provisional two-line `final-gate.log` and its matching checksum so the evidence-baseline validator could validate a complete bundle without a self-referential hash. Replacing that marker with the completed transcript is the only post-gate raw-artifact mutation; the packet checksum, evidence-baseline validator, documentation check, and release-note comparison are rerun afterward against the final bytes.

Issue #44's remaining criteria are met by the raw checked bundles, the hash-linked manifest, and the recorded bit-identical Go 1.26.4/1.26.5 vector comparison. Issue #33's pre-tag lanes are assembled here. Its tag-triggered Release Validation and post-gosec-SARIF Code Scanning check are structurally post-tag, remain blocking publication verification, and are not claimed by this packet.

## Verification

On macOS:

```sh
cd docs/evidence/1f19e27-20260801-candidate
shasum -a 256 -c SHA256SUMS
```

On Linux:

```sh
cd docs/evidence/1f19e27-20260801-candidate
sha256sum -c SHA256SUMS
```

Then verify each bundle digest named in `packet-manifest.txt` and run `scripts/check-evidence-baseline.sh` from the repository root.

## Residual Limitations

The packet aggregates project-generated evidence; it does not convert self-review into independent external or cryptographic review. The GitHub evidence is point-in-time and mutable at the service, the frozen source SHA itself has no direct GitHub checks, the Intel all-target fuzz campaign did not finish cleanly despite its same-host targeted recovery, and autoscaled arm64 capacity remained intermittently unavailable.

No `SHA256SUMS.sig` is included because no release-authorized signing key is used during branch implementation. The checksums provide tamper detection within repository history; the maintainer's future signed release tag remains the release trust root. Before publication, recapture mutable GitHub state if its freshness condition fires, create and verify the signed annotated tag, require every Release Validation job to pass, recheck Code Scanning after SARIF ingestion, and verify the published SBOM, Sigstore bundle, checksum, attestation, prerelease flag, and not-latest flag.
