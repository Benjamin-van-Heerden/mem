---
title: Add init and onboard for folders
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:09:27+02:00"
completed_at: "2026-11-01T15:09:27+02:00"
---

mem init --folder writes .mem/config.toml, .mem/structure.md, .mem/install.md, the instructions block and the agent hooks, with no .gitignore, branches or commit, and refuses --template, --single-branch and branch flags. Onboard in folder mode runs autoUpdate, load, readKnowledge, Refresh, WriteGuide, SyncClaude, SyncCursor and agentsSizeWarning, then prints the context: structure doc, docs, runnables, todos and logs, with no shared-codebase, release, branch, template or spec section, and an agent instruction grounded in that. Make sure no line of onboard runs git. Test end to end in a temp dir with a failing git on PATH.

## Completion Notes

mem init --folder writes the config, install guide, instructions block and agent hooks, with no Git, and refuses Git and template flags; --user names who works there, defaulting to the account name. The structure doc is left to mem structure, as in a Git project. onboardFolder applies updates and prints the context with no shared-codebase, release, spec or staleness sections. applyFolderUpdates keeps a folder's managed files current. onboard.go passed 500 lines, so the context rendering moved to onboard_context.go. TestInitFolderSetsUpAPlainFolderAndOnboardBuildsItsContext and TestInitFolderRefusesGitFlags run with a git that fails the test if called; go test ./internal/cli -run 'Onboard|Init|Sync|Log|Single|Hook|Compact' and go vet pass.
