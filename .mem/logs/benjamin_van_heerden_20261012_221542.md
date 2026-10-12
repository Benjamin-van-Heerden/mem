---
created_at: "2026-10-12T22:15:42+02:00"
user: benjamin_van_heerden
---

# Work Log - Work logs drop the guidance comment; Replay Theater copy experiments end with images and Open in editor

## Overarching Goals

Fix work logs carrying the template's guidance comment, and let the user copy code from Replay Theater's pane. Copying could not be made reliable on the desktop, so the pane settled on images without copying and a button that opens the file in the user's editor.

## What Was Accomplished

### Work logs (mem)

- `internal/work/log.go`: the log template no longer contains the `<!-- A work log records … -->` guidance comment; the same guidance is in `mem log new`'s step 1 ("The log is not updated later, so do not list future work in it") and the todo step. `Finish` (run by `log commit`) now removes such a comment by its opening words, `regexp.MustCompile(`<!-- A work log records [^>]*-->\n\n`)`, so a log started by v0.7.10 and committed after an upgrade is cleaned too; before, it matched one exact text, which had already changed on `dev` ("as fact and as things stand").
- `internal/work/work_test.go`: `TestNewLogsCarryNoGuidanceAndFinishRemovesAnOlderOne` replaces the old Finish test. `go test ./internal/work/`, `go test ./internal/cli/ -run Log` and `go vet` pass. Commit `1011cc4`.
- Committed logs never kept the comment (v0.7.10 strips its own text); the user saw it in Replay Theater's diff of the log file before commit.

### Replay Theater (in `~/.claude/mods`, repo `Benjamin-van-Heerden/claude-mods`), 0.1.0-local.11 to local.18

- Selectable diffs: `Svg isInteractive` drew the frame at a small default size until each piece passed `width` and `height`; then the frames flashed and dropped the selection whenever the pane held the keyboard. Returning the same tree on redraws, opening the pane without `focus` and dropping `autoFocus` did not stop it. Handing the keyboard back after each press in the pane (`$.prompt.read()` then `$.prompt.fill({ text })` from a `ui.press` hook) made selection work, but the first click on any pane button then only focused the pane. Reverted (`c82bc99` reverts `d294530`).
- Copy buttons (`$.ui.copy`) per changed stretch and "Copy all" worked, and the whole file as a native `Code` block with `startLine={1}`; the user preferred no copying over them, so both were removed (`724a25e`).
- Kept from this stretch: one column of line numbers (the file after the edit; removed lines have none, `gutter = numWidth + 3`); the pane never asks for focus; a new **Open in editor** button on the file row runs `$.process.run(['open', step.path])` and toasts on failure; `session.start` calls `$.ui.status(undefined)` to clear the diagnostic status line a temporary version left, which outlives reloads (`402280a`).
- Tests (11): the pane test presses the file name (whole file drawn as SVG) and Open in editor (`open /work/a.go` run); the copy tests went with the code.
- A bug report for Anthropic was drafted and the user sent it: interactive Svg frames in a Pane reload when the pane gains or loses the keyboard, and the frame needs explicit width and height.

### Memories

- `mod-ui` updated twice: code is drawn as images and not selectable (the user preferred that to Copy buttons); handing the keyboard back only moves the problem; files open with `$.process.run(['open', path])`; a pane of buttons asks for no focus.

## Decisions

- Work-log guidance lives in `mem log new`'s stdout, not in the log file.
- Replay Theater's code is not selectable or copyable: every desktop route tried (interactive frames, with or without handing the keyboard back) was unreliable, and the user found Copy buttons not worth the clutter. Opening the file in the editor covers reading and copying the real file.
- `open <path>` (the app macOS associates with the file type) rather than `open -t` (the default plain-text editor).

## Key Files Affected

- `internal/work/log.go`, `internal/work/work_test.go`: template without the comment, `olderGuidance` removal in `Finish`, test.
- `AGENTS.md`: `mod-ui` memory updates.
- `~/.claude/mods/replay-theater/hooks/register.tsx`: Open in editor, one diff block for both views, no copy code, status line cleared at load; `hooks/diff-svg.ts`: single line-number column; `hooks/diff.ts`, `hooks/highlight.ts`: `afterText` and `stretchesOf` added then removed; `tests/`: pane test for the whole file and Open in editor; `README.md`, `.claude-plugin/plugin.json`: kept current.

## Errors and Barriers

- Interactive Svg frames on the desktop: tiny without explicit size; with it, they reload on every change of keyboard focus in the pane, losing selections and swallowing the first click when focus is handed back. No mod-side fix found.
- `$.ui.status` text outlives a mod reload; a version that sets it must clear it.
- In `claude plugin test`, stubbing `ui.invalidate` stops a mounted drawing from redrawing after a press, and hooks must be registered before the test's first `$` call.
