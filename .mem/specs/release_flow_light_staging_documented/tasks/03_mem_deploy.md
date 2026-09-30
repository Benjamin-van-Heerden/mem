---
title: mem deploy
status: todo
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:05:57+02:00"
---

internal/cli/deploy.go: staging to the development upstream head, then production to staging, direct fast-forwards, production tagged with GeneratedMessage; report-style stdout for a human; no-op hops reported; in production_pr repos fail before moving anything, with guidance; hidden --force (MarkHidden, never mentioned in help or instructions) overrides. Tests in a disposable repo.
