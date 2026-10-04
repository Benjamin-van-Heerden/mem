# Status

What is implemented, what is planned, and known limits. [design.md](design.md) describes the intended behavior.

## Implemented

- `init`, which also creates and publishes missing development, staging and production branches (starting a repository without commits with an empty commit), and `import agent-core` (conversion from the Python coding harness).
- `onboard`: Git convergence, managed `AGENTS.md` refresh with a version stamp, the `.mem/local/` ignore entry, project patch hook (no patches yet), hook installation, memories and skills changed by the sync, and context: structure doc, docs, runnables, active spec, open specs and todos with their age, the latest work log with a list of other recent ones, release status.
- `sync`: Git convergence and a push when the branch is ahead and not behind, template sync, and a report of incoming commits, work record changes, changed memories and skills, and a newer mem release. `task complete` and `spec complete` commit their record and run the same catch-up and push; `log new` prints drift nudges.
- Claude Code compaction hook: `init`, `import` and onboard maintain a `SessionStart`/`compact` entry in `.claude/settings.json` that runs `mem hook compact` (the `mem sync` catch-up with a short digest); `[claude] compact_hook = false` removes it.
- Specs, tasks, todos, work logs (facts only, closed with `log commit`) and memories.
- `structure` with drift detection against the last commit that touched `.mem/structure.md`.
- Runnables in `.mem/runnables/`.
- `promote staging|production` with date tags, drafted release notes confirmed with `--confirm`, optional production pull requests completed by fast-forward (`[release] production_pr`), `deploy` for one-step releases, and the `pre-push`/`pre-commit` hooks.
- Templates: `init --template`, `template use|list|promote|reset`, and template sync at onboard with `.mem/templates.lock`; one-time template setups installed by `init` to `.mem/setup.md` and led by onboard and the compaction digest until the file is deleted.
- Self-update: onboard installs a newer release automatically; `mem update` on demand.
- Releases: v0.2.0 is published with binaries for Linux, macOS and Windows, built by GoReleaser from semver tags.

Tested with disposable repositories and bare remotes; CI runs on Linux, macOS and Windows. mem manages its own source repository. praxis-app has been imported; its commit is pending review.

## Planned

- **Template content.** The library at github.com/Benjamin-van-Heerden/mem-templates starts with few templates; profiles for further kinds of project (for example Rust + GPUI desktop, Go) are added as they are needed.

## Known limits

- Hooks and runnables rely on `sh` and executable bits, the compaction hook's command assumes bash (Claude Code falls back to PowerShell on Windows without Git Bash), and template skills are linked into `.claude/skills` with symlinks; Windows behavior is untested.
- `git push --no-verify` and merges made in a Git host's web UI bypass the hooks. Promotion detects the resulting divergence and prints the recovery steps.
- Work logs imported from the Python harness are interpreted in the importing machine's local time zone.
