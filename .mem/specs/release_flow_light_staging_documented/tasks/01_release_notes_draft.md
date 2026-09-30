---
title: Release notes draft
status: completed
created_at: "2026-10-01T00:05:57+02:00"
updated_at: "2026-10-01T00:12:09+02:00"
completed_at: "2026-10-01T00:12:09+02:00"
---

internal/release/notes.go: DraftNotes (tag heading; completed specs with titles and overview line; work log titles and What Was Accomplished headings committed in the range; commits capped at 30) with a <!-- mem:release <sha> --> marker; GeneratedMessage for deploy. Tests with a disposable repo and bare remote.

## Completion Notes

internal/release/notes.go: NotesPath (.mem/local/release-notes.md); DraftNotes writes a <!-- mem:release <sha> --> marker, '# Release <tag>', Specs completed (title and first overview sentence of specs archived as completed in the range), Work (titles of work logs added in the range with their What Was Accomplished ### headings), Commits (capped at 30); GeneratedMessage (tag, completed spec titles, commit subjects) for mem deploy; DraftCommit parses the marker. completedSpecs now reuses specSummaries. notes_test.go covers a range with a completed and an abandoned spec, a log, and the previous release's commits excluded; release tests and go vet pass.
