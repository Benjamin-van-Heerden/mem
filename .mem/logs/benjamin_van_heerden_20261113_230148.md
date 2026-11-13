---
created_at: "2026-11-13T23:01:48+02:00"
user: benjamin_van_heerden
commits: [fa0341d]
---

# Work Log - Runnables become [runnables] commands in the config

## Goal

Let runnables use any interpreter, such as Python or bun scripts, by listing them as commands in `.mem/config.toml`, with a migration that keeps existing projects' runnables working.

## What Was Accomplished

### Runnables from config

- `Config.Runnables` is `map[string]string` (`toml:"runnables"`). `runnables.Run` runs each command in name order through the new `internal/shell` package's `Command` (`sh -c`, or `cmd /C` on Windows), which `runChecks` uses too. The limits stay as they were: 15 s, 20,000 characters, and the first stderr line on failure.
- `.mem/runnables/` remains the place for onboard scripts, but a file there runs only when an entry names it. The exec-bit check is gone.

### Migration (folded into schema 3, still unreleased)

- `moveRunnablesToConfig` (patch 2→3) adds `name = '.mem/runnables/<name>'` for each executable file (every file on Windows, which keeps no exec bit, since the upgrade is shared). It records non-executable files separately, because they never ran.
- When it found anything, it writes `.mem/todos/review_runnables_config.md` ("Review the runnables moved into the config"). The todo explains the new model, lists the migrated and skipped files and asks for a review with the user. Onboard publishes it with the config.
- Patches and `project.Upgrade` now take `*Project`. `applyUpdates` and `applyFolderUpdates` pass the onboarding project, so the context built right after an upgrade uses the new config. Before this, the migrated runnable did not run on the first onboard.

### Templates and import

- Installing or updating a template runnable adds `name = '.mem/runnables/<name>'` when the project has no entry by that name, so a project's own command for it is kept.
- `convertOnboardConfig` returns entries. The old harness `[[runnables]]` keep their commands (`30_<slug>`), and the files and tree scripts run as `sh .mem/runnables/<name>`.

### Docs

The guide's Runnables section was rewritten, and `[runnables]` added to its config example and template notes. README, `docs/design.md`, `docs/status.md` and `.mem/structure.md` were updated to match.

### Verified

- `TestOnboardUpgradesASchema2ProjectsChecksAndRunnables` covers the config rename, the migrated entry running, the non-executable file left out, and the todo being shared and readable.
- The folder test now uses a plain `echo` command, which also runs on Windows. The template and import tests assert their entries.
- `mem precommit` passed.
- End to end: a project made by installed v0.14.1 with a Python runnable, onboarded with `dist/mem-dev`, upgraded, ran it and committed the todo.

## Decisions

- **Config is the only way runnables run.** A folder that also auto-ran its files would have meant two mechanisms.
- **Template runnables keep the file-plus-shebang form.** mem adds the entry, so the template format is unchanged.
- **No per-OS commands.** A runnable whose interpreter is missing only prints a ⚠️ line.
- **The migration lives in schema 3**, because schema 3 has not reached production.

## Key Files Affected

- `internal/runnables/runnables.go`, `internal/shell/shell.go` (new), `internal/cli/check.go`
- `internal/project/project.go`, `internal/project/patches.go`
- `internal/cli/onboard.go`, `onboard_updates.go`, `onboard_context.go`, `folder.go`
- `internal/templates/sync.go`, `internal/importer/importer.go`, `internal/cli/import.go`
- `internal/cli/schema3_test.go` (new), `check_test.go`, `folder_test.go`, `template_test.go`, `internal/importer/importer_test.go`
- `internal/agentsmd/readme.md`, `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`

## Notes

- The mem-templates library's `rust` and `rust-gpui` templates each ship a `#!/bin/sh` runnable. These keep working through the entry mem adds when it installs them.
