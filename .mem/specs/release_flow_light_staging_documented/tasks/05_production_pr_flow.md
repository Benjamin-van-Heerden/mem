---
title: Production PR flow
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:20:15+02:00"
completed_at: "2026-10-01T00:20:15+02:00"
---

In production_pr repos: --confirm pushes promotion/production/<UTC ts> (MEM_PROMOTE), opens 'Release <tag>' PR with the notes, refuses if one is open; promote production with an open promotion PR reports it and, with --confirm, completes it by fast-forwarding production to the PR head (production ancestor of head, head equals snapshot branch), tags with the PR body, deletes the snapshot branch; changes requested refuses. Tests with httptest GitHub + bare remote. Then one real run on a throwaway GitHub repo (ask the user first) to confirm GitHub marks the PR merged after the fast-forward; implement the close-with-comment fallback if not; delete the repo.

## Completion Notes

internal/cli/promote_pr.go: with [release] production_pr, promote production first checks for an open promotion PR (head prefix promotion/<production>/): without --confirm it reports URL, head and approvals; with --confirm it refuses on changes requested or when the snapshot branch no longer matches the PR head, then fast-forwards production to the PR head via release.Prepare(production, head) and Execute, tags with the PR body (heading updated to the actual tag), waits up to 15s for GitHub to mark the PR merged, closes it with a comment naming the tag if not, then deletes the snapshot branch. Without an open PR, --confirm pushes promotion/<production>/<UTC ts> and opens 'Release: staging → production (<date>)' with the notes, deleting the draft; the draft instruction says --confirm opens a PR. Remote identity from the raw remote URL (insteadOf-safe). Bug found and fixed: git tag's default cleanup stripped '#' Markdown headings from release notes; Execute now uses --cleanup=whitespace (tests tightened). promote_pr_test.go against a fake GitHub backed by the bare repo: open, status, changes requested refused, approved completion (main = head, merged, snapshot deleted, tag keeps heading), and the close-with-comment fallback. Real GitHub (throwaway private repo, deleted afterwards): after mem's fast-forward the PR reported state=closed merged=true with merge_commit_sha equal to the head (no new commit), main linear, snapshot deleted, tag notes intact.
