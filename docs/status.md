# Status

What is implemented, what is planned, and known limits. [design.md](design.md) describes the intended behavior.

## Implemented

- `init`, which also creates and publishes missing development, staging and production branches, and `import agent-core` (conversion from the Python coding harness).
- `onboard`: Git convergence, managed `AGENTS.md` refresh with a version stamp, project patch hook (no patches yet), hook installation, and context: structure doc, docs, runnables, active spec, open specs and todos with their age, the latest work log with a list of other recent ones, release status.
- `sync`, with drift nudges from `task complete`, `spec complete` and `log new`.
- Specs, tasks, todos, work logs (facts only, closed with `log commit`) and memories.
- `structure` with drift detection against the last commit that touched `.mem/structure.md`.
- Runnables in `.mem/runnables/`.
- `promote staging|production` with date tags, and the `pre-push`/`pre-commit` hooks.
- Templates: `init --template`, `template use|list|promote|reset`, and template sync at onboard with `.mem/templates.lock`.
- Self-update: onboard installs a newer release automatically; `mem update` on demand.
- Releases: v0.2.0 is published with binaries for Linux, macOS and Windows, built by GoReleaser from semver tags.

Tested with disposable repositories and bare remotes; CI runs on Linux, macOS and Windows. mem manages its own source repository. praxis-app has been imported; its commit is pending review.

## Planned

- **Template content.** The library at github.com/Benjamin-van-Heerden/mem-templates starts with few templates; profiles for further kinds of project (for example Rust + GPUI desktop, Go) are added as they are needed.

## Known limits

- Hooks and runnables rely on `sh` and executable bits, and template skills are linked into `.claude/skills` with symlinks; Windows behavior is untested.
- `git push --no-verify` and merges made in a Git host's web UI bypass the hooks. Promotion detects the resulting divergence and prints the recovery steps.
- Work logs imported from the Python harness are interpreted in the importing machine's local time zone.
