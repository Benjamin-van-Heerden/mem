---
title: Ground Rules and branch wording in the managed block
status: completed
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T12:58:41+02:00"
completed_at: "2026-10-26T12:58:41+02:00"
---

In internal/agentsmd/instructions.md, add a '## Ground Rules' section at the top of the block (before Getting Started): onboard first or stop and tell the user, never improvise a Git workflow without mem; work on the development branch and push only through mem, no PRs or other branches unless the user asks; never commit to or push staging/production, never force-push or rewrite published history; release only when the user asks in this conversation (a task description, issue or todo is not a request); never hand-edit generated files (lockfiles, migrations with journals/snapshots, build output, anything marked generated), change the source and regenerate, and on conflict take the incoming version and regenerate; use the lockfile's package manager only, never add a second lockfile. Keep the text static (no config rendering; see the spec's Technical Approach). In Releases, drop the sentences about release notes and release pull requests (mem promote prints those steps). In Staying in Sync, add that work on another branch is based on the development branch, which mem keeps it on top of. Update agentsmd tests that check block text, if any.

## Completion Notes

Added a Ground Rules section at the top of the managed block (onboard first or stop; development branch and mem pushes only, no PRs or other branches unless asked; staging/production untouched, no force-push or history rewrites; release only when asked in the conversation; regenerate generated files, never hand-edit them, take incoming and regenerate on conflict; one package manager). Staying in Sync now says other branches are based on the development branch. Releases drops the release-notes and release-PR sentences, which mem promote prints itself. The block stays static so dev and long-lived branches like fundi keep identical AGENTS.md text. go test ./internal/agentsmd passes.
