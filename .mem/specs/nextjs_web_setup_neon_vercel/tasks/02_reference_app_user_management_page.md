---
title: 'Reference app: user management page'
status: todo
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:26:00+02:00"
---

src/app/(app)/admin/users: list users, create user (name, email, initial password, role admin/member), change role, ban/unban via better-auth admin APIs in Server Actions guarded by requirePermission('users:manage'); the super admin is read-only in the UI. Verify with a created member who can sign in but gets no access to /admin/users.
