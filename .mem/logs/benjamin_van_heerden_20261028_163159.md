---
created_at: "2026-10-28T16:31:59+02:00"
user: benjamin_van_heerden
---

# Work Log - nextjs-web: React Compiler on through the Rust port in Turbopack

## Goal

Turn on the React Compiler in the nextjs-web template, so every project built from it has it, without `babel-plugin-react-compiler`.

## What Was Accomplished

### The non-Babel compiler

- Next 16.3.0 added `experimental.turbopackRustReactCompiler`: a native Rust port of the React Compiler that runs inside Turbopack, with no `babel-plugin-react-compiler` to install. It needs `reactCompiler` set, only works under Turbopack (Next throws under webpack), and is marked experimental. Source: `node_modules/next/dist/docs/01-app/03-api-reference/05-config/01-next-config-js/turbopackRustReactCompiler.md` in studii-next (Next 16.4.2).
- `enablePreserveExistingMemoizationGuarantees`: the JS compiler (React Compiler 1.0, bundled in `eslint-plugin-react-hooks` 7.1.1) defaults it to `true`. Next passes it to the Babel path only when set, and the Rust binary's default could not be read, so the template sets it explicitly.

### Template change (mem-templates 4630cd8, pushed to `main`)

`skills/nextjs-setup/files/next.config.ts`:

```ts
reactCompiler: { environment: { enablePreserveExistingMemoizationGuarantees: true } },
experimental: {
  authInterrupts: true,
  turbopackRustReactCompiler: true,
},
```

- `skills/nextjs/SKILL.md`: the Next 16 notes, the config example and the "React Compiler and Memoization" section now describe the compiler as on (Rust port, no Babel plugin, no `--webpack`). The section also says `bun run lint` reports skipped components through `eslint-plugin-react-hooks`, to use react-hook-form's `useWatch` instead of `form.watch()`, and when to use `'use no memo'`. The `reactCompiler`/Babel bullet left "Add other options deliberately", as did `authInterrupts`, which the base example now shows.
- `skills/nextjs-setup/SKILL.md`: the `next.config.ts` bullet lists the compiler, and the "Adding a piece" table has a React Compiler row for existing projects.
- `setup.md` is unchanged: the Rust port needs no package.

### Verification

In a throwaway project in the scratchpad, scaffolded through setup steps 1–3 (Next 16.4.3, React 19.3.0, placeholder env, no database):

- `next build` passed, and its output listed `turbopackRustReactCompiler` under Experiments. `bun run typecheck` passed; `bun run lint` had 0 errors (two warnings in Workflow's generated route files).
- Client chunks held 28 `react.memo_cache_sentinel` occurrences in 5 chunks with the compiler on, and 1 (React's runtime) with it off, so the Rust compiler compiles the starter's client components. `babel-plugin-react-compiler` was not installed.
- The build only passed after adding a Tailwind Turbopack loader to the scratch copy's `next.config.ts` (see Notes).

## Decisions

- **Rust port, not the Babel plugin:** the user's choice. Accepted tradeoff: the option is experimental in Next, though the compiler itself is stable.
- **`enablePreserveExistingMemoizationGuarantees: true`:** matches the JS compiler's default and keeps existing `useMemo`/`useCallback`/`memo` semantics, whichever implementation runs.

## Key Files Affected

- `.mem/todos/nextjs_web_next_config_ts_drops.md`: new todo for the Tailwind loader drift.
- Outside this repository, in mem-templates: `nextjs-web/skills/nextjs-setup/files/next.config.ts`, `nextjs-web/skills/nextjs/SKILL.md`, `nextjs-web/skills/nextjs-setup/SKILL.md`.

## Notes

- **The template's `next.config.ts` no longer builds a fresh project, regardless of the compiler.** create-next-app 16.4 compiles Tailwind with `@tailwindcss/turbopack` through `turbopack.rules` in its own `next.config.ts` (no PostCSS config) and sets `partialPrefetching: true`. Setup step 3 overwrites that file, so `next build` fails with "Can't resolve 'tw-animate-css'" and `@apply` parse errors in `globals.css`. studii-next already has a working form of the rule. Recorded as todo `nextjs_web_next_config_ts_drops`.
- **Existing projects don't get the compiler from a sync:** the skills sync into them, but `next.config.ts` does not. studii-next's skill will say the compiler is on while its config doesn't have it, until the config change is applied there.
