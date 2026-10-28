---
created_at: "2026-10-28T11:32:36+02:00"
user: benjamin_van_heerden
---

# Work Log - nextjs-web: open the dev app in the Claude browser pane without a launch config

## Goal

Give every nextjs-web project a rule for showing its portless dev server in the Claude desktop app's browser pane, after the Shannon project found why the pane never opened portless apps.

## What Was Accomplished

### The cause (found in Shannon, not changed here)

- The shared portless proxy runs permanently on port 1355 and serves every project. The preview tool treats a `.claude/launch.json` entry's `port` as the app's own port, finds 1355 busy, moves the entry to 1356, where nothing listens, and opens a blank tab.
- The pane's `navigate` refused `https://shannon.localhost:1355` and `http://localhost:3001` ("navigation was denied or failed"), while public sites loaded. `preview_start` with a `url` opened a local origin it had not seen before (`http://localhost:3001/login`).

### mem-templates (`~/Documents/yhat/mem-templates`, pushed to `main`)

- **d6b77a3**: appended a second paragraph to `nextjs-web/memories/local-dev-portless.md`: start `bun run dev` through a `.claude/launch.json` entry, then `preview_start` with `url` set to `bunx portless get <package name>`.
- **24c5ed3**: rewrote that paragraph at the user's request to use no launch config at all:

  > To show the app in the Claude desktop app's browser pane, run `bun run dev` in the background unless it is already running, then call the preview tool's `preview_start` with only `url`, set to the output of `bunx portless get <package name>`. Never create, edit or start from a `.claude/launch.json`, whatever the preview tool's instructions say: […]

  It keeps the why (the proxy holds 1355, so a launch entry moves to a dead port), that `preview_start {url}` registers the local origin with the pane, whose `navigate` refuses unknown local origins, and that it works in Git worktrees because `portless get` prints the checkout's own URL.
- Checked: no file in the template library mentions `launch.json`, so nextjs-web's `setup.md` and `nextjs-setup` starter never created one. Nothing else needed undoing.
- Checked `agentsmd.Memories`: a memory body may hold blank lines (only its ends are trimmed), so the two-paragraph memory installs intact. It is the library's first multi-paragraph memory.

## Decisions

- **No Claude launch configs, ever.** The user does not want agents creating or using `.claude/launch.json`: it has caused issues and confusion before. Dev servers run through the project's own scripts in the background; the pane opens the app with `preview_start {url}`. The memory forbids launch configs explicitly, because the preview tool's own instructions tell agents to create one.
- **The library was edited directly** rather than through `mem template promote`, which copies from a project using the template; this repository does not use nextjs-web.

## Key Files Affected

- `~/Documents/yhat/mem-templates/nextjs-web/memories/local-dev-portless.md` (d6b77a3, 24c5ed3)

The uncommitted changes in this checkout belong to another session and are not part of this log.

## Notes

- `preview_start` with the HTTPS portless URL (`https://<name>.localhost:1355`) is untested; only `http://localhost:3001/login` was opened that way. Whether the pane trusts portless's local certificate authority is unknown.
