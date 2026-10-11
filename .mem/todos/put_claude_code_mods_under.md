---
title: Put the Claude Code mods under version control
created_at: "2026-10-11T20:42:29+02:00"
---

The permit and mission-control mods live only in /Users/benjamin/.claude/mods/ with no history or backup. They load into every session through CLAUDE_CODE_PLUGIN_DIRS in ~/.claude/settings.json (with CLAUDE_CODE_PLUGIN_DIR_WATCH=1). Decided 2026-10-11 not to bundle them into mem: they are personal workflow on an early-access API. Make ~/.claude/mods a Git repository; optionally push it to a private GitHub repo with .claude-plugin/marketplace.json so other machines can install the mods with /plugin install (then drop them from CLAUDE_CODE_PLUGIN_DIRS so they do not load twice).
