---
title: Instructions and docs
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:21:55+02:00"
completed_at: "2026-10-01T00:21:55+02:00"
---

internal/agentsmd/instructions.md Releases section as in the spec (incl. the explicit mem deploy consent rule and PR completion only on the user's word); docs/design.md, docs/status.md, README.md, .mem/structure.md. Keep planned vs implemented distinct.

## Completion Notes

Managed instructions (internal/agentsmd/instructions.md): production releases refine the drafted notes, show the user and --confirm; PR-based projects complete the PR only when the user says it is approved; mem deploy is the user's command, never run without an explicit request; staging and production move only through mem promote and mem deploy. docs/design.md (production notes draft and --confirm, mem deploy, production_pr flow and its GitHub access, config example), README Releases, docs/status.md, .mem/structure.md (release/notes.go, internal/github, promote_pr.go, deploy.go, the promotion flow) updated. Tests for cli, release, github, project, agentsmd pass; go vet and gofmt clean.
