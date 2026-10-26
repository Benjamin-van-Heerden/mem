---
title: Templates ship runnables
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:15:36+02:00"
completed_at: "2026-10-26T13:15:36+02:00"
---

In mem (this repo): a template may contain runnables/<name>; templates.Sync installs it to .mem/runnables/<name> with the executable bit, records it in .mem/templates.lock and classifies/reconciles it like skills and docs (missing, updated, local edits, both edited), normalising CRLF to LF. Add a RunnablePath beside SkillPath/DocPath in internal/templates/sync.go. mem template promote accepts 'runnable' and mem template list shows runnables. Tests in internal/templates (install, update, local edit kept, promote) and a cli onboard test that runs a template-installed runnable. Update README, docs/design.md, docs/status.md, .mem/structure.md. go test and go vet for internal/templates and internal/cli. Do not release; releasing is the user's call.

## Completion Notes

Templates can ship runnables: a template's runnables/<name> is installed to .mem/runnables/<name> (installRunnable: LF line endings, mode 0755), tracked in .mem/templates.lock as kind runnable, and synced, reset and promoted like docs (RunnablePath; localPath now drives localHash and export). template list shows runnables, and promote/reset help names the kind. Verified with TestSyncInstallsRunnablesExecutableWithLFLineEndings (CRLF template script installed with LF and +x, hidden files skipped, stays in sync, update, two-sided edit flagged) and TestOnboardRunsTemplateRunnablesAndPromoteSendsThemBack (onboard installs, commits as 100755 and runs it; promote pushes an edit back; list shows it in sync); go test for internal/templates and the template CLI tests, go vet for both packages. README, docs/design.md, docs/status.md and .mem/structure.md updated.
