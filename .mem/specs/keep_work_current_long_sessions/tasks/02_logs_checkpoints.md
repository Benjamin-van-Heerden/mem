---
title: Logs at checkpoints
status: todo
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:07:22+02:00"
---

Reframe work logs as checkpoints, not session ends. internal/agentsmd/instructions.md Work Logs: a log records a stretch of work, written when a spec completes, before a release, when the commit nudge says so, and when the user wraps up; remove 'End a session with mem log new'. internal/cli/log.go: remove 'run mem log commit to close the session', the 'Close the session' help, 'Tell the user the session is closed', and the 'before the session ends'/'The session ends with' wording; the commit output reports the log committed and pushed. internal/work/log.go template guidance: 'this stretch of work' instead of 'this session' where it implies an ending. Compaction digest (internal/cli/compact.go) reports work commits since the last log or completed task with the same escalating wording, using the shared counter. Tests for the digest line and log commit output. Update docs/design.md, docs/status.md, README and the structure doc.
