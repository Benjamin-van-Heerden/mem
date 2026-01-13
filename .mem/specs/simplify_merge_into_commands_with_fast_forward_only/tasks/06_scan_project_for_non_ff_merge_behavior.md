---
title: Scan project for non-ff merge behavior
status: completed
created_at: '2026-01-13T14:03:02.788311'
updated_at: '2026-01-13T14:11:40.287043'
completed_at: '2026-01-13T14:11:40.287037'
---
Search the codebase for any instances of merge commands or git config that don't use fast-forward only, and fix them if needed

## Completion Notes

Scanned codebase for merge.ff, ff_only, --no-ff, back-merge patterns. Found issues only in tests/test_merge_into.py: docstring mentions back-merges (lines 5-6) and test_into_main_dry_run_shows_all_steps asserts for back-merge output that no longer exists. These will be fixed in the tests update task.