---
title: Global config and template alignment
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 3
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/3
branch: dev-benjamin_van_heerden-global_config_and_template_alignment
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/4
created_at: '2026-01-06T13:42:32.138219'
updated_at: '2026-01-06T14:32:07.454227'
completed_at: '2026-01-06T14:32:07.453195'
last_synced_at: '2026-01-06T14:24:05.334422'
local_content_hash: 1e393df540afc5a0f6f9f571bae9fc1eb0c0513c06fe90f3a76931c27d94e5ed
remote_content_hash: 1e393df540afc5a0f6f9f571bae9fc1eb0c0513c06fe90f3a76931c27d94e5ed
---
## Overview

Establish a proper global config structure at `~/.config/mem/` that provides defaults for all mem projects, and ensure spec templates are aligned between local creation and GitHub issue sync.

## Goals

- Global `~/.config/mem/config.toml` provides default settings inherited by all projects
- Templates live in `~/.config/mem/templates/` (already working)
- Local `.mem/config.toml` overrides global defaults
- Spec template (`src/templates/spec.md`) matches GitHub issue template (`mem-spec.yml`)
- Specs created via GitHub issues look identical to specs created locally

## Technical Approach

1. **Global config loading**: Update config reading to check `~/.config/mem/config.toml` first, then merge/override with local `.mem/config.toml`

2. **Spec template alignment**: Add "Success Criteria" section to `src/templates/spec.md` to match `mem-spec.yml`

3. **Sync body parsing**: When syncing GitHub issues, the issue body uses the YAML form format. Consider whether to parse this into markdown sections or keep as-is.

## Notes

Current structure:
```
~/.config/mem/
├── config.toml      # Global defaults (empty, to be defined)
└── templates/
    ├── python.md
    ├── rust.md
    └── ...
```

The `generic_templates_location` config already points to `~/.config/mem/templates` so template loading works. The gap is global config defaults.
