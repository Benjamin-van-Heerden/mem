---
created_at: "2026-10-29T10:11:59+02:00"
user: benjamin_van_heerden
commits: [649e4a4]
---

# Work Log - Onboard leaves staging and production before it commits anything

## Goal

Fix the bug the ev-reporting session (local_d6fb92b7) reported on the user's behalf. A fresh clone of a repository whose GitHub default branch is production checks out production, and its first `mem onboard` committed "Update mem project files" there and pushed it. In ev-reporting (nickfols17, mem v0.9.1, 2026-10-26, commit e76c1e3) that ran the production deploy, and left production ahead of development, so promotion could not fast-forward.

## What Was Accomplished

- `internal/cli/onboard_branch.go`: `leaveGuardedBranch` runs right after onboard's `readKnowledge` and before `converge.Sync`. On a guarded branch, unless the project is single-branch or the checkout detached, it:
  - warns about commits on the branch that no remote branch has (`rev-list --count HEAD --not --remotes=<remote>`); they stay on that branch;
  - switches to development: `git switch <dev>`, or `git switch --create <dev> --track <remote>/<dev>` when only the remote's copy exists;
  - reports "Switched from main to dev: …".

  If the switch fails, it prints a ⚠️ ("could not switch … Onboard committed and pushed nothing … resolve it with them") and returns `stuck`.
- `onboard.go`:
  - after a successful switch it reloads the project, since development's config may differ;
  - when stuck, it uses `converge.Local` only, runs `hooks.Sync` (which commits nothing), and skips `converge.Sync`, `applyUpdates` and `syncTemplates`.
- `onboard_updates.go`: `applyUpdates` now runs `hooks.Sync` before its `publish`, so the protection hooks exist before onboard's first commit.
- Tests in `onboard_branch_test.go` (`productionClone`: a shared project with `protect = true` and no `.mem/install.md`, pushed to dev, test and main, then a fresh clone on main):
  - `TestOnboardSwitchesAFreshCloneOffProductionBeforeCommitting`: the clone ends on dev; the project-files commit is on dev and pushed; origin/main and local main are unchanged; a local commit on main gets the ⚠️; the pre-commit hook is installed.
  - `TestOnboardCommitsNothingWhenItCannotLeaveProduction`: an uncommitted README edit that conflicts with dev keeps the clone on main with the ⚠️; main is unchanged, nothing is written besides the README edit, and the hooks are installed.
  - Both tests fail with the onboard changes stashed. `go test ./internal/cli/ -run 'Onboard|Sync|Init|Hook|Template|Compact|Single|Base'` and `go vet` pass, and CI passed on ubuntu, macOS and Windows.
- `docs/design.md` (a new onboard step 0), `docs/status.md` and `.mem/structure.md` describe it.

## Decisions

- **Only onboard switches branches.** `mem sync`, `task complete`, `log commit` and `spec complete` on a guarded branch are covered by the hooks. With the hooks installed first, `protect = true` refuses those commits, and `converge.Push` never pushes a guarded branch anyway. Onboard is the entry point a fresh clone runs.
- **The design doc's new step is numbered 0,** so references to "step 1" (converge) stay correct.

## Key Files Affected

- `internal/cli/onboard_branch.go`, `onboard_branch_test.go` (new); `internal/cli/onboard.go`, `onboard_updates.go`
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`

## Notes

- Not released yet. The fix reaches ev-reporting and other projects only with a mem release, which is the user's call. Until then, nickfols17's ev-reporting checkout needs `git switch development` by hand.
