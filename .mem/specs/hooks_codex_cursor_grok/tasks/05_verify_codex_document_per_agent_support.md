---
title: Verify with Codex and document per-agent support
status: todo
created_at: "2026-10-25T15:33:39+02:00"
updated_at: "2026-10-25T15:33:39+02:00"
---

First ask the user for consent, since this uses their Codex account. Then check against real Codex in a scratch repository initialised with dist/mem-dev, never this checkout: run codex exec with -c to trust the scratch project and --dangerously-bypass-hook-trust. Confirm that the UserPromptSubmit nudge reaches the model, forcing a due state with enough commits, and that the SessionStart compact digest appears after compaction; if codex exec cannot compact on demand, say so and record what was verified. Then document: README (an agent support section; Codex users trust the project and approve hooks in /hooks; Grok needs folder trust; Cursor and Grok get AGENTS.md and the post-commit nudge but no compaction or per-message hooks), docs/design.md (the After compaction and At each user message paragraphs per agent; [agent_hooks]), docs/status.md (implemented vs verified per agent; Windows caveats), .mem/structure.md (internal/agenthooks, the data flow). Keep the distinction between verified and unverified behavior.
