---
title: Add tests for drift + patch
status: todo
created_at: '2026-01-15T12:54:06.824412'
updated_at: '2026-01-15T12:54:06.824412'
completed_at: null
---
Add automated tests for: (1) onboard drift warning without mutation, (2) mem patch config --dry-run does not modify files, (3) mem patch config applies changes preserving user values and unknown keys, and (4) idempotency (second run yields no changes).