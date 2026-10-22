---
created_at: "2026-10-22T13:57:36+02:00"
user: benjamin_van_heerden
commits: [6c9849e]
---

# Work Log - mem v0.8.4: template memory changes become todos; nextjs-web gets a repository layer

## Goal

Give the nextjs-web template a proper data access layer, modelled on the repositories in Benjamin's other projects, and make sure conventions that arrive through a template reach the existing code of projects already using it, not only new work.

## What Was Accomplished

### Repository pattern review

Compared `src/db/repositories` in studii-all-risk (Drizzle), shannon (Tortoise) and ev-ledger (raw psycopg). ev-ledger is the model: functions per table take the cursor first, callers own transactions (`DBTxnCtx`), results are frozen dataclasses, and only `utils/migration_tool.py` runs SQL outside it. shannon's repositories take `conn` too, but 46 modules outside `src/db/` query models directly. studii-all-risk's repositories import the global `db`, so they cannot join a transaction, and `createClaimRecord` computes the next claim number with read-max-then-insert outside any transaction or lock; `src/auth/server.ts` and `src/agents/nathan/run.ts` import `db` directly.

shannon: todo `route_all_database_access_through` committed and pushed on its `dev` (465e5e4), listing the offenders by area and the grep that finds them.

### nextjs-web: every query lives in src/db/repositories (mem-templates c43a99f)

- Starter files: `src/db/repositories/notes.ts` (`listForUser`, `create`, `countSince`) and `users.ts` (`list`, `listIds`, `getRole`); `src/db/client.ts` exports `Tx` and `Db = typeof db | Tx`. `notes/data.ts`, `notes/actions.ts`, `notes/digest-steps.ts`, `users/data.ts`, the cron route and the super-admin guard in `auth/server.ts` call repositories (`import * as notesRepo from "@/db/repositories/notes"`; `notesRepo.listForUser(db, user.id)`).
- New memory `data-access-repositories`; `drizzle-neon` gains a Repositories section (rules, a caller table, a transaction example, exceptions: `src/db/`, the better-auth config, the seed); `nextjs`, `nextjs-caching`, `background-jobs`, `better-auth` and `nextjs-setup` updated to match.
- `setup.md` step 3 adds two `no-restricted-imports` blocks to `eslint.config.mjs`: outside `src/db/**` and `src/features/auth/server.ts` no `drizzle-orm` or `@/db/schema`; no `.tsx` file imports `@/db/*`. Its "Done when" includes `bun run lint`.
- Verified in a scratch app scaffolded with the setup's commands (next 16.4.2, drizzle-orm 0.45.4, better-auth 1.7.12): baseline `tsc` and lint clean; after the change `tsc` clean, lint 0 errors, `next build` succeeds (with a stub `globals.css`, see Notes); two planted violations each raised the expected lint error.

### Convention todos (6c9849e)

- `internal/templates/sync.go`: `Result.Memories []MemoryChange{Name, Template, Body, Previous}`, filled by `noteMemory` when `reconcile` installs a memory (`StateMissing`, Previous empty) or replaces it (`StateUpdated`, Previous read from AGENTS.md before the install).
- New `internal/cli/template_conventions.go`: `conventionTodos` writes "Apply the template convention <name>" per memory (text quoted, previous text under "It used to read:", then the check instruction), skips projects with `.mem/setup.md`, and rewrites an open todo with the same title through `upsertTodo`; it returns `📌` lines and the todo paths. `templateUseTodo` writes one "Apply the <template> template's conventions" todo listing the memories.
- `syncTemplates` adds the todo paths to the published "Sync template items" commit. Onboard (`state.conventionTodos`, `conventionTodoInstruction`), `mem sync` and the compaction digest instruct the agent to raise 📌 items; `template use` adds a numbered instruction for its todo.
- Tests: `TestTemplateMemoryChangesRecordTodosToCheckExistingCode` (changed and added memory → two todos committed with the sync, previous text kept; a second change rewrites the open todo; a pending setup gets none) and the `template use` todo in `TestInitInstallsTheTemplateSetupAndTemplateUseDoesNot`. `go test ./internal/cli/ ./internal/templates/` and `go vet` passed. Docs: README, `docs/design.md`, `docs/status.md`, `.mem/structure.md`.

### Release v0.8.4

CI run 39751641819 green on ubuntu, macOS and Windows. `mem promote staging`, `mem promote production --confirm` (date tag `v2026.10.22.2`), `v0.8.4` on `origin/main` with the confirmed notes; Release run 39774233898 built 7 assets; installed mem updated to v0.8.4.

## Decisions

- Repositories live in `src/db/repositories/<entity>.ts`, as in Benjamin's other projects, take `db: Db` first and never import `db` or open transactions; callers own auth, caching and transactions. Imported as namespaces (`notesRepo`), mirroring ev-ledger's `vouchers_repo`.
- The cross-user guard moves from "unexported cached function" alone to the import rule enforced by ESLint, since a repository function accepts any user id.
- Convention todos are for memories only: skills are reference material and change often. None at `mem init` or during a template setup; one combined todo for `mem template use`. A todo rather than an onboard instruction, so it is shared once through Git and can wait.

## Key Files Affected

- `internal/cli/template_conventions.go` (94 lines): new; convention and template-use todos
- `internal/cli/template_test.go` (61 lines): convention todo tests
- `internal/templates/sync.go` (27 lines): `Result.Memories`, `noteMemory`
- `internal/cli/template.go` (11 lines): `template use` todo and instruction
- `internal/cli/onboard_updates.go`, `onboard.go`, `compact.go`, `sync.go`: publishing the todos and the 📌 instructions
- `README.md`, `docs/status.md`, `docs/design.md`, `.mem/structure.md`
- Outside this repository: `~/Documents/yhat/mem-templates/nextjs-web` (c43a99f) and shannon's `.mem/todos/route_all_database_access_through.md` (465e5e4)

## Notes

- In the scratch app, Turbopack could not resolve `tw-animate-css` from the shadcn `globals.css` although the package was installed under `node_modules`; the build was checked with a one-line `globals.css`. Not investigated further, since the template's CSS was unchanged.
- An earlier count of shannon's out-of-repository queries said 72 files: ugrep printed paths without `./`, so the `grep -v "^./db/"` filter did not exclude `src/db/` itself. The correct count is 46.
- `mem update` again needed retries for a few minutes after the Release workflow finished before it saw the new version.
