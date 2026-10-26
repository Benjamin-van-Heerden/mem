---
title: 'Rust template: defaults, memories, skills and rust-checks'
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

In ~/Documents/yhat/mem-templates/rust: memories cargo-iteration, small-target and crate-boundaries (wording in the spec, a few lines each); fix skills/rust/SKILL.md (Static Verification and Agent Workflow must not default to --all-features or workspace-wide commands; add crate-isolation guidance; point to the memory); new skill rust-build-performance (profile and why with the measured numbers, what fills target/ and when to clean, rust-analyzer sharing target/, reading dependency source at the locked version via cargo metadata / ~/.cargo/registry/src / ~/.cargo/git/checkouts, why not Cranelift/rust-lld/sccache/blaze with the date checked, moving the toolchain pin). New runnables/rust-checks (sh): one status line; warnings with the fix when target/ exceeds the limit set from the reference app's measurement (proposal 10 GB), when the dev profile lines are missing from the workspace Cargo.toml, or when rust-toolchain.toml names an uninstalled toolchain; exits 0 outside a Cargo workspace; no cargo invocations; under 2 s on a target/ over 5 GB (check with the reference app's target/).
