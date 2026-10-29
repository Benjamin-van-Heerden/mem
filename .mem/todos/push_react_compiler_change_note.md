---
title: Push the React Compiler change note after the next mem release
created_at: "2026-10-29T09:03:58+02:00"
---

mem-templates eb333eb on main (nextjs-web/changes/2026-10-29-react-compiler.md and the README's Change notes section) is committed but not pushed: mem reads change notes only from the release that includes the template change notes spec. After that release (check with mem version that the installed mem has it), push mem-templates main. Projects using nextjs-web, such as studii-next, then get the 'Apply the nextjs-web template change: React Compiler through the Rust port in Turbopack' todo at their next onboard. Delete this todo once pushed.
