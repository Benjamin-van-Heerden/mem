---
title: nextjs-web setup with Neon, Vercel and RBAC
status: completed
assigned_to: benjamin_van_heerden
created_at: "2026-09-30T13:30:00+02:00"
updated_at: "2026-09-30T23:39:24+02:00"
completed_at: "2026-09-30T23:39:24+02:00"
---

## Overview

Rework the `nextjs-web` template in mem-templates (`~/Documents/yhat/mem-templates`) so its setup takes a new project from an empty repository to a deployed, branded app with RBAC and a responsive signed-in shell:

1. Scaffold the app and install the stack.
2. Neon: CLI logged in, project created by the agent, a `dev` branch for development and `main` for production. Local Postgres is dropped.
3. Auth barriers and route groups.
4. Branding: the user supplies a logo and answers the design questions.
5. A basic home page and a login page.
6. First production deploy with the Vercel CLI: project created, connected to GitHub, production only.
7. Migration and seed scripts.
8. RBAC by default, with exactly one super admin kept in sync with env vars.
9. The signed-in dashboard and navigation, modeled on shannon's: an icon rail and header on desktop, a sheet on mobile, and a dot-grid or diagonal-hatch background chosen by the user.
10. Verify, then deploy again.

The setup document may contain real code snippets. The verified starter files stay in the `nextjs-setup` skill.

## Goals

- A new web app goes from "empty GitHub repo" to "deployed, branded, super admin can sign in" by following `.mem/setup.md`, with the user acting only at the marked `(you)` points.
- The signed-in shell is responsive and production quality from the start, and its look comes from the user's answers, not generic defaults.
- RBAC and user management exist from day one.
- The setup and skills reflect what was actually run against Neon and Vercel.

## Technical Approach

All changes are in mem-templates `nextjs-web/`, built and verified in a reference app in the scratchpad (as for the first version), then copied into `skills/nextjs-setup/files/`.

### Environments and database

- Development and production both use Neon; local Postgres is no longer a prerequisite. One Neon project per app. Production is branch `main`. Development is branch `dev`, created from `main` and shared by the team; one developer can branch further.
- The agent checks `neon me` and asks the user to run `neon auth` when that fails. It creates the project with `neon projects create --name <app> --region-id <region near Vercel's region> -o json`, then `neon branches create --project-id <id> --name dev`. It writes `.env.local` from `neon connection-string dev --project-id <id> --pooled` (as `DATABASE_URL`) and `neon connection-string dev --project-id <id>` (as `DATABASE_URL_UNPOOLED`). Verify the exact flags against `neon <command> --help` (neon CLI 4.17) while building.
- Production URLs come from the `main` branch the same way and go into Vercel with `vercel env add <NAME> production --value … --yes`.
- `NEXT_PUBLIC_APP_ENV` stays `development | staging | production`. Staging is not deployed in this pass.

### Auth, RBAC and the super admin

- better-auth `admin` plugin, with `adminClient()` on the client. Roles: `super_admin` (exactly one: the account from env), `admin` (manages users) and `member` (default). `emailAndPassword.disableSignUp: true`.
- `src/features/auth/permissions.ts` has a typed `Permission` union, a role → permissions map and `hasPermission`. `getCurrentUser()` returns `role`. `requirePermission(p)` is in `session.ts`. Server Actions call it, and pages hide what the role lacks.
- Env: `SUPER_ADMIN_EMAIL`, `SUPER_ADMIN_NAME`, `SUPER_ADMIN_PASSWORD` (min 12), in `serverEnvSchema` and a `seedEnvSchema` subset.
- `scripts/seed.ts` runs after migrations. It holds the super-admin sync, adapted from anssum's `ensure-super-admin.ts` and `super-admin.ts`: in one transaction with a table lock it creates or updates the super admin (name, `super_admin` role, unbanned, credential password re-hashed only when it no longer verifies), and demotes any other `super_admin` to `admin`. It is idempotent, and further seed data can be added below. `build` becomes `bun scripts/migrate.ts && bun scripts/seed.ts && next build`, and `db:seed` runs it by hand. Changing the env vars and redeploying rotates the credentials.
- Minimal user management in `src/app/(app)/admin/users/`, needing `users:manage`: list users; create a user (name, email, initial password, role `admin`/`member`); change role; ban or unban. It uses better-auth admin APIs from Server Actions. The super admin cannot be changed from the UI.

### Route groups and pages

- `(site)`: a basic home page with the logo, a headline, one paragraph and a sign-in link, bespoke CSS on the tokens; `/login` (email and password, no sign-up).
- `(app)`: `/dashboard` (a welcome and placeholder cards), `/admin/users`.
- `proxy.ts` matches `/dashboard`, `/admin` and other app prefixes. Authorisation stays in `getCurrentUser`/`requirePermission`.

