---
created_at: '2026-01-12T10:09:23.517784'
username: benjamin_van_heerden
spec_slug: remove_important_infos_config_option
---
# Work Log - Remove important_infos config option

## Overarching Goals

Remove the redundant `important_infos` config option from mem. This feature duplicated functionality since important information can be added directly to `AGENTS.md` or similar agent instruction files which are already included in project context.

## What Was Accomplished

### Removed important_infos from config template
Deleted the documentation and commented example for `important_infos` from `src/templates/config.toml`.

### Removed display logic from onboard command
Removed the code in `src/commands/onboard.py` that read the `important_infos` config value and displayed it under an "IMPORTANT INFORMATION" section.

### Cleaned up local config
Removed the `important_infos` field from `.mem/config.toml`. The content was already present in `AGENTS.md` so no migration was needed.

### Verified changes
Ran ad-hoc tests confirming:
- `mem onboard` runs successfully
- The "IMPORTANT INFORMATION" section no longer appears in output
- All `important_infos` references removed from codebase

## Key Files Affected

- `src/templates/config.toml` - Removed important_infos documentation and example
- `src/commands/onboard.py` - Removed display logic (lines 636-645)
- `.mem/config.toml` - Removed important_infos field

## What Comes Next

Spec is complete. All tasks finished:
- [x] Remove important_infos from config template
- [x] Remove important_infos display logic from onboard
- [x] Remove important_infos from local config
- [x] Run tests
