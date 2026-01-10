---
title: Emojify CLI output
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 31
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/31
branch: dev-benjamin_van_heerden-emojify_cli_output
pr_url: null
created_at: '2026-01-10T11:34:28.559274'
updated_at: '2026-01-10T12:06:12.211043'
completed_at: null
last_synced_at: '2026-01-10T11:35:20.385652'
local_content_hash: 5dab4d6100e838e11b35148b48dbe46ba28ed4688cbc1b91e241dc4604630125
remote_content_hash: 5dab4d6100e838e11b35148b48dbe46ba28ed4688cbc1b91e241dc4604630125
---
## Overview

Add tasteful emoji usage throughout the CLI output to improve visual clarity and make the interface more friendly.

## Goals

- Add emojis to all command files in `src/commands/`
- Keep it tasteful - just enough to look nice, not overwhelming
- Use emojis consistently across commands for similar concepts

## Technical Approach

Add emojis to enhance:
- Success messages (already uses ✓ in some places, standardize with ✅)
- Error messages (already uses ❌, keep consistent)
- Section headers in onboard output
- Status indicators (todo, completed, merge_ready, abandoned)
- Action indicators (creating, updating, syncing, etc.)
- Warnings (⚠️)

Files to update:
1. `src/commands/spec.py` - spec management commands
2. `src/commands/task.py` - task management commands
3. `src/commands/subtask.py` - subtask commands
4. `src/commands/sync.py` - GitHub sync (already has some)
5. `src/commands/onboard.py` - context output
6. `src/commands/init.py` - initialization
7. `src/commands/log.py` - work log commands
8. `src/commands/merge.py` - merge operations
9. `src/commands/cleanup.py` - cleanup operations

## Emoji Conventions

| Concept | Emoji |
|---------|-------|
| Success/Complete | ✅ |
| Error | ❌ |
| Warning | ⚠️ |
| Info/Tip | 💡 |
| Creating | ✨ |
| Spec | 📋 |
| Task | ✏️ |
| Branch/Git | 🌿 |
| Sync | 🔄 |
| GitHub | 🐙 |
| Worktree | 📂 |
| Log | 📝 |
| Stop/Important | 🛑 |

## Success Criteria

- All command files have consistent emoji usage
- Output looks cleaner and more visually organized
- Not overdone - emojis enhance, not distract
