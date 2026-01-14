---
title: Add --accept flag to task complete command
status: completed
created_at: '2026-01-14T10:47:56.489498'
updated_at: '2026-01-14T11:00:38.289242'
completed_at: '2026-01-14T11:00:38.289235'
---
Modify the complete command in src/commands/task.py to add an --accept flag (default False). When --accept is False, output agent instructions and do NOT call tasks.complete_task_with_notes(). When --accept is True, proceed with the current completion logic.

## Completion Notes

Added --accept flag with user confirmation flow. Without flag: displays task description and agent instructions, does NOT complete task. With flag: actually completes task. Also displays full task body including amendments so agent can compare work against requirements.