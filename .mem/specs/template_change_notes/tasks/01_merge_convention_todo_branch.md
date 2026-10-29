---
title: Merge the convention-todo branch
status: completed
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:31:27+02:00"
completed_at: "2026-10-29T08:31:27+02:00"
---

Review the diff of branch claude/distracted-khayyam-892c14 (commit 8d75236: convention todos during a pending setup, and the unadopted-tooling wording in internal/cli/template_conventions.go, onboard.go, sync.go, template_test.go, docs and structure doc) against dev. Merge it into dev with --no-ff, resolving conflicts with anything dev gained since (e.g. .mem/structure.md). Run go test ./internal/cli/ -run 'Template|Convention|Onboard|Sync' and go vet ./internal/cli/. Delete todo review_merge_convention_todo_branch and commit its removal with the merge. Do not push the claude/ branch; mem sync pushes dev. Remove the branch's worktree and local branch only if the user agrees.

## Completion Notes

Reviewed and merged claude/distracted-khayyam-892c14 (8d75236) into dev with --no-ff; no conflicts. Deleted todo review_merge_convention_todo_branch in the merge commit. go test ./internal/cli/ -run 'Template|Convention|Onboard|Sync' and go vet ./internal/cli/ pass.
