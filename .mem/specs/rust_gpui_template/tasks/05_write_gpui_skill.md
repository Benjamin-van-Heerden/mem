---
title: Write the gpui skill
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:35:43+02:00"
completed_at: "2026-10-26T13:35:43+02:00"
---

In mem-templates rust-gpui/skills/gpui: SKILL.md (when to use, the rules that matter most, an index) and references/ for the nine topics in the spec, written from the exploration notes. Each reference names the GPUI version and the source files it came from. Every Rust code block is collected into the reference app (for example a crate or module of snippet checks) and passes cargo check against gpui-pre 0.3.10; fix the skill until it does. No code copied from GPL-3.0 Zed crates.

## Completion Notes

mem-templates rust-gpui/skills/gpui (local commit, pushed after review): SKILL.md with the version pin, where the locked source lives, twelve rules, a complete gpui-kit view and an index; nine references built from the exploration notes, Zed paths pointed at the public commit and nothing copied from GPL crates; signature listings marked text. All 26 rust blocks compile against gpui-kit 0.8.1 / gpui-pre 0.3.10 and the #[gpui_kit::test] example passes (typing and Enter update the model); checked in the reference app's snippets crate and again with the shipped scripts/check_examples.py in a fresh crate (36 s). Compiling them together exposed that duplicate action names panic at startup, now documented.
