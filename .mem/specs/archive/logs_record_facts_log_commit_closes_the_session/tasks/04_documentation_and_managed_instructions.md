---
title: Documentation and managed instructions
status: completed
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:12:00+02:00"
completed_at: "2026-09-28T09:12:00+02:00"
---

Update the Work Logs section of internal/agentsmd/instructions.md (logs record facts and are not updated; open work goes into todos; end a session with mem log new then mem log commit), docs/design.md (work records and session continuation), docs/status.md, README.md and .mem/structure.md.

## Completion Notes

Managed instructions: Work Logs says logs record facts, open work is a todo, and a session ends with mem log new then mem log commit; Staying in Sync says mem log commit pushes at session end. design.md: todos cover blockers and waiting decisions; Session continuation describes records as the home of open work and the log new/log commit flow and onboard's log selection. status.md, README.md and .mem/structure.md updated for the new onboard log and todo display and log commit. spec complete now offers mem log new then mem log commit. go vet ./... clean.
