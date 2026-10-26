---
title: Ground rules and branch handling
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-10-26T12:56:41+02:00"
updated_at: "2026-10-26T12:58:18+02:00"
---

## Overview

In studii-next, Cursor cloud agents ignored mem: they opened PRs against production, hand-wrote migrations, hand-edited generated data, used npm in a bun project and ran an unrequested production promotion. The research and the full proposal are in `docs/cloud_agents.md`. The user deferred the cloud-agent parts; this spec covers the rest:

- short ground rules every agent reads first;
- Releases wording that no longer describes setups a project does not use;
- a shared branch that survives several agents pushing at once;
- branches other than the development branch, including long-lived ones like studii-next's `fundi`, which is built on released code.

## Goals

- A **Ground Rules** section at the top of the managed block, covering: onboard first (or stop); development branch only, no PRs or other branches unless asked; never touch staging or production, never force-push; release only when asked in the conversation; never hand-edit generated files, regenerate them on conflict; one package manager.
- The Releases section drops the pull-request and release-notes sentences. `mem promote` already prints those steps for projects that use them.
- Staying in Sync says that work on another branch is based on the development branch.
- `mem sync` (and everything that pushes through `converge.Push`) retries a push the remote rejected because it moved on: fetch, catch up, push again.
- `mem sync` never flattens a local, unpushed merge commit by rebasing it.
- A `[git] base` setting marks a development branch built on released code from another branch (`fundi` with `base = 'main'`). Onboard reports how many of that branch's commits are not merged in yet, and states the merge-only rule.

## Technical Approach

- **Managed block (`internal/agentsmd/instructions.md`):** static text, the same in every project. It is deliberately not rendered from `.mem/config.toml`. A long-lived branch like `fundi` has its own config, so a rendered block would differ between `dev` and `fundi`, and every merge of `main` into `fundi` would conflict in AGENTS.md. Branch-specific guidance goes into onboard's output instead, which is per checkout and not committed.
- **Push retry (`internal/git`, `internal/converge`):**
  - `git.Push` reports a rejection because the remote moved on as a distinguishable error, for example `git.ErrRemoteAhead` wrapped alongside `ErrPush`.
  - `converge.Push` then fetches, runs the same catch-up as `Sync` (fast-forward, or rebase of the local commits), and pushes again, up to 3 attempts.
  - If the catch-up hits a conflict or another nudge, it stops and reports that.
  - `Report.Done` records the rebase, the push and the number of attempts.
- **Local merges (`internal/converge`):**
  - Before rebasing in `update` (diverged branch) and `followDevelopment` (feature branch), check `git rev-list --merges <upstream>..HEAD`.
  - When there are merges, don't rebase. Nudge instead: the unpushed commits include a merge that a rebase would flatten. Tell the user. To bring in the incoming commits and keep the merge, either merge the upstream (`git merge <upstream>`) or redo the merge on top of it, then `git push`.
- **`[git] base` (`internal/project`, `internal/cli/onboard.go`):**
  - `GitConfig.Base string` (`toml:"base,omitempty"`).
  - After the sync, onboard counts `HEAD..<remote>/<base>` and adds an informational line to 🌿 SHARED CODEBASE. It is not a ⚠️ nudge, because the owner decides when to merge.
  - The line says: "<dev> is built on released code from <base>; N commit(s) of <base> are not merged in yet. They come in only when the user asks, by a merge commit: `git fetch <remote> && git merge --no-ff <remote>/<base>`, then `git push`. Never rebase <dev>, never force-push it, and never merge <the main project's development branch> into it."
  - The main project's development branch isn't known from the fundi config, so the wording speaks of "never merge other unreleased work into it".
  - Onboard's agent instruction mentions it only when N > 0.
- Docs: `docs/design.md`, `docs/status.md`, `.mem/structure.md`, README where relevant. Trim `docs/cloud_agents.md` to the research plus the deferred cloud proposal, and record the deferred cloud-agent work as a todo.

## Success Criteria

- `internal/agentsmd/instructions.md` starts the block with `## Ground Rules`, covering the six rules above. The Releases section no longer mentions pull requests or release notes. Staying in Sync names the development branch as the base for other branches.
- A test pushes to a shared remote that moved on after the fetch. `mem sync` (through `converge.Push`) ends with the commits pushed and reports the retry. A conflicting concurrent change still ends with a nudge and no lost commits.
- A test with an unpushed local merge commit on a diverged branch: after `mem sync`, the merge commit is still in HEAD's history (`git rev-list --merges` is non-empty), and a nudge explains how to push it.
- A test with `[git] base = 'main'`: onboard on that development branch reports the count of unmerged commits from `origin/main` and the merge-only wording. Without `base`, onboard prints neither.
- `go test` for `internal/agentsmd`, `internal/converge`, `internal/git`, `internal/project` and `internal/cli` (the focused tests) and `go vet` pass.

## Notes

- The user confirmed the branch rule on 2026-10-26: work branches follow the development branch; long-lived branches take in `main` only by merge commit, never by rebase or force-push.
- Slimming the managed block (todo `slim_managed_agents_md_block`) is deliberately out of scope. The Ground Rules add about 1 KB.
- Cloud agents (environment setup, detection, refusing releases there, `MEM_USER`) are deferred: the user finds cloud agents not yet ready for broad use.
- Template-specific rules, such as generating Drizzle migrations and regenerating them on a journal conflict, belong in mem-templates (nextjs-web), outside this repository.
