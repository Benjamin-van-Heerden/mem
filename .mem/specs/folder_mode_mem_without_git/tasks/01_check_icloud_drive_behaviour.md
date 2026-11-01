---
title: Check iCloud Drive behaviour
status: todo
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T14:52:34+02:00"
---

Run a throwaway test in a new folder under ~/Library/Mobile Documents/com~apple~CloudDocs, and delete it afterwards. Put in place a .mem folder, a .claude/settings.json, a .claude/skills/x/SKILL.md, a relative symlink and an executable script. After iCloud has uploaded them, evict each with 'brctl evict' and read it back, and record: do dot-folders sync, does the symlink survive, does the executable bit survive (internal/runnables runs .mem/runnables/* directly, and work's scripts/build needs it too), and does os.Stat on an evicted file leave it evicted. Also run a Claude Code session in that path and check that AGENTS.md is read and that a hook in .claude/settings.json runs, since the path has spaces and a tilde. Optimize Mac Storage appears to be on (defaults read com.apple.bird optimize-storage returns 1). Record the findings in the spec's Notes. If the executable bit does not survive, the next tasks must run runnables through sh instead. Change no mem code in this task.
