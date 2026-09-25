---
title: Onboard integration
status: todo
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-25T14:00:22+02:00"
---

In applyUpdates, pull the library (skip with --offline, warn and use the cached clone on failure), run the sync, print results under a TEMPLATES section, and publish changed paths (AGENTS.md, .agents/skills, .claude/skills, .mem/docs, .mem/templates.lock, .mem/config.toml) with the existing publish helper, not when those paths already had uncommitted edits. Add an end-to-end test: promote in project B, onboard in project A installs the item.
