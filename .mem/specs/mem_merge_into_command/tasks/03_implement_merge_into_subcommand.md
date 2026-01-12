---
title: Implement merge into subcommand
status: todo
created_at: '2026-01-12T11:09:57.054049'
updated_at: '2026-01-12T11:19:31.089429'
---
Add the 'into' subcommand to merge.py that accepts 'test' or 'main' as target. Implement the full flow: validate current branch, fetch, switch to target, merge, push, back-merge (ff-only), push, switch back. Include clear error messages with recovery recommendations.

## Amendments

Additional details for implementation:

**mem merge into test** (from dev):
1. Check working directory is clean
2. git fetch origin
3. git checkout test && git pull
4. git merge dev (regular merge with merge commit)
5. git push origin test
6. git checkout dev
7. git merge test --ff-only (back-merge)
8. git push origin dev
9. Print success + hint about next step

**mem merge into main** (from dev) - MORE COMPLEX:
- Default behavior (no --force): dry-run mode showing what would happen, then print instruction to use --force
- With --force flag, execute full chain:
  1. Check working directory is clean
  2. git fetch origin
  3. git checkout test && git pull
  4. git checkout main && git pull
  5. git merge test (regular merge with merge commit)
  6. git push origin main
  7. git checkout test
  8. git merge main --ff-only (back-merge)
  9. git push origin test
  10. git checkout dev
  11. git merge test --ff-only (back-merge to sync dev)
  12. git push origin dev
  13. Print success

Use GitPython repo.git.* methods for all git operations.