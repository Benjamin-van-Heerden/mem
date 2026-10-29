---
created_at: "2026-10-29T09:39:22+02:00"
user: benjamin_van_heerden
commits: [f87cf77]
---

# Work Log - Releases fast-forward the local copies of staging and production

## Goal

Stop the local staging and production branches from falling behind the remote. `mem init` creates them, but releases only ever moved the remote copies, so this checkout's `main` and `test` were 17 and 261 commits behind.

## What Was Accomplished

### Cleanup first

Removed the merged convention-todo branch `claude/distracted-khayyam-892c14` and its worktree `.claude/worktrees/work-session-fc7b6c`. The worktree was clean and the branch merged; neither was ever pushed. A desktop-app side session had created both on 2026-10-28. After that, mem and mem-templates each had only their main checkout, and their remotes only `dev`/`test`/`main` and `main`.

### Local stage branches follow releases (f87cf77)

- `release.Execute` now returns `(string, error)`. After the atomic push (now `push`), `followLocal` moves `refs/heads/<branch>` to the released commit with `git update-ref <ref> <new> <old>`. It leaves the local copy alone when:
  - it doesn't exist (nothing is created);
  - it is already there;
  - it is checked out in a worktree (`git worktree list --porcelain`), with a ⚠️ line saying to switch back to development;
  - it has commits the release lacks, with a ⚠️ line.

  On success the line is "The local <branch> branch is at <sha> too."
- `mem promote` (`promote.go`), `mem deploy` (`deploy.go`) and completing a production pull request (`promote_pr.go`) print that line through `printLine` after their "is now at" line.
- Tests:
  - `TestReleasesBringTheLocalStageBranchAlong` in `internal/release` covers each case.
  - `TestProductionReleaseWithoutNotesTagsAGeneratedSummary` in `internal/cli` now checks that both local copies match the remote after promote staging and production, and that the line is printed.
  - `go test ./internal/release/`, `go test ./internal/cli/ -run 'Promot|Deploy|Release'` and `go vet` for both packages pass, and CI passed on ubuntu, macOS and Windows.
- `docs/design.md`, `docs/status.md` and `.mem/structure.md` describe the behaviour.
- This checkout's local `main` and `test` were fast-forwarded once by hand (`git fetch . refs/remotes/origin/<b>:refs/heads/<b>`), since the change only takes effect from the next release.

## Decisions

- **Only after a release, not at every fetch:** the user asked for releases. A teammate's release still leaves another checkout's local copies behind until that checkout releases.
- **A checked-out branch is never moved:** `update-ref` would change the branch under its working tree.
- **No local copy is created:** a project without one keeps having none.

## Key Files Affected

- `internal/release/release.go` (`Execute`, `push`, `followLocal`, `checkedOut`), `release_test.go`, `notes_test.go`
- `internal/cli/promote.go` (`printLine`), `deploy.go`, `promote_pr.go`, `promote_test.go`
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`
- Outside the tree: the worktree `.claude/worktrees/work-session-fc7b6c` and the local branch `claude/distracted-khayyam-892c14` were removed.
