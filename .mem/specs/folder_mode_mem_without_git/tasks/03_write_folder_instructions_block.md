---
title: Write the folder instructions block
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:02:21+02:00"
completed_at: "2026-11-01T15:02:21+02:00"
---

Add internal/agentsmd/instructions_folder.md and embed it; render selects by mode; Install and Refresh take the mode as they take single-branch today. Content as described in the spec's Technical Approach: document work with occasional scripts, no Git, branch, release, testing or module-structure rules, and no rule that gates overwriting or deleting files. Stop and show the user the full text for review before moving on. Test: a folder project gets this block, and a refresh leaves memories and text outside the block alone.

## Completion Notes

Added internal/agentsmd/instructions_folder.md as a separate embedded block. agentsmd.Kind (Staged, SingleBranch, Folder) replaces the single flag in Install, Refresh, ReplaceLegacy and render, and instructionKind in cli picks it from the config. TestFolderProjectsGetTheirOwnInstructions checks the block's content and that a refresh keeps memories and outside text. go test on agentsmd and importer passes. The text still needs the user's review before a release.
