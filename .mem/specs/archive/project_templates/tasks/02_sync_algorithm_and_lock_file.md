---
title: Sync algorithm and lock file
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-28T07:49:23+02:00"
completed_at: "2026-09-28T07:49:23+02:00"
---

Implement the reconcile function in internal/templates for memory, skill and doc items, with .mem/templates.lock (kind, name, template, hash). Follow the spec's sync table exactly, including automatic opt-outs into [templates] exclude and later-template-wins for duplicates. Skills go to .agents/skills/<name>/ with a relative .claude/skills/<name> symlink; failure to link is a warning line, not an error. Memories go through agentsmd.Memories/SetMemory. Return report lines for the caller to print. Cover every table row with a test.

## Completion Notes

internal/templates/sync.go: Sync reconciles memories (via agentsmd), skills (.agents/skills with a relative .claude/skills link, warning when linking fails) and docs (.mem/docs) against .mem/templates.lock, following the spec table: install, update unedited, report local-only edits, flag two-sided edits and pre-existing differing items with promote/reset commands, turn deletions into [templates] exclude entries, drop lock entries for items no longer provided. Returns report lines and changed paths. Verified by TestSyncInstallsUpdatesAndRespectsLocalDecisions and TestSyncLeavesAnExistingDifferentItemAndFlagsIt; go vet clean.
