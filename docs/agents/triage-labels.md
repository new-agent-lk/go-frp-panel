# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those roles to the actual label strings used in this repo's issue tracker.

| Label in mattpocock/skills | Label in our tracker | Meaning                                  |
| -------------------------- | -------------------- | ---------------------------------------- |
| `needs-triage`             | `needs-triage`       | Maintainer needs to evaluate this issue  |
| `needs-info`               | `needs-info`         | Waiting on reporter for more information |
| `ready-for-agent`          | `ready-for-agent`    | Fully specified, ready for an AFK agent  |
| `ready-for-human`          | `ready-for-human`    | Requires human implementation            |
| `wontfix`                  | `wontfix`            | Will not be actioned                     |

When a skill mentions a role (e.g. "apply the AFK-ready triage label"), use the corresponding label string from this table.

Edit the "Label in our tracker" column to match whatever vocabulary you actually use.

## Category labels

`bug` and `enhancement` are the two category roles from the `triage` skill. Both already exist in this repo's tracker.

## Wayfinder labels

`/wayfinder` uses its own label namespace, separate from triage roles. A map issue and its child tickets carry **only** these, never a triage label:

| Label              | Meaning                                              |
| ------------------ | ---------------------------------------------------- |
| `wayfinder:map`    | The map issue itself                                 |
| `wayfinder:research`   | AFK ticket: read docs/APIs to surface a fact     |
| `wayfinder:prototype`  | HITL ticket: cheap concrete artifact to react to |
| `wayfinder:grilling`   | HITL ticket: conversation resolves the decision  |
| `wayfinder:task`       | Manual work needed before a decision can be made |
