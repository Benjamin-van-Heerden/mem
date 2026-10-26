---
created_at: "2026-10-26T14:31:39+02:00"
user: benjamin_van_heerden
spec: rust_gpui_template
---

# Work Log - Rust GPUI template: research, rust defaults, rust-gpui template, templates ship runnables

## Goal

Give mem-templates a template for macOS-first desktop apps built with GPUI. The user named three preconditions:
- iteration that stays fast despite Rust's compile times;
- a `target/` that stays small;
- good crate isolation and idiomatic Rust.

GPUI is barely documented, so its skills had to be written from its source and Zed's.

## What Was Accomplished

### Research (subagents, 2026-10-26)

**Build speed and disk.** Measured on the M4 Pro with a GPUI hello world, then with the gpui-kit reference app:

| | Cargo default dev profile | Template profile |
|---|---|---|
| Cold build | 38.3 s | 37.8 s |
| Cold check after the build | 23.3 s | 22.1 s |
| `target/` after build and check | 3.7 GB | 2.1 GB |

- The template profile is `debug = "line-tables-only"` for own crates and `debug = false` for dependencies.
- Optimising all dependencies took a cold build to 124–152 s.
- `split-debuginfo = "packed"` adds about 2.9 s per link.
- rust-lld from 1.96/1.99 can't link against the Xcode 27 SDK (`arm64e.x1-macos`); fixed in 1.100.
- Cranelift and `-Zthreads` are nightly-only.
- sccache barely helps an incremental loop.
- A shared `build.build-dir` made a second worktree's first build 0.74 s instead of 48 s. It wasn't adopted because worktrees are rare here.

**Agent workflow.** The agent files of Zed, uv, Ruff, rust-analyzer, Nushell and gpui-kit agree on:
- `check -q -p` while iterating, scoped clippy and tests before finishing;
- no `--release`, waiting on cargo's file locks, no `cargo clean`.

Two facts on the agent side:
- Claude Code's Bash tool times out after 2 minutes by default.
- The `rust-analyzer-lsp` plugin shares `target/` with the agent.

**GPUI.**
- Official `gpui` on crates.io stopped at 0.2.2. Apps use `gpui-pre`, weekly snapshots of Zed main published by Longbridge, through `gpui-kit`, the former gpui-component.
- gpui-kit 0.8.1 pins gpui-pre 0.3.10 (zed@08cbb88), which needs Rust ≥ 1.98, and enables `runtime_shaders`. This Mac lacks Xcode's Metal toolchain, and that feature removes the need for it.

**Experimental compilers** the user asked about:
- blaze (formerly ra-jit): a borrow checker that misses about a third of rustc's errors, pinned to 1.99.0, 3–12× cargo's memory.
- rustd: needs exactly 1.99.0, no borrow checker.
- Relax: unreleased, on Makepad's `work` branch.
- Lust: Linux only, and its subset excludes GPUI code.
- The Astral/OpenAI rustc fork: not public; small pieces upstreamed into 1.99–1.101.

The user cancelled blaze and chose sensible defaults.

### mem: templates ship runnables (9940030, a0770a7)

- **Sync.** A template's `runnables/<name>` is installed to `.mem/runnables/<name>` by `installRunnable` in `internal/templates/sync.go`, with LF line endings and mode 0755. It's tracked as kind `runnable` in `.mem/templates.lock`, and synced, reset and promoted like docs: `RunnablePath`, with `localPath` now used by `localHash` and `export`.
- **CLI.** Help for `template`, `promote`, `reset` and `init` names the kind.
- **Tests:**
  - `TestSyncInstallsRunnablesExecutableWithLFLineEndings`: a CRLF script installed with LF, hidden files skipped, stays in sync, updates, a two-sided edit flagged.
  - `TestOnboardRunsTemplateRunnablesAndPromoteSendsThemBack`: committed as 100755, runs at onboard, promote pushes an edit back, list shows it in sync.
  - Windows CI then failed, because shell scripts can't execute there. The test now checks execution only off Windows; that limitation was already in `docs/status.md`.
- **Docs:** README, `docs/design.md`, `docs/status.md`, `.mem/structure.md`.

