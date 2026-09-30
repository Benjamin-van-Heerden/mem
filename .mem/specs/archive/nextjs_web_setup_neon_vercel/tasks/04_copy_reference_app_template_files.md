---
title: Copy the reference app into the template files
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:41:10+02:00"
completed_at: "2026-09-30T13:41:10+02:00"
---

Replace skills/nextjs-setup/files/ in mem-templates with the verified reference app's source (no node_modules, .next, .env.local, drizzle/ migrations or generated auth.ts content beyond what setup regenerates; decide and document whether to ship the generated auth schema). Update the files list in skills/nextjs-setup/SKILL.md.

## Completion Notes

mem-templates c6f28ae (local): skills/nextjs-setup/files/ replaced with the verified ref2 source (49 files): config, vercel.json (fra1, deploymentEnabled dev/test false, cron), .env.example (Neon URLs, SUPER_ADMIN_*), env, db incl. the generated auth.ts (so setup no longer needs the empty-schema bootstrap), migrate/seed/auth-schema scripts, auth with RBAC and super-admin guard, users feature and page, navigation shell, site home and login, placeholder logo, notes example and cron workflow. Not shipped: shadcn ui/utils and globals.css (generated; setup appends the background utilities) and drizzle/. SKILL.md 'What files/ contains' rewritten.
