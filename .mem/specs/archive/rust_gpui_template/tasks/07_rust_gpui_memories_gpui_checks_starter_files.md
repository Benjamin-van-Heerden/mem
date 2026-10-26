---
title: rust-gpui memories, gpui-checks, starter files and setup
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:40:24+02:00"
completed_at: "2026-10-26T13:40:24+02:00"
---

In mem-templates rust-gpui: template.toml description; memories gpui-from-source, gpui-version-pins and gpui-verify-ui (wording in the spec); runnables/gpui-checks (sh; reads only Cargo.lock and Cargo.toml: exactly one gpui-pre package, its version matches the skills' version, runtime_shaders enabled or xcrun metal works; exits 0 outside a GPUI project); skill gpui-setup with files/ from the reference app (Cargo.toml with profiles, lints and pins, rust-toolchain.toml 1.99.1, crates/app, crates/ui with the screenshot test, crates/core, assets) and a SKILL.md describing them; setup.md with the steps in the spec as checkbox headings with Done when lines, (you) steps last.

## Completion Notes

mem-templates rust-gpui (local commit): template.toml; memories gpui-from-source, gpui-version-pins, gpui-verify-ui; runnables/gpui-checks (reads Cargo.lock/Cargo.toml only; checked: in sync, outside a GPUI project, two gpui-pre versions, unpinned gpui-kit, skills written for another version; 0.4 s); skill gpui-setup with files/ (app thin binary with quit mode and hot-reload feature, ui crate with init/actions/menus/MainView, #[gpui_kit::test] behaviour test and harness=false screenshot test, core without GPUI, toolchain 1.99.1, dev profile); setup.md with eight steps, the (you) ones last. The starter built cold in 35 s in a scratch workspace, tests passed (2 clicks counted, light and dark screenshots looked at), clippy -D warnings clean, target/ 2.7 GB. gpui-checks no longer checks runtime_shaders, since gpui-kit always enables it (spec updated).
