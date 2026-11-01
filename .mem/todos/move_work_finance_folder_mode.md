---
title: Move work and finance into folder mode
created_at: "2026-11-01T15:13:25+02:00"
---

work is done: it lives at iCloud Drive/work as a folder project with skills for its Typst formats. What is left for work is the user's decision to delete ~/Documents/work and the GitHub repository Benjamin-van-Heerden/work once the new folder has proven itself.

For ~/Documents/yhat/finance, follow the same steps: copy the folder into iCloud Drive, leaving out .git, .agent_core and .gitignore; compare checksums; run mem init --folder; carry over logs, memories and docs from .agent_core by hand; turn reusable formats into skills under .claude/skills, with each new document self-contained; and list generated output under [structure] ignore. Delete the GitHub repository only when the user says so.
