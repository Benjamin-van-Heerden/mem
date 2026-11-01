---
title: Detect changed files without Git
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:05:23+02:00"
completed_at: "2026-11-01T15:05:23+02:00"
---

Add a stat-only folder walk returning the files changed since a time, skipping dot-files, dot-directories, .mem/, AGENTS.md, CLAUDE.md and [structure] ignore patterns, without structure.Relevant's built-in exclusions. Add a folder variant in internal/checkpoint that fills Work with a file count and sets Level to Suggested from 10 files and Due from 20, with the baseline at the newest log's created_at or the creation of .mem/config.toml. Wire it into hook prompt, hook compact and hook stop, and word the lines without any commit step. Walk the folder for structure.Tree in folder mode. Test: changed PDFs and Markdown count, ignored paths and dot-directories do not, a new log clears the nudge, and git is never run.

## Completion Notes

structure.ChangedSince walks a folder project with stats only, counting PDFs and Markdown and skipping dot-paths, the agent instructions and [structure] ignore patterns. checkpoint.FolderSince and FolderCover measure from the newest log or the config's time, and Work.Folder words the nudges without a commit step. hook prompt, stop and compact use workSince, compact gets a Git-free digest, log new lists the changed files, and Tree and Measure work without Git. TestFolderWorkLogNudgeFollowsChangedFiles, TestFolderCompactionDigestStaysOffGit, TestChangedSinceCountsDocumentsAndSkipsWhatIsNotWork and TestTreeOfAFolderProjectNeedsNoGit pass, as do go test on structure, work and checkpoint, cli -run 'Folder|Log|Compact|Hook', and go vet.
