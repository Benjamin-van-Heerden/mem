---
title: Rust GPUI template
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-10-26T13:04:31+02:00"
updated_at: "2026-10-26T13:08:58+02:00"
---

## Overview

Add a `rust-gpui` template to mem-templates (`~/Documents/yhat/mem-templates`) for macOS-first desktop apps built with Zed's GPUI. A project uses it together with the existing `rust` template: `mem init --template rust --template rust-gpui`.

Rust's compile times and the size of `target/` decide whether agents stay productive in such a project. So the work has three parts:

1. **Rust defaults for every Rust project** (the `rust` template):
   - fast iteration
   - a small `target/`
   - proper crate isolation
   - idiomatic standards
2. **GPUI knowledge.** GPUI is barely documented and changes weekly, so skills are written from its source and Zed's, against one pinned version.
3. **An onboard check.** A script run at every onboard reports `target/` size and catches drift from the defaults. It needs a small mem feature: templates ship runnables.

The research behind the defaults was done on 2026-10-26 and is summarised in Notes.

## Goals

- **Fast iteration.** An agent in a Rust project follows a fast loop by default:
  - `cargo check -q -p <crate>` while iterating;
  - scoped clippy and tests before finishing;
  - one cargo command at a time;
  - long builds in the background.
- **A small `target/`.** It stays small by default, and onboard says when it has grown past a limit and how to fix it.
- **Crate isolation.** Workspaces keep a thin binary crate and domain logic that compiles without the UI framework, so edits rebuild little and core tests don't compile GPUI.
- **Current GPUI, not training data.** Agents write GPUI code against the pinned version's real API and read the locked source when unsure. Every code snippet in the GPUI skills compiles against that version.
- **Setup to a working app.** A new project goes from an empty repository to a running GPUI app, with a headless screenshot test, by following `.mem/setup.md`.
- **Agents check UI changes themselves** by rendering headless screenshots in tests, without asking the user to look.

## Technical Approach

### 1. mem: templates ship runnables

- A template may contain `runnables/<name>`. Sync installs it to `.mem/runnables/<name>` with the executable bit set, records it in `.mem/templates.lock`, and classifies and reconciles it like skills and docs: missing, updated, local edits, both edited.
  - Code: `internal/templates/sync.go`, beside `SkillPath`/`DocPath`.
- `mem template promote runnable <name>` sends a project's runnable back to the library, like the other kinds. `mem template list` shows runnables.
- Onboard already runs `.mem/runnables/*` (`internal/runnables`, 15 s timeout, 20,000-byte cap). Nothing changes there.
- Text read from the library may have CRLF line endings on Windows; normalise to LF, as for other items, or the shebang line breaks.
- Tests: in `internal/templates` (install, update, local edits, promote) and an onboard test that runs a template-installed runnable.
- Docs: README, `docs/design.md`, `docs/status.md`, `.mem/structure.md`, and the Layout section of mem-templates' README.
- Releasing mem is the user's call: `mem promote staging`/`production` only when asked.

### 2. The `rust` template (all Rust projects)

**Profile.** The skill and the setup put this in the workspace `Cargo.toml`. It was measured on this machine; see Notes.

```toml
[profile.dev]
debug = "line-tables-only"     # file:line backtraces for our crates

[profile.dev.package."*"]
debug = false                  # dependencies' debug info is most of target/; also lets `check` reuse what `build` compiled
```

- Leave the linker, `incremental`, `codegen-units` and `split-debuginfo` at their defaults.
- Add `opt-level` only for named hot dependencies. Raising it for all dependencies costs +70–100 s per cold build.
- Keep one `target/` per checkout, Cargo's default. The user rarely works in worktrees, so there is no shared build-dir and no `CARGO_TARGET_DIR`.

**Memories.** Short, always in context. Keep each to a few lines, because AGENTS.md size matters (see the todo `slim_managed_agents_md_block`).

- `cargo-iteration`:
  - While iterating, run `cargo check -q -p <crate>`. Before finishing, run `cargo clippy -q -p <crate> --all-targets -- -D warnings` and the tests of the touched crates.
  - Never use `--release` or `--all-features` while iterating.
  - Run one cargo command per checkout at a time. When cargo says it is blocking on a file lock, wait; never kill it or retry.
  - Run commands likely to take over ~90 s in the background, then read the start of the error output, not the tail.
- `small-target`:
  - Keep the dev profile above.
  - Run `cargo clean` only when the onboard check or the user asks; it costs a cold build.
  - Don't turn on debug info or optimisation for all dependencies.
