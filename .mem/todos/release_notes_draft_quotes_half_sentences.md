---
title: Release notes draft quotes half-sentences from spec overviews
status: open
created_at: "2026-10-01T00:51:05+02:00"
---

Seen in the first real use of the drafted notes (v0.6.0, 2026-10-01): DraftNotes in internal/release/notes.go uses the first sentence of each completed spec's Overview, and overviews that open with a sentence ending in a colon followed by a list come out as fragments ('…a responsive signed-in shell:'). Options: fall back to the spec title only when the sentence ends with ':', or take the Goals' first bullet. Cosmetic; the agent refines the draft anyway.
