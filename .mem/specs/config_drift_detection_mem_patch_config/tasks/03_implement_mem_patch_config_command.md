---
title: Implement mem patch config command
status: todo
created_at: '2026-01-15T12:54:03.461941'
updated_at: '2026-01-15T12:54:03.461941'
completed_at: null
---
Add a new top-level 'mem patch' command group with 'mem patch config' as the first target. Support --dry-run to print proposed changes without applying. Apply changes safely: add missing canonical keys with defaults, preserve known user values, preserve unknown keys by default, and ensure idempotency.