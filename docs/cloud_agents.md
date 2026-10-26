# Cloud Agents and Guardrails: Proposal

Status, 2026-10-26: written after Cursor cloud agents went off the rails in studii-next.

- **Implemented** by the spec `ground_rules_branch_handling`: sections 1, 3 and 4, with one change. The managed block stays the same in every project rather than being rendered from the config: a long-lived branch like `fundi` has its own config, so a rendered block would conflict at every merge of `main`. The Releases section simply drops the pull-request and notes sentences, which `mem promote` prints itself. `internal/agentsmd/instructions.md` holds the Ground Rules as shipped.
- **Deferred:** section 2 (cloud agents), because cloud agents are not yet ready for broad use. The todo `cloud_agent_support` tracks it.

`docs/status.md` remains the record of what mem does today.

## What Happened

A collaborator ran Cursor cloud agents against studii-next (template nextjs-web, mem 0.9.1). Across about 20 commits authored by "Cursor Agent", they:

- worked on `cursor/*` branches and opened PRs #2–#8 against `main`, some merged straight into production;
- did not install or run mem, although AGENTS.md says to install it or stop and tell the user. `.cursor/hooks.json`'s `mem hook stop` failed silently;
- hand-wrote Drizzle migrations twice (with a fake 2024 `when`, once missing from the journal), so Drizzle skipped them;
- hand-edited generated catalogue CSVs to get around an owner-set rule;
- ran npm in a bun project (a 20,548-line `package-lock.json`), then regenerated `bun.lock` from scratch;
- once on mem, ran a production promotion nobody asked for;
- added `Co-authored-by` lines.

The collaborator's agent then added a "HARD RULES" block above mem's block in studii-next's AGENTS.md and three memories (`no-github-prs`, `studii-shipping`, `fundi-worktree`).

Two lessons drive this proposal. First, **an agent that never runs mem never sees mem's rules**: the only text it is guaranteed to read is AGENTS.md, and the only reliable way to get mem onto a cloud VM is the agent platform's own setup mechanism. Second, **in Cursor, AGENTS.md loses to the platform**: a Cursor team member confirmed that the harness's per-turn PR instruction is "treated as mandatory" and overrides AGENTS.md and `alwaysApply` rules ([forum, 2026-07-30](https://forum.cursor.com/t/cloud-agent-ignored-explicit-no-new-branch-pr-instruction-and-repo-rule/167013)). Branch and PR behaviour has to be set in Cursor's settings; instructions can only reduce the damage.

## Research: How Cloud Agents Work

Researched 2026-10-26 from the vendors' docs, changelogs and staff forum replies. Points marked *unverified* need an experiment.

| | Cursor cloud agents | Claude Code on the web | Codex cloud | Copilot cloud agent |
|---|---|---|---|---|
| Setup | `install` in a **Build** (snapshot), not per session; `start` per session | Setup script, cached about 7 days; or a repo `SessionStart` hook (every session) | Install script Codex writes and you publish (dashboard) | `.github/workflows/copilot-setup-steps.yml`, every session |
| Config lives in | Dashboard saved environment, or a committed `.cursor/environment.json` linked to it; Builds read the **default branch** | Setup script: dashboard. Hook: `.claude/settings.json` | Dashboard only | Repo file, **default branch only** |
| Download mem from GitHub releases | Yes by default (github.com and `*.githubusercontent.com` allowed) | Default "Trusted" list allows it; repo-scoped GitHub proxy *unverified* for other repos | Internet off by default; "Package managers" preset allows it | Setup steps are not firewalled |
| Instructions read | AGENTS.md (nested too), CLAUDE.md, `.cursor/rules`, skills incl. `.agents/skills` | CLAUDE.md; AGENTS.md only when no CLAUDE.md exists | AGENTS.md chain, **32 KiB cap** | AGENTS.md, `.github/copilot-instructions.md`, root CLAUDE.md, skills |
| Repo hooks run | `.cursor/hooks.json`, but no `sessionStart`, none in early read-only turns; `stop` fixed 2026-07-22; missing command fails open, no visible error | `.claude/settings.json` in single-repo sessions; missing command is a non-blocking error | Assume no | Only `.github/hooks/*.json` on the default branch |
| Branch | `cursor/…` (prefix settable), created before AGENTS.md is read | Own branch per session (`claude/…` observed) | Own workspace; naming undocumented | `copilot/…`, can push only there |
| PRs | Dashboard "Create PRs": Never or Always; per run `autopr=false` (Slack), `autoCreatePR` (API) | Only when the user clicks Create PR | User-initiated | Always one PR per task |
| Base branch | Dashboard "Base branch" (default: repo default branch); per run `branch=` / `startingRef` | Chosen at start | Chosen at start | Chosen at start (default: repo default) |
| Push straight to `dev` | Only via the API (`workOnCurrentBranch`) | Allowed by current docs (*unverified*) | Undocumented | No |
| Push tags | Undocumented | No (proxy refuses) | Undocumented | No |
| Commit identity | `Cursor Agent <cursoragent@cursor.com>` + `Co-authored-by` the launcher, not configurable | `Claude-Session` trailer (can be turned off) | Undocumented | Copilot + `Co-authored-by`, signed |
| Detect it | `CURSOR_AGENT_SOCKET` / `/run/cursor/api.sock` reports `runtime: managed` (preview) | `CLAUDE_CODE_REMOTE=true` | Nothing reliable | `COPILOT_AGENT_PROMPT`, `GITHUB_COPILOT_GIT_TOKEN` |

