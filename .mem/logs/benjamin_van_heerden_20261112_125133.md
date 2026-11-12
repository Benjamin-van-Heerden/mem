---
created_at: "2026-11-12T12:51:33+02:00"
user: benjamin_van_heerden
commits: [c2f45ab]
---

# Work Log - Onboard follows a development branch deleted from the remote

## Goal

Stop onboard from reviving a deleted development branch. EVH-Platform moved to single-branch mode on `main`, and `dev` and `test` were deleted from GitHub. A clone still on `dev` would keep its old config, and onboard told the agent to publish `dev` again with `git push -u origin dev`.

## What Was Accomplished

### Reproduced first

The repro used a bare copy of EVH-Platform in the scratchpad, with a clone on `dev` and `dev` then deleted from the copy. mem v0.14.0's onboard reported "dev is not on origin yet … publish it", printed release status for `dev`/`test`, and committed the template sync on the dead branch.

### Deleted upstreams in converge

`Report.Gone` names the upstream a branch tracked when that remote branch was deleted. `goneUpstream` reads it with `git for-each-ref --format=%(upstream:short) %(upstream:track)` and looks for `[gone]`; `rev-parse @{upstream}` fails in that state. `standingNudges` words it as deleted. For the development branch the nudge says not to publish it again unless the user asks. For a feature branch it says to switch to development and delete the branch, or to ask before republishing.

### leaveDeletedBranch in onboard

This step runs after `converge.Sync` in `onboard.go` and acts only when the development branch's upstream is gone.

- `movedDevelopment` reads `.mem/config.toml` at `refs/remotes/<remote>/HEAD` through the new `project.ParseConfig` (split out of `readConfig`). It takes that config's development branch when it differs from the local one and exists on the remote.
- `unpublishedCommits` (`rev-list --count HEAD --not --remotes=<remote>`), now shared with `leaveGuardedBranch`, must be 0 for the switch.
- On a switch, onboard reloads the project and runs `converge.Sync` again. Otherwise it is `stuck`, like a failed `leaveGuardedBranch`: only `hooks.Sync` runs, with a ⚠️ line for the user.

### Verified

There are two new tests in `onboard_branch_test.go` (`devDeletedClone`: `sharedProject` moved to single-branch on `main`, then `dev` deleted). One covers the switch, with no `dev` published and no RELEASES section. The other covers a local commit on `dev`, where nothing is committed or published. Both fail without the `onboard.go` change. A `dist/mem-dev` rerun of the EVH repro switched to `main`. `mem precommit` passed.

## Decisions

- **The new development branch comes from the remote default branch's config**, not from guessing. The checkout's own config still names the deleted branch.
- **Switch only when no commit would be stranded**; otherwise commit nothing. Onboard's commits could only be shared by publishing the deleted branch again.
- **The local old branch is left in place.** The switch line tells the agent it can delete it with `git branch -D`.

## Key Files Affected

- `internal/cli/onboard_branch.go`: `leaveDeletedBranch`, `movedDevelopment` and `unpublishedCommits`
- `internal/cli/onboard.go`: calls it after the sync, reloads and syncs again after a switch
- `internal/converge/converge.go`: `Report.Gone`, `goneUpstream` and the nudges
- `internal/project/project.go`: `ParseConfig`
- `internal/cli/onboard_branch_test.go`: the two scenario tests
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`: onboard step and converge report described
- Outside this repository: EVH-Platform became single-branch on `main` (its own work log covers that), and its `dev` and `test` branches were deleted from GitHub

## Notes

- `gofmt -l` still lists `internal/checkpoint/checkpoint.go`. That predates this work and was left alone.
