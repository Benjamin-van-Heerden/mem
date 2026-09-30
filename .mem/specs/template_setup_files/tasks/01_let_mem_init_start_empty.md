---
title: Let mem init start from an empty repository
status: completed
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:57:49+02:00"
completed_at: "2026-09-30T11:57:49+02:00"
---

In internal/cli/init.go ensureBranches: when HEAD does not resolve (unborn branch, e.g. a fresh clone of an empty GitHub repo), run `git symbolic-ref HEAD refs/heads/<production>` and `git commit --allow-empty -m 'Initial commit'`, add a line 'Created the first commit on <production>' to the branch lines, then continue as usual (staging/dev created and published, switch to dev). Remove the 'has no commits yet' error. Test in init_test.go with a clone of an empty bare repo: production has one commit, staging and dev exist locally and on the remote, current branch is dev. Delete todo let_mem_init_start_empty when done.

## Completion Notes

ensureBranches creates an empty first commit on the production branch when HEAD is unborn and the remote has no production branch (firstCommit: write-tree against a temporary empty index, commit-tree, update-ref, symbolic-ref), so staged files stay staged; reported as 'Created an empty first commit on <production>'. TestEnsureBranchesStartsAnEmptyCloneWithAnEmptyCommit covers an empty clone with a staged file; dist/mem-dev init in a fresh clone of an empty bare repo created and published main/test/dev and switched to dev. Todo let_mem_init_start_empty deleted.
