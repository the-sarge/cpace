# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those roles to the actual label vocabulary used in this repo's issue tracker (GitHub Issues on `the-sarge/cpace`).

The repo's live taxonomy (verify with `gh label list`) is **dimensional, not workflow-state**:

- `priority/high` (address before v1.0.0), `priority/medium`
- `kind/hardening`, `kind/coverage`
- `area/framing`, `area/ci`
- `release blocker`, `external review`, `security`
- `wontfix`

| Canonical role    | Label in our tracker | How to express it here                                            |
| ----------------- | -------------------- | ----------------------------------------------------------------- |
| `needs-triage`    | *(unmapped)*         | An issue with no `priority/*` label has not been triaged.          |
| `needs-info`      | *(unmapped)*         | Ask in an issue comment; do not label.                             |
| `ready-for-agent` | *(unmapped)*         | State readiness in a comment; do not label.                        |
| `ready-for-human` | *(unmapped)*         | State it in a comment; do not label.                               |
| `wontfix`         | `wontfix`            | Will not be actioned.                                              |

**Do not create labels.** The taxonomy is the maintainer's; `gh label create` for the unmapped roles would fragment triage. When a skill mentions an unmapped role, record the state in an issue comment instead, or ask the maintainer to extend this mapping.
