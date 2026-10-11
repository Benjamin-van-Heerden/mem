---
created_at: "2026-10-11T23:27:45+02:00"
user: benjamin_van_heerden
---

# Work Log - Session title reminder (mem v0.7.10), permit keeps the keyboard, "you are here" pill in mission-control

## Overarching Goals

Make session naming at onboard reliable after another session went unnamed, find and fix the permit pane's two-click Save from the trace, and make the mission-control band show which session you are in.

## What Was Accomplished

### Why the ascendis-transform session went unnamed, and mem v0.7.10

- The session (`local_ea23a21a-…`, transcript `~/.claude/projects/-Users-benjamin-Documents-acn-capital-ascendis-transform/2b8faa27-….jsonl`) ran v0.7.9 and got the full naming instruction. The user's first message was "Let's get to work", so the rename was deferred to the next message. The context was over 14,000 bytes and went to `.mem/local/onboard.md`, which the agent read after the instruction; its reading ended on the work log, and when the user's next message came it never attempted the rename. It was renamed from here to "Git diff count against main".
- `internal/cli/session_title.go`: `nameSession()` now says "otherwise as the first thing you do after their next message"; new `sessionTitleReminder()` returns `"Remember to name this session. " + nameSession()`.
- `internal/cli/onboard.go`: when the context goes to the file, it ends with a section:

  ```go
  output.Section(&buf, "🏷️ SESSION TITLE")
  fmt.Fprintln(&buf, sessionTitleReminder())
  ```

  Inline contexts are unchanged: the instruction, ending with the same request, is already the last thing printed.
- `internal/cli/session_title_test.go`: `TestALongOnboardContextEndsWithTheSessionTitleReminder` (a 300-line doc pushes the context into the file; the file ends with the section; stdout does not contain the reminder).
- Docs: `docs/design.md` (Session titles), `docs/status.md`, `.mem/structure.md`.
- Focused tests and `go vet` passed; CI green on ubuntu, macos and windows. Released as date tag v2026.10.11.4 and `v0.7.10` (main at cc10cda, 7 release assets); installed mem updated.

### permit: the two-click Save (outside this repository)

- The trace (`~/.claude/mission-control-debug.log`) showed the cause: pressing "Draft a rule" redrew the pane with the spinner, and 13 ms later the pane reported `isFocused=false`; the `$.ui.focus(save)` after the draft returned `{"deny":"that site does not hold the keyboard"}`; the first click on Save then only produced a render with `isFocused=true` (or a `ui.focus` with origin person) and no `ui.press`; the second click pressed. The same happened to "Draft a rule" after a save moved to the next action. On the desktop surface a pane gives up the keyboard when the element holding it leaves the drawing; buttons that stay drawn (‹ ›) kept it.
- Fix in `hooks/register.tsx`: `holdKeyboard($, key?)` calls `openReview($)` (`$.ui.open` with `focus: true`, which the surface grants while the prompt has the keys over an empty composer) and then `$.ui.focus` on the given key. Called after a draft lands (key `save` when the body is non-empty), after a save while actions remain (before `$.prompt.fill`, because text in the composer refuses the request), after Enter in the rule Input (`save`) and after Discard (`draft`). The earlier bare `$.ui.focus(save)` was removed.
- `tests/permit.test.tsx`: `engineBeneath` answers `ui.open` (`{ value: { isPlaced: true as const } }`) and records focus requests; the main test asserts the pane asks for the keyboard after the draft lands. The test kit does not model which site holds the keyboard, so `$.ui.focus` never reaches the test's hook; where the ring lands was checked in the app.
- The user confirmed one click saves. The tracing was removed, the debug log deleted and the `find_why_permits_save_rule` todo deleted.

### mission-control band: this session's pill and a hairline

- First a "▪ Folder: ~/…" line; the user found it unclear. Then a Box with `backgroundColor="subtle"`, a round border and `paddingX={1}`: on the desktop the border added a row above and below and the pill was far too big. Final: a `Button` with `variant="primary"` (drawn white on the desktop, the same size as the session pills) labelled `<project> · <session title>`, pressing does nothing.
- The title: the app's `list_sessions` leaves out the calling session (a trace in the ascendis session showed `listed=42 found=false`), so `ownTitle($)` calls `$.mcp.call('ccd_session_mgmt', 'get_session', { session_id: 'self' })` on each 5 s refresh; `hereLabel(title, cwd)` in `hooks/attention.ts` builds the label (project alone without a title); stored in a new `here` atom (`types/index.d.ts`).
- A hairline under the pill: `<Box width="100%" height={0.05} backgroundColor="subtle" />` on the desktop, a dim `'─'.repeat(e.props.bodyColumns)` on the terminal. The band now renders whenever the session's folder is known.
- Validation, `tsc -p .` and `claude plugin test` (15 tests) pass. `hooks/register.tsx` is about 480 lines.

## Decisions

- The session title reminder goes at the end of the context file only; inline contexts already end with the instruction.
- The permit fix keeps the buttons and their behaviour as they are (the user rejected a single primary button keyed across states) and asks for the keyboard again after redraws that remove the pressed button.
- This session's pill uses the surface's primary button so its size matches the other pills; its label carries the session title as well as the project, as a "you are here" marker.

## Key Files Affected

- `internal/cli/session_title.go`, `internal/cli/onboard.go`, `internal/cli/session_title_test.go`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`; tags `v2026.10.11.4` and `v0.7.10`; `.mem/todos/find_why_permits_save_rule.md` deleted.
- Outside the repository: `~/.claude/mods/mission-control/hooks/register.tsx` (`holdKeyboard`, `ownTitle`, `here` atom, the pill and hairline, tracing removed), `hooks/attention.ts` (`hereLabel`), `types/index.d.ts` (`here`), `tests/permit.test.tsx` (`ui.open` stub, keyboard assertion).

## Errors and Barriers

- A test assertion for `focus:save` failed: the test kit has no keyboard model, so `$.ui.focus` is denied before any test hook sees it; the test asserts only the `ui.open` request.
- `ui.open` hooks in tests must answer `{ value: … }`; returning the result directly was skipped by the kit and failed the type check.
- A trace written only when the stored label changed never fired in other sessions (the atom survives reloads); a once-per-load trace showed the session list lacks the calling session.
- This session's mod reloads only when its turn ends, so other sessions' traces were used to check the change within the turn.
