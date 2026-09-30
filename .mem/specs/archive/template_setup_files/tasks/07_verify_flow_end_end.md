---
title: Verify the flow end to end
status: completed
created_at: "2026-09-30T11:56:31+02:00"
updated_at: "2026-09-30T12:07:04+02:00"
completed_at: "2026-09-30T12:07:04+02:00"
---

Build dist/mem-dev. In the scratchpad: bare repo, clone it empty, dist/mem-dev init --template nextjs-web --template-source ~/Documents/yhat/mem-templates --protect=false (clear a stale cache entry for that local source under ~/Library/Caches/mem/templates if its history diverged). Confirm onboard shows 🏗️ SETUP and the instruction; follow the setup steps as written (local Postgres, bun); confirm bun run typecheck and build pass, AGENTS.md has no nextjs-agent-rules block and there is no CLAUDE.md, also after running next dev from the agent; tick boxes and check onboard progress; delete .mem/setup.md and confirm onboard is silent about setup. Fix anything in the steps that does not work as written. Then ask the user to review and push mem-templates.

## Completion Notes

Built dist/mem-dev; in a clone of an empty bare repo, init --template nextjs-web (local mem-templates) made the first commit, wrote .mem/setup.md, and onboard showed 0 of 10 with the setup-first instruction. Followed setup.md as written with bun 1.4 and local Postgres, ticking and committing each step: all Done-when checks held; typecheck and build passed; /dashboard redirected to /login, sign-up and the dashboard worked; after an agent-run next dev AGENTS.md had no nextjs-agent-rules block and there was no CLAUDE.md; .env.local never committed. The digest showed 'Setup pending: 1 of 10'; with all ticked onboard said to delete the file; after deletion onboard printed nothing about setup. Fixes from the run: the scaffold keeps an existing README (tested with and without one) and step 5 ignores /.swc, which the first build adds (mem-templates b7b7157); init/template instructions now name skill directories by count instead of listing 32 paths (cli and templates tests pass).
