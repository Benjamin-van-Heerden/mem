---
title: mem log commit
status: completed
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:10:13+02:00"
completed_at: "2026-09-28T09:10:13+02:00"
---

Add log commit [<log>] to internal/cli/log.go (logic that is not CLI-specific goes in internal/work or internal/converge). Refuse while {placeholders} remain. Commit changed .mem/ paths with git.CommitPaths as 'Work log: <title>', never code outside .mem/. Then converge.Sync, push when ahead (git.PushTimeout), and report under SHARED CODEBASE with a ⚠️ for uncommitted work outside .mem/. State-specific instruction as in the spec. Tests with a bare remote: placeholders refused; a filled log is committed, rebased onto a teammate's pushed commit and pushed so the branch equals its upstream; uncommitted code stays uncommitted and is reported.

## Completion Notes

mem log commit [<log>] (internal/cli/log.go) closes a session: it picks the given log or the user's newest (work.LatestLog), refuses while template placeholders remain (work.Log.Unfilled compares against the template's placeholder blocks), commits changed .mem/ paths as 'Work log: <heading>' with the new git.Commit (git.CommitPaths now builds on git.Commit and git.Push), runs converge.Sync, pushes when ahead, and reports under SHARED CODEBASE: pushed commits, nudges, and a ⚠️ for uncommitted or untracked files outside .mem/ (ignoring .DS_Store). The instruction says the session is closed only when nothing needs attention. Verified by TestLogCommitRefusesAnUnfilledLog and TestLogCommitCommitsRecordsSyncsAndPushes (rebases onto a teammate's pushed commit, pushes so dev equals origin/dev, leaves scratch.go uncommitted and reports it, then confirms a clean close), plus a manual run in a scratch repo. go vet clean.
