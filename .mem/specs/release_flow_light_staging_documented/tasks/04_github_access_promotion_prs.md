---
title: GitHub access for promotion PRs
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:16:11+02:00"
completed_at: "2026-10-01T00:16:11+02:00"
---

internal/github (new): owner/repo from https/ssh remote URLs; token from GITHUB_TOKEN, GH_TOKEN, then gh auth token; REST create PR, list open PRs by head prefix, get PR with head sha and review state, delete ref; 20s timeouts; errors name the token options. project.ReleaseConfig{ProductionPR} under [release]; mem init --production-pr. Tests against httptest.

## Completion Notes

internal/github (new, net/http only): New parses owner/repo from https (incl. credentials), git@ and ssh:// GitHub remotes, takes the token from GITHUB_TOKEN, GH_TOKEN or 'gh auth token' (error names the options), API base overridable with MEM_GITHUB_API for tests; CreatePR, OpenPRs (by base and head prefix), PR, Reviews (latest verdict per reviewer: approvals, changes requested), Comment, Close; API errors carry GitHub's messages; 20s timeout. Snapshot branches are deleted with git, so no ref API. mem init --production-pr sets [release] production_pr. github_test.go covers the URL forms, a non-GitHub remote, a missing token, PR create/list/review aggregation and error messages against httptest; go vet passes.
