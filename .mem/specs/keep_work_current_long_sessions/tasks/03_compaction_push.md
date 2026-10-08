---
title: Compaction push
status: todo
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:07:22+02:00"
---

The compaction hook (mem hook compact) runs converge.Push after its catch-up under the same conditions as mem sync: fetch succeeded, branch ahead and not behind, not staging or production; uncommitted work is never touched; push failures become digest lines and the hook still exits 0. Update docs/design.md ('never commits or pushes' becomes 'never commits; pushes committed work when safe'), docs/status.md and the structure doc. Test: a compaction catch-up pushes an ahead branch and reports it.
