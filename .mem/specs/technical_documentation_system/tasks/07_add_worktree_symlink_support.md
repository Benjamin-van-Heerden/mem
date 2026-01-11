---
title: Add worktree symlink support
status: todo
created_at: '2026-01-11T12:42:11.623592'
updated_at: '2026-01-11T12:42:11.623592'
completed_at: null
---
Modify src/commands/spec.py assign command:
- After worktree creation, read worktree.symlink_paths from config
- For each path in symlink_paths:
  - Check if path exists in main repo
  - If exists, create symlink in worktree pointing to main repo path
  - Handle errors gracefully (log warning, continue)
- Example: .mem/docs/data/ in main repo -> symlink in worktree

Update src/templates/config.toml:
- Add [worktree] section with symlink_paths example:
  [worktree]
  # Paths to symlink from main repo into worktrees
  # Useful for shared data like vector DBs, node_modules, .venv
  # symlink_paths = [".mem/docs/data"]