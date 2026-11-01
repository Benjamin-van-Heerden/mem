---
created_at: "2026-11-01T15:13:14+02:00"
user: benjamin_van_heerden
spec: folder_mode_mem_without_git
---

# Work Log - Folder mode: mem without Git; CI on test and pull requests only; slim-block todo dropped

## Goal

Give mem a mode for plain folders in iCloud Drive with no Git or GitHub, keeping automated context and memories, for the user's `work` and `finance` folders, which are not codebases. Before that: drop the todo to slim the managed `AGENTS.md` block and stop running CI on every push.

## What Was Accomplished

### Slim-block todo dropped (cc16a3a)

Traced `slim_managed_agents_md_block` to the agent hook research of 2026-10-25: Codex reads only the first 32 KiB of `AGENTS.md`. mem's block comes first, so a cut only ever reaches memories, and `agentsSizeWarning` plus `project_doc_max_bytes` handle that. The todo is deleted, and `docs/cloud_agents.md` no longer suggests moving the general principles out.

### Instruction tidy-ups (0eba771)

In `internal/agentsmd/instructions.md`: the working-tree guideline moved to Working Guidelines, error handling to Code Quality Standards, Source Formatting became its own section, and a repeated sentence about work logs went.

### CI (4f3e3b7)

`.github/workflows/ci.yml` now runs on pushes to `test` and on pull requests only. Pushes to `dev` and `main`, and tags, run nothing. `.mem/structure.md` and the release process say to wait for the run on `test` after `mem promote staging`, before `mem promote production`. Confirmed on GitHub that the push of this commit started no run. The `test` trigger has not fired yet.

### Folder mode (spec `folder_mode_mem_without_git`, 7 tasks)

- **iCloud check:** a throwaway folder under `~/Library/Mobile Documents/com~apple~CloudDocs`, tested with `brctl evict`. Dot-folders upload, executable bits survive eviction and download, a `stat` or walk leaves an evicted file evicted, and only reading its content downloads it. A real `claude -p` run on that path read `AGENTS.md` and ran a `UserPromptSubmit` hook. Details are in the spec's Notes.
- **`internal/project`:** `Config.Mode` (`ModeFolder`), `Config.User`, `Folder()`. `Load` walks up to `.mem/config.toml` and returns a folder project without running Git; everything else goes through `git.Toplevel`. `validate` rejects `[git]`, `[release]` or `[templates]` in a folder config, and a missing `user`. `project.User` now takes the `Project`.
- **Refusals:** `refuseInFolder` as the root `PersistentPreRunE` returns `project.ErrFolder` for `sync`, `promote`, `deploy`, `import`, `template`, `spec`, `task`, `log commit` and the Git hooks.
- **Instructions:** `internal/agentsmd/instructions_folder.md` (6.6 KB) is a block of its own. `agentsmd.Kind` (`Staged`, `SingleBranch`, `Folder`) replaces the `single` flag, and `instructionKind` picks it from the config.
- **Change detection:** `structure.ChangedSince` over the stat-only `folderFiles` walk, which counts everything except dot-paths, the agent instructions and `[structure] ignore` patterns. `checkpoint.FolderSince` and `FolderCover` measure from the user's newest log or the config's time, with the existing 10 and 20 file thresholds. `Work.Folder` words the lines without a commit step. `workSince` serves `hook prompt`, `stop` and `compact`, and `compact` prints `writeFolderDigest`.
- **`mem init --folder [--user]`** and **`onboardFolder`**, in `internal/cli/folder.go`. `onboard.go` passed 500 lines, so the context rendering moved to `onboard_context.go`, with `emitContext` and `writeSpecs` extracted.
- **Wording:** `todo delete`, `memory set`, `log new` and `structure` have folder variants, and `mem structure` writes `FolderTemplate` and treats the doc as kept by hand.

Verified: `internal/cli/folder_test.go` runs against a fake `git` that fails the test if called, and `TestFolderCommandsNeverTalkAboutGit` checks the output of a whole session. `dist/mem-dev` ran through init, todo, memory, structure, onboard, log, refusals and hooks in a real iCloud folder, since deleted. `go test` on project, agentsmd, checkpoint, structure, work, importer and the relevant cli tests, and `go vet`, pass. The full suite has not run: CI no longer runs on `dev`, so its first full run comes with the next `mem promote staging`.

## Decisions

- **The name lives in the config, not the OS.** The spec first took the OS account's full name, but release builds have `CGO_ENABLED=0`, where `os/user` gives no full name, and a name differing between machines would split one person's logs. `init --folder --user` sets it.
- **A separate instruction file rather than template branches:** the user wants a different `AGENTS.md` for these projects, with little dev work assumed.
- **No rule gating overwrites or deletes:** the user rejected it because it would make work sluggish with stops and starts.
- **No templates, no migration support, no conflict-copy detection, no stray `.git` check, no specs or tasks:** folder projects are hyper-specific, and the user moves them by hand. The user accepted these defaults by starting the spec.
- **No history or backups in mem:** the project rules forbid a separate version-control system. The user chose iCloud Drive and will delete the GitHub repositories.
- **`mem init` does not write `.mem/structure.md`:** as in Git projects, `mem structure` creates it and onboard offers that. This changed the spec's first draft.
- **CI on `test` rather than `main`:** the same saving, but a failure shows up before `main` moves.

## Key Files Affected

- `internal/project/project.go`; `internal/agentsmd/agentsmd.go`, `instructions.md`, `instructions_folder.md` (new)
- `internal/cli/folder.go` (new), `folder_test.go` (new), `onboard_context.go` (new), `onboard.go`, `init.go`, `root.go`, `hook.go`, `compact.go`, `log.go`, `structure.go`, `todo.go`, `memory.go`, `onboard_updates.go`
- `internal/structure/folder.go` (new), `structure.go`; `internal/checkpoint/folder.go` (new), `checkpoint.go`; `internal/work/log.go` (a file with no lines shows no line count, also for binary files in Git projects)
- `.github/workflows/ci.yml`
- `README.md`, `docs/design.md`, `docs/status.md`, `docs/cloud_agents.md`, `.mem/structure.md`, `AGENTS.md` (Developing mem)
- Outside the repository: `~/.claude/projects/-Users-benjamin-Documents-yhat-mem/memory/mem-release-process.md`

## Notes

- `~/Documents` is a plain local folder, so `work` and `finance` are not in iCloud yet, and Optimize Mac Storage appears to be on.
- Not tested: a second device, working offline with evicted files, and whether a symbolic link reaches another device. Skills in folder projects should be real files.
- `work`'s two reference docs in `.agent_core/docs` are `.typ`, and `mem import` and onboard read only `.md`, so they must become skills by hand.
- All 30 Typst documents in `work` import from `src/` (1,701 lines). Inlining all of it would make each document long, so skills should hold simplified versions; old documents keep building only while `src/` stays.
- Installed mem on PATH is still v0.12.1, which does not know folder mode.
