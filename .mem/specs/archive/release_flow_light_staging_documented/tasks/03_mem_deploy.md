---
title: mem deploy
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:15:02+02:00"
completed_at: "2026-10-01T00:15:02+02:00"
---

internal/cli/deploy.go: staging to the development upstream head, then production to staging, direct fast-forwards, production tagged with GeneratedMessage; report-style stdout for a human; no-op hops reported; in production_pr repos fail before moving anything, with guidance; hidden --force (MarkHidden, never mentioned in help or instructions) overrides. Tests in a disposable repo.

## Completion Notes

internal/cli/deploy.go: mem deploy pushes unpushed development commits (plain push, points to mem sync on failure; warns that uncommitted changes are not deployed), then fast-forwards staging and production via release.Prepare/Execute, tagging production with release.GeneratedMessage; hops already current are reported; diverged stages stop with the existing divergedError; stdout is a report with no agent instruction. project.ReleaseConfig{ProductionPR} under [release] (added here, used by task 4/5). In production_pr repos deploy fails before moving anything; the hidden --force (MarkHidden) overrides. deploy_test.go: one run pushes dev and moves test and main to the dev head with a generated tag message, a second run reports both current; a PR-required repo refuses without moving staging, --force is absent from --help and deploys. Output checked by eye in a scratch repo.
