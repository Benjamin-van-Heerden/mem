---
created_at: "2026-10-28T16:50:37+02:00"
user: benjamin_van_heerden
---

# Work Log - nextjs-web: Tailwind Turbopack loader and Partial Prefetching in next.config.ts

## Goal

Fix the template's `next.config.ts` so a project set up from nextjs-web with today's create-next-app builds again (todo `nextjs_web_next_config_ts_drops`).

## What Was Accomplished

### Template change (mem-templates d74d624, pushed to `main`)

`skills/nextjs-setup/files/next.config.ts` now carries what create-next-app 16.4 puts in its own config, which setup step 3 overwrites. It uses studii-next's form of the loader rules:

```ts
partialPrefetching: true,
turbopack: {
  rules: {
    "*.css": [
      { condition: { path: /\.module\.css$/ }, loaders: ["@tailwindcss/turbopack"], as: "*.module.css" },
      { condition: { not: { path: /\.module\.css$/ } }, loaders: ["@tailwindcss/turbopack"], as: "*.css" },
    ],
  },
},
```

- `skills/nextjs-setup/SKILL.md`: the `next.config.ts` bullet names Partial Prefetching and the Tailwind loader. The Caching row of the "Adding a piece" table adds `partialPrefetching: true`, and the paragraph under the table points to `adopting-partial-prefetching.md` for apps whose links use `prefetch={true}`. A new Known sharp edges bullet says there is no PostCSS config, both rules must stay, and an app scaffolded with `postcss.config.mjs` keeps it and leaves out the `turbopack` block.
- `skills/nextjs/SKILL.md`: a Next 16 note on what `partialPrefetching` changes for `<Link>`, and the option added to the config example.
- The starter has no `<Link prefetch={true}>`, so Partial Prefetching changes nothing it relies on.

### Verification

A fresh project scaffolded in the scratchpad through setup steps 1–3, with the template's `next.config.ts` unmodified (`cmp`), on Next 16.4.3 with placeholder env and no database:

- `next build` passed, with "Partial Prefetching enabled" and `turbopackRustReactCompiler` listed, and no `partialPrefetching` warning.
- The built CSS has the `site.module.css` classes (`dO3plW_hero` and the rest, also in the home page HTML), Tailwind utilities, the template's `bg-dot-grid` utility, and no `@apply`.
- The React Compiler output was unchanged (28 memo-cache sentinels in 5 client chunks). `bun run typecheck` passed, and `bun run lint` had 0 errors (the same two warnings in Workflow's generated files).

## Key Files Affected

- `.mem/todos/nextjs_web_next_config_ts_drops.md`: deleted, done.
- Outside this repository, in mem-templates: `nextjs-web/skills/nextjs-setup/files/next.config.ts`, `nextjs-web/skills/nextjs-setup/SKILL.md`, `nextjs-web/skills/nextjs/SKILL.md`.
