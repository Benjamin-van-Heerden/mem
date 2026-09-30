---
title: Document setup files
status: completed
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T12:01:53+02:00"
completed_at: "2026-09-30T12:01:53+02:00"
---

docs/design.md: Templates section (setup.md at the template root, installed once by init to .mem/setup.md, never synced or locked, deleted when done; template use does not install it) and the .mem/ tree; init on an empty repository. docs/status.md: implemented list. README.md: the init flow (empty repo → mem init --template → setup) and the .mem/ layout. Keep the planned/implemented distinction. Do not add setup to the managed AGENTS.md instructions: onboard's instruction carries it.

## Completion Notes

docs/design.md: .mem/setup.md in the layout, init's empty first commit, and a Setup paragraph in the Templates section (format, init joins setups, onboard/digest lead with it, deleting ends it, not a synced/locked/excluded item, template use only mentions it); template list's setup column. docs/status.md implemented list. README: the new-app flow (empty repo → clone → mem init --template) and the layout. .mem/structure.md updated for setup.go, onboard_setup.go, the onboard/digest flow and init. Managed AGENTS.md instructions unchanged: onboard's instruction carries the setup.