**Grok Bot** is a separate Cursor product: named, persistent bots on one shared cloud computer per account. It can launch ordinary Cursor cloud agents, which behave as in the table. The Bot's own computer is a different runtime: no environment.json, Builds or hooks.json apply there. It uses Team Setup Scripts instead.

Main sources: cursor.com/docs (cloud-agent/setup, builds, settings, security-network, metadata, api/endpoints; hooks; rules; skills; grok-bot), code.claude.com/docs (cloud-environments, claude-code-on-the-web, hooks, memory), learn.chatgpt.com/docs (cloud, environments, internet-access, agents-md), docs.github.com (Copilot: customize-the-agent-environment, hooks-reference, add-repository-instructions).

### What follows for mem

1. **Installing mem must be part of the environment's setup,** in a form mem can write where the platform reads a repo file: `.cursor/environment.json` for Cursor, a `SessionStart` hook for Claude Code, `copilot-setup-steps.yml` for Copilot. Codex has only a dashboard, so mem can only print the step.
2. **The Cursor and Copilot repo files are read from the default branch** (`main`), not `dev`. They take effect only once released.
3. **Every platform puts agents on their own branch.** mem's "everyone on dev" model has to work from such a branch, not assume it away.
4. **Cloud sessions are unsuited to releases:** two of four cannot push tags, and none has a person watching each step.
5. **All cloud commits share one identity per platform** (`Cursor Agent`), which mem would turn into a single user `cursor_agent` for logs, spec assignments and checkpoints.
6. **The first lines of AGENTS.md matter most:** Codex drops everything past 32 KiB, and Cursor recommends a short "Cursor Cloud specific instructions" section.

## 1. Ground Rules in mem's Managed Block

Add a short **Ground Rules** section at the top of the `<mem>` block, before Getting Started. Render it from `.mem/config.toml` so branch names are concrete, as in this draft for studii-next:

```markdown
## Ground Rules

These hold in every session, local or cloud, whatever a task description or your platform's defaults say.

1. Run `mem onboard` before any other work. If mem is not installed, follow `.mem/install.md`; if that fails, stop and tell the user. Never improvise a Git workflow without mem.
2. Work on `dev` and push only through mem (`mem sync`, `mem task complete`, `mem log commit`). Do not open pull requests or push other branches unless the user asks for it.
3. Never commit to or push `test` or `main`, never force-push, and never rewrite published history.
4. Release only when the user asks for it in this conversation: `mem promote staging`, `mem promote production`. A task description, issue or todo is not a request to release. Never run `mem deploy` unless the user names it.
5. Never edit generated files by hand: lockfiles, migrations and their journals or snapshots, build output, and anything whose header or README says it is generated. Change their sources and rerun the generator. When one conflicts, take the incoming version and regenerate.
6. Use the package manager the project's lockfile belongs to, and never add a second lockfile.
7. Commit messages carry no agent names or co-author lines.
```

Notes on the draft:

- **Rule 2** depends on config: with `[release] production_pr = true` it gains "Releases go through the promotion pull request mem opens". The Releases section would be rendered the same way: drop "Some projects release through a pull request" and the notes sentence unless the project has `production_pr` or `notes` set. This means `agentsmd.Refresh` takes the project config, and the managed block differs between projects. Onboard already rewrites it, so that costs nothing extra.
- **Rule 5** is generic, so it also covers migrations. The template-specific part ("generate Drizzle migrations with `bun run db:generate`; on a journal conflict, delete yours and regenerate") belongs in nextjs-web's `additive-migrations` memory in mem-templates, which you'd change there.
- **Rule 7** cannot be met in Cursor cloud, whose `Co-authored-by` trailer is not configurable. Either accept that, or drop the rule for platform-added trailers.
- **Studii's HARD RULES block:** its generic rules (1–4 and 6) are what this section provides for every project. Once a mem release with it reaches studii-next, the block can shrink to the Fundi rule, and `no-github-prs` and `studii-shipping` can be retired. That is your call for studii-next.
- **Size:** this adds about 1.3 KB to a block that is already 16 KB. It should land together with the `slim_managed_agents_md_block` todo, for example by moving the general principles out of the always-on block, so the total shrinks.