### mem-templates `rust` (0833b2e)

- **Memories:**
  - `cargo-iteration`: check while iterating, scoped clippy and tests, one cargo at a time, wait on locks, long commands in the background, read errors from the top;
  - `small-target`;
  - `crate-boundaries`.
- **The `rust` skill:** Static Verification and Agent Workflow no longer default to `--all-features` or the whole workspace. Crate Boundaries explains crate boundaries as compile boundaries: a thin binary, inward dependencies, `[workspace.dependencies]`/`[workspace.lints]`, `default-members`.
- **New skill `rust-build-performance`:**
  - the profile and the measured numbers;
  - the iteration loop, long commands, locks and rust-analyzer;
  - what fills `target/` and when to clean;
  - reading locked dependency source through `cargo metadata`;
  - moving the toolchain pin;
  - a dated table of rejected options.
- **`runnables/rust-checks`** (sh, no cargo): `target/` size against a 10 GB limit, the two profile lines, and that the pinned toolchain is installed. It takes 0.11 s on a 5.8 GB target directory.

### mem-templates `rust-gpui` (c65062c, 1d90ee3, 5798940, 056a706, b95830e)

- **Exploration.** Five subagents read gpui-pre 0.3.10, gpui-kit 0.8.1, a sparse clone of zed@08cbb88 and real apps (zedis, OpenLogi, dbflux, disktree; licences checked). They wrote about 1,800 lines of notes with `file:line` citations.
- **Skill `gpui`.** `SKILL.md` has the pin, where the locked source is, twelve rules, a complete gpui-kit view and an index. There are nine references: entities-and-contexts, elements-and-styling, actions-keys-focus, events-and-subscriptions, async-and-tasks, custom-elements-and-text-input, testing, app-architecture and outdated-patterns. Nothing is copied from Zed's GPL crates.
  - All 26 `rust` blocks compile against the pin, and the `#[gpui_kit::test]` example passes.
  - `scripts/check_examples.py <scratch dir> [gpui-kit version]` rebuilds that check in a fresh crate in about 36 s.
- **Vendored skills.** `gpui-kit` and `gpui-kit-design-guides` from longbridge/gpui-kit v0.8.1 (99576db), with LICENSE, source frontmatter and an "In this template" section.
- **`macos-app-packaging`.** `scripts/bundle-macos.sh <crate> "<App Name>" <bundle id> [icon.png]` does a release build, writes the `.app` and its Info.plist, builds an icns, and signs ad hoc or with `SIGN_IDENTITY`. Signing, notarization, DMG, universal binaries and update options are documented.
- **`gpui-hot-reload`.** Verified with dioxus-cli 0.7.10, installed with the user's approval:
  - Setup: the app crate enables `gpui-pre`'s `hot-reload` feature through a renamed optional dependency.
  - A render edit was patched in 317–385 ms in the same process ("GPUI hot reload applied").
  - Adding a struct field was patched too, and crashed the app.
  - `dx` builds into its own directories: 71 s and 2.5 GB.
- **`gpui-setup` starter files:**
  - workspace with `default-members = app`, exact pins, lints and the dev profile, plus `rust-toolchain.toml` (1.99.1);
  - `crates/app` (`with_assets`, `QuitMode::LastWindowClosed`, the `hot-reload` feature);
  - `crates/ui` (`init` with the Quit action, cmd-q and the menu; `open_main_window`; `MainView` with a focus handle, theme colours and a gpui-kit `Button`; a `#[gpui_kit::test]` click test; a `harness = false` screenshot test in light and dark);
  - `crates/core` (a GPUI-free `Counter`).
- **Memories:** `gpui-from-source`, `gpui-version-pins`, `gpui-verify-ui`.
- **`runnables/gpui-checks`:** one `gpui-pre` in `Cargo.lock`, at the version the skill was written for, and `gpui-kit` pinned exactly.
- **`setup.md`:** eight steps, the `(you)` ones last.
- **README:** a rust-gpui row and `runnables/` in the layout.

### Verification

