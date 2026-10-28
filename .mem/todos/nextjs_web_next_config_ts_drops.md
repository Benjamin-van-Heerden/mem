---
title: nextjs-web next.config.ts drops the Tailwind Turbopack loader
created_at: "2026-10-28T16:31:56+02:00"
---

create-next-app 16.4 compiles Tailwind with @tailwindcss/turbopack through turbopack.rules in its own next.config.ts (no PostCSS config) and sets partialPrefetching: true. Setup step 3 copies the template's next.config.ts over it, so a project set up from nextjs-web today fails next build ('Can't resolve tw-animate-css', '@apply' parse errors in globals.css) and warns about partialPrefetching. Fix: add studii-next's turbopack.rules form (separate rules for *.module.css and other *.css, so CSS module class names stay exported, which site.module.css needs) and partialPrefetching: true to skills/nextjs-setup/files/next.config.ts, and check whether setup step 1's note about next.config.ts still fits. Verify with a fresh scaffold through setup steps 1-3 and next build.
