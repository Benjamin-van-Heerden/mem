---
title: Report incoming commits and record changes
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:09:43+02:00"
completed_at: "2026-09-29T14:09:43+02:00"
---

Extend converge.Sync / Report to carry the upstream revision before and after the fetch. Add a function in internal/work that, given a reader for file contents at two revisions and the changed .mem/ paths (git diff --name-status oldUpstream newUpstream -- .mem/specs .mem/todos), describes record changes in plain terms: todo opened / claimed by X / closed; spec created / started by X / completed or abandoned (moved to archive/); task completed. Git plumbing (diff, git show rev:path, git log for author+subject of incoming commits, capped at 10 with 'and N more') lives in the caller. Render an '📥 INCOMING' section in mem sync. Tests: teammate pushes a commit, opens a todo, completes a task; mem sync lists them; a second mem sync lists nothing.

## Completion Notes

converge.Sync records the upstream revision before and after the fetch (Report.Before/After). work.RecordChanges compares specs, tasks and todos between two revisions by kind and slug (archived specs count as one record) and describes them: todo opened/claimed/closed, spec drafted/started/completed/abandoned/removed, task added/completed. mem sync prints an INCOMING section with up to 10 commits and the record changes, and no longer says to re-read AGENTS.md. Verified with a unit test for RecordChanges and an end-to-end test where a teammate starts a spec, completes a task and opens a todo; a second sync reports nothing.
