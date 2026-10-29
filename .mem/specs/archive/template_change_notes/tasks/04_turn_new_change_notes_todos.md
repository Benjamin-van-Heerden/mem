---
title: Turn new change notes into todos
status: completed
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:49:32+02:00"
completed_at: "2026-10-29T08:49:32+02:00"
---

In internal/cli, beside conventionTodos (as merged from the convention-todo branch): changeTodos(p, res.Changes) creating one todo per note via upsertTodo, titled 'Apply the <template> template change: <title>', body: 'The <template> template changed. Its change note <name>:' + quote(note) + the changeCheck paragraph from the spec (apply to existing code; tell the user where the project diverged or needs unadopted tooling; delete when the project matches or the user decides it does not apply), with the pending-setup variant when .mem/setup.md exists (follow the branch's pattern). Print a 📌 line per todo. Call it from syncTemplates after conventionTodos, appending lines and paths. Generalise conventionTodoInstruction and the templateInstructions step so the 📌 instruction covers both convention and change todos with explicit, state-specific wording; rename helpers where 'convention' no longer fits. End-to-end tests driving New() with a bare template library: init with an existing note creates no todo; a note added after init becomes the todo at onboard with the 📌 line and instruction, committed with the lock; a second onboard creates nothing; template use records without a change todo; pending-setup wording. Run focused tests and go vet ./internal/cli/ ./internal/templates/.

## Completion Notes

Added internal/cli/template_changes.go: changeTodos records one todo per new change note ('Apply the <template> template change: <title>', the note quoted, changeCheck, a pending-setup variant) with a 📌 line; syncTemplates and template use call it. Fixed a gap in task 3: template use records unseen notes of the project's other templates, so it now turns them into todos too. Generalised the 📌 helpers (templateTodoMark, hasTemplateTodo, templateTodoInstruction, onboard state templateTodos) and the template use and digest wording to cover conventions and change notes. Tests in template_changes_test.go: a note present at init records nothing; a later note becomes one committed todo with the lock, once; template use records the new template's notes and turns another template's note into a todo; pending-setup wording. go test ./internal/cli/ -run 'Template|Convention|Onboard|Sync|Compact' and go vet ./internal/cli/ ./internal/templates/ pass.
