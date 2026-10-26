---
title: Explore GPUI and Zed source
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

Sparse-clone zed-industries/zed at 08cbb88 (the gpui-pre 0.3.10 snapshot) into the scratchpad (crates/gpui, gpui_macros, gpui_platform, gpui_apple, and for patterns workspace, settings, theme, editor, ui) and use ~/.cargo/registry/src/*/gpui-pre-0.3.10 and gpui-kit 0.8.1. Run subagents in parallel, one per topic of the gpui skill's references (entities-and-contexts, elements-and-styling, actions-keys-focus, events-and-subscriptions, async-and-tasks, custom-elements-and-text-input, testing, app-architecture, outdated-patterns), each writing notes to the scratchpad with file paths and line references, the current API signatures and pitfalls. Zed's app crates are GPL-3.0: notes describe patterns, they do not copy code. Also read Zed's .rules GPUI section, gpui-kit's skills and its docs on testing, packaging and auto-update, and real apps (zedis, OpenLogi) for app architecture.
