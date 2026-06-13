# Code-review followups (2026-05-28)

This doc tracks items from the 2026-05-27 multi-agent code review that are **not** ADR-worthy on their own — they are implementation/hardening tasks rather than architectural decisions — but should be filed as GitHub issues for visible tracking before v1.0.0.

The ADR-worthy items from the same review live in `docs/adr/0002` through `docs/adr/0007`. The cheap-and-safe internal/test/doc fixes from the same review landed on branch `code-review/safe-fixes-2026-05-28`.

## Items

### F1. Decoder: reject empty `sid` field when local `AllowEmptySessionID == false`

**Severity:** Medium. **Surfaced as review item M4.**

`framing.go:146-160` `readField` accepts a zero-length `sid` value from the wire. The local responder's SID-equality check at `api.go:163` then rejects the message with the generic `ErrMessage: session id mismatch`. The protocol-level acceptance set is unchanged either way; the distinction is **error-message specificity**.

A peer-sent empty `sid` against a non-empty-local-`SessionID` responder is currently indistinguishable in logs from any other SID mismatch. Operators triaging "did the attacker probe whether we accept empty SIDs?" cannot tell from the error code alone.

**Proposed fix:**
- Pass `cfg.AllowEmptySessionID` into the wire decoder (or the immediate `Respond` caller of `decodeMessageA`).
- If the local responder requires non-empty `SessionID` and the peer sent a zero-length `sid`, reject with a more specific wrapped error such as `ErrMessage: peer sent empty session id`.
- Alternative (smaller fix): document explicitly at `docs/security-assessment.md` that decoder accepts empty fields and the protocol-level mismatch is the actual rejection point.

**File as issue with labels:** `area/framing`, `kind/hardening`, `priority/medium`.

---

### F2. Decoder: aggregate message size cap in `newMessageReader`

**Severity:** Medium. **Surfaced as review item M6.**

**Status:** Addressed by PR #92. `newMessageReader` now has a 128 KiB aggregate invalid-message backstop, while valid package-owned message shapes remain governed by per-field caps and exact public-share/tag lengths.

Before PR #92, `framing.go:130-144` performed only a 3-byte header check then began field-by-field parsing. Per-field size caps were enforced individually (max 64 KiB for AD, max 1 KiB for sid, etc.) but no upfront `len(in) > maxMessageLen` check existed.

Before the fix, a server calling `Respond` in a tight loop under a flood of maximally-sized (~66 KiB) malformed messages allocated field-by-field on every call before any cryptographic filter. Per-field caps were the only throttle.

**Implemented fix:**
- Added `maxMessageLength = 128 << 10`, a round cap well above every valid package-owned message shape.
- Checked `if len(in) > maxMessageLength` in `newMessageReader` after header version/suite/role validation and rejected early with `ErrMessage`.
- No valid message length is reduced; this rejects only invalid oversized messages earlier.

**Tracking:** formerly file as issue with labels `area/framing`, `kind/hardening`, `priority/medium`; PR #92 resolves it.

---

### F3. CI: upload Scorecard SARIF to Code Scanning

**Severity:** High. **Surfaced as review item H6.b.**

`.github/workflows/scorecard.yml:32-35` runs Scorecard with `results_format: json` and `publish_results: true` but never uploads SARIF to Code Scanning, never uploads the JSON as a workflow artifact, and never fails on findings. Findings are visible only on the externally hosted OpenSSF page.

**Proposed fix:**
- Switch `results_format: json` → `results_format: sarif`.
- Add upload via `github/codeql-action/upload-sarif` (SHA-pinned; reuse the same version pin as `sast-gate.yml`).
- Ensure `permissions: security-events: write` is set at job level.
- Reference the upstream `ossf/scorecard-action` README pattern.

**File as issue with labels:** `area/ci`, `kind/hardening`, `priority/high`.

---

### F4. CI: release-time license assertion

**Severity:** Medium. **Surfaced as review item H6.d.**

`.github/workflows/dependency-gate.yml:31-38` runs `actions/dependency-review-action` only on `pull_request` events. License drift on `main` between PRs (theoretically possible via transitive dependency floating, though `go.sum` pins prevent it in practice) is not caught at release time.

**Proposed fix:**
- Add a `go-licenses check` (or equivalent) step to `.github/workflows/release.yml`.
- Configure the allowed-licenses list (currently `BSD-3-Clause, MIT, Apache-2.0, ISC` per `dependency-gate.yml:35`) as a shared constant or via a project-level config.
- The step runs after `verify-tag` (see ADR-0007) and before SBOM generation.
- Block the release if any direct or transitive dep has an unlisted license.

**File as issue with labels:** `area/ci`, `kind/hardening`, `priority/medium`.

---

### F5. CI: expand cross-platform smoke matrix

**Severity:** Medium. **Surfaced as review item M8.**

`.github/workflows/cross-platform.yml:52-56` runs `go build ./...` and `go vet ./...` on macOS and Windows. It does not:

- Run `go test ./...` (only build).
- Cover 32-bit (`linux/386`, `linux/arm`).
- Cover big-endian (`linux/s390x`, `linux/ppc64`).

For a crypto library with byte-ordering-sensitive framing (LEB128, length-prefixed concatenation, fixed-size point/tag fields), build-only on two LE 64-bit OSes is a thin smoke test.

**Proposed fix:**
- Add `go test ./...` to the existing macOS/Windows matrix entries.
- Add a `linux/s390x` matrix entry via QEMU (`docker run --platform linux/s390x ...`) running `go test ./...`. This catches byte-order bugs.
- Add a `linux/386` matrix entry via QEMU running `go test ./...`. This catches 32-bit `int` truncation bugs (relates also to review item M5 about `readLEB128` `int` accumulator safety).
- Frame as "advisory" if QEMU runs are flaky on the runner — gate-status can be informational while the matrix stabilises.

**File as issue with labels:** `area/ci`, `kind/coverage`, `priority/medium`.

---

## Notes

- Filing these as issues is preferred over expanding them into ADRs because they are not load-bearing for the v1.0.0 freeze beyond a "should land before v1.0.0" milestone. None of them changes the public API surface or the protocol semantics.
- Review item M5 (`readLEB128` `int` accumulator widening) is intentionally **not** in this list. The current `maxLEB128BytesForField = 3` constant makes the accumulator safe on every architecture; the proposed widening is a defense-in-depth measure that only matters if a future change raises the constant. If F5 lands and adds 32-bit coverage, M5 becomes a real bug detector; until then it is a comment-on-the-static-assertion task and doesn't need its own issue.
- Review items M14 (password double-clear documentation) and parts of M12 (doc accuracy items) are addressed by the safe-fixes branch; they do not need followups.
- Review item H6.e (CHANGELOG `Unreleased` staleness) is also addressed by the safe-fixes branch.
