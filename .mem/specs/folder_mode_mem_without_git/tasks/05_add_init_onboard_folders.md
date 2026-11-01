---
title: Add init and onboard for folders
status: todo
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T14:52:34+02:00"
---

mem init --folder writes .mem/config.toml, .mem/structure.md, .mem/install.md, the instructions block and the agent hooks, with no .gitignore, branches or commit, and refuses --template, --single-branch and branch flags. Onboard in folder mode runs autoUpdate, load, readKnowledge, Refresh, WriteGuide, SyncClaude, SyncCursor and agentsSizeWarning, then prints the context: structure doc, docs, runnables, todos and logs, with no shared-codebase, release, branch, template or spec section, and an agent instruction grounded in that. Make sure no line of onboard runs git. Test end to end in a temp dir with a failing git on PATH.
