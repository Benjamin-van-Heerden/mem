---
title: Production PR flow
status: todo
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:05:57+02:00"
---

In production_pr repos: --confirm pushes promotion/production/<UTC ts> (MEM_PROMOTE), opens 'Release <tag>' PR with the notes, refuses if one is open; promote production with an open promotion PR reports it and, with --confirm, completes it by fast-forwarding production to the PR head (production ancestor of head, head equals snapshot branch), tags with the PR body, deletes the snapshot branch; changes requested refuses. Tests with httptest GitHub + bare remote. Then one real run on a throwaway GitHub repo (ask the user first) to confirm GitHub marks the PR merged after the fast-forward; implement the close-with-comment fallback if not; delete the repo.
