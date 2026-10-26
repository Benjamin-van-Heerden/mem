---
title: Explore GPUI and Zed source
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:24:11+02:00"
completed_at: "2026-10-26T13:24:11+02:00"
---

Sparse-clone zed-industries/zed at 08cbb88 (the gpui-pre 0.3.10 snapshot) into the scratchpad (crates/gpui, gpui_macros, gpui_platform, gpui_apple, and for patterns workspace, settings, theme, editor, ui) and use ~/.cargo/registry/src/*/gpui-pre-0.3.10 and gpui-kit 0.8.1. Run subagents in parallel, one per topic of the gpui skill's references (entities-and-contexts, elements-and-styling, actions-keys-focus, events-and-subscriptions, async-and-tasks, custom-elements-and-text-input, testing, app-architecture, outdated-patterns), each writing notes to the scratchpad with file paths and line references, the current API signatures and pitfalls. Zed's app crates are GPL-3.0: notes describe patterns, they do not copy code. Also read Zed's .rules GPUI section, gpui-kit's skills and its docs on testing, packaging and auto-update, and real apps (zedis, OpenLogi) for app architecture.

## Completion Notes

Sparse clone of zed@08cbb88 plus the gpui-pre 0.3.10 and gpui-kit 0.8.1 registry sources explored by five parallel subagents; notes (scratchpad gpui-notes/, ~1,800 lines) cover the nine reference topics with file:line citations, current signatures, pitfalls and Zed patterns described without copying GPL code. Licences checked: gpui-pre crates and gpui-kit Apache-2.0; third-party apps used for architecture are MIT/Apache. Key findings: AsyncApp updates are infallible, FluentBuilder needs gpui_kit::prelude, flex_col needs flex, svg needs a text colour, application() registers no assets, test-support needs harness = false screenshots on the main thread.
