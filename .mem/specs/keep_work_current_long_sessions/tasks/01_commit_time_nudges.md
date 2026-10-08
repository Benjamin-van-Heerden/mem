---
title: Commit-time nudges
status: completed
created_at: "2026-10-08T13:07:21+02:00"
updated_at: "2026-10-08T13:13:50+02:00"
completed_at: "2026-10-08T13:13:50+02:00"
---

Install a post-commit hook in every project (internal/hooks), independent of protect, running `mem hook post-commit` and always exiting 0; silent when mem is missing. The command reads local state only (no fetch). Lines: (a) on every work commit, the count of the user's work commits since their last work log or completed task, escalating: 1-2 state the count, 3-4 add 'consider writing one now (`mem log new`)', 5+ say 'you should stop and write one now (`mem log new`)'; (b) unpushed commits on the current branch at 3 or more: 'run `mem sync` to share them'; (c) structure.Measure says the doc is stale. Work commits exclude mem's own record commits (claims, task/spec completion, spec start, logs, project file updates); a completed task resets the count like a log. Put the counting in a shared function the compaction digest reuses. Commits mem makes itself set an environment variable the hook honours, so they print nothing. Tests: escalation at 1/3/5, reset after a log and after a completed task, unpushed at 3, silence for mem commits. Update docs and the structure doc.

## Completion Notes

Added internal/checkpoint (WorkSinceLog walks HEAD's non-merge commits until the user's last work log, Complete task/spec commit or mem setup; counts the user's commits that touch paths outside .mem/ and are not onboard's ProjectFilesCommit; LogLine escalates at 3 and 5; CommitLines adds unpushed commits at 3 from cached refs and a stale structure doc). internal/hooks installs post-commit in every project regardless of protect; its script exits at once under MEM_GIT (git.InternalEnv, now set on every Git command mem runs) and when the mem on PATH lacks the subcommand. mem hook post-commit prints mem: lines and always exits 0. Verified: checkpoint test (escalation 1/3/5, record/other-author/project-file commits ignored, reset after a log and a completed task), hooks test (post-commit kept with protect off), cli test (3 commits: suggestion and unpushed lines; silent under MEM_GIT), full internal/cli suite, go vet; and real git commits in a scratch repo: escalating output with the dev build, silence with the installed v0.7.5. Docs: design (Checkpoint nudges), README, status, structure doc.
