---
title: Fix status label ordering
status: completed
subtasks: []
created_at: '2026-01-07T11:42:13.162217'
updated_at: '2026-01-07T12:02:56.215464'
completed_at: '2026-01-07T12:02:56.215459'
---
mem spec complete should set merge_ready label, not mem merge. Currently merge runs sync which updates labels after PR is already merged - this is backwards

## Completion Notes

spec complete now syncs merge_ready label to GitHub immediately