---
title: Add completion date ordering to spec list
status: todo
created_at: '2026-01-14T10:48:03.915230'
updated_at: '2026-01-14T10:48:03.915230'
completed_at: null
---
Modify the list command in src/commands/spec.py. When listing completed specs (-s completed), sort by completed_at datetime. Display the completion date in human-readable format next to each spec.