### Signed-in shell (modeled on shannon `src/frontend/components/navigation/`)

- `src/components/navigation/routes.ts`: `{ href, label, icon, permission? }[]` and a page-title lookup.
- `app-sidebar.tsx`: a fixed 64px icon rail (`md:` and up) with the logo on top, lucide icons, tooltips, `aria-current` and the active state. `app-mobile-menu.tsx`: a left sheet with labelled links and sign-out, below `md`. `app-header.tsx`: a sticky header with the page title, the user's email and sign-out.
- Cache Components: the layout does not await the session. The chrome renders statically; the parts that need the user (email, permission-filtered items) are in components inside `<Suspense>`, fed by `getCurrentUser()` (`use cache: private`).
- Backgrounds: two utilities in `globals.css` built on tokens, `.bg-dot-grid` (radial-gradient dots) and `.bg-hatch` (repeating-linear-gradient diagonals, as in the user's screenshot). The app shell uses the one the user picks, stored as a `data-background` attribute or a class on the shell.

### Branding and design questions

- The logo goes in `public/` as `logo.svg` (or `logo.png`) and is used in the rail, the mobile menu, the home page and the login page. `src/app/icon.*` is derived from it where possible.
- At the branding step the agent asks, and records in `globals.css` tokens and a short "Design decisions" note in `.mem/docs/design.md`:
  - corners: sharp (`--radius: 0`), slightly rounded or rounded
  - fonts: sans only, or a serif display with sans or mono UI (with suggested next/font pairs)
  - accent colour
  - light, dark or both
  - background: dot grid or diagonal hatch

### Deployment (production only)

- The agent checks `vercel whoami` and asks the user to run `vercel login` when that fails. The GitHub repository must exist (`git remote get-url origin`).
- `vercel project add <app>`, `vercel link --yes --project <app>`, `vercel git connect`. Set production env vars from the Neon `main` branch and generated secrets. `vercel.json` sets `"git": { "deploymentEnabled": { "<dev branch>": false, "<staging branch>": false } }` from `.mem/config.toml`, so only the production branch deploys.
- First deploy: `mem promote production` if the flow allows it this early, otherwise `vercel deploy --prod`. Determine which while verifying and write that into the setup.
- Vercel's production branch is mem's production branch (`main`).

### Setup document

`nextjs-web/setup.md` is rewritten in the order of the Overview, as checkbox steps with "Done when" lines and code snippets where they remove guesswork (env schema additions, seed script outline, vercel.json, CLI commands). `(you)` marks logging in to Neon, Vercel and GitHub, the logo and the design answers.

### Skills and memories

- `nextjs-setup`: the files list, deployment (production only, GitHub-connected, `deploymentEnabled`), "Adding a piece".
- `drizzle-neon`: the Neon CLI workflow, `dev`/`main` branches, seeding, no local Postgres.
- `better-auth`: roles, permissions, the super-admin sync, user management.
- `nextjs-env`: the super-admin variables.
- `design-system`: the app shell, backgrounds, the recorded design decisions.
- Memories are updated where a convention changes.

## Success Criteria

- The reference app built from the new starter files passes `bun run typecheck` and `bun run build` against a Neon `dev` branch. Locally:
  - the super admin from env can sign in
  - `/admin/users` creates a member, who can sign in but cannot open `/admin/users`
  - sign-up is refused
  - `bun scripts/seed.ts` is idempotent and changing `SUPER_ADMIN_PASSWORD` rotates the password
- The shell is checked in a browser at desktop and phone widths: rail with tooltips, header, mobile sheet, active states, both background utilities.
- `nextjs-web/setup.md` follows the order above, and each step has a "Done when" line.
- End to end, from an empty GitHub repository through `mem init --template nextjs-web` to a production deployment where the super admin signs in, with the steps followed as written. This creates real resources: a throwaway Neon project, a Vercel project and a GitHub repository. Ask the user before creating them, and delete them afterwards if the user wants.
- Skills and memories are updated. mem-templates is committed and pushed after the user reviews.

## Notes

- Answers from the user (2026-09-30): development on a Neon branch instead of local Postgres; Vercel connected to GitHub, with production deploying from `main` and other branches not deploying; super admin synced from env on every deploy; a minimal user-management page in the first pass; code snippets in the setup document are welcome.
- `git.deploymentEnabled` in vercel.json maps branch names to booleans and leaves unlisted branches enabled (vercel.json schema); list mem's non-production branches explicitly.
- shannon's chrome awaits the session at the top of its layout; under Cache Components it must stream instead (see the `nextjs-caching` and `better-auth` skills).
- anssum's super-admin sync uses role `admin` as the unique role; here the unique role is `super_admin`, so there can be other admins.
