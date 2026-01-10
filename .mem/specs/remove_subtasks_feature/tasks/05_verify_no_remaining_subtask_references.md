---
title: Verify no remaining subtask references
status: todo
subtasks: []
created_at: '2026-01-10T15:55:12.902177'
updated_at: '2026-01-10T15:55:12.902177'
completed_at: null
---
Search the codebase for any remaining references to 'subtask' using grep. Check src/commands/onboard.py and any other files. Remove any stray references found. Run 'uv run python main.py --help' and verify subtask command is gone.