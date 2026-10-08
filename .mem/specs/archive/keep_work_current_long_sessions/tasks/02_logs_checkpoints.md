---
title: Logs at checkpoints
status: completed
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:17:01+02:00"
completed_at: "2026-10-08T13:17:01+02:00"
---

Reframe work logs as checkpoints, not session ends. internal/agentsmd/instructions.md Work Logs: a log records a stretch of work, written when a spec completes, before a release, when the commit nudge says so, and when the user wraps up; remove 'End a session with mem log new'. internal/cli/log.go: remove 'run mem log commit to close the session', the 'Close the session' help, 'Tell the user the session is closed', and the 'before the session ends'/'The session ends with' wording; the commit output reports the log committed and pushed. internal/work/log.go template guidance: 'this stretch of work' instead of 'this session' where it implies an ending. Compaction digest (internal/cli/compact.go) reports work commits since the last log or completed task with the same escalating wording, using the shared counter. Tests for the digest line and log commit output. Update docs/design.md, docs/status.md, README and the structure doc.

## Completion Notes

Instructions (Work Logs) now describe a log as a stretch of work written at checkpoints (spec completion, release, the commit nudge, the user wrapping up) and say a log does not end the session; removed 'End a session' and 'Ask the user before ending a session'. log.go: log/log new/log commit help and steps no longer mention closing the session; uncommitted-work nudge reworded; final instruction says the log is committed and pushed and to carry on. spec complete calls a completed spec a checkpoint and asks for a log. Log template guidance says 'a stretch of work since the previous log'. Compaction digest prints 'Work log: <checkpoint.LogLine>' and, from 5 (checkpoint.LogRequired), instructs writing a log now. Verified: new TestCompactHookAsksForAWorkLogOnceWorkHasGathered (2 commits: line without instruction; 5: escalated line and instruction), updated log commit test, full internal/cli suite, checkpoint/agentsmd/work tests, go vet. Docs: README, design, status, structure doc.
