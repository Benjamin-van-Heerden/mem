---
title: Find why permit's Save rule needs two clicks
created_at: "2026-10-11T22:30:05+02:00"
---

In the mission-control mod (~/.claude/mods/mission-control), the permit review pane's 'Save rule' button needs two clicks after a draft is ready, on the desktop app. The pane already holds focus from the 'Draft a rule' click, so it is not the first click moving focus (user confirmed). Temporary tracing is in hooks/register.tsx (the trace() helper marked 'DEBUG (temporary)', calls in makeDraft and save(), ui.press and ui.focus hooks on the pane, a trace line in the pane render), writing to ~/.claude/mission-control-debug.log. Once the user has reproduced it: read the log to see whether the first click reaches ui.press, whether save() runs and with which draft, whether $.ui.focus(save) was denied and whether a render replaced the button in between; fix the cause, add a test, remove the tracing, delete the log.
