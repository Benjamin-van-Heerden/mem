---
title: Build and measure the reference GPUI app
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

In the scratchpad, build a reference workspace with Rust 1.99.1 (rust-toolchain.toml): crates/app (thin bin: gpui_kit::application(), gpui_kit::init, open a window), crates/ui (a gpui-kit view with a button and an Input, plus a headless screenshot test with harness = false using HeadlessAppContext and capture_screenshot that writes a PNG), crates/core (domain logic with no GPUI, one unit test), assets/. Pins: gpui-kit =0.8.1; gpui = { package = gpui-pre, =0.3.10 } and gpui_platform = { package = gpui-pre-platform, =0.3.10, features font-kit and runtime_shaders } only if used directly. Use the dev profile from the spec. Measure with /usr/bin/time: cold build, cold check after build, incremental check and build after a one-line edit in ui, the test run, and du -sh of target/ broken down by deps/build/incremental. Look at the screenshot PNG. If newer gpui-kit/gpui-pre versions exist, ask the user whether to move the pin first. These files become the gpui-setup starter files later; record the numbers in the spec Notes.
