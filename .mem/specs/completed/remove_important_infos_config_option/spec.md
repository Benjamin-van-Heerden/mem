---
title: Remove important_infos config option
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 39
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/39
branch: dev-benjamin_van_heerden-remove_important_infos_config_option
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/41
created_at: '2026-01-12T09:57:25.820188'
updated_at: '2026-01-12T10:11:34.365699'
completed_at: '2026-01-12T10:11:34.365085'
last_synced_at: '2026-01-12T09:58:56.554756'
local_content_hash: a028611a9b6135a2c824c1af7c38e1b341f5b8559b58b238d4ef05b67346b58c
remote_content_hash: a028611a9b6135a2c824c1af7c38e1b341f5b8559b58b238d4ef05b67346b58c
---
## Overview

Remove the `important_infos` config option from mem. This feature is redundant since important information can be added directly to `AGENTS.md` (or similar agent instruction files) which are already included in project context. The config-based approach adds unnecessary complexity.

## Goals

- Remove `important_infos` from the config template
- Remove the display logic from the onboard command
- Clean up the local `.mem/config.toml` (move content to `AGENTS.md` if needed)

## Technical Approach

1. Remove the `important_infos` section from `src/templates/config.toml`
2. Remove the display logic in `src/commands/onboard.py` (lines ~638-645)
3. Update `.mem/config.toml` to remove the `important_infos` field (the content is already in `AGENTS.md`)

## Success Criteria

- `important_infos` no longer appears in the config template
- `mem onboard` no longer has an "IMPORTANT INFORMATION" section from config
- Tests pass
- Existing projects with `important_infos` in their config won't break (the field is simply ignored)

## Notes

The current `important_infos` content in `.mem/config.toml` reads:
```
- Remember, we are using `mem` to develop `mem` itself...
```

This exact content is already present in `AGENTS.md` under the "IMPORTANT INFORMATION" section, so no migration is needed - just deletion.
