---
created_at: "2026-10-27T12:08:59+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.10.0 released; rust-gpui interact-then-screenshot workflows; LearnHarness fork set up with mem

## Goal

- Release the rust_gpui_template work.
- Give rust-gpui a verified way for agents to drive a UI and screenshot each step.
- Fork amosblomqvist/learnharness for the user's own use, tracking upstream through mem.

## What Was Accomplished

### Release v0.10.0

- CI was green on ubuntu, macOS and Windows for dev at c1a6058.
- `mem promote staging`, then a corrected work log (8d73226), then staging again and production. The notes were refined and confirmed by the user.
- `mem promote production --confirm` created the date tag `v2026.10.26.2` on `main` at 8d73226.
- `v0.10.0` was tagged on the same commit with the same notes (`--cleanup=whitespace`) and pushed. The Release workflow built 7 assets, and `mem update` moved the installed mem from v0.9.1 to v0.10.0.
- The release also carried another session's completed spec `ground_rules_branch_handling` (Ground Rules, push retry, merge-safe sync, `[git] base`) and its `log commit` change. The notes covered both.

### rust-gpui: interact-then-screenshot workflows (mem-templates fc1289b)

- **The starter's `crates/ui/tests/screenshots.rs`** has named cases (`type Case = (&'static str, fn(&str))`) with a substring filter:
  - `main_view_light` and `main_view_dark` stills;
  - `increment_flow`, a workflow: open headlessly, capture `increment_flow_1_start.png`, then `window.click("increment", cx)` twice inside `cx.update_window`, then capture `increment_flow_2_after_two_clicks.png`. It asserts `view.read(cx).count() == 2` and that the two images differ.
  - `shoot` draws before capturing, because `capture_screenshot` reads the last drawn frame.
- **`image = { version = "0.25", default-features = false }`** is a dev-dependency of the ui crate, because gpui doesn't re-export `RgbaImage`. GPUI already builds that crate with PNG support.
- **The starter files are now rustfmt-formatted.** `cargo fmt --check` had failed on all of them.
- **Verified** in a scratch copy: all workspace tests pass, and clippy `-D warnings` and `cargo fmt --check` are clean. The second PNG shows "Clicked 2 times" with the button in its hover shade, since the pointer stays where it clicked.
- **Docs:**
  - The `gpui` skill's `testing` reference has a "Workflows: interact, then screenshot each step" section with a compile-checked example (now 27 examples, all compiling).
  - Rule 12 of the `gpui` skill, the `gpui-setup` skill and the `gpui-verify-ui` memory ask for workflow cases for interactions.

### LearnHarness fork (outside this repository)

- **Upstream:** amosblomqvist/learnharness is AGPL-3.0, TypeScript (Next.js inside Electron), v0.3.8, about 10 days old with 740 commits.
- **The fork:**
  - `gh repo fork` made github.com/Benjamin-van-Heerden/learnharness (public, as forks are), cloned to `~/Documents/yhat/learnharness`.
  - Remotes: `origin` is the fork; `upstream` is fetch-only, its push URL `DISABLED`.
  - `release` was created from `main` and pushed, then `mem init --no-template --production release --staging test --development dev` ran.
  - `.mem/config.toml` gained `base = 'main'` under `[git]`.
  - `.mem/runnables/upstream`:
    1. runs `gh repo sync Benjamin-van-Heerden/learnharness --branch main`, which fast-forwards the mirror and warns if `main` has diverged;
    2. fetches `origin/main` and reports where it stands and how many commits `dev` lacks;
    3. after a merge, asks for `npm install` and `npm test`.

    It runs in about 2 s.
  - The setup was committed on `dev` (5d8caa3) and pushed. Onboard shows the Base line and the runnable output.
- **Checks:**
  - `npm ci`: 34 s, 1.0 GB of `node_modules`.
  - `npm test`: 2,339 tests, 2,338 pass. The failure is `tests/subscription-auth.test.ts:35`, `ERR_PACKAGE_PATH_NOT_EXPORTED` for `@earendil-works/pi-ai`. This machine runs Node 22.19.0 and upstream's CI runs Node 26; that the version is the cause is unverified.
  - Typecheck and lint are clean.
- **The "LearnHarness fork" session** (local_707259bb) was started in `~/Documents/yhat/learnharness`, with a brief to onboard, report the failure and ask the user what to change. The first start was refused because the folder wasn't trusted; it started on the user's request. It was then detached from this session so the sidebar shows it under its own folder.

### Discussion with the user

- **gpui-kit vs hand-rolled styling.** Keep gpui-kit: GPUI has no text input, overlays, lists or tables, and gpui-kit's test API drives the headless tests. Own the look through theme tokens, or through custom `RenderOnce` components on `gpui-base`'s behaviour.
- **Driving the app.** Headless test workflows are the primary route. Live-window capture isn't available, because this session lacks Screen Recording permission. Computer use is available but untested on GPUI.
- **Whether Rust/GPUI is wise.** For the starter, the edit loop is about 1 s incremental and `target/` 3–7 GB per checkout. The bigger risk is GPUI's snapshot distribution. The user is comfortable with GPUI.
- **Fork or rewrite.** Fork learnharness rather than rewrite it in GPUI. Its core features (KaTeX, Mermaid, React Flow, pdf.js, agent-generated React components through `react-jsx-parser`) depend on a web runtime.

## Decisions

- **Fork learnharness, keep TypeScript.** No GPUI rewrite: its value is in prompts and agent orchestration, and its rendering needs the web stack. The user's use is personal, possibly shared; if it's distributed, changes stay AGPL-3.0 and the app is renamed.
- **The fork's branches:**
  - `main` is a pure mirror of upstream, never committed to;
  - `dev`, `test`, `release` are mem's branches, `release` serving as production;
  - `[git] base = 'main'`: upstream comes in only when the user asks, by `git merge --no-ff origin/main`.
- **The upstream runnable syncs the mirror itself** with `gh repo sync`, a fast-forward only, rather than only reporting. Fast-forwarding a branch that is never committed to is safe, and keeps the Base count current.
- **The starter's screenshot cases take their name as an argument,** so each case writes `<name>.png` or `<flow>_<n>_<what>.png` and the filter matches on names.

## Key Files Affected

- `.mem/logs/benjamin_van_heerden_20261026_143139.md`: corrected the note on what task commits swept in (8d73226)
- `~/Documents/yhat/mem-templates` (fc1289b):
  - `rust-gpui/skills/gpui-setup/files/crates/ui/tests/screenshots.rs`, `crates/ui/Cargo.toml`: the workflow case and the `image` dev-dependency
  - `rust-gpui/skills/gpui-setup/files/crates/ui/src/{lib,main_view}.rs`, `tests/main_view.rs`: rustfmt
  - `rust-gpui/skills/gpui/references/testing.md`, `rust-gpui/skills/gpui/SKILL.md`, `rust-gpui/skills/gpui-setup/SKILL.md`, `rust-gpui/memories/gpui-verify-ui.md`
- `~/Documents/yhat/learnharness` (the new fork): `AGENTS.md` (mem block above upstream's rules), `.gitignore`, `.claude/settings.json`, `.cursor/hooks.json`, `.mem/config.toml`, `.mem/runnables/upstream`
- Tags `v2026.10.26.2` and `v0.10.0` on `main`

## Notes

- **The release process held with the dev build's new features.** Re-promoting staging after the log correction made `mem promote production` redraft the notes ("the previous draft was for another commit").
- **`cargo binstall` takes `-y` or `--no-confirm`, not both:** passing both errors with "cannot be used multiple times".
- **Auto mode blocked running the fork's test suite** as external code until the user added an allow rule. The first test run's grep for `real` matched test names, so totals came from a second run with `tail`.
- **`mem init` refuses a production branch that doesn't exist** ("the production branch release does not exist locally or on origin"). Create and push it first.
