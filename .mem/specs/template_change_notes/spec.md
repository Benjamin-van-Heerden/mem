---
title: Template change notes
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-10-29T08:24:47+02:00"
updated_at: "2026-10-29T08:29:49+02:00"
---

## Overview

A template can change in ways that existing projects must act on: a setting in a starter file, a setup step, a dependency, a skill whose new guidance contradicts what projects already built. Today the only change that reaches an existing project as work is a changed memory, through the convention todo (`internal/cli/template_conventions.go`). Updated skills are replaced quietly ("Updated skill X from T", plus the CHANGED SKILLS listing), `setup.md` is copied once by `init`, and the project's own files built from starter files never sync. An agent in the project is never asked to change anything.

The case that prompted this: nextjs-web turned on the React Compiler (mem-templates 4630cd8). New projects get it from `skills/nextjs-setup/files/next.config.ts`, and existing ones get the updated `nextjs` skill, which now says the compiler is on. Their own `next.config.ts` stays as it was, and nothing tells them.

Change notes let the template author say so. A note is a Markdown file in the template, `<template>/changes/<name>.md`, saying what changed and what an existing project should do. At sync, every note a project has not seen becomes a todo, and the project's `.mem/templates.lock` records it as seen. A project that receives the template in its current state (`init`, `template use`) records its notes as seen without todos, since it already starts from that state.

## Goals

- A template author can mark a change that existing projects must act on, by adding one file to the template.
- Each project using the template gets that note as a todo exactly once, at its next onboard, `mem sync`, `task complete` or `spec complete` (wherever `syncTemplates` runs).
- Projects created or extended from the template after the note was written get no todo for it.
- The todo is self-contained: a fresh session can act on it without the template library or this spec.
- It works the same way as the convention todos and builds on the convention-todo branch (`claude/distracted-khayyam-892c14`, 8d75236), including its wording for a setup that is still pending.

## Technical Approach

### Note format (template library)

- Path: `<template>/changes/<name>.md`. The name is the file name without `.md`, a date and a slug, for example `2026-10-29-react-compiler`. Notes are processed in name order, so the date prefix keeps them chronological. Files starting with `.` and non-`.md` files are ignored, as with other item kinds.
- Content: a first-line `# <title>` heading, then the body. The body says what changed in the template, what an existing project should change, how to check it, and when it does not apply. Without a heading, the title is the name.
- Notes are per template, keyed `<template>/<name>`. Unlike items, a later template does not override an earlier one's note of the same name. Both are separate notes.
- Normalise CRLF to LF when reading (see the CRLF convention in `.mem/structure.md`).
- A note is not an item. It is not installed into the project, not hashed, and not shown by `mem template list` as an item.

### Library (`internal/templates`)

- A `Change` type: `Template`, `Name`, `Title`, `Body`, plus `Key()` returning `<template>/<name>`.
- `Library.Changes(names []string) ([]Change, error)`: reads `changes/` of each named template, skipping templates without the directory, sorted by template order, then name. An unknown template name errors, as `Items` does.

### Lock and Sync

- `.mem/templates.lock` gains `[[change]]` tables with `template` and `name`, kept sorted by key. `readLock` and `writeLock` handle both tables, and the header comment covers both. The `state` gains the seen set and marks the lock changed when it changes. go-toml v2 ignores unknown tables, so an older mem binary reads a new lock without error. It drops the `[[change]]` tables when it rewrites the lock, though, and the notes then fire again under a newer mem. Accept this: onboard updates mem itself first.
- `Sync` takes the templates whose notes should be recorded without todos. An options struct or a parameter is fine; pick whichever reads better at the three call sites:
  - `init` (`internal/cli/init.go`): every template it installs.
  - `template use` (`internal/cli/template.go`): the newly used template.
  - `syncTemplates` (`internal/cli/onboard_updates.go`): none.
- For every note of a used template that the lock has not seen: record it as seen, and unless its template is in the adopted set, append it to a new `Result.Changes []Change`. Note files are not repository paths that changed. The lock is, through `LockPath` in `Result.Paths`, so `syncTemplates` publishes it.
- Drop seen records for notes the library no longer provides, and for templates the project no longer uses, without a line. Deleting a note is the author's call, and a renamed note fires again, which is acceptable.
- A project with no `[[change]]` records yet, such as every existing project when this ships, gets every existing note as a todo at its next sync. That is intended: notes start with this feature, so each one describes a change existing projects lack.

### Todos and output (`internal/cli`)

Build on the convention-todo branch's `template_conventions.go`, merged first (task 1).

- `changeTodos(p project.Project, changes []templates.Change) (lines, paths []string, err error)`, alongside `conventionTodos`, called from `syncTemplates` right after it, with its lines and paths appended the same way.
- One todo per note through `upsertTodo`, titled `Apply the <template> template change: <title>`. Body:
  - `The <template> template changed. Its change note <name>:`, then the note quoted with `quote()`.
  - Then a fixed paragraph, `changeCheck`: "The project was built from the template before this change. Apply it to this project's existing code as the note describes. Where the project has diverged from the template so that the note does not fit, tell the user and ask how to proceed. Where the change needs something the project has not adopted, tell the user and ask whether to adopt it, as its own todo or spec. Delete this todo once the project matches the note, or once the user decides it does not apply here."
  - While `.mem/setup.md` is pending, follow the branch's pattern. Add that the change arrived after the setup was copied, so apply it in the remaining setup steps, or to what earlier steps built, and check it once the setup is done.