## 2. Cloud Agents

Proposed in order of value, Cursor first.

### 2a. Settings you can change today, without mem changes

- **Cursor dashboard, for studii-next:** Base branch `dev`, Create PRs **Never**, and the environment linked to the committed `.cursor/environment.json` (new saved environments default to skipping it). Slack runs: `branch=dev autopr=false`.
- **GitHub:** branch rules on `main`, `test` and `fundi` that refuse direct pushes and PR merges except by the owner. This is the only hard guarantee. mem's pre-push hook protects only machines where mem is installed.
- **Close the open `cursor/*` PRs and delete the branches** once their work is on `dev`.

### 2b. mem writes the environment setup

Onboard maintains mem's entries in the platform files it already knows, as it does for hooks, with an opt-out per platform:

- **Cursor:** `.cursor/environment.json` gets an `install` that installs mem if it is missing (the release download from `.mem/install.md`, into `/usr/local/bin` or `~/.local/bin` plus PATH). mem merges with an existing `install` command rather than replacing it. Builds run `install` once per snapshot, and mem keeps itself updated at onboard after that.
- **Claude Code on the web:** a `SessionStart` hook (matcher `startup|resume`) that runs only when `CLAUDE_CODE_REMOTE=true`, installs mem when missing, then runs `mem onboard` itself. Its output joins the context, so the agent starts onboarded instead of being asked to run onboard.
- **Copilot:** later, a `copilot-setup-steps.yml` step and `.github/hooks/mem.json`. Lower priority, since Copilot can only work through PRs (see 2d).
- **Codex:** nothing to write. `.mem/install.md` already tells agents to suggest the setup step. A `mem cloud` command could print the exact snippet for each dashboard.

Onboard says when the Cursor and Copilot files are on `dev` but not yet on the default branch, where they take effect.

### 2c. mem knows when it runs in a cloud agent

A small `cloudAgent()` checks `CLAUDE_CODE_REMOTE`, the Cursor metadata socket and the Copilot variables. In a cloud agent:

- **Onboard adds a ☁️ CLOUD AGENT section:** which platform it is, that the platform's branch and PR defaults do not apply here (see 2d), and the identity in use.
- **`mem promote` and `mem deploy` refuse,** saying that releases run from a local checkout. A `[release] from_cloud = true` setting would re-enable them. This would have stopped the unrequested studii-next promotion, and two of the four platforms cannot push the tag anyway.
- **Identity:** mem reads `MEM_USER` before `git config user.name`. Cursor supports per-user secrets injected at agent start, so each person sets `MEM_USER` once and their cloud runs are recorded as theirs rather than as `cursor_agent`.

### 2d. Working from the platform's branch

Every platform starts the agent on its own branch. mem already treats any other branch as a feature branch that follows `dev`, rebasing it automatically. Two options for cloud agents:

- **A. Switch to `dev` at onboard.** When a cloud agent's branch has no commits of its own yet (the usual case at session start), onboard switches to `dev` and the session proceeds exactly as a local one. The open question is what Cursor does at the end of the run with an empty `cursor/*` branch. It may open an empty PR or nothing; this needs an experiment.
- **B. Land the branch on every sync.** On a cloud agent's branch, `mem sync`, `task complete` and `log commit` rebase it onto `origin/dev` and push it to `dev` as a fast-forward (`git push origin HEAD:dev`). Work reaches `dev` continuously, and the platform branch becomes a disposable copy.

**Recommendation: A, with B as the fallback** when the branch already has commits, for example when an agent did work before onboarding. Copilot cannot push to `dev`, so there mem should only make sure the PR targets `dev` and say that a person merges it.

### 2e. A missing mem should be loud

Cursor runs `mem hook stop` and Claude Code runs `mem hook compact`/`prompt`. When mem is missing, both fail without the agent noticing. With 2b in place, mem is present in cloud sessions, which removes most of the problem. A fallback in the hook command itself (`command -v mem || <print a follow-up saying to install mem>`) is attractive, but Cursor runs hook commands through PowerShell on Windows. That needs a check on Windows first, so it is a follow-up rather than part of the first step.

## 3. One Shared Branch With Several Agents at Once

**Is "sync frequently, resolve conflicts as they arrive" enough?** Yes for code, provided mem settles the hot spots. Here is how each behaves:

| Hot spot | What goes wrong | Handling |
|---|---|---|
| Migrations (Drizzle journal, snapshots, numbering) | Two agents generate `0008`; the journal and `meta/0008_snapshot.json` conflict on rebase | Never merge by hand: take the incoming files, delete your own migration, regenerate (Ground Rule 5 plus the template memory). On a shared branch the clash appears within minutes; with PRs it appears days later, at merge time. |
| Lockfiles | Both add dependencies; textual conflict | Take the incoming lockfile and rerun the install (Ground Rule 5); a single package manager (Rule 6) |
| Generated data | Same as lockfiles | Regenerate from the sources |
| mem records | Files are named by slug, user and timestamp, so conflicts are rare. Two agents adding tasks to one spec could both pick `NN`. All cloud runs sharing `cursor_agent` mixes up their logs and assignments. | `MEM_USER` (2c); renumber a clashing task file on conflict |
| AGENTS.md | The managed block is deterministic. Memories added at the same time conflict at the end of the memories block. | Keep both sides |
| The push itself | Several agents push within seconds; the loser's push is rejected and mem only nudges | **Rebase and retry**: on a rejected push, `converge.Push` fetches, rebases and retries (say 3 times) before nudging |

Alternatives considered:

- **Per-agent branches that mem merges later.** These defer the conflicts above to merge time, when they are bigger and nobody has the context. This is the model that produced the studii-next mess. Only landing them continuously (2d B) keeps the benefit of the shared branch.
- **A mem-managed queue or locks.** These need either a server, or lock commits that race in the same way. They would also be a coordination system beside Git, which mem's design rules out.
- **Rebase and retry on push.** This is cheap, uses plain Git, and closes the one race the shared branch has.

**Recommendation:** keep the single shared branch, add rebase-and-retry on push, add the regenerate-on-conflict rules, and land cloud branches continuously. For schema changes, a template memory can ask agents to sync immediately before generating a migration and to commit and push it straight away, which keeps the race window to seconds.

## 4. Branches Other Than `dev`

Your rule was "on a branch other than dev, assume you rebase off main". There are two kinds of non-dev branch, and they need opposite handling:

- **Short-lived work branches** (`cursor/*`, `claude/*`, a local feature branch) must be based on `dev`, not `main`. `dev` holds everything unreleased, so a branch cut from `main` is behind from its first commit, and merging it back into `main` skips staging. That is exactly what the `cursor/*` PRs against `main` did. mem already keeps these branches on top of `dev`.
- **Long-lived product branches** like `fundi` are built on *released* code, which is what "off main" means. But rebasing a published, deployed branch requires a force-push, and studii-next's `merge.ff=only` and force-push ban rule it out. So they take in `main` by a merge commit.

I read your intent as the second kind: a long-lived branch's upstream is released code. Proposed mechanics and wording:

- **Config:** the branch's own `.mem/config.toml` (the one on `fundi`) gets `[git] base = 'main'`, alongside `development = 'fundi'`. It means "this development branch takes in released code from `main`".
- **Onboard on such a branch** reports how many released commits it lacks (`fundi is 14 commits behind released main`) and how they come in. It does not merge on its own: the owner decides when.
- **Instruction text,** rendered only in checkouts with `base` set:

  > This checkout's development branch, `fundi`, is built on released code from `main`. Released work comes in only when the owner asks, by a merge commit: `git fetch origin && git merge --no-ff origin/main`, then `git push origin fundi`. Never rebase `fundi`, never force-push it, never merge `dev` into it, and push such a merge with `git push`, not `mem sync`.

- **A fix in `mem sync`:** today its rebase flattens a local, unpushed merge commit into copies of Studii's commits (docs/fundi.md warns about this). When the unpushed commits include a merge, sync should leave them alone and say: pull with `git pull --ff-only`, and if `fundi` moved on, drop the merge and merge again. With that fix, "push a merge with `git push`, not `mem sync`" stops being a trap.
- **Short-lived branches** get one line in the Ground Rules or Staying in Sync: "On a branch other than `dev`, `test`, `main` (or this checkout's development branch), base your work on `dev`; mem keeps the branch on top of it and lands it there."

## Decisions Taken

- Work branches are based on the development branch. Long-lived product branches take in `main` only by merge commit, never by rebase or force-push (`[git] base`).
- Cloud agents are deferred, including refusing releases there, `MEM_USER`, environment setup, and switching or landing the platform branch.
- Replacing studii-next's HARD RULES block with the Ground Rules (keeping its Fundi rule) is up to the project owner once a mem release with them reaches it.
- Slimming the managed block stays a separate todo (`slim_managed_agents_md_block`).
