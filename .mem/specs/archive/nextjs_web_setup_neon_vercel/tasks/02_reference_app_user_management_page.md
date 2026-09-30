---
title: 'Reference app: user management page'
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:36:46+02:00"
completed_at: "2026-09-30T13:36:46+02:00"
---

src/app/(app)/admin/users: list users, create user (name, email, initial password, role admin/member), change role, ban/unban via better-auth admin APIs in Server Actions guarded by requirePermission('users:manage'); the super admin is read-only in the UI. Verify with a created member who can sign in but gets no access to /admin/users.

## Completion Notes

ref2: src/features/users/{schema,data,actions,create-user-form,user-row-actions}.tsx and src/app/(app)/admin/users/page.tsx. getUsers checks users:manage, then an unexported use cache read tagged 'users'; Server Actions share a zod-generic run() (permission, validation, better-auth admin API with request headers, APIError -> {ok:false}, updateTag). The super admin row is read-only ('Managed by environment'). proxy matcher adds /admin. Verified in the browser as the super admin: created a member through the form (list refreshed), banned and unbanned via the row buttons; via curl: the member signs in and sees the dashboard, /admin/users renders the forbidden page, the member's direct create-user call is 403, a banned admin cannot sign in (403).
