---
title: Add mem hook stop, the Cursor work-log follow-up
status: todo
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:40:06+02:00"
---

New hidden subcommand mem hook stop in internal/cli/hook.go. It reads Cursor's stop input JSON from stdin: status (completed, aborted or error) and loop_count. It always prints a JSON object and exits 0. It prints {"followup_message": text} only when status is completed, loop_count is 0, checkpoint.Since gives Level() == Due, and Work.Summary() differs from the summary recorded in .mem/local/stop_nudge; after nudging it writes the summary there. Below Due it removes the file so the next due state nudges again. Otherwise it prints {}. Encode with SetEscapeHTML(false). Add checkpoint.TurnEndLine(w Work) next to MidTurnLine/TurnStartLine, empty unless Due. Its text: an automatic follow-up from mem, not from the user; the work summary; if the turn finished a piece of work, write the log now (mem log new, then mem log commit), then end the turn; if the turn ended waiting for the user's answer or in the middle of a change, say in one line that a log is due and end the turn without writing it; do not start other work. Tests: a cli test drives the command with stdin in a project with a due log (from enough work commits) and checks the follow-up, the dedup on a second call, aborted, loop_count 1, invalid stdin and a non-mem directory each giving {}. A checkpoint test covers TurnEndLine's levels.
