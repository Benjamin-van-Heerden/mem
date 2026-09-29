---
title: Verify the compaction hook in Claude Code
status: open
created_at: "2026-09-29T14:27:47+02:00"
---

mem v0.5.0 installs a SessionStart/compact hook (.claude/settings.json) that runs mem hook compact. It is tested by running the command directly, not yet inside a real Claude Code compaction. After the next onboard installs it in this repository, compact a session (/compact) and check that the digest reaches the agent's context, that the sync ran, and that the timeout and bash guard behave.
