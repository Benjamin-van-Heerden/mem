---
created_at: "2026-10-06T22:04:35+02:00"
user: benjamin_van_heerden
---

# Work Log - Install guide for agents without mem, mem v0.7.4

## Overarching Goals

Close the gap where an agent without mem installed (a teammate's, or a cloud agent such as Cursor's) reads AGENTS.md, fails to run `mem onboard` and carries on without mem. Prompted by the question of what Marco's agents do in studii-tan after the migration.

## What Was Accomplished

### Install guide (4267a7f)

- `internal/agentsmd/install.md`, embedded and written to `.mem/install.md` by `agentsmd.WriteGuide` (compares content, writes only on change; no version stamp, so releases do not churn it). Covers `go install …@latest`; the release binary for macOS/Linux (tag from the `releases/latest` redirect, `mem_<tag>_<os>_<arch>`, checked with `sha256sum` or `shasum -a 256` against `checksums.txt`, installed to `~/.local/bin`); Windows PowerShell (tag from the GitHub API, `%LOCALAPPDATA%\Programs\mem\mem.exe`, `Get-FileHash` check); PATH for the session, and permanently only after asking the user; then `mem version` and `mem onboard`.
- `init` and `import` write it and list it in their output; onboard's `applyUpdates` writes it and publishes it unless it had uncommitted edits.
- `internal/agentsmd/instructions.md` Getting Started: if `mem` is not found, read `.mem/install.md`, follow it, run `mem onboard`; if it cannot be installed, tell the user before any other work.
- The macOS/Linux steps were run verbatim with `sh` in a scratch HOME: installed v0.7.3 and the checksum passed. The Windows steps were not run.
- Test `TestOnboardGivesTeammatesTheInstallGuide`: a project without the guide gets it at onboard, and a teammate's pull receives it and the AGENTS.md pointer. README (install section, layout), `docs/design.md` layout, `docs/status.md` and the structure doc updated.

### Release and rollout

- mem v0.7.4 (date tag v2026.10.06.3) with the install guide and the `todo list` change (claimed todos always listed, `--all` and `work.OpenTodos` removed). Release workflow built 7 assets; installed mem updated v0.7.3 → v0.7.4.
- studii-tan onboard committed and pushed `.mem/install.md` and the refreshed AGENTS.md (9416d13 on its `dev`).

### Marco's agents in studii-tan (findings)

- On `dev`: AGENTS.md has the mem block; with mem installed, onboard identifies `marco_booyse`; without it, agents now follow `.mem/install.md`. Cursor cloud agents commit as "Cursor Agent".
- `main` and `fundi-main` still carry an older prototype's `AGENTS.md` (also saying `mem onboard`) and `.mem/config.toml`; current mem stops there with "no upgrade path from project schema 0" and changes nothing (checked in a disposable clone). `main` changes when production is next promoted.

## Decisions

- The install steps live in `.mem/install.md`, not AGENTS.md; AGENTS.md carries one pointer line.
- Agents install mem per user without sudo and ask before persistent PATH or shell profile changes; if installing fails they tell the user rather than continuing without mem.
- No version stamp in the guide.

## Key Files Affected

- `internal/agentsmd/install.md` (new), `agentsmd.go` (`GuidePath`, `WriteGuide`), `instructions.md`.
- `internal/cli/onboard_updates.go`, `init.go`, `import.go`, `onboard_test.go`.
- `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`.
- Tags: v2026.10.06.3 and v0.7.4.

## Errors and Barriers

- None in this stretch.
