---
created_at: "2026-10-11T21:56:07+02:00"
user: benjamin_van_heerden
---

# Work Log - Session titles in v0.7.7 and v0.7.8, artifacts guidance, the mission-control mod

## Overarching Goals

Release the session-title instructions and test them, keep agents from creating hosted artifacts, and give the user one place to see which sessions need attention, are working, or were touched last, across projects, without keeping the desktop app's sidebar open; then polish the permit mod's review flow.

## What Was Accomplished

### mem: CI fix, releases v0.7.7 and v0.7.8

- `internal/cli/main_test.go` (new): `TestMain` sets `GIT_CONFIG_COUNT`/`KEY`/`VALUE` for `maintenance.auto=false`, `gc.auto=0`, `receive.autogc=false`. Newer Git (2.55 on CI) repacked in the background after pushes and fetches and broke `t.TempDir` cleanup (`TestOnboardRemovesTodoClaimsFromOlderProjects`, `TestSpecCompletionWaitsUntilConflictsAreResolved`) on ubuntu and macos. CI green on all three platforms afterwards; the todo for it was deleted with the fix.
- v0.7.7 (date tag v2026.10.11.1): session titles `<project> · <the work>` at onboard, compaction and `log new`.
- `internal/agentsmd/instructions.md`: new "Documents, Reports and Visualizations" section: no hosted artifacts or canvases even when a tool or skill suggests one; Markdown in the repository, self-contained HTML files built from real data for richer reports; conversation-only answers stay in chat.
- Session titles name only the work (the user groups the sidebar by project folder again): `nameSession()` asks for "a short title for the work, a few words such as \"Tugela charger startup issue\""; `retitleSession()` asks to look up the title and rename it "to a short title for the current work"; `renderOnboardInstruction` lost its name parameter. Docs and structure doc updated.
- v0.7.8 (date tag v2026.10.11.2): plain titles and the artifacts guidance; 7 release assets; installed mem updated. `mem log new` output in this repository now shows the retitle step.
- Desktop sessions were renamed twice with `set_session_title`: to `<project> · <title>` and then back to plain titles.
- A throwaway mem project for testing titles was created at `/Users/benjamin/Documents/zzz-mem-title-test/title-test` (local bare remote, project name `title-test`).

### Mission Control mod (outside this repository)

- Probe findings: a mod can call the desktop app's session tools with `$.mcp.call('ccd_session_mgmt', 'list_sessions', …)` (sessionId, title, cwd, isRunning, lastActivityAt, link; no unread or attention state). A `Link` to a session's `claude://` address does not switch sessions inside the app; `$.process.run(['open', link])` does.
- Iterations the user rejected: a pane with an SVG activity timeline and bordered cards (timeline flashed on 5 s redraws, cards looked heavy), then a compact list (still not liked, and it showed irrelevant sessions because "seen" was only known for sessions opened from the panel). A pane cannot match the native sidebar's styling.
- Chosen design: each session reports its own activity to `~/.claude/mission-control/<session id>.json` (`working` on prompt submit and after tool calls, `waiting` on PermissionRequest, PermissionDenied and AskUserQuestion, `finished` on turn complete), keyed by `CLAUDE_CODE_HOST_SESSION_ID` (the app's `local_…` id) or `CLAUDE_CODE_SESSION_ID`. Every session polls the reports and the app's session list every 5 s and draws one band above the prompt: Blocked (this session's auto mode denials), Needs you, Working, Recent (three last touched); each pill opens the session; opening from the band marks a report acknowledged. Status line: "N blocked · N working elsewhere · N need you". The session being viewed is never listed.
- Fixes found in use: a session waiting on a question was shown as working (the app's isRunning stays on mid-turn); the app's flag also lagged after turns ended. A session's own report now decides; the flag counts only for sessions that never reported. Toasts were removed (the app places plugin toasts at the transcript's top right); the band's row gap reduced; wrapped pills get a row gap.
- The permit mod was merged into mission-control (one band, one status entry; two mods drawing the same band gave uneven padding). `claude plugin validate` requires each event registered once per module and follows `$` only into functions of the registering file, so everything touching `$` is in `hooks/register.tsx` (about 430 lines) with plain logic in `hooks/attention.ts` and `hooks/rules.ts`.
- Review pane redesign: one blocked action at a time with "‹ 1 of N ›", heading from the classifier's category, the command in a `Code` block, "Draft a rule" or "Dismiss"; the draft in a bordered "Proposed allow rule" box with Edit (Input), Applies ("Only in <starting folder>" or "Everywhere"), Save rule, Redraft, Discard; an SVG spinner while drafting; focus moved to Save when a draft is ready; save updates the pane before writing the file (restores the action with the reason on failure); the pane closes when nothing is left.
- Installed permanently: `~/.claude/mods/mission-control` (version 0.3.0) loads into every session through `CLAUDE_CODE_PLUGIN_DIRS` in `~/.claude/settings.json`, with `CLAUDE_CODE_PLUGIN_DIR_WATCH=1` for live reload. The separate `permit` folder was deleted. Verified with `claude plugin validate`, `tsc` against the build's types and `claude plugin test` (14 tests).
- In use: permit saved an `autoMode.allow` rule for the old harness's promotion command, scoped to ev-reporting (the user confirmed rules are scoped to the session's starting folder).

## Decisions

- No Mission Control replacement pane: the native sidebar (grouped by folder) stays the list, and the band shows attention, work and recent sessions with their project.
- Custom sidebar groups were not used: they cannot sub-group by folder, so they would need the project back in titles.
- Session titles carry only the work; the band adds the project.
- Toasts are not used: their position cannot be set.
- Each session's own report outranks the app's isRunning flag.
- permit and mission-control are one mod; allow rules are scoped to the session's starting folder.
- The mods are not bundled into mem: they are personal workflow on an early-access API, and Claude Code's plugin marketplaces already share plugins. They are to be put under version control on their own (todo).
- Artifacts: mem's instructions advise against hosted artifacts in every project.

## Key Files Affected

- mem: `internal/cli/main_test.go` (new), `internal/cli/session_title.go`, `internal/cli/session_title_test.go`, `internal/cli/onboard.go`, `internal/cli/compact.go`, `internal/cli/compact_test.go`, `internal/cli/log.go`, `internal/agentsmd/instructions.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`, `.mem/todos/` (CI flake todo deleted; plugin-sharing todo replaced by the version-control todo). Tags v2026.10.11.1/v0.7.7 and v2026.10.11.2/v0.7.8.
- Outside the repository: `~/.claude/mods/mission-control/` (`.claude-plugin/plugin.json`, `hooks/register.tsx`, `hooks/attention.ts`, `hooks/rules.ts`, `types/index.d.ts`, `tests/*.test.ts(x)`), `~/.claude/settings.json` (`env` block; one rule added by permit), `~/.claude/mission-control/` (session reports).

## Errors and Barriers

- A Python edit run through an unquoted shell heredoc mangled template strings in the mod; the file was rewritten whole.
- Rewriting the attention tests by cutting a text range also dropped the recent-sessions test; it was restored.
- `go run - ` does not read source from stdin; a first check of the hook command ran an empty command and was redone.
- On macOS `/home/...` counts as a network location for `$.fs`; mod tests use `/Users/tester`.
- `mem promote production --confirm` removes `.mem/local/release-notes.md`; the semver tag's notes were taken from the date tag's annotation.
- The rule permit saved earlier used the session's folder (ev-reporting) although the blocked command ran `cd` into ev-ledger; kept as intended by the user.
