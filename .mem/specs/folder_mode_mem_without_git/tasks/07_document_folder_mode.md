---
title: Document folder mode
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:13:07+02:00"
completed_at: "2026-11-01T15:13:07+02:00"
---

Update docs/design.md (the 'Plain files, ordinary Git' principle gets a carve-out for folder mode, and the sync and release sections say they apply to Git projects), docs/status.md, README.md, .mem/structure.md and the mem init help. Add a line on folder mode to the Developing mem section of AGENTS.md if a future session would otherwise miss it. Run dist/mem-dev through a whole session by hand in a disposable iCloud folder: init, onboard, memory, todo, log.

## Completion Notes

README, docs/design.md (a Folder projects section and a carve-out in the plain-files principle), docs/status.md, .mem/structure.md, the mem init help and the Developing mem section of AGENTS.md describe folder mode. Ran dist/mem-dev through init, todo, memory, structure, onboard, log new, refused commands and hook prompt in a throwaway iCloud Drive folder, which was deleted afterwards; a runnable's output reached onboard and a command run from a subfolder found the project. go test on project, agentsmd, checkpoint, structure, work, importer and cli -run 'Onboard|Init|Sync|Log|Single|Hook|Compact|Folder|Structure|Todo|Memory|Template', and go vet, pass.
