---
title: Add mem hook stop, the Cursor work-log follow-up
status: completed
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:46:01+02:00"
completed_at: "2026-10-25T15:46:01+02:00"
---

New hidden subcommand mem hook stop in internal/cli/hook.go. It reads Cursor's stop input JSON from stdin: status (completed, aborted or error) and loop_count. It always prints a JSON object and exits 0. It prints {"followup_message": text} only when status is completed, loop_count is 0, checkpoint.Since gives Level() == Due, and Work.Summary() differs from the summary recorded in .mem/local/stop_nudge; after nudging it writes the summary there. Below Due it removes the file so the next due state nudges again. Otherwise it prints {}. Encode with SetEscapeHTML(false). Add checkpoint.TurnEndLine(w Work) next to MidTurnLine/TurnStartLine, empty unless Due. Its text: an automatic follow-up from mem, not from the user; the work summary; if the turn finished a piece of work, write the log now (mem log new, then mem log commit), then end the turn; if the turn ended waiting for the user's answer or in the middle of a change, say in one line that a log is due and end the turn without writing it; do not start other work. Tests: a cli test drives the command with stdin in a project with a due log (from enough work commits) and checks the follow-up, the dedup on a second call, aborted, loop_count 1, invalid stdin and a non-mem directory each giving {}. A checkpoint test covers TurnEndLine's levels.

## Completion Notes

Added the hidden mem hook stop: it reads Cursor's stop input ({status, loop_count}) and always prints a JSON object. It sends followup_message only for a completed turn with loop_count 0, a due work log, and a work summary that differs from the one recorded in .mem/local/stop_nudge; below the due level it removes the record. checkpoint.TurnEndLine carries the wording: an automatic follow-up from mem, not the user; write the log now if the turn finished a piece of work, otherwise say a log is due and end the turn. Verified: TestStopHookAsksForADueWorkLogOncePerStateOfWork (below due, aborted, error, loop_count 1, invalid stdin, the follow-up, the dedup, re-asking after more work, outside a mem project), TurnEndLine level checks in the checkpoint test; go test and go vet for internal/cli and internal/checkpoint pass.
