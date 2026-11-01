---
title: Move work and finance into folder mode
created_at: "2026-11-01T15:13:25+02:00"
---

A manual migration, after a mem release with folder mode. For each of ~/Documents/work and ~/Documents/yhat/finance: move the folder into iCloud Drive, carry over logs, memories and docs from .agent_core by hand, run mem init --folder, remove .agent_core, .git, .gitignore and the old AGENT_CORE block in AGENTS.md, and delete the GitHub repository. work's two .typ reference docs in .agent_core/docs become Typst skills under .claude/skills (letterhead, layout, invoice), with each new document self-contained; recurring work such as setting up a new client becomes skills too, with scripts for the mechanical steps; set [structure] ignore = ['_generated/**'] in work.
