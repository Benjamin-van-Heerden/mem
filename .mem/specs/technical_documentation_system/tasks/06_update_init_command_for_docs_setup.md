---
title: Update init command for docs setup
status: todo
created_at: '2026-01-11T12:42:04.237446'
updated_at: '2026-01-11T12:42:04.237446'
completed_at: null
---
Modify src/commands/init.py:
- Add check for VOYAGE_AI_API_KEY and OPENROUTER_API_KEY env vars
- Show warning if missing: '⚠️ Document functionality requires VOYAGE_AI_API_KEY and OPENROUTER_API_KEY'
- Create .mem/docs/ directory structure (docs/, docs/summaries/, docs/data/)
- Add '.mem/docs/data/' to project .gitignore if not already present