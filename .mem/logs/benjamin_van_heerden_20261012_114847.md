---
created_at: "2026-10-12T11:48:47+02:00"
user: benjamin_van_heerden
---

# Work Log - Replay Theater mod installed, its hint drawn in mission-control's band

## Overarching Goals

Install the Replay Theater mod from anthropics/claude-code-playground (step through the last turn's file edits in a pane) next to mission-control, without the two fighting over the band above the prompt.

## What Was Accomplished

### Replay Theater installed

- Read the mod's source on GitHub first (`claude-code/mods/replay-theater`, public, Apache-2.0): it records Edit/Write/MultiEdit calls in `tool.call` and always calls `next(e)`, reads a file's old text with `$.fs` before a Write, registers `/replay`, and draws a band hint and a `Pane`. No network, processes or writes.
- With the user's approval, downloaded `.claude-plugin/plugin.json`, `hooks/hooks.json`, `hooks/replay-theater.mjs`, `README.md` and `.gitignore` (17,111 bytes, sizes matching GitHub; screenshots left out) to `~/.claude/mods/replay-theater`.
- `~/.claude/settings.json`: `CLAUDE_CODE_PLUGIN_DIRS` is now `/Users/benjamin/.claude/mods/mission-control:/Users/benjamin/.claude/mods/replay-theater`.

### Sharing the band

- The conflict: mission-control's `AbovePrompt` hook draws whenever the session has a working directory and never calls `next`, and Replay Theater's hint did the same after any editing turn, so whichever ran first took the whole band.
- Replay Theater (local changes, version `0.1.0-local.1`, noted at the top of the module): a contract `types/index.d.ts` declaring `PluginState['replay-theater'] = { edits: ReplayEdits }` and `"types"` in `plugin.json`; `$.state.set(EDITS, 0)` in `session.start` (a reload empties the replay) and `$.state.set(EDITS, state.replay.length)` in `turn.complete`; its band hook now only draws the replay when no pane could be placed, and `hintText` is gone.
- mission-control: `replayEdits = { plugin: 'replay-theater', key: 'edits' }` read with `$.state.get` in the band hook; while above zero a row `▶ Replay: N edits in the last turn` with a `Replay` button (hotkey `r`) running `$.command.run({ command: 'replay' })`; the count also keeps the band from returning `next(e)` when nothing else shows. `"dependencies": ["replay-theater"]` in its `plugin.json`, so the engine lays Replay Theater's contract into its types.
- Checks: `claude plugin validate` passes for both (CLI 2.1.281 and the desktop app's 2.1.295); mission-control type-checks with the dependency's contract in a scratch tsconfig; `claude plugin test` passes its 15 tests. A backup of mission-control's hooks, types and manifest from before the edit was kept in the session scratchpad.

### Why the user saw nothing

- The open sessions started before the plugin list changed: a session reads `CLAUDE_CODE_PLUGIN_DIRS` at start and watches only those folders, so they hot-reloaded mission-control but never loaded Replay Theater.
- Verified with the app's build (`~/Library/Application Support/Claude/claude-code/2.1.295/.../claude -p "/replay" --debug-file …` in a scratch folder): both modules loaded and were admitted, `/replay` was listed, and it answered "Replay Theater: no edits in the last turn." The engine also laid `replay-theater` into mission-control's type root.

## Decisions

- mission-control draws the replay hint (the user's choice among: hint in mission-control's band, Replay Theater as-is with `/replay` only, or merging it into mission-control). Cross-mod state reads (`$.state.get` of another plugin's value) and `$.command.run` make this work without touching Replay Theater's pane.
- Replay Theater loads after mission-control in `CLAUDE_CODE_PLUGIN_DIRS`.

## Key Files Affected

- `~/.claude/mods/replay-theater/` (new): the five downloaded files, `types/index.d.ts`, and edits to `hooks/replay-theater.mjs` and `.claude-plugin/plugin.json`.
- `~/.claude/mods/mission-control/hooks/register.tsx`: `replayEdits` and the Replay row in the band; `.claude-plugin/plugin.json`: `dependencies`.
- `~/.claude/settings.json`: `CLAUDE_CODE_PLUGIN_DIRS`.
- No changes in this repository.

## Errors and Barriers

- Replay Theater's first contract ended with `export {}`; validate refuses that ("a contract exports types and nothing else"), so it exports `type ReplayEdits` instead.
- `tsc -p` on mission-control failed until the dependency's contract was laid and named in `types`; the engine does both when it loads the two mods together, so the check used a scratch tsconfig rather than editing the generated one.
- Debug logs were off (`~/.claude/debug` held only an old `latest` link), so the load was confirmed with a headless session and `--debug-file`.
