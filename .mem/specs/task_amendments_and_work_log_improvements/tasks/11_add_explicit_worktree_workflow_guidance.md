---
title: Add explicit worktree workflow guidance
status: completed
subtasks: []
created_at: '2026-01-09T16:07:53.375176'
updated_at: '2026-01-09T16:35:22.073591'
completed_at: '2026-01-09T16:35:22.073586'
---
Add clear messaging throughout the spec workflow: 1) After spec new - warn to only create tasks from main repo, must start new session in worktree to work on spec. 2) After spec assign - show exact worktree path, emphasize creating new agent session there, warn against working in main repo. 3) Make it impossible to misunderstand: agents cannot continue working from main repo after assign.

## Completion Notes

Added clear messaging in spec new (warning about starting new session after assign) and enhanced spec assign output (🛑 THIS SESSION MUST END HERE, WHY A NEW SESSION section explaining the isolation, IN THE WORKTREE SESSION section with commands)