---
title: Raise Codex's AGENTS.md limit and warn on large AGENTS.md
status: todo
created_at: "2026-10-25T15:33:39+02:00"
updated_at: "2026-10-25T15:33:39+02:00"
---

Codex stops reading AGENTS.md at project_doc_max_bytes (32,768 by default), which cuts off memories; shannon was at 33,456 bytes on 2026-10-25. init, import and onboard ensure .codex/config.toml has a top-level project_doc_max_bytes, whatever the hook switches say. If it is missing, insert project_doc_max_bytes = 65536 before the first [table] header line, creating the file if needed. Never change an existing value. Edit line by line so comments and other settings survive. Publish with the Codex hook file, with the same ignored and uncommitted-edits handling. Verified with codex 0.153.4 (codex debug prompt-input, scratch CODEX_HOME with the scratch repo trusted): the project setting is honoured. Onboard also prints one warning line (⚠️) in its updates when AGENTS.md exceeds 40,000 bytes, naming the size and the three largest memories with their sizes, and suggesting trimming or retiring memories (mem memory remove). Tests: insertion into a file with comments and tables, an existing value kept, the size warning.
