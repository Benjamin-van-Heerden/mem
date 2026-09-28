---
title: Logs record facts; log commit closes the session
status: draft
created_at: "2026-09-28T09:00:00+02:00"
updated_at: "2026-09-28T09:00:00+02:00"
---

## Overview

Work logs currently end with a "What Comes Next" section, and onboard asks the agent to summarize "what the recent work logs say comes next". Logs are never edited after they are written, so every pending item they mention stays "pending" forever: in praxis-app, an onboard reported seven pending items from logs, six of which later sessions had already done. Open work needs a primitive with a lifecycle, and mem already has one: todos.

This spec makes logs statements of fact about a finished session, moves anything still to be done into todos, and turns the end of a session into an explicit step, `mem log commit`, which commits the session's mem records, converges the checkout with the shared codebase and pushes. Onboard reads open work from specs, todos and computed release state, and uses logs only as background.

## Goals

- Log template records facts only: goals, what was done, decisions, failed approaches, files affected. No section for future work.
- Open work lives in todos. Ending a session prompts the agent, lightly, to delete todos it completed and create todos for anything left open.
- `mem log commit` finalizes the session: it refuses while the log still has placeholders, commits the log with the other `.mem/` records, syncs with the upstream and pushes, and reports anything that keeps the checkout from matching the shared codebase.
- Onboard summarizes project state from open specs, open todos (with their age) and release status. It shows the user's latest log in full and lists other recent logs by title only.

## Technical Approach

### Log template (`internal/work/log.go`)

Sections: Overarching Goals, What Was Accomplished, Decisions, Key Files Affected, Errors and Barriers (failed approaches and unresolved problems, as facts). Remove "What Comes Next". The template's guidance says the log is a record of the session and is not updated later, and that anything still to be done belongs in a todo.

### `mem log new` (`internal/cli/log.go`)

Instruction steps:

1. Fill in every placeholder in the log.
2. "If this session completed any open todos, delete them now (`mem todo delete <todo>`). Record anything still to be done, including blockers and decisions waiting on the user, as todos (`mem todo new`)." It lists the open todos inline so the agent can check them without another command.
3. Structure doc staleness, as today.
4. Run `mem log commit`.

The unpushed-commit and drift nudges after `log new` stay as they are.

### `mem log commit`

1. Find the log: the argument, or the current user's newest log.
2. Refuse, with the placeholder names, while the log still contains `{…}` template placeholders.
3. Commit the `.mem/` paths that have changes (the log, todos, specs, structure doc) with `git.CommitPaths`, message `Work log: <log title>`. Code outside `.mem/` is never committed by mem.
4. `converge.Sync`: fetch, and fast-forward or rebase onto the upstream when safe (the existing logic, including abort on conflict).
5. Push when the branch has an upstream and is ahead, with `git.PushTimeout`.
6. Report under 🌿 SHARED CODEBASE: what was committed, synced and pushed, and every nudge. Uncommitted changes outside `.mem/` get a ⚠️: "the session ends with uncommitted work in <n> file(s); commit it now or tell the user why it stays uncommitted." Feature-branch drift and being off the development branch are reported as today.
7. The instruction is state-specific: when everything is committed, synced and pushed, "Tell the user the session is closed: the log is committed and <branch> matches <upstream>." Otherwise, one step per remaining ⚠️.

### Onboard (`internal/cli/onboard.go`)

- Recent logs: the current user's latest log in full; other logs from the last 14 days listed as `date  user  title  (mem log show <log>)`, at most 10.
- Open todos show their age (for example `12 days`) and who claimed them.
- The instruction step becomes: "Summarize the project state from open specs, open todos and release status. Work logs are history: use them for background, not as a list of open work."

### Managed instructions (`internal/agentsmd/instructions.md`)

Work Logs section: a log records what happened in a session and is not updated later; open work goes into todos; end a session with `mem log new`, then `mem log commit`.

### Documentation

`docs/design.md` (work records, session continuation), `docs/status.md`, `README.md`, `.mem/structure.md`.

## Success Criteria

- The log template has no future-work section, and `mem log new` prints the open todos and the todo prompt, ending with `mem log commit`.
- `mem log commit` refuses a log with placeholders; on a filled log in a disposable repository with a bare remote, it commits the `.mem/` changes, rebases onto a teammate's pushed commit, pushes, and leaves the branch equal to its upstream. A test covers this.
- `mem log commit` with uncommitted code outside `.mem/` leaves that code uncommitted and reports it with ⚠️. A test covers this.
- Onboard prints only the user's latest log in full, lists other recent logs by title, shows todo ages, and no longer tells the agent to take next steps from logs.
- Managed instructions, design, status, README and the structure doc describe the new behavior; `go vet` is clean for affected packages.

## Notes

- Agreed with the user (2026-09-28): logs are statements of fact; open work and blockers are todos; the todo prompt at session end is a light instruction, not an enforced step; the log stays because its finalization step is where mem keeps everyone on the same codebase.
- Not linking todos to logs in frontmatter: Git history records where a todo came from, and a link would need maintenance without solving staleness.
- Existing logs (for example praxis-app's 18 imported logs) keep their "What Comes Next" sections. After this change onboard shows only the latest log in full, and the agent is told logs are history. The user may want to turn praxis's still-open items into todos once.
- Related but separate: `mem sync` reporting what teammates changed since the last sync (for long sessions that go through compaction), and a post-compaction hook. Not part of this spec.
