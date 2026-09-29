---
created_at: "2026-09-29T21:41:32+02:00"
user: benjamin_van_heerden
---

# Work Log - Slug fix, compaction hook verified, nextjs-web template built

## Overarching Goals

Clear the open todos, prepare the migration of projects from the Python harness (`.agent_core`) to mem, then build the `nextjs-web` template in mem-templates into a full stack template (many skills, memories and a setup doc), so that a new web app can be started with `mem init` alone.

## What Was Accomplished

### mem

- Record slugs: `recordSlug` (`internal/work/markdown.go`) splits titles on `[\s/\\.:,;]+`, so hyphenated words stay one unit, and drops filler and repeated words. Tests added to `TestLongTitlesGetShortUniqueSlugs`. Commit b267552.
- The post-compaction hook was verified in a real Claude Code `/compact`: the "🔄 MEM AFTER COMPACTION" digest reached the agent's context. Todo `verify_compaction_hook_claude_code` closed (38608dc).
- Migration survey: 29 repositories contain `.agent_core`; several have uncommitted work. A trial import of ev-server produced 1 open spec, 11 archived specs and 15 logs, and left `.agent_core` and its `.gitignore` entries in place.

### mem-templates: `nextjs-web` (pushed: 8e98d87, 82845c8, 49064f6)

A reference app was built in a scratch directory with bun, Next 16.3.7 (`cacheComponents: true`), Drizzle 0.45 on local Postgres, better-auth 1.7.6, shadcn 4.21 (`radix-nova`), Workflow 4.8 and a Vercel cron route. It was verified end to end: `bun run build` (Turbopack, `/dashboard` partially prerendered), the proxy redirecting to `/login`, sign-up and sign-in, a `use cache` read staying cached after a direct insert and refreshed by `updateTag` after a Server Action, the cron route rejecting requests without `CRON_SECRET` and starting a workflow whose steps ran.

The template now holds:

