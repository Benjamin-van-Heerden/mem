---
created_at: "2026-09-29T14:27:40+02:00"
user: benjamin_van_heerden
spec: mem_sync_catches_up_mid
---

# Work Log - Onboard and sync report what changed; post-compaction hook

## Overarching Goals

Remove onboard's instruction to re-read AGENTS.md and surface changed memories and skills instead; keep `.mem/local/` ignored; clear two small todos and release v0.4.2; then make `mem sync` the mid-session catch-up, add a Claude Code post-compaction hook that runs it, and release v0.5.0.

## What Was Accomplished

### Onboard: changed memories and skills, `.mem/local/` ignore (v0.4.2)

- The "Read AGENTS.md again" step was removed. `readKnowledge` (`internal/cli/onboard_knowledge.go`) snapshots project memories from AGENTS.md and `SKILL.md` files under `.agents/skills` and `.claude/skills` before the sync and after the template sync; `renderKnowledgeChanges` prints 🧠 CHANGED MEMORIES (full text) and 🛠️ CHANGED SKILLS (path and description) at the top of the project context, so they land in `.mem/local/onboard.md` when the context is long. stdout keeps only a one-line instruction to follow them.
- `applyUpdates` calls `ensureIgnored` for `/.mem/local/` (it now reports whether it changed `.gitignore`) and publishes the change; it warns with `git rm -r --cached .mem/local` when files there are tracked. Previously only `init` and `import` added the entry.
- `applyUpdates`, `syncTemplates`, `hasWarning` and `afterUpdates` moved to `internal/cli/onboard_updates.go` to keep `onboard.go` under 500 lines.
- `mem log commit` calls `Log.Finish`, which removes the template's guidance comment (`logGuidance`) before committing.
- Spec, task and todo slugs come from `work.recordSlug`: filler words dropped, at most five words; full titles still resolve.
- Released v0.4.2.

### Spec `mem_sync_catches_up_mid` (archived)

- `converge.Sync` records the upstream revision before and after the fetch (`Report.Before`/`After`); `upstream` and `upstreamRevision` were factored out of `inspect`.
- `work.RecordChanges(paths, before, after Revision)` (`internal/work/changes.go`) compares specs, tasks and todos at two revisions by kind and slug, so a spec moving to `archive/` is one record, and describes: todo opened / claimed by X / closed; spec drafted / started by X / completed / abandoned / removed; task added (to an existing spec) / completed. `parseMarkdown` was split out of `ReadMarkdown`.
- `catchUp` (`internal/cli/sync.go`), shared by `mem sync` and `mem hook compact`: snapshot knowledge → `converge.Sync` → reload project → `syncTemplates` → `afterUpdates` → `newerRelease` nudge → `readIncoming` (`git log Before..After` capped at 10, record changes via `git diff --name-only` and `git show rev:path`) → snapshot again.
- `mem sync` prints the report, 🧩 TEMPLATES, 📥 INCOMING and changed memories/skills with grounded instructions; its "re-read AGENTS.md" line is gone.
- `newerRelease` (`internal/cli/update.go`) reports a newer release and `mem update` without replacing the binary.
- Hidden `mem hook compact` (`internal/cli/compact.go`) prints a digest: branch and what the sync did, the user's active spec with progress and next task, claimed todos, up to 5 commits and 10 record changes, template lines, changed memories and skills, nudges, one instruction paragraph. Any error prints one line and exits 0.
- New package `internal/claude`: `SyncCompactHook(root, enabled)` maintains `{"matcher": "compact", "hooks": [{"type": "command", "command": "command -v mem >/dev/null 2>&1 && mem hook compact || true", "timeout": 60}]}` under `hooks.SessionStart` in `.claude/settings.json`, using an order-preserving JSON `object` and encoding without HTML escaping; writes only on change; never creates the file just to opt out. `project.ClaudeConfig` adds `[claude] compact_hook` (unset means on). Called by `init`, `import` and `applyUpdates`, which publishes it.
- Released v0.5.0 (date tag v2026.09.29.1); the installed mem updated from v0.4.2 to v0.5.0 via `mem update`.

### Verification

- New tests: `onboard_test.go` (changed memories and skills land in onboard.md only and are not repeated; ignore entry restored and pushed; hook installed, published and removed on opt-out), `sync_test.go` (teammate's commit, spec start, task completion, todo and memory reported once; promoted template memory installed and reported; newer release reported without a download), `compact_test.go` (digest contents, under 3000 bytes, no onboard context; exit 0 outside a mem project), `internal/claude/settings_test.go`, `internal/work/changes_test.go` and slug and log guidance tests in `work_test.go`.
- CI passed on ubuntu, macOS and Windows before each promotion; both release builds published 7 assets.

## Decisions

- Changed memories and skills are shown in the project context rather than asking the agent to re-read AGENTS.md; changes to mem's own instructions take effect in the next session.
- The onboard context file is where context belongs, the exception to instructions living in stdout; only the agent instruction stays in stdout.
- Template sync is part of `mem sync`; there is no `mem template sync`.
- mem leans into Claude Code. The post-compaction hook complements compaction with a small digest and version control actions (fetch, fast-forward or safe rebase, template sync); it does not restate session progress, the structure doc, docs or logs, and never commits or pushes the user's work.
- The hook uses `SessionStart` with the `compact` matcher because its plain stdout reaches Claude's context; `PostCompact` output reaches only the user. It is installed in the shared `.claude/settings.json`, on by default.
- "Incoming" is `upstream before fetch..upstream after fetch`, independent of fast-forward or rebase.
- `mem sync` reports a newer release but only onboard and `mem update` replace the binary.

## Key Files Affected

- `internal/cli/onboard.go`, `onboard_updates.go` (new, moved from onboard.go), `onboard_knowledge.go` (new), `onboard_test.go` (new).
- `internal/cli/sync.go` (rewritten around `catchUp`, `readIncoming`, `renderIncoming`), `sync_test.go` (new).
- `internal/cli/compact.go`, `compact_test.go` (new); `hook.go` registers `compact`.
- `internal/cli/update.go`: `newerRelease`.
- `internal/cli/init.go`, `import.go`: `ensureIgnored` returns whether it changed; `claude.SyncCompactHook`.
- `internal/cli/log.go`, `internal/work/log.go`: `Log.Finish`, `logGuidance`.
- `internal/work/markdown.go` (`recordSlug`, `parseMarkdown`), `changes.go`, `changes_test.go` (new), `spec.go`, `task.go`, `todo.go`, `work_test.go`.
- `internal/converge/converge.go`: `Report.Before/After`, `upstream`, `upstreamRevision`.
- `internal/claude/settings.go`, `settings_test.go` (new); `internal/project/project.go`: `ClaudeConfig`.
- `internal/cli/template_test.go`: expectation for the changed-memories section.
- `.mem/structure.md`, `docs/design.md`, `docs/status.md`, `README.md`.

## Errors and Barriers

- `echo ========` fails in zsh: a word starting with `=` is expanded as a command path.
- `TestPromotedItemsReachOtherProjectsAtOnboard` still expected "Read AGENTS.md again" and was pushed failing, because only `TestOnboard*` had been run; fixed in a follow-up commit before the v0.4.2 release.
- The five-word slug cap cuts titles mid-phrase (e.g. `mem_sync_catches_up_mid`, `add_mem_hook_compact_compact`).
- The test helper `run` forces the Git author to "Test", so commit authors in sync tests are asserted as "Test".
