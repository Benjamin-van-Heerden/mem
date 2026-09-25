---
title: mem template promote and reset
status: todo
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-25T14:00:22+02:00"
---

promote <memory|skill|doc> <name> [--to <template>]: copy the project item into the library clone, commit 'Promote <kind> <name> from <project>', push, record the hash in the lock. Require --to when the item is untracked and the project uses more than one template. reset <kind> <name>: overwrite the local item with the template copy and record the hash. Test against a bare library repository.
