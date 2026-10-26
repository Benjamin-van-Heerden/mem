---
title: Cloud agent support
created_at: "2026-10-26T13:04:58+02:00"
---

Deferred on 2026-10-26: cloud agents (Cursor cloud agents, Claude Code on the web, Codex cloud, Copilot) are not yet ready for broad use. When they are, revisit docs/cloud_agents.md section 2 and its research table: mem writes the setup that installs mem (.cursor/environment.json install, a Claude Code SessionStart hook gated on CLAUDE_CODE_REMOTE, Copilot setup steps); detection of a cloud agent; mem promote/deploy refusing there ([release] from_cloud to opt in); MEM_USER as mem's identity so cloud commits by 'Cursor Agent' are not one shared user; switching to the development branch at onboard or landing the platform branch on it at each sync; loud failure when a hook finds mem missing. Re-check the research first: these platforms change monthly.
