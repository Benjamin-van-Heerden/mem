---
title: 'Production release: draft then --confirm'
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:13:39+02:00"
completed_at: "2026-10-01T00:13:39+02:00"
---

promote production writes or keeps .mem/local/release-notes.md (keep when the marker matches the plan's commit, replace and say so otherwise), prints the plan and the three-step instruction; --confirm reads the draft, refuses when missing/empty/stale, releases with it and deletes it. Remove --notes. Staging closing wording: preview only if the project deploys staging (covers todo promote_staging_wording_when_staging). Update promote tests.

## Completion Notes

promote production without --confirm prints the plan and writes .mem/local/release-notes.md (writeDraft: keeps a draft whose marker matches the plan's commit, redrafts otherwise and says so) with the three-step instruction; --confirm (confirmedNotes) refuses a missing, empty or stale draft, tags with the notes without the marker, and removes the draft after the release. --notes removed. Staging's closing instruction no longer promises a deployed preview (todo promote_staging_wording_when_staging deleted); production's says to check deployments where CI deploys. promote_test.go covers: confirm without draft refused, nothing pushed on the first run, an edited draft kept on re-run, the tag carrying the edited notes, the draft removed, a stale draft refused. cli and release tests, go vet pass.
