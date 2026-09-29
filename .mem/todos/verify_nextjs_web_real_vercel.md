---
title: Verify nextjs-web on a real Vercel deployment
status: open
created_at: "2026-09-29T21:42:21+02:00"
---

The nextjs-web template (mem-templates) was verified locally only: build, sign-in, cached reads with updateTag, cron starting a Workflow. Unverified on Vercel: bun install/build via vercel.json, the Neon Marketplace integration and its DATABASE_URL/DATABASE_URL_UNPOOLED names, per-preview Neon branches, running scripts/migrate.ts in the Vercel build, crons firing only on production, and Workflow step maxDuration. Deploy the setup skill's starter app to a throwaway Vercel project with staging (test branch) and production, then correct the nextjs-setup, drizzle-neon and background-jobs skills where reality differs.
