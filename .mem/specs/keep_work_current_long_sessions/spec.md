---
title: Keep work current in long sessions and on feature branches
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-10-08T12:54:40+02:00"
updated_at: "2026-10-08T13:07:26+02:00"
---

## Overview

mem keeps a project's records and branches in step at a few moments: onboard, `mem sync`, task and spec completion, `mem log commit`, and the catch-up after Claude Code compacts. Two ways of working slip between those moments.

**Long sessions.** Many sessions never end: work continues across compactions for hours or days. Onboard runs once, `mem log new` is tied to "ending a session" and rarely happens, the compaction hook never pushes, and plain coding never reaches task or spec completion. The result is missing work logs, commits teammates cannot see, and a structure doc that goes stale without a warning. Every migration so far also showed the same decay in records: todos claimed months ago, spec tasks still `todo` after their work shipped, handovers behind the code.

**Feature branches.** On a branch other than the development branch, mem keeps the branch in step with its own upstream and reports how many commits it is ahead of or behind development. It does not say whether that drift will conflict, does not bring development in, and records changed on the branch (a todo claim, a started spec) stay invisible to teammates until the branch merges.

This spec makes mem measure and act at the moments that do happen often, commits and compactions, and makes feature branches follow development continuously so merging back is a fast-forward. It adds no new kinds of record.

## Goals

- Every commit checks, from local state only, whether commits are waiting to be pushed, how much work has gathered since the user's last work log or completed task, and whether the structure doc is stale. The work-log line appears on every work commit and escalates; the others appear only past a threshold. It works in any harness, because it is a Git hook.
- Work logs are written at checkpoints during long sessions, not only when a session ends; nothing in `mem log new`, `mem log commit` or the instructions says a log closes or ends the session.
- The compaction catch-up pushes committed work when that is safe, as `mem sync` does, and asks for a work log when enough has gathered since the last one.
- On a feature branch, mem rebases the branch onto a newer development branch when the rebase applies cleanly, force-pushing it with a lease when the branch is pushed and only the user has committed to it. When the rebase would conflict, or others have committed to the branch, it only nudges, naming the files.
- A checkout whose feature branch was rewritten on the remote catches up by replaying only its genuinely new commits.
- On any branch, work records are read from and written to `origin/<development>`, so claims and specs reach teammates at once and a feature branch never changes records itself.
- Onboard flags records and branches that have probably stopped being true: long-claimed todos, active specs that have not moved, unmerged branches nobody has touched, and how long work has waited unreleased.

## Technical Approach

### Commit-time nudges

- `internal/hooks` installs a `post-commit` hook in every project, independent of `protect` (which keeps governing `pre-push` and `pre-commit`). Like the others it runs `mem hook post-commit` and exits 0 when mem is missing; it always exits 0 itself.
- `mem hook post-commit` reads local state only: no fetch, no network, so a commit stays fast. It prints at most three lines, each only past its threshold:
  - commits on the current branch not on its upstream (from cached refs) at or above a threshold: "N commits on dev are not pushed: run `mem sync` to share them";
  - on every work commit, the number of the user's work commits since their last work log or completed task, escalating: 1 and 2 state the count ("2 commits since your last work log"); 3 and 4 add "consider writing one now (`mem log new`)"; 5 and more say "you should stop and write one now (`mem log new`)". Work commits exclude mem's own record commits (claims, task and spec completion, spec start, logs, project file updates). A completed task resets the count like a log, because its completion note already records what was done and how it was verified;
  - `structure.Measure` reports the doc stale: "The structure doc is stale (N code files changed since it was updated)".
- Commits mem makes itself (record commits, publishing) set an environment variable the hook checks, so they stay silent.

### Logs at checkpoints

- `internal/agentsmd/instructions.md` (Work Logs) and the log template describe a log as the record of a stretch of work, written at checkpoints: when a spec completes, before a release, when the commit nudge says so, and when the user wraps up. The session wording goes: `mem log new`'s "run `mem log commit` to close the session", `mem log commit`'s "Close the session" help and "Tell the user the session is closed", its "before the session ends" lines, and the instructions' "End a session with `mem log new`".
- The compaction digest (`internal/cli/compact.go`) reports work commits since the user's last work log or completed task, with the same escalating wording.

### Compaction pushes

