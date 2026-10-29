---
created_at: "2026-10-29T09:03:58+02:00"
user: benjamin_van_heerden
spec: template_change_notes
---

# Work Log - Template change notes: template changes reach existing projects as todos

## Goal

Give templates a way to tell projects built from an earlier state of them that they must change their own files, which no sync touches. The prompting case was the React Compiler: the nextjs-web template turned it on, but existing projects got only the updated skill.

## What Was Accomplished

### Before the spec

- Read the sync code to answer the user's questions. Only memory changes became todos. Skill updates replaced copies silently, `setup.md` was copied once, and starter files never synced. Nothing was planned for the rest. The user rejected adding a memory just to trigger a todo, and asked for a spec.
- Wrote spec `template_change_notes` with five tasks and pushed the draft.

### Task 1: convention-todo branch merged

`claude/distracted-khayyam-892c14` (8d75236) merged into dev with `--no-ff` (3367809), with no conflicts. Todo `review_merge_convention_todo_branch` was deleted in the merge commit. Its worktree (`.claude/worktrees/work-session-fc7b6c`) and local branch are still in place.

### Tasks 2–3: library and lock (`internal/templates`)

- `changes.go`: `Change` (`Template`, `Name`, `Title`, `Body`; `Key()` is `template/name`) and `Library.Changes(names)`. It reads `<template>/changes/*.md` in template order, then name order, ignores dotfiles and non-`.md` files, normalises CRLF, and takes the title from a first-line `# ` heading, falling back to the name. `reconcileChanges` records every unseen note and lists it in `Result.Changes` unless its template is adopted, and drops records of notes no longer provided.
- `lock.go`: the lock code moved out of `sync.go`, which was at 503 lines and is now 464. It adds `[[change]]` records (`LockChange`: template, name), omitted when empty, so a lock without notes keeps its old layout.
- `Sync(p, lib, adopted ...string)`: `init` passes every template it installs, `template use` the new one, and `syncTemplates` none.

### Task 4: todos (`internal/cli`)

- `template_changes.go`: `changeTodos` records one todo per new note, titled "Apply the <template> template change: <title>". Its body quotes the note, then gives `changeCheck`: apply the change, say so where the project has diverged or needs tooling it never adopted, and delete the todo when the project matches. While `.mem/setup.md` is pending, the todo uses a setup variant. Each todo prints a 📌 line.
- `syncTemplates` calls `changeTodos` after `conventionTodos`. `template use` calls it too: its sync records the unseen notes of the project's other templates, so without the call those notes would be marked seen and lost. This was found while writing task 4, after task 3 was committed.
- The 📌 helpers now serve both kinds of todo: `templateTodoMark`, `hasTemplateTodo`, `templateTodoInstruction` and onboard's `templateTodos`, with the `template use` and compaction digest wording generalised.

### Task 5: docs and the first note

- `docs/design.md`, `docs/status.md` and `.mem/structure.md` describe change notes. The managed instructions are unchanged.
- In mem-templates, eb333eb on `main`, committed but not pushed: a Layout line and a "Change notes" section in the README, and `nextjs-web/changes/2026-10-29-react-compiler.md`. The note covers:
  - Next 16.3 or later;
  - the two `next.config.ts` settings;
  - removing `babel-plugin-react-compiler`;
  - fixing what `bun run lint` reports (`useWatch` instead of `form.watch()`);
  - typecheck and build, and a UI check by the user;
  - that it does not apply to webpack builds.

### Verification

- Focused tests:
  - `TestChangesListsNotesInTemplateThenNameOrder` and `TestSyncReturnsEachChangeNoteOnceExceptForAdoptedTemplates` in `internal/templates`;
  - `TestTemplateChangeNotesBecomeTodosOnceForProjectsBuiltBeforeThem`, `TestTemplateUseRecordsTheNewTemplatesNotesButNotTheOthers` and `TestTemplateChangesArrivingDuringTheSetupAreLedIntoIt` in `internal/cli`;
  - the existing `Template|Convention|Onboard|Sync|Compact` tests.

  `go vet` passes for both packages.
- End to end with `dist/mem-dev` and isolated `HOME`, on a disposable project from a bare copy of mem-templates at d74d624, with the setup marked done. Onboard reported "No open todos." Moving the copy to eb333eb produced the 📌 line, the todo `apply_nextjs_web_template_change_react` quoting the note, and a `[[change]]` record. The "Sync template items" commit included the lock and the todo, and the tree was clean.
- CI on dev passed on ubuntu, macOS and Windows for 4b14570 and f1370b1.

## Decisions

- **Change notes, not a todo per skill update:** most skill edits need no change in the project. The author decides which changes need action.
- **Silent recording for `init` and `template use`:** those projects start from the template's current state.
- **Notes are permanent:** a project never re-reads a seen note, so a correction is a new note. Records of notes deleted from the library are dropped silently, so a renamed note fires again.
- **`Library.Changes` does not re-check template names:** `Sync` calls `Items` first, which rejects unknown templates. This is a deviation from the spec text.
- **No note for the Tailwind/`partialPrefetching` change (d74d624):** existing projects keep create-next-app's own `next.config.ts`, which already has both.
- **The note waits for a release:** an installed mem without this feature ignores `changes/`. Recorded as todo `push_react_compiler_change_note`.

## Key Files Affected

- `internal/templates/changes.go`, `changes_test.go`, `lock.go` (new); `sync.go`, `promote.go`
- `internal/cli/template_changes.go`, `template_changes_test.go` (new); `template_conventions.go`, `template.go`, `init.go`, `onboard.go`, `onboard_updates.go`, `sync.go`, `compact.go`, `template_test.go`
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`
- `.mem/todos/push_react_compiler_change_note.md` (new); `.mem/todos/review_merge_convention_todo_branch.md` (deleted in the merge)
- mem-templates (unpushed eb333eb): `README.md`, `nextjs-web/changes/2026-10-29-react-compiler.md`

## Notes

- **The 📌 OPEN TODOS heading also contains 📌.** Tests must look for a specific 📌 line, not the emoji alone.
- **A stray `cat > /tmp/x` with no input blocked a test command** until it was stopped. The CLI tests had already passed in that run.
