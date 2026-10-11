---
title: Share Claude Code plugins and mods through mem templates
created_at: "2026-10-11T14:19:49+02:00"
---

Proposed 2026-10-11, waiting on the user's go-ahead and on where the permit mod lives. Make the template library a Claude Code plugin marketplace (.claude-plugin/marketplace.json, plugins/<name>/), add a plugin template item kind (template promote/reset), and have template sync write extraKnownMarketplaces and enabledPlugins into the project's .claude/settings.json, touching only entries mem added; report changes at onboard/sync; docs. autoMode rules stay per user (Claude Code never reads them from projects or plugins). Candidates: the permit and mission-control mods, which now live in /Users/benjamin/.claude/mods/ (not under version control) and load into every session through CLAUDE_CODE_PLUGIN_DIRS in ~/.claude/settings.json (with CLAUDE_CODE_PLUGIN_DIR_WATCH=1). Decide whether they live in mem-templates or a repo of their own; once installed from a marketplace, drop them from CLAUDE_CODE_PLUGIN_DIRS so they do not load twice.
