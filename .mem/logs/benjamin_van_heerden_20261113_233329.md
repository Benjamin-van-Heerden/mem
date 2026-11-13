---
created_at: "2026-11-13T23:33:29+02:00"
user: benjamin_van_heerden
commits: [f36534d, 0148bb7, dc295c2, a476507]
---

# Work Log - mem v0.15.0 released; todo priorities

## Goal

Release the guide, checks and runnables changes as v0.15.0, check that folder and single-branch projects were handled, and add todo priorities, with the runnables review as a critical todo.

## What Was Accomplished

### Windows CI fix and folder audit

- `f36534d`: Windows CI failed the upgrade test. On Windows the schema 3 patch lists every file in `.mem/runnables/`, by design, because there is no executable bit. The test now expects that there (`notes.txt` listed).
- The audit found the folder block still calling runnables "small executable scripts". `0148bb7` rewrote that line for `[runnables]`, updated the `mem template --help` text and made the review todo's wording precise.
- Already in place: `check` refused in folder projects, the guide pointer and guide writing there, and the runnables upgrade running there. In single-branch projects the instructions name `task complete`/`spec complete` as running the checks, and those commands do.
- The tool configuration rule stays out of the folder block on purpose.

### Release v0.15.0

- Staging at `0148bb7`. CI and Windows passed on the earlier `f36534d`, and the `0148bb7` changes are wording only.
- `mem promote production --version v0.15.0 --confirm`, with notes on schema 3, the guide, checks, runnables and the tool rule. The Release workflow published 7 assets, and `mem update` installed v0.15.0.

### Todo priorities (dc295c2)

- `work.TodoMeta.Priority` (`priority`, omitempty) is one of `work.Priorities` (`critical`, `high`, `medium`, `low`). `Todo.Priority()` defaults to `medium`. `Todos` sorts by priority rank, then creation time. `NewTodo` takes a priority, and `SetTodoPriority` changes it.
- CLI:
  - `todo new --priority`, validated with an `unknown priority` error.
  - The new `todo priority <todo> <p>` (`high → critical` output).
  - A PRIORITY column in `todo list` and onboard, `Priority:` in `todo show`, and the priority shown in `log new`'s todo list.
  - `log new` suggests `--priority`.
- Onboard: `writeContext` collects critical todos into `contextState.critical`. `criticalTodoInstruction` comes after the template-todo step and before setup, in both `renderOnboardInstruction` and `renderFolderInstruction`. It asks for the critical todos first, even when the user has already named other work, unless the user decides otherwise.
- Managed instructions (both blocks): agents choose the priority unless the user names one, with one-line meanings for each level.
- Schema 4: the schema 3 todo is now written with `priority: critical`. `prioritiseRunnablesReview` adds `priority: critical` after the title of an existing `review_runnables_config.md` that v0.15.0 wrote without one.
- Tests:
  - `TestTodoPrioritiesOrderTheTodosAndCriticalOnesLeadOnboard`.
  - `TestOnboardMakesARunnablesReviewFromSchema3Critical`.
  - The schema 2 upgrade test now expects the critical onboard step. It moved to `schema_upgrade_test.go`.
  - The folder test (`a476507`) covers the critical step in a folder project.
- Docs: the guide's Todos section, `docs/design.md`, `docs/status.md` and `.mem/structure.md`.

### Verified

Focused tests, then the full `go vet`/`go test` run. Both passed.

## Decisions

- **No migration for existing todos.** A missing priority reads as `medium`.
- **The schema 4 patch exists only for the review todo.** v0.15.0 shipped that todo without a priority, and the user wants it critical.
- **Template convention and change-note todos stay at the default priority.**
- **Critical todos come before a pending setup** in the onboard instruction, since setup stops the instruction early.

## Key Files Affected

- `internal/work/todo.go`, `internal/cli/todo.go`, `internal/cli/onboard.go`, `onboard_context.go`, `folder.go`, `log.go`
- `internal/project/patches.go`, `project.go`
- `internal/agentsmd/instructions.md`, `instructions_folder.md`, `readme.md`
- `internal/cli/template.go`, `template_conventions.go`
- Tests: `todo_test.go`, `schema_upgrade_test.go` (was `schema3_test.go`), `folder_test.go`, `log_test.go`, `internal/work/work_test.go`
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`

## Notes

- This repository's own config is still schema 2 with `[precommit]`. The installed v0.15.0 reads it, and the next onboard upgrades it.
- Release notes went out as drafted after the user first chose to edit them and then released without changes.
