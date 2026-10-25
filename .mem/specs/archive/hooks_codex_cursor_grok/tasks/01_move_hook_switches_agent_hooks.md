---
title: Move the hook switches to [agent_hooks]
status: todo
created_at: "2026-10-25T15:33:39+02:00"
updated_at: "2026-10-25T15:33:39+02:00"
---

Replace project.Config's Claude ClaudeConfig ([claude] compact_hook/prompt_hook) with AgentHooks AgentHooksConfig ([agent_hooks] compact/prompt; unset means on; CompactEnabled/PromptEnabled helpers). Raise project.Schema to 3 and add patch 2: moveClaudeHooks in internal/project/patches.go. It reads .mem/config.toml as raw TOML, moves the values and drops an empty [claude] table. Upgrade writes p.Config after each patch, so the patch must also update the in-memory config: change the patch signature if needed. Update every caller (onboard_updates.go, claudeHooksLine wording). Test: a schema 2 config with [claude] compact_hook = false upgrades to [agent_hooks] compact = false with no [claude] table, and onboard then leaves the compact entry out.
