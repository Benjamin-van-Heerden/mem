---
title: Update onboard command with docs section
status: completed
created_at: '2026-01-11T12:41:58.922628'
updated_at: '2026-01-11T13:45:03.121762'
completed_at: '2026-01-11T13:45:03.121756'
---
Modify src/commands/onboard.py:
- Add TECHNICAL DOCUMENTATION section after IMPORTANT FILES
- List each indexed document with its full summary content
- Check for unindexed docs (files without hash entries)
- Show warning if unindexed docs exist: '⚠️ Unindexed docs: {names}. Run mem docs index'
- Handle case where docs dir doesn't exist (skip section silently)

## Completion Notes

Added TECHNICAL DOCUMENTATION section to onboard output showing indexed docs with summaries and warning for unindexed docs. Fixed circular import by inlining config reader in docs.py