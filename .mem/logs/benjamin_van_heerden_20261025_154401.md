---
created_at: "2026-10-25T15:44:01+02:00"
user: benjamin_van_heerden
commits: ["4960796", c810112]
---

# Work Log - Worktrees cleaned up, onboard warns in a worktree, mission-control shows the branch

## Goal

The desktop app had been running sessions in Git worktrees on throwaway `claude/work-session-…` branches, which work against mem's single shared checkout on the development branch. Clean up the old worktrees, have onboard flag a worktree, and show the checked-out branch in mission-control's band.

## What Was Accomplished

### Old worktrees removed (about 6.6 GB)

Every `.claude/worktrees` and `.git/worktrees` under `~/Documents` was checked for uncommitted files and for commits not on a remote (`git log <branch> --not --remotes`, then `git cherry` for rebased copies). The only unpushed commits were superseded AGENTS.md stamp refreshes; Urbion-AI's detached worktree's two commits were on `origin/dev` under other hashes. Removed with `git worktree remove` and `git branch -D`: mem (3), ev-reporting (2, plus one orphaned `claude/admiring-gauss-df40bd` branch and two empty leftover folders), ascendis-transform (3), Urbion-AI (1). Left alone: this session's worktree and `~/Documents/Studii/studii-next-fundi`, a hand-made worktree on `fundi` tracking `origin/fundi`. The eight desktop sessions that pointed at the removed worktrees were archived; the app also archived the side session "Reporting app permissions and repo split" with its parent, and Benjamin chose to leave it archived.

Projects whose recent sessions ran in worktrees: mem, ev-reporting, ascendis-transform (Urbion-AI only in an older one). The desktop app's worktree mode is set per project; Benjamin is turning it off.

### Onboard warns in a linked worktree (4960796, c810112)

- `git.MainCheckout(ctx, root)`: `rev-parse --path-format=absolute --git-dir --git-common-dir`; when they differ, the first `worktree` line of `git worktree list --porcelain`.
- `converge.Report.MainCheckout`, set in `inspect`; `converge.Worktree(r)` words the nudge: the main checkout, the branch the work sits on until merged into the development branch, and "Tell the user they are working in a worktree. If they intended it, carry on; if not, they should continue in the main checkout and turn off worktree sessions for this project in their coding agent."
- Only onboard adds it, first among the 🌿 SHARED CODEBASE nudges, so the existing "Tell the user about each ⚠️ item" step covers it; `mem sync` and `log new` do not repeat it.
- `TestOnboardWarnsWhenTheCheckoutIsALinkedWorktree`; Windows CI failed once because Git prints `C:/…`, fixed by comparing with `filepath.ToSlash`. CI run for c810112 green on ubuntu, macOS and Windows. Docs: `docs/status.md`, `.mem/structure.md`.

### mission-control 0.3.5 to 0.3.9 (claude-mods, outside this repository)

- A `branch` state value (`types/index.d.ts`), refreshed with the band every 5 s from `$.process.run(['git', 'branch', '--show-current'], { cwd })`; `''` outside a repository or on a detached HEAD.
- Desktop: this session's pill is now a Box in the primary button's colours (`backgroundColor="text"`, `Text color="inverseText"`) holding the session label and the branch as `<Text mono size="caption" dimColor>`; 0.3.9 drops the round border and keeps `paddingX={0.5}` only. Terminal: the native primary Button with the branch as dim text beside it.
- `tests/branch.test.tsx`: desktop and terminal drawings with a branch, and a pill without one outside a repository.

## Decisions

- The worktree warning is an onboard nudge only, not a standing nudge, so it does not repeat at every `mem sync`.
- Old worktree branches whose only unpushed commit was an AGENTS.md stamp refresh were deleted: later refreshes on `dev` supersede them.
- On the desktop the band draws this session's pill itself, because a native Button cannot style part of its text.

## Key Files Affected

- `internal/git/git.go` (18 lines): `MainCheckout`
- `internal/converge/converge.go` (15 lines): `Report.MainCheckout`, `Worktree`
- `internal/cli/onboard.go` (3 lines): adds the worktree nudge
- `internal/cli/onboard_test.go` (18 lines): worktree test
- `docs/status.md` (1 line), `.mem/structure.md`
- Outside this repository: `~/.claude/mods/mission-control/hooks/register.tsx`, `types/index.d.ts`, `tests/branch.test.tsx`, `.claude-plugin/plugin.json` and the mods `README.md` (commits 89a50fc to 7c7f0b4)

## Notes

- The desktop app (2.1.308 engine) draws a Button's `label` and ignores its children when both are given, and with children only it flattens them into one plain style: `mono`, `size` and `dimColor` inside a Button do not show. The `claude` on PATH is 2.1.281: it refuses a Button with mixed children and does not know `Text mono`, so test the mods with the app's binary (`~/Library/Application Support/Claude/claude-code/2.1.308/…/claude.app/Contents/MacOS/claude plugin test .`).
- A Box with `borderStyle` on the desktop takes about twice a text line's height; a framed chip beside the pill (0.3.7) and a bordered pill (0.3.8) were both too tall.
- `claude plugin test` gives the test no `$.process`; to see what `$.process.run` returns in the app, a temporary probe wrote results to `~/.claude/mc-branch-probe-<session>.json`. Other idle sessions reload a saved mod within seconds (`CLAUDE_CODE_PLUGIN_DIR_WATCH=1`), so their probe files answered within the turn. The probe is removed; a session mid-turn keeps running its old copy until it reloads.
- While the next-step box shows reply buttons, the desktop app does not ask the band to draw at all (found earlier in the ascendis session); a recap-only box leaves the band visible.
- The `mod-ui` project memory still says Text has no size and is proportional; on 2.1.308 Text has `size` and `mono`.
