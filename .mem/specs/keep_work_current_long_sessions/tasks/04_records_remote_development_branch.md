---
title: Records on the remote development branch
status: todo
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:07:22+02:00"
---

On a branch other than development, record commands (spec, task, todo, log, memory) read records from origin/<development> after a fetch and write each change as a commit built directly on origin/<development> with Git plumbing (temporary GIT_INDEX_FILE read from that tree, changed record files added, write-tree, commit-tree), pushed to development as a fast-forward with a lease; no branch switch, the feature branch's files untouched. On a rejected push, fetch and rebuild once, then report. The structure doc stays on the branch with its code. The branch's .mem/ copy catches up at its next rebase onto development. Tests: claiming a todo on a feature branch creates a commit on origin/dev, leaves the working tree unchanged, and todo list on the branch shows origin/dev's records; a concurrent push to dev is handled. Update docs and the structure doc.
