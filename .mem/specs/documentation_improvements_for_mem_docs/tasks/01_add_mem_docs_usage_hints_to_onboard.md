---
title: Add mem docs usage hints to onboard
status: todo
created_at: '2026-01-13T18:07:42.494115'
updated_at: '2026-01-13T18:07:42.494115'
completed_at: null
---
In src/commands/onboard.py, add usage hints for mem docs commands (search, list, index) to the 'Key commands' section in the About mem block around line 290-310. Add entries like: '- mem docs search "query" - Semantic search across indexed documentation', '- mem docs list - List all documents', '- mem docs index - Index documents'