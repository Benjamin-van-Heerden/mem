---
title: Let mem init start from an empty repository
status: open
created_at: "2026-09-29T21:42:21+02:00"
---

Cloning a freshly created, empty GitHub repo and running `mem init` fails with: 'this repository has no commits yet; make a first commit, then run `mem init`'. For the flow 'create empty GitHub repo -> mem init -> setup runs', init should either create the initial commit itself (e.g. an empty commit or one containing the files it writes) on the production branch before creating dev/test, or print the exact commands. Test in a disposable clone of an empty bare repo.
