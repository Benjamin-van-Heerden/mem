---
title: Fix status label ordering
status: todo
subtasks: []
created_at: '2026-01-07T11:42:13.162217'
updated_at: '2026-01-07T11:42:13.162217'
completed_at: null
---
mem spec complete should set merge_ready label, not mem merge. Currently merge runs sync which updates labels after PR is already merged - this is backwards