---
title: Let mem init start from an empty repository
status: todo
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:56:30+02:00"
---

In internal/cli/init.go ensureBranches: when HEAD does not resolve (unborn branch, e.g. a fresh clone of an empty GitHub repo), run `git symbolic-ref HEAD refs/heads/<production>` and `git commit --allow-empty -m 'Initial commit'`, add a line 'Created the first commit on <production>' to the branch lines, then continue as usual (staging/dev created and published, switch to dev). Remove the 'has no commits yet' error. Test in init_test.go with a clone of an empty bare repo: production has one commit, staging and dev exist locally and on the remote, current branch is dev. Delete todo let_mem_init_start_empty when done.
