---
title: Add --accept flag to task complete command
status: todo
created_at: '2026-01-14T10:47:56.489498'
updated_at: '2026-01-14T10:47:56.489498'
completed_at: null
---
Modify the complete command in src/commands/task.py to add an --accept flag (default False). When --accept is False, output agent instructions and do NOT call tasks.complete_task_with_notes(). When --accept is True, proceed with the current completion logic.