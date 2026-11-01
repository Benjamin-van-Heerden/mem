---
title: Reword records and logs for folders
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:10:23+02:00"
completed_at: "2026-11-01T15:10:23+02:00"
---

todo delete, memory set, structure and log new print commit or push wording today; give each folder-mode wording that stops at the file being saved. work.NewLog fills Key Files Affected from the changed files and leaves commits empty. structure creates the doc when missing and says it is kept by hand, with no drift. Test: stdout of each folder-mode command holds no mention of commits, pushes, branches or releases.

## Completion Notes

Folder-mode stdout stops at the saved file: todo delete and memory set drop their commit wording, and mem structure writes a folder template (FolderTemplate) with its own instructions, then says the doc is kept by hand. log new was done with the change detection task. TestFolderCommandsNeverTalkAboutGit drives init, onboard, structure, todo, memory, log and hooks with a blocked git and checks their output never mentions commits, pushes, branches, releases, repositories, the codebase or Git; go test on cli (-run 'Folder|Structure|Todo|Memory') and structure, and go vet, pass.
