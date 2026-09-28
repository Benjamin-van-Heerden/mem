---
title: mem log commit
status: todo
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:02:02+02:00"
---

Add log commit [<log>] to internal/cli/log.go (logic that is not CLI-specific goes in internal/work or internal/converge). Refuse while {placeholders} remain. Commit changed .mem/ paths with git.CommitPaths as 'Work log: <title>', never code outside .mem/. Then converge.Sync, push when ahead (git.PushTimeout), and report under SHARED CODEBASE with a ⚠️ for uncommitted work outside .mem/. State-specific instruction as in the spec. Tests with a bare remote: placeholders refused; a filled log is committed, rebased onto a teammate's pushed commit and pushed so the branch equals its upstream; uncommitted code stays uncommitted and is reported.
