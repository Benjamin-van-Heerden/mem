---
title: Templates ship runnables
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

In mem (this repo): a template may contain runnables/<name>; templates.Sync installs it to .mem/runnables/<name> with the executable bit, records it in .mem/templates.lock and classifies/reconciles it like skills and docs (missing, updated, local edits, both edited), normalising CRLF to LF. Add a RunnablePath beside SkillPath/DocPath in internal/templates/sync.go. mem template promote accepts 'runnable' and mem template list shows runnables. Tests in internal/templates (install, update, local edit kept, promote) and a cli onboard test that runs a template-installed runnable. Update README, docs/design.md, docs/status.md, .mem/structure.md. go test and go vet for internal/templates and internal/cli. Do not release; releasing is the user's call.
