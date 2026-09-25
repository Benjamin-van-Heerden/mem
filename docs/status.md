# Status

What is implemented, what is planned, and known limits. [design.md](design.md) describes the intended behavior.

## Implemented

- `init`, which also creates and publishes missing development, staging and production branches, and `import agent-core` (conversion from the Python coding harness).
- `onboard`: Git convergence, managed `AGENTS.md` refresh with a version stamp, project patch hook (no patches yet), hook installation, and context: structure doc, docs, runnables, active spec, open specs and todos, recent logs, release status.
- `sync`, with drift nudges from `task complete`, `spec complete` and `log new`.
- Specs, tasks, todos, work logs and memories.
- `structure` with drift detection against the last commit that touched `.mem/structure.md`.
- Runnables in `.mem/runnables/`.
- `promote staging|production` with date tags, and the `pre-push`/`pre-commit` hooks.
- Releases: v0.2.0 is published with binaries for Linux, macOS and Windows, built by GoReleaser from semver tags.

Tested with disposable repositories and bare remotes; CI runs on Linux, macOS and Windows. mem manages its own source repository. praxis-app has been imported; its commit is pending review.

## Planned

- **Self-update.** A check at onboard for a newer mem release.
- **Templates.** Project profiles (for example Next.js web app, Rust + GPUI desktop app, general Python, Go and Rust) kept in a GitHub template library, supplying initial memories, docs and skills, with a way to promote a project's memory or doc back into its template and receive template updates at onboard.

## Known limits

- Hooks and runnables rely on `sh` and executable bits; Windows behavior is untested.
- `git push --no-verify` and merges made in a Git host's web UI bypass the hooks. Promotion detects the resulting divergence and prints the recovery steps.
- Work logs imported from the Python harness are interpreted in the importing machine's local time zone.
