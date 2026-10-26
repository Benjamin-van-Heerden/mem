---
title: Retry a push the remote rejected because it moved on
status: todo
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T12:58:18+02:00"
---

git.Push: make a rejection because the remote has new commits distinguishable (e.g. ErrRemoteAhead, still matching ErrPush). converge.Push: on that error, fetch, run the same catch-up as Sync (update: fast-forward or rebase local commits), and push again, up to 3 attempts; stop with the existing nudge if the catch-up conflicts or leaves the branch behind. Report the catch-up and the retried push in Done. Test in internal/converge (or cli sync tests): a teammate pushes between our fetch and push (simulate by pushing from a second clone after inspect, or by calling Push with a stale report), and the commits end up pushed; a conflicting teammate change ends with a nudge and the local commits intact.
