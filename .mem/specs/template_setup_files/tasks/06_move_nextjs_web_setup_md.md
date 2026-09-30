---
title: Move nextjs-web to setup.md
status: todo
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:56:30+02:00"
---

In ~/Documents/yhat/mem-templates: replace nextjs-web/docs/setup.md with nextjs-web/setup.md in the checkbox format (## [ ] N. Title, a 'Done when:' line per step, (you) steps last: Vercel/Neon linking, env vars per environment, first admin, brand/example replacement with the user). Move the numbered 'New project' procedure out of skills/nextjs-setup/SKILL.md into it; the skill keeps the files, 'Adding a piece', deployment reference and sharp edges and points to the setup. Step 1: create-next-app into scaffold-tmp with --skip-install --disable-git --no-agents-md, merge only .gitignore, fix the package name, then immediately add agentRules: false to next.config.ts. For projects scaffolded before mem init: delete CLAUDE.md and the nextjs-agent-rules block in AGENTS.md. Update the library README (layout: setup.md; drop the docs/setup.md convention). Commit locally; push only after the user reviews.
