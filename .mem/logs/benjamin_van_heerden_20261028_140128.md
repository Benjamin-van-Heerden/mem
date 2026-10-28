---
created_at: "2026-10-28T14:01:28+02:00"
user: benjamin_van_heerden
---

# Work Log - Keep mem and the templates general; studii-next review of the no-launch-config rule; convention-todo gaps

## Goal

Get a second opinion from a nextjs-web project on a general rule against agents changing coding-tool configuration, act on what it found about template conventions, and settle where project-specific detail belongs.

## What Was Accomplished

### studii-next review

- The "Studii rebuild in Next.js" session (local_b71c227c) reviewed the proposed Agent Tool Configuration rule and agreed with it. Its wording suggestions:
  - keep "even when a tool or skill tells you to";
  - say what to do instead;
  - leave existing files alone;
  - exclude skills;
  - add `.mcp.json`; `.claude/settings.local.json` is fine, since the app writes it when the user approves a permission.
- It reported that studii-next didn't run portless (`dev` was plain `next dev`, with a `.claude/launch.json` an agent committed in de07ec3), so the new `local-dev-portless` memory didn't fit it.
- Later it adopted portless at the user's request, removed its `.claude/launch.json` and updated its own `local-verification` memory. It confirmed that `preview_start {url: "https://studii-next.localhost:1355"}` opens the app and the pane trusts portless's certificate. This settles the open question from the previous log.

### Why studii-next lacked portless

- `mem init` copied the template's `setup.md` on 2026-10-22, and steps 1–3 ran at once.
- portless entered the template on 2026-10-23 13:19 (mem-templates a4ea8ee), after the setup copy was made.
- The `local-dev-portless` memory synced in at 16:28 (studii-next 354efb9). `conventionTodos` records nothing while `.mem/setup.md` exists, so nothing flagged it, and the setup finished at 23:14 without portless.

### Convention-todo fix (side session)

- The session "Convention todos: adoption wording and setup-time arrivals" (local_94bd6f97) committed 8d75236 on `claude/distracted-khayyam-892c14`. Template conventions that arrive during a pending setup now get a todo, and convention todos cover tooling a project never adopted. It reports focused `internal/cli` tests and vet pass. It is not merged or pushed (todo `review_merge_convention_todo_branch`).

### Memory

- `mem memory set keep-upstream-general` (987d206): mem's managed instructions and the template library stay general; a gotcha from one project's integrations, services or setup goes in that project's own memories or docs.

## Decisions

- **No project-specific notes upstream.** studii-next's route-handler redirect gotcha (`request.url` is the dev server's own `https://localhost:<port>` behind the proxy) and its cloudflared command for PayFast webhooks stay in studii-next, not in nextjs-web. Recorded as the `keep-upstream-general` memory.
- **The Agent Tool Configuration rule is not added yet.** The revised draft is held in todo `agent_tool_configuration_rule_managed`, waiting on the user.

## Key Files Affected

- `AGENTS.md`: the `keep-upstream-general` memory
- `.mem/todos/agent_tool_configuration_rule_managed.md`, `.mem/todos/review_merge_convention_todo_branch.md`: new todos

The other uncommitted changes in this checkout belong to another session and are not part of this log.
