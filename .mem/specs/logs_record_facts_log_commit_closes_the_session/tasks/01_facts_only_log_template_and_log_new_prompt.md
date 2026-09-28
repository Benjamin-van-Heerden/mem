---
title: Facts-only log template and log new prompt
status: completed
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:05:37+02:00"
completed_at: "2026-09-28T09:05:37+02:00"
---

In internal/work/log.go replace the template: Overarching Goals, What Was Accomplished, Decisions, Key Files Affected, Errors and Barriers; no future-work section; guidance says the log is a record of the session and anything still to do belongs in a todo. In internal/cli/log.go make log new list open todos inline and instruct: fill the log; delete todos this session completed and create todos for anything left open (light wording, see spec); structure staleness as today; finish with mem log commit.

## Completion Notes

The log template (internal/work/log.go) now has Overarching Goals, What Was Accomplished, Decisions, Key Files Affected and Errors and Barriers, with a header comment saying the log records facts, is not updated later and that open work belongs in todos; What Comes Next is gone. mem log new lists open todos, then instructs: fill the placeholders with facts; delete todos the session completed and record anything still to be done as todos; structure staleness as before; finish with mem log commit. Verified with dist/mem-dev in a scratch repo with one open todo; go vet clean, work tests pass.
