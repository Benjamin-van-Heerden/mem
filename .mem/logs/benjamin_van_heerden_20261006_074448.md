---
created_at: "2026-10-06T07:44:48+02:00"
user: benjamin_van_heerden
---

# Work Log - Completion pushes, Urbion-AI migration, mem v0.7.0 and v0.7.1

## Overarching Goals

Work through the open todos (record completion that commits and pushes, release-notes half-sentences), then prove the Python-harness import on a long-running project: Urbion-AI (a year of Agent Core use). Iterate on a testbed copy until the migration works, fix what it surfaces in mem, release, and migrate the real repository with the user.

## What Was Accomplished

### Release notes overview sentences (c7ace4a)

- `firstSentence` in `internal/release/notes.go` returns "" when the first Overview sentence ends with ':', so such specs are listed by title only. Test fixture added in `notes_test.go`.

### Task and spec completion commit, sync and push (cbb685b)

- `converge.Push`: after a successful fetch, pushes a branch that is ahead and not behind (never staging or production), replacing the unpushed nudge with "Pushed N commit(s)" or a failure nudge. Used by `log commit` (replacing its inline push), `mem sync`, `task complete` and `spec complete`. The compaction hook (`catchUp` alone) still never pushes.
- `internal/cli/sync.go`: `share` (catch-up, push, report, returns instruction lines) and `commitRecord` (commits a record on its own, warns when other files stay uncommitted). `task complete` commits the spec directory as "Complete task <slug>", `spec complete` commits the old and archived directories as "Complete spec <slug>"; both print the report and 📥 INCOMING before the instruction. `uncommittedFiles` extracted in `log.go`.
- `spec start` instruction, `internal/agentsmd/instructions.md` (Staying in Sync, task/spec complete lines), `docs/design.md`, `docs/status.md` and the structure doc updated. Tests: `converge` push after rebase, `cli/task_test.go` (teammate push during task and spec completion, archived spec reaches the teammate, sync pushes).
- User decision: `mem sync` pushes when safe.

### Urbion-AI testbed and the fixes it surfaced

- Studied `.agent_core/`: 22 completed specs (119 tasks), 8 todos all under `todos/claimed/` with GitHub issue links, 7 memories, 120 logs (62 titled `# Work Log - …`, 39 with a plain `Work Log - …` line, 19 untitled with top-level section headings), users `benjamin_van_heerden`, `ubuntu`, `assistant`, `claude`; config with worktree symlink paths; `.gitignore` with legacy-mem lines and `.claude` ignored; `CLAUDE.md` symlinked to `AGENTS.md`.
- Testbed at `~/Documents/Urbtec/Urbion-AI-memtest` (bare clone as stand-in remote plus a working clone), rebuilt by a reset script for four passes: import, the agent's cleanup, commit, push, onboard, then a working day with a teammate clone (todo claim, spec with two tasks, spec complete, todo delete, log new/commit, promote staging, production draft, teammate onboard, compaction digest). Deleted afterwards.
- Fixes (fc7e1fb, 07f94ca, 7ad3d51, 1b81b39):
  - importer reads `todos/claimed/*.md` and `specs/abandoned/*/spec.md` (both were silently dropped); `withIssue` keeps `issue_url` as "GitHub issue: <url>"; the harness placeholder description is dropped; `titledLog` gives every log a `# Work Log - <title>` heading (plain title line promoted; untitled logs named "Session of <date>" with headings outside code fences moved down a level when the log used `# ` sections).
  - import instructions: `rm -rf .agent_core` after `git rm` (the ignored `tmp/` was otherwise committed once its ignore rules were dropped) and `git grep -n -e agent_core -e harness/main.py -- ':!.mem'` for leftover references.
  - onboard and `log new` list claimed todos with their claimer (onboard filtered them out despite its CLAIMED BY column; `log new` hid exactly the todos a session usually completes).
  - release notes: `predates` leaves out specs completed and logs created before the range's first commit (imported history listed all 22 specs and 120 logs).
  - `git.Push` errors match `ErrPush` through `pushError` and describe a rejected push as "the remote has commits this checkout does not have yet"; `publish` points to `mem sync`.
- Releases: mem v0.7.0 (date tag v2026.10.04.1) with the completion, todo, import and notes changes; `docs/status.md` brought up to date first.

### Real Urbion-AI migration (Urbion-AI 87ce92e and follow-ups, pushed to origin/dev)

