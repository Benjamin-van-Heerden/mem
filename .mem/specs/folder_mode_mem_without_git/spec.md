---
title: 'Folder mode: mem without Git'
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-11-01T14:51:36+02:00"
updated_at: "2026-11-01T14:55:42+02:00"
---

## Overview

Folder mode is a mem project with no Git and no GitHub: a plain folder in iCloud Drive, used for document and operations work rather than software. mem gives such a folder its automated context, memories, todos and work logs, and iCloud does the syncing. mem never calls the `git` executable in this mode.

It exists for two projects today, `~/Documents/work` (Yhat Solutions documents, invoices and client work in Typst) and `~/Documents/yhat/finance` (bookkeeping and tax records). Neither is a codebase, and in both Git serves only as save points ("checkpoint.", "things"). Those projects move across by hand: mem ships no migration for them.

## Goals

- `mem init --folder` sets up a folder that is not a Git repository, and every folder-mode command works on a machine without Git installed.
- `mem onboard` builds the context from the memories, the structure doc, docs, runnables, todos and work logs.
- Folder projects get their own managed instructions, written for document work with occasional scripts, not a trimmed copy of the coding block.
- Skills are the way a folder project keeps reusable formats. mem reads them from `.claude/skills` and `.agents/skills` as it does today, and installs none.
- Commands that only make sense with Git refuse with one clear line.

## Technical Approach

**Mode.** `.mem/config.toml` holds a top-level `mode = "folder"` and no `[git]`, `[release]` or `[templates]` tables. `Config.validate` rejects a folder config that has them, as it does for a half-applied single-branch config, and `WriteConfig` must not write an empty `[git]`. `project.Project` gets a `Folder()` helper. The mode is always explicit, never inferred from a missing `.git`.

**Finding the project.** `project.Load` walks up from the directory to the nearest `.mem/config.toml`. If that config has `mode = "folder"`, its directory is the root and Git is not touched. Otherwise `Load` carries on with `git.Toplevel` exactly as today, so Git projects behave the same.

**Identity.** `.mem/config.toml` holds a top-level `user`, and in folder mode `project.User` returns its slug. Nothing is read from `git config`. `mem init --folder --user "<name>"` sets it, defaulting to the OS account name. The OS full name is not used: release builds have `CGO_ENABLED=0`, where `os/user` gives no full name, and a name that changed between machines would split the logs of one person.

**Saving records.** Folder-mode commands write the file and stop. `publish`, `commitRecord`, `converge` and the Git hooks are never reached. Stdout must not mention commits, pushes or branches: `todo delete`, `memory set` and `structure` print commit wording today and need folder variants, and instructions must stay grounded in what actually happened.

**Commands.**

| Command | In folder mode |
|---|---|
| `init --folder` | Writes `.mem/config.toml`, `.mem/install.md`, the instructions block and the agent hooks. No `.gitignore`, no branches, no commit. Cannot be combined with `--template`, `--single-branch` or the branch flags |
| `onboard` | Self-update, then the steps below |
| `memory`, `todo`, `log new`, `log list`, `log show`, `update`, `version`, `hook prompt`, `hook compact`, `hook stop` | Work, without Git |
| `structure` | Creates the doc when missing, then says the doc is kept by hand. No drift measurement |
| `log commit`, `sync`, `promote`, `deploy`, `import`, `template`, `spec`, `task`, `hook pre-push`, `hook pre-commit`, `hook post-commit` | Refuse with a single line from one shared error, such as `release.SingleBranchError` is for releases |

**Onboard.** The flow becomes: `autoUpdate`, load, `readKnowledge`, apply the updates that make sense (`agentsmd.Refresh`, `agentsmd.WriteGuide`, `agenthooks.SyncClaude`, `agenthooks.SyncCursor`), `agentsSizeWarning`, then the context: structure doc, `.mem/docs/*.md`, runnable output, open todos and work logs. There is no 🌿 SHARED CODEBASE, 🚀 RELEASES, branch handling, template sync or spec section. The context is written to `.mem/local/onboard.md` when long, as today. `.mem/local/` syncs through iCloud like everything else: it holds nothing that matters and a conflict there does no harm.

**Change detection.** Work-log nudges and the file list in a new log come from file timestamps, not commits. A new function in `internal/structure` (or a small package beside it) walks the folder and returns the files changed since a given time, using `os.Stat` only so that it never makes iCloud download a file. The walk skips dot-files and dot-directories, `.mem/`, `AGENTS.md`, `CLAUDE.md` and `[structure] ignore` patterns. It must not use `structure.Relevant`'s built-in exclusions, which treat Markdown and binaries as not work: here PDFs and Markdown are the work. The baseline is the newest log's `created_at`, or the moment `.mem/config.toml` was created when there is no log.

