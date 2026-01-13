---
title: Simplify _merge_into_main to remove back-merges
status: completed
created_at: '2026-01-13T13:52:15.434323'
updated_at: '2026-01-13T14:09:01.239124'
completed_at: '2026-01-13T14:09:01.239116'
---
Remove steps 7-10 (back-merge to test, back-merge to dev, push both). Update dry-run output. Update success message to: main is now at the same commit as test.

## Completion Notes

Removed back-merge steps (switch to test, merge main into test, push test, switch to dev, merge test into dev, push dev). Simplified flow to 6 steps. Updated dry-run output, success message, and command docstring.