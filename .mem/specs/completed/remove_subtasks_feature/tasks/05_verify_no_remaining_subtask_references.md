---
title: Verify no remaining subtask references
status: completed
subtasks: []
created_at: '2026-01-10T15:55:12.902177'
updated_at: '2026-01-10T16:19:09.683576'
completed_at: '2026-01-10T16:19:09.683568'
---
Search the codebase for any remaining references to 'subtask' using grep. Check src/commands/onboard.py and any other files. Remove any stray references found. Run 'uv run python main.py --help' and verify subtask command is gone.

## Completion Notes

Searched codebase for subtask references. Cleaned up: spec.py (removed subtask display in show command and subtask check in complete), onboard.py (removed subtask concept from primitives and workflow hints), README.md (updated primitives from 5 to 4, removed subtask sections), deleted test_task_integrity.py, updated test_spec_subdirectories.py. Verified no subtask references remain in src/, tests/, main.py, or README.md.