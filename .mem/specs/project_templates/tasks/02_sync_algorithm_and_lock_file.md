---
title: Sync algorithm and lock file
status: todo
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-25T14:00:22+02:00"
---

Implement the reconcile function in internal/templates for memory, skill and doc items, with .mem/templates.lock (kind, name, template, hash). Follow the spec's sync table exactly, including automatic opt-outs into [templates] exclude and later-template-wins for duplicates. Skills go to .agents/skills/<name>/ with a relative .claude/skills/<name> symlink; failure to link is a warning line, not an error. Memories go through agentsmd.Memories/SetMemory. Return report lines for the caller to print. Cover every table row with a test.
