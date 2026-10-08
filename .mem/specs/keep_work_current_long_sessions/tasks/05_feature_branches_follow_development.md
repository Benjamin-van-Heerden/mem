---
title: Feature branches follow development
status: todo
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:07:22+02:00"
---

In converge.Sync on a branch other than development/staging/production, after the upstream catch-up and with a clean working tree: if origin/<development> has commits the branch lacks, git rebase onto it; on the first conflict abort and nudge with the conflicting files and the manual command. After a clean rebase: no upstream, or no upstream commits outside development, needs nothing more; pushed and every upstream commit outside development authored by the user (author email): push with --force-with-lease=<branch>:<fetched upstream commit>; commits by others: undo the rebase and nudge that the branch is shared. Report 'rebased onto dev (N new commits); run the tests before continuing'. When the current branch's upstream was rewritten (old upstream not an ancestor of the new), replay only commits after the fork point (git rebase --fork-point). Never rebase or force-push development, staging or production. Spec completion on a feature branch instructs merging into development now. Tests: clean rebase; force-push with lease and a second clone catching up with only its new commit replayed; conflicting rebase aborted with files named; shared branch not rewritten. Update docs and the structure doc.