- `crate-boundaries`:
  - Keep the binary crate thin.
  - Domain logic goes in crates that don't depend on the UI framework or I/O-heavy dependencies.
  - Split out a crate when a part compiles and tests on its own.
  - Tests live in the crate they test, so `-p` scopes them.

**Skills.**
- `rust`: fix the Static Verification and Agent Workflow sections, which tell agents to run `--all-features` and whole-workspace commands. Point to the iteration memory and add crate-isolation guidance. Keep the rest.
- `rust-build-performance` (new). Covers:
  - the profile and why, with the measurements;
  - what fills `target/` and when to clean;
  - rust-analyzer: it shares `target/` by default, so expect the occasional lock wait; a separate target dir doubles disk use;
  - reading a dependency's API at the locked version (`cargo metadata` → `manifest_path`, `~/.cargo/registry/src`, `~/.cargo/git/checkouts`);
  - why not Cranelift, rust-lld, sccache or blaze, each with the date checked;
  - moving the toolchain pin.

**Runnable `rust-checks`.** It prints one short status line, and ⚠️ lines with the fix only when something is wrong:
- `target/` larger than the limit. Set the limit from the reference app's measurement; the proposal is 10 GB. The fix is "run `cargo clean` now, before starting work, when no cargo command is running".
- The dev profile lines are missing from the workspace `Cargo.toml`.
- `rust-toolchain.toml` names a toolchain that `rustup toolchain list` doesn't have.

It must finish well inside 15 s on a 10+ GB `target/`: `du -sk`, no cargo invocations. It must also exit 0 with a short note outside a Cargo workspace.

### 3. GPUI exploration and skills (`rust-gpui` template)

**Version.** Pin `gpui-kit = "=0.8.1"`, which pins `gpui-pre =0.3.10`, the 2026-10-25 snapshot of zed@08cbb88. When GPUI is needed directly:

```toml
gpui = { package = "gpui-pre", version = "=0.3.10" }
gpui_platform = { package = "gpui-pre-platform", version = "=0.3.10", features = ["font-kit", "runtime_shaders"] }
```

- Without `font-kit` on macOS, no glyphs render.
- `runtime_shaders` removes the need for Xcode's Metal toolchain download. This machine lacks it: `xcrun metal` fails.

If newer versions exist when implementation starts, decide with the user whether to move the pin before writing the skills.

**Sources to explore:**
- `~/.cargo/registry/src/*/gpui-pre-0.3.10/` and the other `gpui-pre-*` crates;
- gpui-kit 0.8.1 and its story app;
- a sparse clone of zed-industries/zed at 08cbb88 in the scratchpad: `crates/gpui`, `crates/gpui_macros`, `crates/gpui_platform`, `crates/gpui_apple`, and the app crates for patterns: `workspace`, `settings`, `theme`, `editor`, `ui`;
- Zed's `.rules` GPUI section.

Explore with subagents, one topic each, producing notes in the scratchpad. Then write the skills.

**Licences.**
- `gpui`, `gpui_macros` and `gpui_platform` are Apache-2.0, and so is gpui-kit.
- Most of Zed's app crates (`workspace`, `editor`, `ui`, …) are GPL-3.0. The skills describe their patterns in our own words and code, and copy nothing from GPL crates.

**Skill `gpui`.** A `SKILL.md` overview plus `references/`:
- `entities-and-contexts`: `Entity`, `Context<T>`, `App`, `Window`, `WeakEntity`, update/read rules, the reentrancy panic, `Global`.
- `elements-and-styling`: `div()`, layout, the Tailwind-like style methods, `RenderOnce`/`IntoElement`, lists, `uniform_list`, images and SVG, assets.
- `actions-keys-focus`: `actions!` and `derive(Action)`, key bindings and contexts, `FocusHandle`, dispatch.
- `events-and-subscriptions`: `EventEmitter`, `subscribe`, `observe`, keeping `Subscription`s.
- `async-and-tasks`: `cx.spawn`, `background_spawn`, `Task` cancellation on drop, executor timers.
- `custom-elements-and-text-input`: the `Element` trait phases, `EntityInputHandler`, and when to use gpui-kit's `Input` instead.
- `testing`: `#[gpui::test]`, `TestAppContext`, `VisualTestContext`, simulated input, `run_until_parked`, headless screenshots (`HeadlessAppContext`, `capture_screenshot`, `harness = false`).
- `app-architecture`: state in entities and globals, views, panes and docks, settings and themes, as Zed and real apps (zedis, OpenLogi) structure them.
- `outdated-patterns`: the table of pre-2025 and 0.2.2 idioms with their current forms.

Rules for the skill:
- Each reference names the GPUI version and the source files it was written from.
- Every Rust code block compiles against the pinned version: collected in the reference app and checked with `cargo check`.

