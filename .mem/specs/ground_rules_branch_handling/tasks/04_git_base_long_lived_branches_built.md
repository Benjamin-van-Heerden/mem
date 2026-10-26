---
title: '[git] base for long-lived branches built on released code'
status: todo
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T12:58:18+02:00"
---

Add GitConfig.Base (toml base, omitempty) in internal/project. In onboard (after the sync), when base is set and the current branch is the development branch, count HEAD..<remote>/<base> commits and add an informational line (not a warning nudge) to the SHARED CODEBASE section: <dev> is built on released code from <base>; N commit(s) of <base> are not merged in yet; they come in only when the user asks, by a merge commit: git fetch <remote> && git merge --no-ff <remote>/<base>, then git push; never rebase <dev>, never force-push it, never merge other unreleased work into it. When N is 0 say it is up to date with <base> and still state the merge-only rule. Test with cli onboard: a project with base = main on a development branch behind origin/main shows the count and wording; without base nothing is shown.
