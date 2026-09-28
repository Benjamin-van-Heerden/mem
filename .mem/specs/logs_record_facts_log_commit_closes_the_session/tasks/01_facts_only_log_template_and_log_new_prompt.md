---
title: Facts-only log template and log new prompt
status: todo
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:02:02+02:00"
---

In internal/work/log.go replace the template: Overarching Goals, What Was Accomplished, Decisions, Key Files Affected, Errors and Barriers; no future-work section; guidance says the log is a record of the session and anything still to do belongs in a todo. In internal/cli/log.go make log new list open todos inline and instruct: fill the log; delete todos this session completed and create todos for anything left open (light wording, see spec); structure staleness as today; finish with mem log commit.
