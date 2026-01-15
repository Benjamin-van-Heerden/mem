---
title: Add tests for drift + patch
status: completed
created_at: '2026-01-15T12:54:06.824412'
updated_at: '2026-01-15T15:24:00.582653'
completed_at: '2026-01-15T15:24:00.582646'
---
Add automated tests for: (1) onboard drift warning without mutation, (2) mem patch config --dry-run does not modify files, (3) mem patch config applies changes preserving user values and unknown keys, and (4) idempotency (second run yields no changes).

## Completion Notes

Created tests/test_config_drift.py with 10 tests covering: drift detection (clean config, unknown top-level, unknown nested, unknown in list items, validation errors), config generation (validates, preserves values), and patch logic (removes unknown, preserves files, idempotent). All tests pass.