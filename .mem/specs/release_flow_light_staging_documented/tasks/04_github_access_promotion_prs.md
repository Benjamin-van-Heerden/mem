---
title: GitHub access for promotion PRs
status: todo
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:05:57+02:00"
---

internal/github (new): owner/repo from https/ssh remote URLs; token from GITHUB_TOKEN, GH_TOKEN, then gh auth token; REST create PR, list open PRs by head prefix, get PR with head sha and review state, delete ref; 20s timeouts; errors name the token options. project.ReleaseConfig{ProductionPR} under [release]; mem init --production-pr. Tests against httptest.
