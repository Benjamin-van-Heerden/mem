---
title: gpui-kit, hot reload and packaging skills
status: completed
created_at: "2026-10-26T13:07:01+02:00"
updated_at: "2026-10-26T14:31:06+02:00"
completed_at: "2026-10-26T14:31:06+02:00"
---

In mem-templates rust-gpui/skills: vendor longbridge/gpui-kit's skills/gpui-kit at the commit matching 0.8.1 with its LICENSE and the source commit in frontmatter (as nextjs-web's vendored skills do), plus a short local section on how this template uses it. gpui-hot-reload: first verify in the reference app that gpui-pre 0.3.10 has the hot-reload feature and that dx serve --hot-patch (dioxus-cli 0.7.10, a global install: ask the user first) patches a render change; document what works, the limits (custom Element phases, struct layout changes) and the version to retry with if it does not work. macos-app-packaging: .app bundle with Info.plist and icon, codesign --options runtime inside-out, notarytool and stapling, auto-update options, from gpui-kit's packaging docs and Zed's script/bundle-mac; verify up to an unsigned bundle of the reference app that launches.

## Completion Notes

mem-templates rust-gpui/skills (pushed): gpui-kit and gpui-kit-design-guides vendored from longbridge/gpui-kit at v0.8.1 (99576db) with LICENSE and source in frontmatter, plus an 'In this template' section; macos-app-packaging with scripts/bundle-macos.sh (release build, .app with Info.plist, icns from a PNG, ad hoc or Developer ID signing), verified on the reference app and in the e2e repo (bundle opened, process ran), signing/notarization and updates documented; gpui-hot-reload verified with dioxus-cli 0.7.10 (installed with the user's approval): a render edit hot-patched in 317-385 ms in the same process with GPUI logging 'hot reload applied'; adding a struct field was also patched and crashed the app (panic in gpui-pre view.rs:185), now documented; dx builds add ~2.5 GB and 71 s cold.
