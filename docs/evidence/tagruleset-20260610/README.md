# v* Tag-Authority Ruleset Evidence

Date: 2026-06-10

Main at capture: `8e57063` (post PR #71 merge)

These files record the ADR-0007 tag-authority control going active in its
final form, per the *Tag authority control* acceptance criterion and
`docs/release-checklist.md` §3.

Ruleset history: "Protect release tags" (id `16048307`) has been ACTIVE on
`refs/tags/v*` since 2026-05-06 with `update` + `deletion` rules and an empty
bypass list. On 2026-06-10 the missing `creation` rule was added
(maintainer-authorized), completing the criterion: creation, update, and
deletion of `refs/tags/v*` are restricted with **no routine bypass actors**
(`current_user_can_bypass: never`). Releasing a `v*` tag therefore requires a
deliberate, documented break-glass repository-admin change (temporarily
adding a bypass actor or disabling the ruleset), per the ADR.

## Files

| File | Contents |
| --- | --- |
| `rulesets-list.json` | `gh api /repos/the-sarge/cpace/rulesets` — all repository rulesets at capture. |
| `ruleset-16048307.json` | `gh api /repos/the-sarge/cpace/rulesets/16048307` — full ruleset detail after the `creation` rule was added. |
| `ruleset-negative-test.log` | Negative authorization test: a repository-admin attempt to push `refs/tags/v0.0.0-ruleset-test` rejected with `GH013 — Cannot create ref due to creations being restricted`. |
| `SHA256SUMS` | SHA-256 digests for the files above. |

## Verification

On macOS:

```sh
cd docs/evidence/tagruleset-20260610
shasum -a 256 -c SHA256SUMS
```

Live re-verification: `gh api /repos/the-sarge/cpace/rulesets/16048307` should
show `enforcement: active`, rules `creation`/`update`/`deletion`, conditions
including `refs/tags/v*`, and `bypass_actors: []`.
