---
title: Onboard integration
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-28T07:52:18+02:00"
completed_at: "2026-09-28T07:52:18+02:00"
---

In applyUpdates, pull the library (skip with --offline, warn and use the cached clone on failure), run the sync, print results under a TEMPLATES section, and publish changed paths (AGENTS.md, .agents/skills, .claude/skills, .mem/docs, .mem/templates.lock, .mem/config.toml) with the existing publish helper, not when those paths already had uncommitted edits. Add an end-to-end test: promote in project B, onboard in project A installs the item.

## Completion Notes

onboard runs syncTemplates after the managed-block updates: opens the project's library (pulling unless --offline; a failed pull or clone becomes a ⚠️ line, not an error), syncs, prints the result under 🧩 TEMPLATES, and publishes the changed paths with the existing publish helper unless AGENTS.md, .agents/skills, .claude/skills, .mem/docs, the lock or the config already had uncommitted edits (then it says to commit them together). The onboard instruction adds a step for ⚠️ template items. Verified by TestOnboardDrawsInNewTemplateItemsAndPublishesThem: init --template installs a memory, a skill pushed to the bare library arrives at onboard, is reachable through .claude/skills, is committed and pushed with no unpushed warning, and --offline does not pull while a normal onboard does. go vet clean.
