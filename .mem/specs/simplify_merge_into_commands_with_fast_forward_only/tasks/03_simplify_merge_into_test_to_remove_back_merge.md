---
title: Simplify _merge_into_test to remove back-merge
status: completed
created_at: '2026-01-13T13:52:13.036215'
updated_at: '2026-01-13T14:06:13.801709'
completed_at: '2026-01-13T14:06:13.801700'
---
Remove steps 6-8 (switch to dev, back-merge, push dev). Update dry-run output. Update success message to: test is now at the same commit as dev.

## Completion Notes

Removed back-merge steps (merge test into dev, push dev). Simplified flow to 6 steps. Updated dry-run output and success message.