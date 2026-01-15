---
title: Implement mem patch config command
status: completed
created_at: '2026-01-15T12:54:03.461941'
updated_at: '2026-01-15T15:20:53.209930'
completed_at: '2026-01-15T15:20:53.209924'
---
Add a new top-level 'mem patch' command group with 'mem patch config' as the first target. Support --dry-run to print proposed changes without applying. Apply changes safely: add missing canonical keys with defaults, preserve known user values, preserve unknown keys by default, and ensure idempotency.

## Completion Notes

Created src/commands/patch.py with 'mem patch config' command. Features: removes unknown keys, adds missing keys with defaults, preserves user values, idempotent, supports --dry-run.