- Each todo prints a `📌` line under 🧩 TEMPLATES: `📌 The <template> template changed (<title>), so todo <slug> asks for the change to be applied to this project.`, or the pending-setup variant.
- `hasConventionTodo` already matches any `📌` line, so onboard, `share` and `templateInstructions` already add an instruction step. Generalise `conventionTodoInstruction` (the branch's `func(setupPending bool) string`) and the `templateInstructions` step to cover both kinds. Keep the wording explicit: tell the user about each 📌 item, what the todo asks for, and ask whether to do it now or later. Rename the helpers if `convention` no longer fits, for example `templateTodoMark` and `hasTemplateTodo`.
- The compaction digest needs nothing new: it does not report convention todos either, and the todos show under OPEN TODOS at onboard.

### Documentation

- `docs/design.md`, template library section: the `changes/` directory, the note format, the lock's `[[change]]` records, when notes become todos, and why `init` and `template use` record them silently.
- `docs/status.md`: the implemented behaviour in the Templates bullet.
- `.mem/structure.md`: the `internal/templates` entry (`Library.Changes`, `Result.Changes`, the lock) and the Templates data flow (`changeTodos`).
- The managed instructions (`internal/agentsmd/instructions.md`) stay unchanged. The stdout instruction carries the behaviour (see "Persistent agent context" in AGENTS.md).

### Template library (mem-templates, separate repository)

- README: a "Change notes" section. Write a note when a template change needs action in existing projects: a starter file, a setup step, a dependency, or a skill whose new guidance contradicts what projects already built. Don't write one for wording, new knowledge that needs no change, or changes that only matter to new projects. Give the format and an example, and say that a note is permanent: editing a published note does not reach projects that have already seen it, so write a new one instead.
- The first note: `nextjs-web/changes/2026-10-29-react-compiler.md`. It turns the compiler on in an existing project, following mem-templates 4630cd8:
  - the `next.config.ts` settings: `reactCompiler: { environment: { enablePreserveExistingMemoizationGuarantees: true } }` and `experimental.turbopackRustReactCompiler: true`;
  - remove `babel-plugin-react-compiler` if installed;
  - fix what `bun run lint` reports from `eslint-plugin-react-hooks`, for example `form.watch()` to `useWatch`;
  - check the affected UI;
  - not applicable to a project that builds with webpack.

  The Tailwind loader and `partialPrefetching` change (mem-templates d74d624) gets no note: existing projects keep create-next-app's own `next.config.ts`, which already has both.
- Commit and push mem-templates on `main` (its convention: `nextjs-web: …` subjects, no attribution). Push the note only after the mem release that reads notes, so projects don't see a `changes/` directory their mem ignores. That is harmless, but the note would then fire later as intended anyway.

## Success Criteria

- `Library.Changes` returns each used template's notes from `changes/*.md` in template then name order, with the title from the first `# ` heading (or the name), CRLF normalised, ignoring dotfiles and non-`.md` files.
- `.mem/templates.lock` holds `[[change]]` records (`template`, `name`) sorted by key. A lock without them still reads, and an unknown table in the lock does not break `readLock`.
- `Sync` on a project that has not seen a note returns it in `Result.Changes` and records it as seen. A second `Sync` returns nothing. A note whose template is in the adopted set is recorded with nothing returned. Records of notes the library no longer provides are dropped.
- End to end (tests in `internal/cli` driving `New()` against a bare template library):
  - `init --template T` with a note already in T records it and creates no todo.
  - A note added to T after `init` becomes the todo `Apply the T template change: <title>` at the next onboard, with the note quoted in its body, a `📌` line under 🧩 TEMPLATES and an instruction step that mentions it. It is committed with the lock by the sync's publish.
  - A second onboard creates nothing.
  - `template use T` on an existing project records T's notes and creates no change todo.
  - With `.mem/setup.md` present, the todo and the `📌` line use the pending-setup wording.
- The convention-todo branch is merged into dev, and its tests pass, before the change-note code lands.
- `go test ./internal/templates/ ./internal/cli/` (focused on the new and touched tests) and `go vet ./internal/templates/ ./internal/cli/` pass. CI is green on ubuntu, macOS and Windows after the push.
- `docs/design.md`, `docs/status.md` and `.mem/structure.md` describe the implemented behaviour.
- mem-templates has the README section and `nextjs-web/changes/2026-10-29-react-compiler.md`, pushed after a mem release that includes this feature. That release is the user's call.

## Notes

- **Why not a todo for every skill update:** most skill edits are wording or new knowledge that needs no change in the project, and a todo for each would be noise. The author knows which changes need action; the note records that judgement.
- **Why not a memory:** memories are lasting conventions, and the user rejected using one only to trigger a todo.
- **Why notes are permanent:** projects that have seen a note never read it again, so a correction needs a new note.
- **Before this ships:** existing nextjs-web projects (such as studii-next) have the updated `nextjs` skill but not the compiler setting. The React Compiler note reaches them once this ships. Until then it can be applied by hand.
- **Out of scope:** a command to write notes (they are written by hand in the library), notes in `mem template list`, notes promoted from a project with `mem template promote`, and checking vendored skills against their recorded upstream source.
