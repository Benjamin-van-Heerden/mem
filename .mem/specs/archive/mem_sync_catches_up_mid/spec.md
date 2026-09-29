---
title: mem sync catches up mid-session
status: completed
assigned_to: benjamin_van_heerden
created_at: "2026-09-29T14:03:33+02:00"
updated_at: "2026-09-29T14:19:59+02:00"
completed_at: "2026-09-29T14:19:59+02:00"
---

## Overview

`mem sync` becomes the mid-session catch-up with the shared codebase: it converges Git as today, syncs template items, and reports what arrived (teammates' commits, work record changes, changed memories and skills, a newer mem release). A Claude Code `SessionStart` hook with the `compact` matcher runs the same catch-up after every compaction and prints a short digest, so each compaction becomes a point where mem converges the checkout and nudges version control.

Compaction itself is good and stays as it is. The hook adds only a small digest of shared state and the sync actions; it does not restate session progress, the structure doc, docs or logs.

Closes the todos `mem_sync_reports_what_teammates_changed`, `decide_on_mem_template_sync` and `consider_a_post_compaction_hook`.

## Goals

- `mem sync` reports what the sync brought in, grounded in actual Git state, and no longer tells the agent to re-read AGENTS.md.
- `mem sync` syncs template items (the same step as onboard); there is no separate `mem template sync`.
- After a compaction in Claude Code, the agent receives a compact digest: branch state and what the sync did, the active spec and its next task, the user's claimed todos, incoming changes, changed memories in full, nudges and a short instruction.
- mem installs and maintains the hook in the project's shared `.claude/settings.json`, and a project can opt out.

## Technical Approach

### Incoming changes (`mem sync`)

- Resolve the upstream ref (`@{upstream}`) before `converge.Sync` fetches and again after. Incoming is `oldUpstream..newUpstream`: exactly what others pushed since this checkout last fetched, independent of whether the branch fast-forwarded or rebased. `converge.Report` can carry both revisions. Record changes are also read between these two revisions. When there was no upstream before (first sync), report nothing incoming.
- Commits: author and subject of incoming commits, capped (e.g. 10, then "and N more").
- Work records: compare `.mem/specs/**`, `.mem/todos/*.md` between the two revisions (`git diff --name-status` plus frontmatter read with `git show <rev>:<path>`), and describe changes in plain terms: todo opened / claimed by X / closed (deleted); spec created / started by X / completed or abandoned (moved to `archive/`); task completed. Put the classification in `internal/work` (it owns the record formats), fed by a small reader so `work` does not depend on Git; the Git plumbing stays in the caller.
- Memories and skills: reuse `readKnowledge` / `renderKnowledgeChanges` from `internal/cli/onboard_knowledge.go` (snapshot before convergence, compare after the template sync).
- Newer mem: `selfupdate.Latest` with a short timeout; when newer, report it and name `mem update`. `mem sync` does not replace the binary mid-session (onboard keeps auto-updating).
- Templates: call the existing `syncTemplates` (it publishes template changes, or reports uncommitted edits).
- Instruction: grounded in what happened (review incoming changes relevant to current work, follow changed memories, tell the user about ⚠️ items). Remove "re-read AGENTS.md if it changed".

### Compaction hook

- Hidden command `mem hook compact`: reads the hook's stdin JSON (ignore it apart from tolerating it), runs the same catch-up as `mem sync`, and prints a compact digest as plain text (stdout reaches Claude's context). Keep it under ~3 KB: cap lists, never include the structure doc, docs or logs. Exits 0 on any mem error, printing a one-line note instead, so compaction is never disrupted.
- Digest contents: branch line with ahead/behind and what the sync did; active spec assigned to the user and its next pending task; todos claimed by the user; incoming changes (as in `mem sync`); changed memories with full text and changed skills by name/path; nudges; an instruction to continue with the current work, follow changed memories and raise ⚠️ items.
- Installation: `internal/hooks` (or a new `internal/claude` package) merges an entry into `.claude/settings.json`:
  `{"hooks": {"SessionStart": [{"matcher": "compact", "hooks": [{"type": "command", "command": "command -v mem >/dev/null 2>&1 && mem hook compact || true", "timeout": 60}]}]}}`
  mem's entry is identified by its command containing `mem hook compact`; it is added once, updated in place, and removed when opted out. Other settings and hooks are preserved. Installed by `mem init`, `mem import` and onboard's `applyUpdates`, published with the other mem project files (same pattern as the `.gitignore` entry, including the "already had uncommitted edits" case).
- Opt-out: `[claude] compact_hook = false` in `.mem/config.toml` (absent means on). No schema bump: the field is optional.

## Success Criteria

- `mem sync` after a teammate pushed a commit, opened a todo, completed a task and set a memory lists that commit, "todo … opened", "task … completed" and the memory with its text; an immediate second `mem sync` reports nothing incoming.
- `mem sync` output never contains "re-read AGENTS.md".
- `mem sync` in a project using templates installs a newly promoted template memory and reports it.
- `mem hook compact` in the same scenario prints the digest (branch, active spec and next task, claimed todos, incoming changes, memory text) under 3 KB, and exits 0 even when the project cannot be loaded.
- `mem init` and onboard create or update `.claude/settings.json` with exactly one mem `SessionStart`/`compact` entry, preserving existing keys and hooks; with `[claude] compact_hook = false` onboard removes the entry.
- Focused tests for the above pass on CI (ubuntu, macos, windows); `go vet` is clean; `.mem/structure.md`, `docs/design.md` and `docs/status.md` describe the new behaviour.

## Notes

- Claude Code docs (code.claude.com/docs/en/hooks): `SessionStart` matchers are `startup`, `resume`, `clear`, `compact`, `fork`; plain stdout is added to Claude's context; default timeout is 600 s; hooks run via bash, or PowerShell on Windows without Git Bash (the `command -v` guard assumes bash). `PostCompact` exists but its output reaches only the user, so it is not suitable.
- Re-marshalling `.claude/settings.json` with `encoding/json` maps reorders keys alphabetically. Prefer preserving the user's key order (decode with an ordered representation, or only rewrite when mem's entry actually changes).
- The user wants mem to lean into Claude Code; other agents keep working from AGENTS.md alone.
- The hook only runs `mem sync`-level actions (fetch, fast-forward or safe rebase, template sync). It never commits the user's work or pushes code.
