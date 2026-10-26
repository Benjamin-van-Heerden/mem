---
title: Never flatten a local merge commit in sync
status: todo
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T12:58:18+02:00"
---

In internal/converge, before rebasing in update (diverged branch) and in followDevelopment (feature branch), check git rev-list --merges <upstream or development>..HEAD. If there are merge commits, do not rebase: add a nudge saying the unpushed commits include a merge that a rebase would flatten, tell the user, and bring the incoming commits in with a merge (git merge <upstream>) or redo the merge on top of them, then git push. Test: a diverged branch with a local merge commit keeps the merge after Sync and gets the nudge; a diverged branch without merges is still rebased as before.
