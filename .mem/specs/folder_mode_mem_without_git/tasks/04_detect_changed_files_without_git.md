---
title: Detect changed files without Git
status: todo
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T14:52:34+02:00"
---

Add a stat-only folder walk returning the files changed since a time, skipping dot-files, dot-directories, .mem/, AGENTS.md, CLAUDE.md and [structure] ignore patterns, without structure.Relevant's built-in exclusions. Add a folder variant in internal/checkpoint that fills Work with a file count and sets Level to Suggested from 10 files and Due from 20, with the baseline at the newest log's created_at or the creation of .mem/config.toml. Wire it into hook prompt, hook compact and hook stop, and word the lines without any commit step. Walk the folder for structure.Tree in folder mode. Test: changed PDFs and Markdown count, ignored paths and dot-directories do not, a new log clears the nudge, and git is never run.
