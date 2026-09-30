---
created_at: "2026-10-01T00:50:57+02:00"
user: benjamin_van_heerden
---

# Work Log - Template setup files, nextjs-web on Neon/Vercel with RBAC, release flow and mem v0.6.0

## Overarching Goals

Make templates able to scaffold a whole app reproducibly (a one-time setup document that onboard leads with), rebuild the nextjs-web setup around the user's actual stack (Neon, Vercel, RBAC with one super admin, a shannon-style signed-in shell) and prove it with an unaided agent, then rework releases so the user can deploy without an agent, production releases by agents are documented, and some repositories can require pull requests, and release all of it as mem v0.6.0.

## What Was Accomplished

### Spec `template_setup_files` (archived)

- `mem init` works in a repository without commits: `firstCommit` (`internal/cli/init.go`) builds an empty commit on the production branch from the empty tree (`write-tree` against a temporary `GIT_INDEX_FILE`, `commit-tree`, `update-ref`, `symbolic-ref`), so staged files stay staged.
- `internal/templates/setup.go`: `SetupPath` (`.mem/setup.md`), `Library.Setup(names)` joins the templates' root `setup.md` files in order (CRLF normalised), `SetupProgress` counts `## [ ]`/`## [x]` step headings; `Template.HasSetup`.
- `mem init --template` writes `.mem/setup.md`; `template use` never installs it, only mentions it; `template list` has a SETUP column. Not a template item: never synced, locked or excluded; deleting the file ends the setup.
- `internal/cli/onboard_setup.go`: 🏗️ SETUP section first in the context with "N of M steps done"; the onboard instruction leads with the setup and ends there while steps remain; with all boxes ticked it says to delete the file. The compaction digest adds "Setup pending: N of M". Steps marked `(you)` are asked for, then the agent continues.
- Init and template instructions name skill directories by count (`describePaths`) instead of listing ~35 paths.

### Spec `nextjs_web_setup_neon_vercel` (archived; mem-templates pushed)

- Reference app built against a throwaway Neon project: better-auth `admin` plugin with roles `super_admin`/`admin`/`member`, `disableSignUp`, a `before` hook (`protectSuperAdmin`) refusing any `/admin/*` call that targets the super admin or assigns `super_admin`; `permissions.ts`; `requirePermission` with `forbidden()` (auth interrupts); `syncSuperAdmin` + `scripts/seed.ts` run in every build after migrations (`build = migrate && seed && next build`); `/admin/users` (create, change role, ban/unban via Server Actions over the better-auth admin API); navigation shell ported from shannon (64px icon rail with tooltips, header, mobile sheet) streaming user-dependent parts under Cache Components; `bg-dot-grid`/`bg-hatch` utilities; basic home and login pages. Verified with curl and the browser (desktop and phone).
- Template: starter files replaced; `setup.md` rewritten as 12 checkbox steps in the user's order (scaffold, stack, starter files, Neon, auth barriers, branding (you), home/login, production deploy, migrations/seed, roles, dashboard/background (you), verify) with code snippets; skills (`better-auth` rewritten, `drizzle-neon`, `nextjs-env`, `nextjs-setup`, `design-system`) and memories (new `single-super-admin`) updated.
- End-to-end test: a fresh subagent given only the empty repo, a mem shim and "start a new web app with mem's nextjs-web template" took it to a deployed app ("Delta") on Vercel via three `mem promote` releases, with this session playing the user. Its 20 findings were fixed: `.env.local` snippet bug, `"framework": "nextjs"` in vercel.json, production `SUPER_ADMIN_PASSWORD` set by the user (`vercel env add … --sensitive`), repository must live where Vercel's GitHub app is installed, `scripts/smoke.ts` (`bun run smoke [url]`) for checks without seeing secrets, README `if/then`, font snippet, truthful migration output, `sslmode=require` → `verify-full` transform, notes on blocked install scripts, tooltip hint, `VERCEL_OIDC_TOKEN`, forbidden pages answering 200, account creation on production as `(you)`.
- All throwaway resources torn down (Vercel project, two Neon projects, GitHub repositories, local clones and credential files).

### Spec `release_flow_light_staging_documented` (archived)

- `internal/release/notes.go`: `DraftNotes` (marker `<!-- mem:release <sha> -->`, specs completed with title and first overview sentence, work log titles with their accomplishment headings, commits), `GeneratedMessage`, `DraftCommit`.
- `mem promote production` writes/keeps `.mem/local/release-notes.md` and asks the agent to refine, show and `--confirm`; `--confirm` refuses missing, empty or stale drafts and removes the draft after release. `--notes` removed. Staging wording no longer promises a deployed preview.
- `mem deploy` (`internal/cli/deploy.go`): pushes development, fast-forwards staging and production, tags with `GeneratedMessage`, report-style output; refuses `[release] production_pr` repositories unless given the hidden `--force`.
- `internal/github`: owner/repo from GitHub remote URLs, token from `GITHUB_TOKEN`/`GH_TOKEN`/`gh auth token`, pull request create/list/get/reviews/comment/close; `MEM_GITHUB_API` for tests. `mem init --production-pr`.
- `internal/cli/promote_pr.go`: `--confirm` pushes `promotion/<production>/<ts>` and opens a pull request with the notes; with one open, `--confirm` completes it by fast-forwarding production to its head (refusing on changes requested or a moved snapshot), tags with its body, waits for GitHub to mark it merged (else comments and closes), then deletes the snapshot branch.
- Verified on a throwaway GitHub repository: after the fast-forward GitHub reported the pull request `merged=true` with `merge_commit_sha` equal to the head.
- Bug fixed: `git tag` stripped `#` Markdown headings from release notes; `Execute` now tags with `--cleanup=whitespace`.

