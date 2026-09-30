---
title: 'Production release: draft then --confirm'
status: todo
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:05:57+02:00"
---

promote production writes or keeps .mem/local/release-notes.md (keep when the marker matches the plan's commit, replace and say so otherwise), prints the plan and the three-step instruction; --confirm reads the draft, refuses when missing/empty/stale, releases with it and deletes it. Remove --notes. Staging closing wording: preview only if the project deploys staging (covers todo promote_staging_wording_when_staging). Update promote tests.