- The compaction hook runs `converge.Push` after its catch-up, under the same conditions as `mem sync`: the fetch succeeded, the branch is ahead and not behind, and it is not staging or production. Uncommitted work is never touched. The design doc's "never commits or pushes" line becomes "never commits; pushes committed work when safe".

### Feature branches follow development

- `converge.Sync`, on a branch other than development, staging or production, after its existing upstream catch-up and with a clean working tree: if `origin/<development>` has commits the branch lacks, attempt `git rebase origin/<development>`. On the first conflict, `git rebase --abort`, then nudge with the conflicting files and the command to do it by hand.
- After a successful rebase:
  - a branch with no upstream, or whose upstream has no commits outside development, needs nothing more;
  - a pushed branch whose commits outside development are all by the user (author email, as Git identity) is pushed with `--force-with-lease=<branch>:<the upstream commit just fetched>`;
  - a pushed branch with commits by others is not rewritten: undo the rebase (reset to the pre-rebase commit) and nudge that the branch is shared.
  - The report says "rebased onto dev (N new commits); run the tests before continuing", because code can break without a Git conflict.
- When the upstream of the current branch was rewritten (its old commit is no longer an ancestor of the new one), `update` replays only the local commits after the fork point (`git rebase --fork-point`) instead of every local commit.
- Development, staging and production are never rebased or force-pushed.
- Spec completion on a feature branch instructs the agent to merge the branch into development now.

### Records live on the remote development branch

- On a branch other than development, record commands (specs, tasks, todos, logs, memories) read records from `origin/<development>` after a fetch, and write each change as a commit built directly on `origin/<development>` with Git plumbing (a temporary index read from that tree, the changed files added, `write-tree`, `commit-tree`), pushed to development as a fast-forward with a lease. No branch is switched and the feature branch's files are not touched.
- The feature branch's copy of `.mem/` catches up when it is next rebased onto development; since the branch never changes records itself, that rebase cannot conflict there.
- The structure doc is not a record: it describes the branch's code and is updated on the branch with that code.
- If the push to development is rejected (someone pushed meanwhile), fetch and rebuild the commit once, then report the failure.

### Onboard staleness

- Open todos claimed more than 30 days ago are flagged in the todo table, with an instruction to check each against the code and the user.
- Active specs with no completed task and no change to their record for 14 days are flagged the same way.
- Remote branches other than development, staging and production that are not merged into development and have no commit in 14 days are listed with age and last author: merge, delete or keep, with the user.
- Release status adds the age of the oldest commit on development that is not on staging.

## Success Criteria

- Each work commit prints the escalating work-log line (1, 3 and 5 commits checked); a commit after a log or a completed task starts again at 1; unpushed commits at 3 and a stale structure doc print their lines; commits made by mem print nothing.
- The compaction hook pushes a branch that is ahead and not behind and reports how many commits have gathered since the last work log.
- In a test repository: a feature branch behind development with no conflicts is rebased; if pushed and authored only by the user it is force-pushed with a lease, and a second clone of that branch with a new local commit catches up with only that commit replayed; a conflicting rebase is aborted with the files named; a branch with another author's commits is not rewritten.
- A todo claimed on a feature branch is a commit on `origin/<development>` immediately, the feature branch's working tree is unchanged, and listing todos on the branch shows records from `origin/<development>`.
- Onboard flags a todo claimed 31 days ago, a stalled active spec, a stale unmerged branch, and shows the age of unreleased work.
- Instructions, `docs/design.md`, `docs/status.md` and the structure doc describe the new behaviour; focused tests and `go vet` pass for the affected packages.

## Notes

Decided with the user:

- A clean development branch is rebased into feature branches; a conflicting one is only a nudge.
- Force-pushing a rewritten feature branch is acceptable, guarded by the lease, the authorship check and fork-point catch-up on other checkouts.
- Long-session and feature-branch work belong in one spec.

- Thresholds: unpushed commits nudge at 3; the work-log line shows on every work commit, suggests a log from 3 and says to stop and write one from 5; todos claimed over 30 days ago; specs and branches with no movement for 14 days.
- Records are read from and written to the remote development branch, not the feature branch.
- mem does not keep compaction summaries.

Claude Code's PreCompact hook cannot reach the model or shape the summary (it can only block compaction), so nothing here runs before compaction.
