---
title: End-to-end verification and publishing
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

Build dist/mem-dev. In a disposable Git repository run dist/mem-dev init --template rust --template rust-gpui --template-source <local mem-templates checkout> and follow .mem/setup.md as written: the app builds and runs, tests pass, the screenshot PNG shows the window's content. Check dist/mem-dev onboard runs rust-checks and gpui-checks, and that rust-checks warns with a temporarily lowered limit. Fix the template where the steps were wrong. Update mem-templates README (rust-gpui row, runnables/ in Layout). Show the user the result, then commit and push mem-templates after their review. Narrow the todo add_rust_gpui_and_go_templates to Go.
