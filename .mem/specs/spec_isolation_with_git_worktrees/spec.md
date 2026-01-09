---
title: Spec isolation with git worktrees
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 25
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/25
branch: dev-benjamin_van_heerden-spec_isolation_with_git_worktrees
pr_url: null
created_at: '2026-01-09T13:33:59.872799'
updated_at: '2026-01-09T15:22:33.739693'
completed_at: null
last_synced_at: '2026-01-09T13:36:05.249356'
local_content_hash: 961b3fa1877241d8768159de00eaf7b34e8ca098f59be7c66781c597ce12a130
remote_content_hash: 961b3fa1877241d8768159de00eaf7b34e8ca098f59be7c66781c597ce12a130
---
## Overview

Use git worktrees to isolate spec work into separate directories. Each spec gets its own worktree with a dedicated feature branch, enabling parallel work on multiple specs with separate agent sessions.

## Goals

- Enable parallel spec work with multiple agents (each agent in its own worktree)
- Centralized coordination from the main repo (merging, syncing, onboarding)
- Eliminate the need for `mem spec activate/deactivate` - being in a worktree means the spec is active
- Clean separation between specs without branch switching chaos

## Technical Approach

### Directory structure

```
/path/to/project/                    # Main repo, stays on 'dev'
/path/to/project-worktrees/          # Sibling directory for worktrees
  ├── user_auth/                     # Worktree on 'dev-ben-user_auth' branch
  └── fix_sync/                      # Worktree on 'dev-ben-fix_sync' branch
```

### Command changes

1. **`mem spec new "title"`** - Also creates a git worktree + feature branch. The worktree is created in a sibling directory (`../<project>-worktrees/<slug>/`). Outputs the path so user can `cd` there or start a new agent session.

2. **Remove `mem spec activate/deactivate`** - No longer needed. The active spec is determined by which worktree you're in (detected via `.git` file pointing to main repo).

3. **`mem spec show`** - When run from a worktree, shows that spec. When run from main repo, shows all specs or requires a slug.

4. **`mem spec complete <slug>`** - Can be run from main repo or from the worktree. Commits, pushes, creates PR, then removes the worktree (keeps the branch for the PR).

5. **`mem merge`** - Run from main repo. Merges completed PRs, deletes merged branches.

### Detection logic

Detect if we're in a worktree by checking if `.git` is a file (not a directory). If it's a file, it contains a path to the actual git dir, which tells us the main repo location.

```python
# In a worktree, .git is a file containing: "gitdir: /path/to/main/.git/worktrees/<name>"
# In main repo, .git is a directory
```

### Workflow

1. User runs `mem spec new "feature"` from main repo
2. Worktree created at `../project-worktrees/feature/` with branch `dev-user-feature`
3. User opens new terminal/agent session in that directory
4. Work happens there, commits go to the feature branch
5. When done, run `mem spec complete feature` (from anywhere)
6. PR created, worktree removed, branch remains for PR review
7. After merge, `mem merge` cleans up the branch

## Success Criteria

- Can create a spec and have worktree + branch created automatically
- Can work in worktree with full `mem` functionality (tasks, logs, etc.)
- Can complete spec and have PR created, worktree cleaned up
- Main repo stays on dev, can coordinate multiple specs
- `mem onboard` from main repo shows all active specs with their worktree paths

## Notes

- Agents can't persist `cd` between commands, so each spec needs its own agent session in its worktree directory
- This enables true parallel development with multiple agents
- The worktrees directory is a sibling to avoid cluttering the project
