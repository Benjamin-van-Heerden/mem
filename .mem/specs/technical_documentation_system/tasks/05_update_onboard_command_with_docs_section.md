---
title: Update onboard command with docs section
status: todo
created_at: '2026-01-11T12:41:58.922628'
updated_at: '2026-01-11T12:41:58.922628'
completed_at: null
---
Modify src/commands/onboard.py:
- Add TECHNICAL DOCUMENTATION section after IMPORTANT FILES
- List each indexed document with its full summary content
- Check for unindexed docs (files without hash entries)
- Show warning if unindexed docs exist: '⚠️ Unindexed docs: {names}. Run mem docs index'
- Handle case where docs dir doesn't exist (skip section silently)