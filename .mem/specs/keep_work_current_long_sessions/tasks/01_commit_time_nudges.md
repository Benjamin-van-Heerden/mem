---
title: Commit-time nudges
status: todo
created_at: "2026-10-08T13:07:21+02:00"
updated_at: "2026-10-08T13:07:21+02:00"
---

Install a post-commit hook in every project (internal/hooks), independent of protect, running `mem hook post-commit` and always exiting 0; silent when mem is missing. The command reads local state only (no fetch). Lines: (a) on every work commit, the count of the user's work commits since their last work log or completed task, escalating: 1-2 state the count, 3-4 add 'consider writing one now (`mem log new`)', 5+ say 'you should stop and write one now (`mem log new`)'; (b) unpushed commits on the current branch at 3 or more: 'run `mem sync` to share them'; (c) structure.Measure says the doc is stale. Work commits exclude mem's own record commits (claims, task/spec completion, spec start, logs, project file updates); a completed task resets the count like a log. Put the counting in a shared function the compaction digest reuses. Commits mem makes itself set an environment variable the hook honours, so they print nothing. Tests: escalation at 1/3/5, reset after a log and after a completed task, unpushed at 3, silence for mem commits. Update docs and the structure doc.
