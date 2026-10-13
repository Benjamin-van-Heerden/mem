---
created_at: "2026-10-13T22:26:09+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.8.0 released; Replay Theater's file row trimmed and leftover panes closed on reload

## Goal

Release the 17 commits on `dev` as mem v0.8.0, and tidy two things in Replay Theater the user noticed while using it.

## What Was Accomplished

### mem v0.8.0

- CI green on ubuntu, macOS and Windows for `0dfe3de`. `mem promote staging` moved `test` from `cc10cda` to `0dfe3de`; `mem promote production` drafted notes, which were rewritten to cover only mem's own changes (path-scoped task commits, size-based work-log nudges with the `mem hook prompt` UserPromptSubmit hook and `[structure] include`, checkpoint-shaped logs without the guidance comment, up to three recent logs in onboard), shown to the user and released with `--confirm`: `main` at `0dfe3de`, date tag `v2026.10.13.1`.
- `v0.8.0` tagged on `origin/main` with the same notes (`git tag -a --cleanup=whitespace -F`), pushed; the Release workflow (run 38459494426) succeeded with 7 assets. `mem update` moved the installed mem from v0.7.10 to v0.8.0. This log is the first written with v0.8.0's template.
- `--confirm` deletes `.mem/local/release-notes.md`, so the semver tag's message was read back from the date tag with `git for-each-ref refs/tags/v2026.10.13.1 --format='%(contents)'` (it carries no marker line).

### Replay Theater (`~/.claude/mods`, repo `Benjamin-van-Heerden/claude-mods`)

- 0.1.0-local.19 (`4aab2a9`): the dim text after the counts on the file row (tool, note, "click the name for the whole file") removed; it wrapped badly in a narrow pane. The tool and note still reach Haiku.
- 0.1.0-local.20 (`7463f9c`): `session.start` calls `closeReplay($)`. A reload empties the replays but the host keeps an open pane, which then showed "No edits to replay yet" after turns without edits; it is now closed. Test `a reload closes a pane the host kept open` raises `session.start` and expects `ui.close` for `replay-theater`. 12 tests pass.

## Decisions

- v0.8.0 rather than v0.7.11: task completion, the log template and onboard behave differently, and every project gains a hook.
- The release notes leave out the mods work and memory or log commits, which change nothing in mem.
- Replay Theater keeps opening on any Edit or Write in a turn, notes and logs included; excluding `.mem/` or files outside the project was offered and not asked for.

## Key Files Affected

- No files in this repository changed besides this log; the release moved `test` and `main` and added tags `v2026.10.13.1` and `v0.8.0`.
- `~/.claude/mods/replay-theater/hooks/register.tsx`: file row without the step text; `closeReplay($)` in `session.start`.
- `~/.claude/mods/replay-theater/tests/pane.test.tsx`: the reload test.
- `~/.claude/mods/replay-theater/.claude-plugin/plugin.json`: versions local.19 and local.20.
