---
title: gpui-kit, hot reload and packaging skills
status: todo
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T13:07:01+02:00"
---

In mem-templates rust-gpui/skills: vendor longbridge/gpui-kit's skills/gpui-kit at the commit matching 0.8.1 with its LICENSE and the source commit in frontmatter (as nextjs-web's vendored skills do), plus a short local section on how this template uses it. gpui-hot-reload: first verify in the reference app that gpui-pre 0.3.10 has the hot-reload feature and that dx serve --hot-patch (dioxus-cli 0.7.10, a global install: ask the user first) patches a render change; document what works, the limits (custom Element phases, struct layout changes) and the version to retry with if it does not work. macos-app-packaging: .app bundle with Info.plist and icon, codesign --options runtime inside-out, notarytool and stapling, auto-update options, from gpui-kit's packaging docs and Zed's script/bundle-mac; verify up to an unsigned bundle of the reference app that launches.
