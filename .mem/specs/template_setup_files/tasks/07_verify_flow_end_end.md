---
title: Verify the flow end to end
status: todo
created_at: "2026-09-30T11:56:31+02:00"
updated_at: "2026-09-30T11:56:31+02:00"
---

Build dist/mem-dev. In the scratchpad: bare repo, clone it empty, dist/mem-dev init --template nextjs-web --template-source ~/Documents/yhat/mem-templates --protect=false (clear a stale cache entry for that local source under ~/Library/Caches/mem/templates if its history diverged). Confirm onboard shows 🏗️ SETUP and the instruction; follow the setup steps as written (local Postgres, bun); confirm bun run typecheck and build pass, AGENTS.md has no nextjs-agent-rules block and there is no CLAUDE.md, also after running next dev from the agent; tick boxes and check onboard progress; delete .mem/setup.md and confirm onboard is silent about setup. Fix anything in the steps that does not work as written. Then ask the user to review and push mem-templates.
