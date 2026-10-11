---
title: Fix flaky TempDir cleanup in cli tests on macOS CI
created_at: "2026-10-11T14:19:49+02:00"
---

CI run 38132015281 (dev, 2026-10-11) failed on macos-latest: TestOnboardRemovesTodoClaimsFromOlderProjects passed its assertions but t.TempDir cleanup failed with 'unlinkat …/teammate/.git/objects/pack: directory not empty'. Likely a background git process (auto gc / maintenance, gc.autoDetach) still writing into the test repository. Consider setting gc.auto=0 and maintenance.auto=false for test repositories (e.g. in sharedProject or via GIT_CONFIG_* env in tests) and confirm with repeated CI runs.
