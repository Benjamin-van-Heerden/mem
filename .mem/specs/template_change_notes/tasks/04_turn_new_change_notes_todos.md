---
title: Turn new change notes into todos
status: todo
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:27:22+02:00"
---

In internal/cli, beside conventionTodos (as merged from the convention-todo branch): changeTodos(p, res.Changes) creating one todo per note via upsertTodo, titled 'Apply the <template> template change: <title>', body: 'The <template> template changed. Its change note <name>:' + quote(note) + the changeCheck paragraph from the spec (apply to existing code; tell the user where the project diverged or needs unadopted tooling; delete when the project matches or the user decides it does not apply), with the pending-setup variant when .mem/setup.md exists (follow the branch's pattern). Print a 📌 line per todo. Call it from syncTemplates after conventionTodos, appending lines and paths. Generalise conventionTodoInstruction and the templateInstructions step so the 📌 instruction covers both convention and change todos with explicit, state-specific wording; rename helpers where 'convention' no longer fits. End-to-end tests driving New() with a bare template library: init with an existing note creates no todo; a note added after init becomes the todo at onboard with the 📌 line and instruction, committed with the lock; a second onboard creates nothing; template use records without a change todo; pending-setup wording. Run focused tests and go vet ./internal/cli/ ./internal/templates/.
