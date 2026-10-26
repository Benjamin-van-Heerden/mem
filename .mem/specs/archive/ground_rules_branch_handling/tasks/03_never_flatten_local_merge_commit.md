---
title: Never flatten a local merge commit in sync
status: completed
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T13:02:01+02:00"
completed_at: "2026-10-26T13:02:01+02:00"
---

In internal/converge, before rebasing in update (diverged branch) and in followDevelopment (feature branch), check git rev-list --merges <upstream or development>..HEAD. If there are merge commits, do not rebase: add a nudge saying the unpushed commits include a merge that a rebase would flatten, tell the user, and bring the incoming commits in with a merge (git merge <upstream>) or redo the merge on top of them, then git push. Test: a diverged branch with a local merge commit keeps the merge after Sync and gets the nudge; a diverged branch without merges is still rebased as before.

## Completion Notes

update (diverged branch) and followDevelopment (feature branch) now check for merge commits among the local commits (hasMerges: git rev-list --merges) and, when there are any, leave the branch unchanged with a nudge to bring the incoming commits in with git merge --no-ff (--no-ff so it works under merge.ff=only), then git push, instead of rebasing the merge flat. Tests: TestSyncKeepsALocalMergeCommitInsteadOfRebasingIt and TestSyncLeavesAFeatureBranchWithAMergeUnrebased; go test ./internal/converge, the cli Sync/Branch/Spec tests and go vet pass.
