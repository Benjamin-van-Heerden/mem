---
title: "Release flow: light staging, documented production, mem deploy, optional PRs"
status: draft
created_at: "2026-10-01T09:00:00+02:00"
updated_at: "2026-10-01T09:00:00+02:00"
---

## Overview

Rework releases around how they are actually used:

- **Staging is light.** `mem promote staging` fast-forwards staging with no notes and no second step, as it does now.
- **Production is careful and documented when an agent does it.** `mem promote production` drafts release notes itself from the completed specs, the work logs and the commits since the last release, into `.mem/local/release-notes.md`. The agent refines the draft, shows the user, and runs `mem promote production --confirm`, which releases with those notes. This replaces the two passes with a notes file "outside the repository".
- **Some repositories require a PR into production.** With `[release] production_pr = true`, `--confirm` opens a pull request from a snapshot branch instead of pushing, with the notes as its description. Once the PR is reviewed, `mem promote production` completes it by fast-forwarding production to the PR's head, which keeps history linear; GitHub then marks the PR as merged. GitHub's merge button is never used.
- **`mem deploy` is the user's command.** It moves dev → staging → production in one go, with no notes and no questions, and tags production with a generated summary. A human can release without an agent. AGENTS.md states that agents never run it unless the user explicitly asks for it.

The three-branch model (development, staging, production), fast-forward only, stays as it is.

## Goals

- The user can deploy everything with one command and no agent.
- Agents release production with reviewed notes in one clear two-step flow: draft, then confirm.
- Repositories can require a PR into production without breaking the linear history.
- No reliance on `gh` being logged in, except where a repository opts into PRs.

## Technical Approach

### Release notes draft (`internal/release/notes.go`, new)

- `DraftNotes(p, plan)` builds Markdown for the release:
  - `# Release <tag>`
  - "Specs completed": title and overview line of each spec archived as completed in the range (existing `completedSpecs`, extended with titles).
  - "Work": the titles of work logs committed in the range, and their "What Was Accomplished" headings.
  - "Commits": subject and author, capped at 30.
- `.mem/local/release-notes.md` is the draft's fixed location (already ignored). `mem promote production` writes it, or keeps an existing draft whose `<!-- mem:release <to-sha> -->` marker matches the plan, so an agent's edits survive a re-run. A draft for another commit is replaced, with a line saying so.
- `GeneratedMessage(plan)` is the short automatic tag message used by `mem deploy`: completed spec titles and commit subjects.

### `mem promote production` (`internal/cli/promote.go`)

- Without `--confirm`: print the plan, write or keep the draft, and instruct:
  1. Refine `.mem/local/release-notes.md` into a short summary of what the release delivers for its users; remove the "Commits" section unless it is useful.
  2. Show the user the notes and ask them to confirm the release.
  3. After they confirm, run `mem promote production --confirm`.
- With `--confirm`: read the draft (refusing when it is missing, empty or marked for another commit), then either push and tag as now, or open the PR (below). Delete the draft after a successful release.
- `--notes <file>` is removed; `--confirm` replaces it.
- Staging is unchanged, except that its closing instruction no longer promises a deployed preview (see Notes).

### Required PRs (`[release] production_pr = true`)

- `project.ReleaseConfig{ ProductionPR bool }` under `[release]` in `.mem/config.toml`; `mem init --production-pr` sets it.
- GitHub access (`internal/github`, new, small): owner/repo from the remote URL (https and ssh forms); token from `GITHUB_TOKEN`, then `GH_TOKEN`, then `gh auth token` if `gh` is on PATH; a missing token is an error that names these options and is only raised in PR-required repositories. REST calls: create pull request, list open pull requests by head prefix, get pull request (state, head sha, review decision via reviews), delete branch ref. 20-second timeouts.
- `--confirm` in such a repo: push `promotion/production/<UTC timestamp>` at the plan's commit (allowed by the pre-push hook through `MEM_PROMOTE`), open a PR into production titled `Release <tag>` with the notes as body, print its URL, delete the draft. Refuse when a promotion PR is already open.
- Completing: `mem promote production` in a repository with an open promotion PR reports it (URL, approvals, changes requested) and, with `--confirm`, completes it: production must still be an ancestor of the PR head and the head must equal the snapshot branch; push production to the head, tag with the PR body as notes, delete the snapshot branch. When the PR has "changes requested", refuse. The instruction tells agents to complete a PR only when the user says it is approved.
- Verify with a real throwaway GitHub repository that fast-forwarding the base to the PR head marks the PR merged; if GitHub does not, close the PR with a comment linking the release tag and say so.

### `mem deploy` (`internal/cli/deploy.go`, new)

- Promotes staging to the development branch's upstream head, then production to staging, each as a direct fast-forward, and tags production with `GeneratedMessage`. It prints the same plans as promote and one summary. It does nothing more when a hop is already current.
- In a repository with `production_pr = true` it promotes staging and then refuses the production hop, pointing at `mem promote production` (open question below).
- It works for a human at a terminal: no agent instruction is required to finish, and its stdout reads as a report.

### Instructions and docs

- `internal/agentsmd/instructions.md`, Releases:
  - Preview release: `mem promote staging`.
  - Release: `mem promote production`, then follow its instructions: refine the notes, show the user, and confirm.
  - A full deployment, when an agent is asked for it, is `mem promote staging` then `mem promote production`, with the notes step.
  - "`mem deploy` releases everything without review. It is the user's command: never run it unless the user explicitly asks you to run `mem deploy`."
  - In repositories that require PRs, a production release opens a PR; complete it only when the user says it is approved.
- `docs/design.md` (Releases), `docs/status.md`, `README.md` (Releases section, `mem deploy`, `production_pr`), `.mem/structure.md`.

## Success Criteria

- `mem deploy` in a disposable repository with a bare remote moves staging and production to the development head in one run, tags production with a generated message, and needs no further command; in a `production_pr` repository it moves staging only and explains why.
- `mem promote production` writes `.mem/local/release-notes.md` with the completed specs, work log titles and commits of the range; a re-run keeps an edited draft for the same commit and replaces one for another commit; `--confirm` tags with the draft's content and deletes it; `--confirm` without a valid draft refuses.
- `mem promote staging` is unchanged apart from its closing wording.
- In a `production_pr` repository (tests against an `httptest` GitHub API and a bare remote): `--confirm` pushes a snapshot branch and creates a PR with the notes; a second `--confirm` while it is open completes it by fast-forward, tags, and deletes the snapshot branch; "changes requested" refuses; a missing token errors only there.
- One real run against a throwaway GitHub repository confirms the PR shows as merged after the fast-forward (or the fallback is implemented); the repository is deleted afterwards.
- AGENTS.md instructions, docs and README describe the flow; focused tests and `go vet` pass.

## Notes

- User decisions (2026-10-01): keep staging (for eventual preview deployments; one code path); `mem deploy` runs the whole sequence and agents need explicit consent to run it; staging promotions are light, production promotions careful and documented when agents do them; some repositories must require PRs into production; do not rely on `gh` being logged in; PRs must not create merge commits.
- Open question for the user: in a `production_pr` repository, should `mem deploy` refuse the production hop (current plan) or open the PR itself?
- Todo `promote_staging_wording_when_staging` is covered by the staging wording change: "Staging is at <sha>. Check the preview if this project deploys staging; release with `mem promote production`." Delete the todo when done.
- Every GitHub merge method (merge commit, squash, rebase) rewrites or adds commits, which is why mem completes promotion PRs itself.
