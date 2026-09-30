---
title: Update skills and memories
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:45:52+02:00"
completed_at: "2026-09-30T13:45:52+02:00"
---

In mem-templates nextjs-web: drizzle-neon (Neon CLI workflow, dev/main branches, seed, no local Postgres), better-auth (roles, permissions, super-admin sync, user management), nextjs-env (SUPER_ADMIN_*), design-system (app shell, backgrounds, design decisions doc), nextjs-setup (deployment: production only, GitHub-connected, deploymentEnabled). Update memories whose conventions changed. Grep for stale mentions of local Postgres/createdb.

## Completion Notes

mem-templates 1b4686e (local): better-auth rewritten (config with admin plugin and disableSignUp, roles/permissions table, the super admin sync and guard, user management with the run() helper, session reads, proxy matcher, schema regeneration incl. bootstrap); drizzle-neon (Neon dev/main, no local Postgres, a Seed section, Neon CLI commands incl. the verified 'neon branches reset dev --parent', staging as a later addition); nextjs-env (SUPER_ADMIN_* group, production-only vercel env add via stdin, no env pull over .env.local); nextjs-setup deploy section (CLI project, GitHub connect, deploymentEnabled, env vars from Neon main, logs); design-system (Design decisions in .mem/docs/design.md, the signed-in shell, backgrounds). Memories: server-boundary (requirePermission, permissions not roles), additive-migrations (seed upserts), design-tokens (design.md), new single-super-admin. Grep finds no stale local-Postgres/createdb mentions in house skills.