- **Reference app** (gpui-kit 0.8.1, 796 crates):
  - incremental `check -p ui` 0.3–0.6 s, `build` 0.8–1.5 s, core test 0.7 s;
  - the screenshot test: 20 s first build, about 3 s later, a 960×640 PNG;
  - release build 62 s.
- **End to end:** `dist/mem-dev init --template rust --template rust-gpui --template-source ~/Documents/yhat/mem-templates` in a disposable repository with a bare remote, then `setup.md` followed as written:
  - renaming with macOS grep and sed;
  - cold build 40 s;
  - tests, both screenshots looked at, clippy clean;
  - both runnables in sync, and `rust-checks` asking for `cargo clean` over a lowered limit;
  - the app launched, the bundle opened, hot reload worked.
- **Pushed:** mem-templates main, with the user's approval.

## Decisions

- **No experimental compiler** (user, 2026-10-26). Faster compilers are expected soon, so the iteration command lives in one memory, `cargo-iteration`.
- **One `target/` per checkout, no shared build-dir** (user): worktrees are rare here.
- **Templates ship runnables** (user) rather than a setup step copying scripts, so checks stay current through template sync.
- **Two templates used together:** `rust` holds what every Rust project needs, `rust-gpui` what's specific to GPUI.
- **`gpui-kit` as the only GPUI dependency,** pinned exactly. It's the app-facing facade (its own "public facade decision"), and it pins gpui-pre.
- **Rust 1.99.1 pinned and made the machine's default** (user). Move to 1.100 after 2026-11-12, for incremental GC and rust-lld working with Xcode 27.
- **`gpui-checks` doesn't check `runtime_shaders`:** gpui-kit 0.8.1 always enables it.
- **Hot reload is opt-in** (an app feature plus `dx`), because of its build and disk cost and the struct-change crash.
- **The starter has no logger dependency.** The hot-reload skill says how to add one to see GPUI's messages.

## Key Files Affected

- `internal/templates/library.go`: the `Runnable` kind, `runnables/` dir, item naming and paths
- `internal/templates/sync.go`: `installRunnable`, `RunnablePath`, `localHash` through `localPath`
- `internal/templates/promote.go`: runnable path and export, the error message
- `internal/templates/sync_test.go`, `internal/cli/template_test.go`: the runnable tests
- `internal/cli/template.go`, `internal/cli/init.go`: help text
- `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`
- `.mem/specs/archive/rust_gpui_template/`, `.mem/todos/add_rust_gpui_and_go_templates.md` (now Go only)
- Outside this repository:
  - `~/Documents/yhat/mem-templates`: `rust/` (memories, skills/rust, skills/rust-build-performance, runnables/rust-checks), `rust-gpui/` (template.toml, setup.md, memories, runnables/gpui-checks, skills gpui, gpui-kit, gpui-kit-design-guides, gpui-setup, gpui-hot-reload, macos-app-packaging), `README.md`
  - Rust 1.99.1 (with rust-analyzer) as the rustup default, and `~/.cargo/bin/dx` (dioxus-cli 0.7.10)

## Notes

- **Duplicate action names panic at startup** ("Action with name `app::Quit` already registered"). Compiling all skill examples into one crate found it, and the `actions-keys-focus` reference now says so.
- **Adding gpui-kit's `test-support` to an existing lockfile failed to resolve** (`taffy`, `unicode-properties`), because the feature pulls the optional `gpui-fast` crates into resolution. A fresh `cargo generate-lockfile` worked, so the starter ships the dev-dependency from the start.
- **`border_1()` without a `border_color` draws nothing,** and `svg()` needs a size and a text colour.
- **This session can't capture windows** (`screencapture -l` failed for lack of Screen Recording permission). Hot reload was confirmed through GPUI's log lines instead, with a temporary `env_logger` in the disposable app.
- **The build-speed subagent's first crates.io API call sent the user's email in its User-Agent header.** Later calls used a generic one. The user was told.
- **Another session committed in this checkout at the same time** (log commit changes, "Turn off Git maintenance for test remotes"). Its first commit broke `TestLogCommitCommitsRecordsSyncsAndPushes` in CI until its own fix. Task commits named their paths, except task 6's, which committed the todo and `docs/status.md` meant for task 8; nothing of the other session's was in the tree.
