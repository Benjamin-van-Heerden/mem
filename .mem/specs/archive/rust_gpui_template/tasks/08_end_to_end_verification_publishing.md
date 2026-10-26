---
title: End-to-end verification and publishing
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T14:31:20+02:00"
completed_at: "2026-10-26T14:31:20+02:00"
---

Build dist/mem-dev. In a disposable Git repository run dist/mem-dev init --template rust --template rust-gpui --template-source <local mem-templates checkout> and follow .mem/setup.md as written: the app builds and runs, tests pass, the screenshot PNG shows the window's content. Check dist/mem-dev onboard runs rust-checks and gpui-checks, and that rust-checks warns with a temporarily lowered limit. Fix the template where the steps were wrong. Update mem-templates README (rust-gpui row, runnables/ in Layout). Show the user the result, then commit and push mem-templates after their review. Narrow the todo add_rust_gpui_and_go_templates to Go.

## Completion Notes

dist/mem-dev init --template rust --template rust-gpui --template-source <local mem-templates> in a disposable repo with a bare remote: 6 memories, 8 skills, both runnables installed and committed as 100755; onboard led with the setup and ran both runnables. Setup followed as written: renaming with macOS grep/sed worked (gpui-demo, gpui_demo, Gpui Demo), cold build 40 s, workspace tests passed (behaviour test, light/dark screenshots looked at), clippy -D warnings clean, rust-checks and gpui-checks reported in sync, the app launched, rust-checks with a lowered limit asked for cargo clean, step 7's bundle opened, step 8's hot reload verified with dx. mem-templates README lists rust-gpui and runnables/; all mem-templates commits pushed with the user's approval; docs/status.md lists the templates; the old todo now covers only Go (both committed with the previous task).