### Release mem v0.6.0

- CI had been failing on Windows since the setup-files push: `Library.Setup` copied CRLF from Windows checkouts into `.mem/setup.md`; fixed by normalising line endings (b484c01), CI green on all three platforms.
- `mem promote staging`, then the new flow with `dist/mem-dev`: drafted notes refined and shown to the user, `mem promote production --confirm` (v2026.10.01.1), then `v0.6.0` tagged with `--cleanup=whitespace`; Release workflow succeeded with 7 assets; installed mem updated v0.5.0 → v0.6.0.

## Decisions

- Template setup is a single `setup.md` with checkbox step headings, not a spec or a new command set: the file is the state and deleting it ends the setup.
- create-next-app's `AGENTS.md`/`CLAUDE.md` are not wanted: scaffold with `--no-agents-md` and set `agentRules: false` immediately; mem projects keep no `CLAUDE.md`.
- nextjs-web: shadcn on Radix; development and production both on Neon (no local Postgres); Vercel connected to GitHub with only production deploying (`git.deploymentEnabled`); exactly one super admin owned by `SUPER_ADMIN_*` and synced on every build; minimal user management from day one.
- End-to-end validation of templates uses a fresh subagent with vague instructions; this session plays the user without steering. When the subagent asked this session to sign in with a password it had been denied, that was declined and surfaced to the user.
- Keep staging (for eventual preview deployments; one code path). Staging promotions are light; production promotions by agents are documented and confirmed by the user; `mem deploy` is the user's command and agents need explicit consent to run it. Some repositories require production pull requests; mem completes them by fast-forward, never GitHub's merge button. No reliance on `gh` being logged in outside those repositories. In PR-required repositories `mem deploy` fails up front; a hidden `--force` overrides.
- Check CI after every push to mem, not only before a release.

## Key Files Affected

- mem: `internal/cli/init.go` (`firstCommit`, setup install, `--production-pr`, `describePaths` use), `internal/cli/template.go` (`describePaths`, setup note, SETUP column), `internal/cli/onboard.go`, `onboard_setup.go` (new), `compact.go`, `internal/templates/setup.go`/`setup_test.go` (new), `internal/templates/library.go` (`HasSetup`), `internal/release/notes.go`/`notes_test.go` (new), `internal/release/release.go` (`--cleanup=whitespace`, `completedSpecs` via `specSummaries`), `internal/cli/promote.go` (draft/`--confirm`), `promote_pr.go`, `deploy.go` (new) with tests, `internal/github/` (new), `internal/project/project.go` (`ReleaseConfig`), `internal/agentsmd/instructions.md` (Releases), `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md`.
- mem-templates `nextjs-web/`: `setup.md` (new, 12 steps), `skills/nextjs-setup/files/` (replaced: RBAC, seed, smoke, users, navigation shell, env with super admin and sslmode transform, vercel.json with framework and deploymentEnabled), `skills/{better-auth,drizzle-neon,nextjs-env,nextjs-setup,design-system}/SKILL.md`, `memories/` (updated, `single-super-admin.md` new), `docs/setup.md` removed, `README.md` (setup convention).
- Todos: deleted `let_mem_init_start_empty`, `promote_staging_wording_when_staging`, `verify_nextjs_web_real_vercel`; added `let_task_spec_completion_commit`, `release_notes_draft_quotes_half_sentences`.

## Errors and Barriers

- `sed` edits with `$`/backticks and a following regex `grep` gave misleading "not applied" results; a second edit then doubled a replacer (`() => () =>`) in the setup snippet, which the subagent caught. Use fixed-string checks (`grep -F`) and one edit tool.
- zsh: an unquoted `$M` holding `env VAR=1 cmd` is not word-split (the command never ran); a function parameter named `path` clobbers `PATH`. Use shell functions and other names.
- `rm -rf $S/*` with an unguarded variable is blocked by the safety check; use literal paths or `"${S:?}"`.
- better-auth CLI cannot load configs importing `server-only` (worked around with `auth:schema` earlier); Next.js embeds the `forbidden.tsx` boundary in page HTML, so "No access" text cannot signal a denied page (the smoke script checks the user's email instead).
- A Vercel project created before any code gets the "Other" preset (404s); team variables are stored as sensitive and cannot be read back.
- `vercel project rm` has no `--yes`; confirm through stdin.
- Postgres resolves a missing table even in an unused `CASE` branch; check `to_regclass` in a separate query.
- CI went red on Windows for a day because it was not checked after pushes.
