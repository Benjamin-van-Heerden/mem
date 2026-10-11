---
title: Share Claude Code plugins and mods through mem templates
created_at: "2026-10-11T14:19:49+02:00"
---

Proposed 2026-10-11, waiting on the user's go-ahead and on where the permit mod lives. Make the template library a Claude Code plugin marketplace (.claude-plugin/marketplace.json, plugins/<name>/), add a plugin template item kind (template promote/reset), and have template sync write extraKnownMarketplaces and enabledPlugins into the project's .claude/settings.json, touching only entries mem added; report changes at onboard/sync; docs. autoMode rules stay per user (Claude Code never reads them from projects or plugins). First plugin: the permit mod, currently only in a session dev-mods folder (/Users/benjamin/.claude/dev-mods/0546f051-a81d-4cd0-8520-9ac080e5a0e7/permit/) — decide whether it lives in mem-templates or its own repo.
