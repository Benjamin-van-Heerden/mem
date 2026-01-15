---
title: Update config.toml template
status: completed
created_at: '2026-01-15T12:53:56.877602'
updated_at: '2026-01-15T13:07:57.895856'
completed_at: '2026-01-15T13:07:57.895843'
---
Update the config template used by mem (init/templates) so it matches current supported .mem/config.toml schema/behavior, including symlink-related config if still supported.

## Completion Notes

Updated src/templates/config.toml to reflect recommended generic_templates (python/general) and made worktree.symlink_paths default to empty (optional) while keeping the supported key documented.