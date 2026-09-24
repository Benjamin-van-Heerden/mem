# Status

What is implemented, what is planned, and known limits. [design.md](design.md) describes the intended behavior.

## Implemented

- `init` and `import agent-core` (conversion from the Python coding harness).
- `onboard`: Git convergence, managed `AGENTS.md` refresh with a version stamp, project patch hook (no patches yet), hook installation, and context: structure doc, docs, runnables, active spec, open specs and todos, recent logs, release status.
- `sync`, with drift nudges from `task complete`, `spec complete` and `log new`.
- Specs, tasks, todos, work logs and memories.
- `structure` with drift detection against the last commit that touched `.memr/structure.md`.
- Runnables in `.memr/runnables/`.
- `promote staging|production` with date tags, and the `pre-push`/`pre-commit` hooks.

Tested with disposable repositories and bare remotes on macOS. praxis-app has been imported; its commit is pending review.

## Planned

- **Releases and self-update.** Published binaries (GoReleaser is configured) and a check at onboard for a newer memr.
- **Templates.** Project profiles (for example Next.js web app, Rust + GPUI desktop app, general Python, Go and Rust) kept in a GitHub template library, supplying initial memories, docs and skills, with a way to promote a project's memory or doc back into its template and receive template updates at onboard.

## Known limits

- Hooks and runnables rely on `sh` and executable bits; Windows behavior is untested.
- `git push --no-verify` and merges made in a Git host's web UI bypass the hooks. Promotion detects the resulting divergence and prints the recovery steps.
- Work logs imported from the Python harness are interpreted in the importing machine's local time zone.