`checkpoint` gets a folder variant that fills `Work` with a file count and no commits or lines. `Level` is `Suggested` from 10 files and `Due` from 20, and the existing `TurnStartLine`, `TurnEndLine` and `MidTurnLine` word it. Log wording in folder mode ends at "fill it in": there is no commit step. `work.NewLog` gets the changed files for Key Files Affected and an empty `commits`.

**Instructions.** A second embedded file, `internal/agentsmd/instructions_folder.md`, is a block of its own, not a template branch of `instructions.md`. `render` selects by mode, and `Install` and `Refresh` keep working on the `<mem>` tags with the version stamp. It covers: run `mem onboard` first, what onboard gives you, memories, todos, work logs, skills as the source of reusable formats, documents live in the folder and hosted pages do not, and the working conduct the user cares about (the key principle, planning before larger changes, surfacing uncertainty, keeping changes to what was asked, preserving what you did not make, stopping when told to). It holds no Git, branch or release rules, no testing philosophy and no file-size or module-structure rules. It does not forbid or gate overwriting or deleting files: stopping to ask would make the work sluggish. The user reviews this text before it lands.

## Success Criteria

- `mem init --folder` in an empty directory with no working `git` on `PATH` succeeds, and `onboard`, `memory set`, `todo new`, `todo delete`, `log new`, `structure` and `hook prompt` all run there and exit 0.
- No folder-mode code path runs `git`: a test puts a failing `git` first on `PATH` and drives those commands.
- Output of every folder-mode command contains no mention of commits, pushes, branches or releases.
- Each refused command exits non-zero with the shared one-line error and changes nothing.
- A folder project's `AGENTS.md` block comes from `instructions_folder.md`, and `onboard` refreshes it without touching memories or text outside the block.
- Changing 20 PDF or Markdown files in a folder project gives a `Due` nudge from `mem hook prompt`, and a new log clears it. Files under `[structure] ignore` and dot-directories do not count.
- `mem log new` lists the changed files under Key Files Affected.
- Git projects are unchanged: `go test ./internal/cli/ -run 'Onboard|Init|Sync|Log|Single'`, `go test ./internal/project/ ./internal/agentsmd/ ./internal/checkpoint/ ./internal/structure/` and `go vet` on those packages pass.
- `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md` and the `mem init` help describe folder mode.

## Notes

Decided with the user on 2026-11-01:

- iCloud Drive is the storage. mem stays unaware of it apart from never forcing downloads during a change scan.
- No templates: folder projects are too specific for shared templates. This also removes the only other reason to need Git.
- No migration support. `work` and `finance` move by hand, so there is no `import --folder`, no handling of `.agent_core`, and no checks for a stray `.git`.
- The GitHub repositories for `work` and `finance` are deleted, not archived, once the files are safe in iCloud. No backup or history feature is added to mem.
- No "no undo" rule in the instructions.

Recommended in this draft, for the user to confirm:

- Specs and tasks are left out of folder mode. Neither project has ever used them, and todos and logs cover the same ground with far less surface.
- Conflict copies are not detected. One person across devices will rarely hit them, and iCloud keeps both files.
- Work-log nudges stay, based on file counts, because logs in `work` stopped in June without them.

Out of scope, part of the manual `work` migration: Typst house formats as skills (letterhead, layout, invoice), so that each document is self-contained with no imports, and a change to a skill cannot break past documents.

iCloud Drive findings, 2026-11-01, on one Mac with Optimize Mac Storage on:

- `.mem/` and `.claude/` upload like any other folder: `brctl evict` removed their files' local copies and reads brought them back.
- Executable bits survive eviction and download, so runnables and `scripts/build` run as they are. Runnables need no `sh` wrapper.
- `os.Stat` and a `find` walk leave an evicted file evicted; only reading its content downloads it. The stat-only change scan will not trigger downloads.
- A symlink stays in place locally. Whether it reaches another device was not tested, so folder projects should keep real files under `.claude/skills/`. Nothing in folder mode creates symlinks.
- Claude Code run under `~/Library/Mobile Documents/com~apple~CloudDocs/` reads `AGENTS.md` and runs hooks from `.claude/settings.json`; the spaces and tilde in the path cause no trouble.
- Not tested: a second device, and working offline with evicted files.