**Skill `gpui-kit`.** Vendor longbridge's own skill (`longbridge/gpui-kit/skills/gpui-kit`) with its LICENSE and the source commit in frontmatter, as the nextjs-web skills do. Add a short local section on how this template uses it.

**Skill `gpui-hot-reload`.** Subsecond hot-patching (zed-industries/zed#65738, merged 2026-10-17):
- the `hot-reload` feature;
- `dx serve --hot-patch` with dioxus-cli 0.7.10 (a global install the user approves);
- its limits: custom `Element` phases, struct layout changes.

First verify that gpui-pre 0.3.10 has the feature and that `dx` copes with the `gpui-pre` package rename. If it doesn't work, the skill says so and names the version to retry with.

**Skill `macos-app-packaging`.** Bundling a `.app` with an Info.plist and icon, `codesign --options runtime` signing inside-out, notarisation with `notarytool` and stapling, and the auto-update options. Based on gpui-kit's packaging docs and Zed's `script/bundle-mac`. Verify only up to an unsigned bundle that launches; signing needs the user's Developer ID.

**Memories:**
- `gpui-from-source`:
  - Never write GPUI from memory: training data mixes the 2024 Model/View API, 0.2.2 and main.
  - Use the `gpui` skill, and read the locked `gpui-pre`/gpui-kit source when unsure.
  - Never translate an older GPUI example by analogy.
- `gpui-version-pins`:
  - `gpui-kit` and `gpui-pre*` have exact pins and are bumped together.
  - `Cargo.lock` holds exactly one `gpui-pre`.
  - After a bump, re-check the skills' "outdated patterns" against the new source.
- `gpui-verify-ui`: check UI changes with a headless screenshot test and look at the PNG before saying a view works.

**Runnable `gpui-checks`.** It reads only `Cargo.lock` and `Cargo.toml`, no cargo invocations:
- exactly one `gpui-pre` package in `Cargo.lock`;
- its version matches the one the skills were written for;
- `runtime_shaders` is enabled, or `xcrun metal --version` works.

### 4. Setup and starter files

`rust-gpui/setup.md` uses checkbox steps with "Done when" lines, as in nextjs-web. The verified starter files live in the skill `gpui-setup`, under `files/`.

**Starter workspace:**
- `Cargo.toml`: workspace, profiles, `[workspace.lints]`, exact GPUI pins.
- `rust-toolchain.toml`: `1.99.1`, components clippy, rustfmt and rust-src.
- `crates/app`: a thin binary that runs the application, initialises gpui-kit and opens the main window.
- `crates/ui`: GPUI views on gpui-kit, and a headless screenshot test with `harness = false`.
- `crates/core`: domain logic with no GPUI dependency and a unit test.
- `assets/`: an icon and an SVG.

**Steps:**
1. Check prerequisites: rustup, and Xcode or its Command Line Tools.
2. Copy the starter files and rename the crates to the project's name.
3. First build, in the background. Record its time and the size of `target/`.
4. Tests, including the screenshot test. Look at the PNG.
5. Optional: hot reload.
6. (you) The app's display name, bundle identifier and icon.

### 5. Verification

Build the reference app in the scratchpad with Rust 1.99.1 (installed on 2026-10-26). Measure:
- the cold build, the incremental `check` and `build`, and the test run;
- the size of `target/` with the profile.

Those numbers set the `rust-checks` limit and go into `rust-build-performance`.

Then, end to end:
- In a disposable repository, run `dist/mem-dev init --template rust --template rust-gpui --template-source <local mem-templates checkout>` and follow `.mem/setup.md` as written.
- Check that onboard runs both runnables, and that `rust-checks` warns when `target/` is over a temporarily lowered limit.

## Success Criteria

- **mem.** A runnable in a template is installed to `.mem/runnables/` with its executable bit, tracked in `.mem/templates.lock`, updated on the next sync, and kept when edited locally. `mem template promote runnable <name>` and `template list` handle it. The new tests pass, with `go test` and `go vet` clean for `internal/templates` and `internal/cli`. The docs describe the feature.
- **The `rust` template:**
  - the memories `cargo-iteration`, `small-target` and `crate-boundaries`;
  - the updated `rust` skill, with no `--all-features` or workspace-wide commands as the default loop;
  - the new `rust-build-performance` skill;
  - the runnable `rust-checks`, which runs in under 2 s on a `target/` of more than 5 GB and warns when the limit is exceeded or the profile is missing.
- **The `rust-gpui` template:**
  - the skills `gpui` (with all nine references), `gpui-kit` (with its licence and source commit), `gpui-hot-reload`, `macos-app-packaging` and `gpui-setup`;
  - the memories `gpui-from-source`, `gpui-version-pins` and `gpui-verify-ui`;
  - the runnable `gpui-checks` and the `setup.md`.
- **Snippets compile.** Every Rust code block in the `gpui` skill is checked to compile against gpui-pre 0.3.10 in the reference app.
- **The reference app works.** It builds, its core and UI tests pass, and the screenshot test writes a PNG that shows the window's content. Its measured numbers are recorded in `rust-build-performance`.
- **End to end.** In a disposable repository, `mem init --template rust --template rust-gpui` and the setup as written produce a running app. Onboard prints both runnables' status.
- **Publishing.** The mem-templates README lists `rust-gpui` and documents `runnables/`. mem-templates is committed and pushed after the user reviews.

## Notes

### Research, 2026-10-26

Measured on an M4 Pro (12 cores, 24 GB), macOS 27, Xcode 27, Rust 1.96.1, with a GPUI 0.2.2 hello world and `runtime_shaders`.

**Cargo default vs the profile above:**

| | Cargo default | Profile above |
|---|---|---|
| Cold build | ~55 s | ~44–55 s |
| `target/` after build and check | 2.7 GB | 1.4 GB |
| Cold check after a build | ~35 s | ~20 s |
| Incremental build | ~1 s | ~1 s |

**Other settings:**
- `opt-level = 1` or `3` for all dependencies: a cold build of 124–152 s. Zed optimises only the hot crates: taffy, resvg and proc macros.
- `split-debuginfo = "packed"` adds ~2.9 s per link through dsymutil.
- rust-lld from 1.96.1 can't link against the Xcode 27 SDK (`arm64e.x1-macos`). That's fixed in Rust 1.100 (2026-11-12), and it was no faster than Apple's ld anyway.

**Cargo and Rust features:**
- `build.build-dir` is stable since 1.91. Cargo has no GC for `target/`: `cargo clean gc` covers only `~/.cargo`. cargo-sweep 0.8.0 doesn't support build-dir.
- Rust 1.100 garbage-collects old incremental sessions (rust#162240): 27–31% smaller `target/` in that patch's own tests. Move the pin to 1.100 after it ships, and update the skill's numbers.
- Nightly-only, so not used: Cranelift (it fails on `pathfinder_simd` on arm64), `-Zthreads`, `-Zfine-grain-locking`. sccache can't cache incremental, bin or proc-macro crates.

**GPUI:**
- Official `gpui` on crates.io stopped at 0.2.2 (2025-10-22).
- `gpui-pre` is a near-weekly snapshot of Zed main published by Longbridge (huacnlee), so it is a third-party dependency.
- gpui-pre 0.3.10 needs Rust ≥ 1.98: it uses `map_or_default`. Zed pins 1.98.1.
- `create-gpui-app` is abandoned.

**Experimental compilers, checked 2026-10-26.** The user cancelled blaze for now, and none of the others is usable:
- blaze/ra-jit: a borrow checker that misses about a third of rustc's errors; pinned to 1.99.0; 3–12× cargo's memory.
- rustd: needs exactly 1.99.0 and has no borrow checker.
- Relax: unreleased, on Makepad's `work` branch.
- Lust: Linux only, and its subset excludes GPUI's code.
- The Astral/OpenAI rustc fork: not public; small pieces are upstreamed into 1.99–1.101.

Faster compilers are expected soon, so the iteration command lives in one memory (`cargo-iteration`), which can change in one place.

**Agent guidance in real repositories** (Zed, uv, Ruff, rust-analyzer, Nushell, gpui-kit) agrees on:
- narrow `-p` checks and tests;
- no release builds while iterating;
- waiting on cargo's locks;
- no `cargo clean`;
- reading the snapshots before accepting them.

Zed's `.rules`, AGENTS.md and CLAUDE.md are identical; its GPUI section is the best short primer. gpui-kit ships agent skills and `https://gpui-kit.com/llms.txt`.

### Decisions (user, 2026-10-26)

- No blaze or other experimental compiler for now. Rely on sensible defaults.
- `target/` must stay small: a memory plus an onboard check.
- One `target/` per checkout, because worktrees are rare.
- Templates ship runnables (a mem feature) rather than a setup step copying the script.
- Rust 1.99.1 is the default toolchain on this machine (with clippy, rustfmt, rust-src and rust-analyzer).

### Follow-ups

- The todo `add_rust_gpui_and_go_templates` keeps only the Go part once this spec completes.
- The scratchpad bench (`gpui-bench/RESULTS.md`) is session-only; its numbers are in these notes.
