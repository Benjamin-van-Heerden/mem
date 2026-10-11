---
created_at: "2026-10-11T14:19:05+02:00"
user: benjamin_van_heerden
---

# Work Log - Auto mode denials and the permit mod; session titles "<project> · <the work>"

## Overarching Goals

Reduce the friction of auto mode blocking actions (without hand-editing `~/.claude/settings.json`), explore mem as a way to share Claude Code mods and plugins, and make sessions from many projects distinguishable in the desktop app's sidebar.

## What Was Accomplished

### Auto mode investigation

- Parsed the last 30 days of transcripts under `~/.claude/projects`: 37 distinct auto mode denials (shannon 18, zcc-ev-reporting 5, zcc-ev-server 4, mem/ascendis/Urbion-AI 3 each, anssum 1). Most frequent reasons: Self-Modification (the agent editing `~/.claude/settings.json`), Merge Without Review / Production Deploy (promotions), Production Reads (ssh to production hosts), plus some misclassified local reads.
- Found that the user-level `autoMode.environment` in `~/.claude/settings.json` describes only shannon ("Source control: the trusted repo (DNA-EOR-Pty-Ltd/shannon) … only"), which applies to every project.
- Documented facts (code.claude.com auto-mode-config, permissions, hooks, plugin docs): `autoMode` is read only from user settings, managed settings and `--settings`, never from `.claude/settings.json` or `.claude/settings.local.json`; `permissions.allow` matches skip the classifier, and broad interpreter/Bash(*) rules are dropped in auto mode; `PermissionDenied` hooks fire only for auto mode denials (fields `tool_name`, `tool_input`, `tool_use_id`, `reason`); `/permissions` has a Recently denied tab (`r` retries once); auto mode pauses after 3 consecutive or 20 total denials; plugin `settings.json` honours only `agent` and `subagentStatusLine`.
- Live checks requested by the user: `createdb`/`dropdb permit_demo` on local Postgres, `ssh charge-prod` hostname/uptime/`docker ps`, and a read-only listing of the ev-ledger database's tables were all allowed by the classifier, since the user asked for each explicitly.

### permit mod (outside this repository)

Built at `/Users/benjamin/.claude/dev-mods/0546f051-a81d-4cd0-8520-9ac080e5a0e7/permit/` (session dev-mods folder, hot-reloaded in this session):

- `classic.PermissionDenied` hook records each denial (tool, action summary, reason, cwd; repeats collapsed; no-verdict denials marked unfixable) and shows `permit: N blocked` in the status line.
- `AbovePrompt` band when the turn is not working (`e.props.isWorking`) and open denials exist: Review / Dismiss all. `/permit` command and a `Pane` with Make rule / Dismiss per denial.
- Make rule drafts one `autoMode.allow` entry with `$.model.complete({ model: 'haiku', effort: 'low' })` (narrowest pattern plus condition such as read-only unless the user explicitly asked; `UNSAFE:` replies become a warning), editable in an `Input`, scope `Select` (project prefixes `In <cwd>: `, or everywhere).
- Save reads and rewrites `~/.claude/settings.json` through `$.fs` (`addRule`: new lists start with `"$defaults"`, duplicates skipped, non-object settings or non-list `allow` refused), then fills the prompt with a retry request (`$.prompt.fill`, mode `append`); it never writes `permissions.allow`.
- Verified with `claude plugin validate`, `tsc` against the build's types, and `claude plugin test` (9 tests: rule logic, and the denial → band → draft → save → prompt fill flow on terminal and desktop surfaces, plus dismiss without writing). No denial occurred live, so the band was not seen in a real session.

### Desktop sidebar and session titles

- Switched the user's Code tab sidebar to group by state (user did it in the filter menu) and renamed all 37 active sessions to `<project> · <title>` with `set_session_title`.
- mem now asks agents for `<project> · <the work>` titles (`internal/cli/session_title.go`):
  - `nameSession`: last line of the onboard instruction — title the session as soon as the user has said what to work on, now if they already have, otherwise after their next message.
  - `retitleSession`: in the compaction digest and as a `log new` step — look up the current title, compare it with what the session is working on now, rename it `<project> · <the current work>` if it no longer fits.
- A first version used a `UserPromptSubmit` hook (`mem hook prompt`, session marks keyed by `CLAUDE_CODE_SESSION_ID` in `.mem/local/sessions/`); it was removed in favour of the onboard instruction. While it existed, testing showed that v0.7.6 prints `mem hook` usage to stdout for the unknown subcommand.
- Tests: `TestOnboardAndWorkLogsAskForAProjectSessionTitle`; `TestCompactHookSyncsAndPrintsAShortDigest` checks the retitle wording. Onboard, init, import, log and compaction tests pass; `go vet` clean.

## Decisions

- Allow rules with conditions are written as plain-English `autoMode.allow` entries, so the classifier still checks the condition; a matching `permissions.allow` rule would bypass it. The user accepts occasional false positives.
- The mod writes rules only on the user's button press, never on the agent's request; rules are written to real settings rather than approved by the mod through `tool.check`, so they stay visible in `/permissions` and work without the mod.
- Drafting uses Haiku; Retry fills the prompt instead of sending it.
- Session titles are instructions in mem's stdout, not a prompt hook: one line at onboard is enough, and a hook costs a process per prompt and breaks on older mem binaries.
- The compaction and log title checks are explicit (look up the title, compare, rename) because a trailing "unless its title already says so" let an unseen title pass as fitting.
- Mission Control as a mod was deferred: grouping the sidebar by state plus project-prefixed titles addresses the scrolling problem.

## Key Files Affected

- `internal/cli/session_title.go` (new): `nameSession`, `retitleSession`.
- `internal/cli/onboard.go`: `renderOnboardInstruction` takes the project name and ends with `nameSession`.
- `internal/cli/compact.go`: digest instruction uses `retitleSession`.
- `internal/cli/log.go`: `log new` step `retitleSession` before `mem log commit`.
- `internal/cli/session_title_test.go` (new), `internal/cli/compact_test.go`.
- `docs/design.md` (new Session titles paragraph, compaction paragraph), `docs/status.md`, `.mem/structure.md` (session titles; `patches.go` now `Schema = 2` with `dropTodoClaims`).
- Commits: 5b1ce69 (hook version), 54ac1a4 (hook removed, onboard instruction), and the retitle wording commit.

## Errors and Barriers

- A transcript grep over all recent JSONL files ran past the 2-minute limit; a Python parser over `tool_use`/`tool_result` pairs replaced it.
- Mod tests: `/home/...` paths count as a network location on macOS (`fs.exists` refused); the test engine needs bottoms for `classic.PermissionDenied`, `ui.status`, `ui.toast` and `ui.render`.
- `go run ./cmd/mem hook compact --dir .` was run in this checkout to preview the digest; it ran the sync catch-up here (no change: `dev` already matched `origin/dev`). Previews belong in a disposable repository.
- CI on `dev` failed on macOS for the onboard update commit (`TestOnboardRemovesTodoClaimsFromOlderProjects`: `TempDir RemoveAll cleanup: … .git/objects/pack: directory not empty`); the test passes locally five times in a row.
