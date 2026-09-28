---
title: mem template promote and reset
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-28T07:53:57+02:00"
completed_at: "2026-09-28T07:53:57+02:00"
---

promote <memory|skill|doc> <name> [--to <template>]: copy the project item into the library clone, commit 'Promote <kind> <name> from <project>', push, record the hash in the lock. Require --to when the item is untracked and the project uses more than one template. reset <kind> <name>: overwrite the local item with the template copy and record the hash. Test against a bare library repository.

## Completion Notes

templates.Promote copies a project memory, skill or doc into its template (from the lock, the project's single template, or --to; a new template gets a template.toml and is added to the project's use list), pulls, commits with the project's Git identity, pushes with the push time limit (resetting the cached clone if the push fails), and records the item in the lock. templates.Reset installs the template's copy, records it, and lifts an exclusion. CLI: mem template promote <kind> <name> [--to] and mem template reset <kind> <name>; mem template list now prints hints with these commands per item state. Verified by TestPromotedItemsReachOtherProjectsAtOnboard (memory promoted in project b arrives in project a at onboard; a skill promoted --to a new template creates it and does not reach a) and TestResetTakesTheTemplateCopyAndLiftsAnExclusion (restores an excluded doc, replaces local skill edits, requires --to for untracked items with several templates). go vet clean.
