---
title: 'Reference app: Neon, RBAC, super admin and seed'
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:34:00+02:00"
completed_at: "2026-09-30T13:34:00+02:00"
---

In the scratchpad reference app (rebuild from the current nextjs-setup files if gone): switch to a Neon dev branch (ask the user before creating a throwaway Neon project), add SUPER_ADMIN_* to the env schemas, better-auth admin plugin with roles super_admin/admin/member and disableSignUp, permissions.ts, getCurrentUser returning role, requirePermission, scripts/seed.ts with the super-admin sync (adapted from anssum ensure-super-admin.ts/super-admin.ts, unique role super_admin), build = migrate && seed && next build, db:seed. Regenerate the auth schema (auth:schema) and migrations. Verify: seed idempotent, password rotation, super admin signs in, sign-up refused.

## Completion Notes

Scratchpad reference app ref2 (fresh Radix build from the template files) on the throwaway Neon project tiny-poetry-63499905, branch dev (URLs via neon connection-string --pooled / direct). Added SUPER_ADMIN_* (superAdminEnvSchema, seedEnvSchema), permissions.ts (roles super_admin/admin/member, assignableRoles, Permission app:use/users:manage, hasPermission), better-auth admin plugin (defaultRole member, adminRoles super_admin+admin, adminAc/userAc) with disableSignUp and a before-hook that refuses any /admin/* call targeting the super admin or assigning super_admin; getCurrentUser returns role, requirePermission uses forbidden() (experimental.authInterrupts, src/app/forbidden.tsx); super-admin.ts syncSuperAdmin (lock, upsert, demote others to admin, rehash only when the password no longer verifies) and scripts/seed.ts; build = migrate && seed && next build, db:seed. Verified against Neon: typecheck and build pass; seed idempotent; sign-up 400; super admin signs in and sees the dashboard; create-user admin 200; super_admin assignment and set-role/ban/update on the super admin 403 even for an admin and for the super admin itself; changing SUPER_ADMIN_PASSWORD + seed rotates it (old 401, new 200); a stray super_admin is demoted to admin.
