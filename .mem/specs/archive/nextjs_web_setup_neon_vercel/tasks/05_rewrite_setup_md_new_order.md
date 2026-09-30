---
title: Rewrite setup.md in the new order
status: completed
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:43:41+02:00"
completed_at: "2026-09-30T13:43:41+02:00"
---

nextjs-web/setup.md: scaffold+stack; Neon (neon me / neon auth (you), projects create, dev branch, connection strings into .env.local, first migrate+seed); auth barriers and route groups; branding (you): logo into public/, design questions (corners, fonts, accent, light/dark, dot grid vs hatch) recorded in tokens and .mem/docs/design.md; basic home and login; first production deploy (vercel whoami / login (you), GitHub remote, project add, link, git connect, production env vars from Neon main, deploymentEnabled for mem's dev/staging branches, first deploy); migration and seed scripts; RBAC and super admin; dashboard and navigation with the chosen background; verify and redeploy. Code snippets where they remove guesswork; Done-when per step.

## Completion Notes

mem-templates: nextjs-web/setup.md rewritten as 12 steps in the agreed order (scaffold; stack and shadcn; copy starter files with the background utilities and deploymentEnabled; Neon with the CLI, dev branch and a bun snippet that writes .env.local without printing secrets, then migrate/seed/typecheck; auth barriers and route groups; branding (you) with the design questions and .mem/docs/design.md; home and login; production deploy with vercel project add/link/git connect, env vars from Neon main via vercel env add on stdin, mem promote staging/production; migrations and seed; roles and user management; dashboard and navigation with the background choice (you); verify and redeploy), each with code and a Done-when line. The env snippet was tested (uses a replacer so $ in URLs is safe). Library README: (you) steps are asked for, not pushed last. mem: onboard's setup instruction now says steps marked (you) need the user and the agent asks and continues (was 'stop and hand them over'); design.md updated; onboard setup test passes. Real-run verification in task 7.