- `mem import agent-core` with v0.7.0; removed `.agent_core/`, `CLAUDE.md` and the legacy-mem `pre-merge-commit` hook from `.git/hooks`; `.gitignore` ignores only `.claude/settings.local.json` and lost its legacy-mem and Agent Core lines; `pyproject.toml` lost the `.agent_core` ruff/ty excludes; `docs/handover.md` rewritten for mem (install, onboard, `mem promote staging`/`production`, `mem deploy`, hooks) and README/structure doc wording updated.
- Todo review against the code: deleted 7 (sender domains, payer throttle, trash dedup, non-releasable review cases, multi-month anomalies, old task runner, Windows firewall); kept `profile_memory_on_deployed_server`; added `smartpop_retry_hygiene_when_vend` and `smartpop_more_descriptive_vend_error`. The latest log's What Comes Next items needed no todos.
- Removed the What Comes Next sections from the 120 imported logs (diff checked to contain only removals and trailing-whitespace trims).

### mem v0.7.1 (27bdf3d, 20ee211; date tag v2026.10.05.1)

- `removeClaudeLink` (`internal/cli/root.go`) deletes a `CLAUDE.md` that is a link to `AGENTS.md` at `init` and `import`; a `CLAUDE.md` with content stays.
- Onboard does not publish `.claude/settings.json` when `.gitignore` ignores it and says the hook stays on this machine; `settingsIgnored` and `sharedSettingsWarning` replace the earlier warning helper.
- Import removes each log's "What Comes Next" section (`withoutNextSteps`, fence-aware), shows the newest one under 🔜 WHAT COMES NEXT and instructs the agent to review it and the imported todos with the user.
- Release workflow succeeded with 7 assets; installed mem updated v0.6.2 → v0.7.0 → v0.7.1.

## Decisions

- `mem sync` pushes after a successful fetch when the branch is ahead and not behind; staging and production are never pushed by it.
- mem-templates is treated as an extension of mem, so its todo (Rust GPUI and Go templates) stays in this repository; the praxis-app todo was deleted here because it belongs in praxis-app.
- `.claude/settings.json` is committed and only `.claude/settings.local.json` ignored (Claude Code's convention); onboard installs the compaction hook and Git hooks in every checkout either way.
- No `CLAUDE.md` in mem projects; a link to `AGENTS.md` is removed automatically.
- Urbion-AI keeps `production_pr` off (the default): production releases use drafted notes confirmed by the user.
- Imported logs lose their "What Comes Next" sections; the import is the point where open items are reviewed into todos.
- A log from 2025-12-11 that uses bold text instead of headings keeps its next-steps list; it is too old to appear in onboard.

## Key Files Affected

- `internal/converge/converge.go` (`Push`), `converge_test.go`.
- `internal/cli/sync.go` (`share`, `commitRecord`), `task.go`, `spec.go`, `log.go` (`converge.Push`, `uncommittedFiles`, claimed todos in `log new`), `onboard.go` (claimed todos), `onboard_updates.go` (ignored settings), `import.go` (notes, 🔜 section, numbered steps, `settingsIgnored`), `init.go`, `root.go` (`publish` message, `removeClaudeLink`); tests `task_test.go` (new), `log_test.go`, `onboard_test.go`.
- `internal/importer/importer.go` (`claimed/`, `abandoned/`, `withIssue`, `titledLog`, `withoutNextSteps`, placeholder description, `Summary.LatestLog/NextSteps`), `importer_test.go`.
- `internal/release/notes.go` (`firstSentence` colon rule, `predates`), `notes_test.go`.
- `internal/git/git.go` (`pushError`), `git_test.go`.
- `internal/agentsmd/instructions.md`, `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md`.
- Todos: deleted `let_task_spec_completion_commit`, `release_notes_draft_quotes_half_sentences`, `turn_praxis_apps_open_log_items_into_todos`.
- Tags: v2026.10.04.1 and v0.7.0, v2026.10.05.1 and v0.7.1.

## Errors and Barriers

- zsh does not word-split a variable holding `env VAR=1 cmd`; use a shell function (`m() { MEM_NO_UPDATE=1 …/mem-dev "$@"; }`).
- `git rev-parse 'v0.7.0^{}'` failed in this shell; `git ls-remote --tags origin` confirmed the tags instead.
- Asking the user about `.gitignore`, the handover doc and the old hook without first restating where the migration stood confused them; the state and each question's context had to be explained before they could decide.
