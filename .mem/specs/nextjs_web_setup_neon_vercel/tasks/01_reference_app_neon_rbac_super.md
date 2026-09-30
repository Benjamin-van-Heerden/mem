---
title: 'Reference app: Neon, RBAC, super admin and seed'
status: todo
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:26:00+02:00"
---

In the scratchpad reference app (rebuild from the current nextjs-setup files if gone): switch to a Neon dev branch (ask the user before creating a throwaway Neon project), add SUPER_ADMIN_* to the env schemas, better-auth admin plugin with roles super_admin/admin/member and disableSignUp, permissions.ts, getCurrentUser returning role, requirePermission, scripts/seed.ts with the super-admin sync (adapted from anssum ensure-super-admin.ts/super-admin.ts, unique role super_admin), build = migrate && seed && next build, db:seed. Regenerate the auth schema (auth:schema) and migrations. Verify: seed idempotent, password rotation, super admin signs in, sign-up refused.
