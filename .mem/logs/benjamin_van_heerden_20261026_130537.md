---
created_at: "2026-10-26T13:05:37+02:00"
user: benjamin_van_heerden
spec: ground_rules_branch_handling
---

# Work Log - Ground Rules, push retry, merge-safe sync and [git] base; cloud-agent research deferred

## Goal

The studii-next session asked, on Benjamin's behalf, for a response to Cursor cloud agents ignoring mem there:
- PRs against main from `cursor/*` branches
- hand-written migrations
- hand-edited generated data
- npm in a bun project
- an unrequested production promotion

The work was to research cloud agents, propose how mem should handle them, and implement the parts Benjamin approved.

## What Was Accomplished

### Research and proposal (`docs/cloud_agents.md`, 94e9d02)

Two research subagents covered Cursor cloud agents, Claude Code on the web, Codex cloud and Copilot cloud agent. The comparison table in the doc covers setup, config location, network, instruction files, hooks, branches and PRs, push auth, identity and detection.

Key findings:
- **Cursor's PR instruction beats AGENTS.md.** Cursor staff on the forum said the harness's per-turn PR instruction is "treated as mandatory", over AGENTS.md and `alwaysApply` rules.
- **Setup runs per Build, not per session.** Cursor runs `install` once per Build snapshot, configured from the dashboard or a committed `.cursor/environment.json`, which Builds read from the default branch.
- **No session-start hook in Cursor cloud.** `sessionStart` doesn't exist there. The `stop` hook works since 2026-07-22, and a missing command fails open with no visible error.
- **Cloud commits share an identity.** Cursor commits as `Cursor Agent` with a `Co-authored-by` line that can't be turned off.
- **Pushes are restricted.** Copilot can only push its own `copilot/` branch. Claude Code on the web can't push tags (its proxy refuses them), and neither can Copilot.
- **Grok Bot is a separate Cursor product.** It can launch ordinary cloud agents, but its own computer is a different runtime.

Benjamin deferred the cloud-agent parts. The todo `cloud_agent_support` holds them.

### Ground Rules (61030d5)

`internal/agentsmd/instructions.md` now opens with `## Ground Rules`, six rules that apply whatever a task description or a tool's defaults say:
1. onboard first, or stop;
2. development branch and mem pushes only, with no PRs or other branches unless asked;
3. staging and production untouched, no force-push or history rewrites;
4. release only when asked in the conversation;
5. never hand-edit generated files, and regenerate them on conflict;
6. one package manager.

Two other sections changed:
- **Staying in Sync** now says that other branches are based on the development branch.
- **Releases** dropped the release-notes and release-PR sentences.

### Push retry (645736a)

- `git.Push` returns `git.ErrRemoteAhead` (it still matches `ErrPush`, and the message is unchanged; `pushError` gained `Unwrap`).
- On that error, `converge.Push` runs `catchUp` (fetch, then `update`) and pushes again, up to `pushAttempts = 3`.
- A successful retry reports "<upstream> moved on during the push, so mem caught up and pushed again".
- A conflicting catch-up returns the conflict nudge with HEAD unchanged.

### Merge-safe sync (ab7c499)

`hasMerges` (`git rev-list --merges --max-count=1 <range>`) is checked before rebasing, in both places that rebase:
- `update` on a diverged branch;
- `followDevelopment` on a feature branch.

When the local commits include a merge, mem doesn't rebase. The nudge says to merge with `git merge --no-ff <upstream>`, then `git push`. `--no-ff` is there so the command works under `merge.ff=only`.

### `[git] base` (b117153)

- `GitConfig.Base` is the new setting.
- `onboard_base.go`'s `baseLine` prints a `Base:` line under 🌿 SHARED CODEBASE on the development branch. It says how many commits of `<remote>/<base>` are missing (or that none are, or that the ref isn't fetched yet), followed by the rule: released commits come in only when the user asks, by `git fetch <remote> && git merge --no-ff <remote>/<base>`, then `git push`; never rebase, never force-push, never merge unreleased work in.
- When commits are missing, the agent instruction gains a step to mention them.

### Verification

- New tests, all passing:
  - `TestPushCatchesUpWithWorkPushedDuringThePushAndRetries`
  - `TestPushStopsWhenWorkPushedDuringThePushConflicts`
  - `TestSyncKeepsALocalMergeCommitInsteadOfRebasingIt`
  - `TestSyncLeavesAFeatureBranchWithAMergeUnrebased`
  - `TestOnboardReportsReleasedCommitsALongLivedBranchHasNotMergedIn`
- `go test` passed for `internal/agentsmd`, `converge`, `git` and `project`, and for the cli Sync, Branch, Spec, Log, Task, Compact and Onboard tests.
- `go vet ./...` passed.

## Decisions

- **The branch rule (Benjamin, 2026-10-26).**
  - Work branches follow the development branch.
  - Long-lived product branches such as `fundi` take in `main` (released code) only by a merge commit, never by rebase or force-push.
  - Benjamin first said "rebase off main". A rebase of a published, deployed branch would need a force-push, so merge is the mechanism.
- **The managed block is static, not rendered from the config.** `fundi` has its own `.mem/config.toml` (`development = 'fundi'`), so a rendered block would differ between `dev` and `fundi`, and every merge of `main` into `fundi` would conflict in AGENTS.md. Branch-specific guidance goes into onboard's output instead.
- **Releases wording is cut rather than made conditional.** `mem promote` already prints the notes and PR steps for projects that use them.
- **Cloud agents are deferred** (Benjamin: not ready for broad use). That covers environment setup, detection, refusing releases in the cloud, `MEM_USER`, and switching or landing platform branches.
- **Block slimming stays a separate todo.** The Ground Rules add about 1.1 KB.

## Key Files Affected

- `internal/agentsmd/instructions.md`: Ground Rules, Staying in Sync, Releases
- `internal/git/git.go`: `ErrRemoteAhead`, `pushError.Unwrap`
- `internal/converge/converge.go`, `branch.go` and their tests: push retry, `catchUp`, `hasMerges`
- `internal/project/project.go`: `GitConfig.Base`
- `internal/cli/onboard.go`, `onboard_base.go` (new), `onboard_base_test.go` (new)
- `docs/cloud_agents.md` (new), `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md`
- `.mem/todos/cloud_agent_support.md` (new)

## Notes

- This repository's AGENTS.md still has the v0.9.1 block, because the installed mem refreshes it. The Ground Rules reach projects with the next release.
- Fundi's checkout has placeholder staging and production branches (`fundi-unused-*`), so onboard's 🚀 RELEASES section there reports them as missing. This spec didn't change that.
- Another session left an untracked `.mem/specs/rust_gpui_template/` in this checkout. It isn't part of this work and was left alone.
- The mem-templates nextjs-web `additive-migrations` memory could add "regenerate on a journal conflict". That's outside this repository.