- House skills: `nextjs` (the Python harness's `coding_next.md`, updated for 16.3: `retry`, `instant`, segment configs removed under Cache Components, `preferredRegion` deprecated), `nextjs-setup` (procedure plus `files/`, the verified reference app), `nextjs-env`, `nextjs-caching`, `drizzle-neon`, `better-auth`, `design-system`, `background-jobs`.
- Upstream skills, taken fresh from GitHub with frontmatter reduced to name, description, licence and source commit, and a `LICENSE` file where required: `next-best-practices`, `next-cache-components`, `next-upgrade`, `workflow`, `shadcn` (vercel/vercel-plugin c632a50, Apache-2.0), `react-best-practices` (vercel-labs/agent-skills 063bee9, MIT), `neon-postgres` (neondatabase/agent-skills 80164a2, Apache-2.0), `vercel-cli` (vercel/vercel c628be7, Apache-2.0). `shadcn` was corrected against shadcn 4.21 (CLI flags, `--base base|radix|aria`, the font setup) and its Vercel design defaults replaced by a pointer to `design-system`; a single-argument `revalidateTag` in `next-best-practices` was updated.
- Memories: `bun-only`, `nextjs-house-skills`, `nextjs-server-boundary`, `nextjs-typed-env`, `additive-migrations`, `design-tokens`, `nextjs-cache-rules`, `nextjs-local-docs`.
- `docs/setup.md`: a one-time instruction printed at every onboard until deleted; it points to `nextjs-setup` and ends by deleting itself.
- README: the licence convention for copied skills and the `setup.md` convention.

Key patterns in the starter files:

```ts
// src/features/auth/session.ts: the data access layer's entry point
export async function getCurrentUser(): Promise<CurrentUser> {
  "use cache: private";
  const session = await auth.api.getSession({ headers: await headers() });
  if (!session) redirect("/login");
  return { id: session.user.id, name: session.user.name, email: session.user.email };
}
```

- `"build": "bun scripts/migrate.ts && next build"`; `scripts/migrate.ts` takes `pg_advisory_lock` on the unpooled URL and runs Drizzle's migrator.
- `src/env/{schema,client,server}.ts`; `next.config.ts` parses both schemas so the build fails on a bad environment; `databaseEnvSchema` is a separate object because zod 4 refuses `.pick()` on refined schemas.
- `"auth:schema": "bun --conditions=react-server scripts/auth-schema.ts"` calls `generateDrizzleSchema` from the `auth` package, because the better-auth CLI cannot load configs that import `server-only`.
- `next.config.ts` sets `agentRules: false`.

### Verified flows

- `mem init --template nextjs-web` in a disposable repo installs 16 skills, 8 memories and the doc; TypeScript, ESLint and `next build` ignore `.agents/`.
- Full flow from `mem init` in a repo with one commit: onboard printed `setup.md`; the skill's steps as written ended in a passing typecheck and build; `next dev` started by the agent did not add its block to AGENTS.md; after deleting `.mem/docs/setup.md`, onboard added `doc:setup` to `[templates] exclude` and it did not return.
- `mem init` in a clone of an empty repository fails: "this repository has no commits yet".

## Decisions

- The template enables Cache Components: in Next 16.3, `use cache` requires `cacheComponents: true` and `unstable_cache` is replaced, so it is the only supported way to cache server functions.
- Stack: bun, Vercel, Neon (local Postgres in development), Drizzle, better-auth in our own database, zod-typed env modules, Workflow for jobs, Vercel Cron that only starts workflows.
- Migrations run in the build and must be additive; one-off seeding is a hand-run script.
- One set of design tokens; shadcn for the signed-in `(app)` route group, bespoke CSS on the same tokens for the public `(site)` group.
- shadcn on Radix (`--base radix`), not the Base UI default of `init -d`, matching the user's existing projects and Vercel's AI Elements.
- House skills win over upstream skills; upstream skills are refreshed from their recorded source rather than edited, with local changes noted under `metadata.modified`.
- Reproducible scaffolding uses a template doc, `docs/setup.md`, deleted after setup; no template repositories or new mem mechanism.
- mem projects keep no `CLAUDE.md` and no Next.js agent block: create-next-app's `AGENTS.md` and `CLAUDE.md` are discarded, and `agentRules: false` keeps `next dev` (when started by an agent) from appending its block.
- Migration: an importer `--commit` mode will clean `.gitignore`, remove `.agent_core` and commit only the migration paths; no template assignment during import; pushes after user approval.

## Key Files Affected

- mem: `internal/work/markdown.go`, `internal/work/work_test.go`; `.mem/todos/` (closed `improve_slugs_cut_five_words` and `verify_compaction_hook_claude_code`; added `let_mem_init_start_empty` and `verify_nextjs_web_real_vercel`).
- mem-templates `nextjs-web/`: `template.toml` (new description); `memories/*.md` (7 new); `docs/setup.md` (new); `skills/nextjs/SKILL.md` (rewritten); new skill directories `nextjs-setup` (with `files/`: `next.config.ts`, `vercel.json`, `.env.example`, `drizzle.config.ts`, `scripts/{migrate,auth-schema}.ts`, `src/env/*`, `src/db/*`, `src/proxy.ts`, `src/app/**`, `src/features/{auth,notes}/*`), `nextjs-env`, `nextjs-caching`, `drizzle-neon`, `better-auth`, `design-system`, `background-jobs`, and the eight upstream skills.
- mem-templates `README.md`: nextjs-web description, licence and setup.md conventions.

## Errors and Barriers

- The better-auth CLI refuses configs whose import graph includes `server-only`, even with `NODE_OPTIONS=--conditions=react-server` (it loads through jiti). Solved with a bun script under `--conditions=react-server` calling the CLI's API. `@better-auth/cli` is stuck at 1.4 and its types mismatch better-auth 1.7; the current CLI is the `auth` package.
- `drizzle-kit` runs under node and does not see `.env.local`, including via `bun --bun`; `drizzle.config.ts` calls `loadEnvConfig` from `@next/env` (named import; the default import is undefined).
- `shadcn init` writes `--font-sans: var(--font-sans)` while create-next-app names fonts `--font-geist-*`, so text fell back to serif. Renaming the next/font variables to `--font-sans`/`--font-mono` fixes it (verified in the browser); the upstream shadcn skill's claim that `@theme inline` cannot use them is wrong for current versions.
- A Base UI `Button` wrapping a `Link` needs `nativeButton={false}`; resolved by the move to Radix and `buttonVariants()`.
- create-next-app refuses a directory containing `.mem/`, `.agents/` or `AGENTS.md`; the setup skill scaffolds into `scaffold-tmp` and moves the files in.
- `next dev` appends its agent block to AGENTS.md only when started by an AI agent and the block is missing; it does not overwrite the file.
- The shell sets `PYTHONPYCACHEPREFIX=.cache/pycache` (relative), so Python runs wrote bytecode into the working directory; those files were removed, and later runs used `PYTHONDONTWRITEBYTECODE=1`.
- Amending an already-cloned local commit made mem's cached clone of the local template source unable to fast-forward; that cache entry was deleted.
