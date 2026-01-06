---
title: Add GitHub PR query function
status: completed
subtasks: []
created_at: '2026-01-06T15:12:55.915501'
updated_at: '2026-01-06T15:14:47.712856'
completed_at: '2026-01-06T15:14:47.712849'
---
Create a function to query GitHub for open PRs ready to merge.

The function should:
1. Query GitHub API for open PRs where base branch is 'dev'
2. Filter to PRs with '[Complete]:' in title (our convention)
3. Return list of PRs with: number, title, author, linked issue, checks status

Location: src/utils/github/api.py

Function signature:
def list_merge_ready_prs(repo) -> list[dict]:
    '''List open PRs targeting dev that are ready to merge.'''

Each returned dict should contain:
- number: PR number
- title: PR title (with [Complete]: prefix stripped for display)
- author: GitHub username
- issue_number: Linked issue number (extracted from 'Closes #X' in body)
- checks_passing: bool
- html_url: Link to PR
- head_branch: Branch name to delete after merge

## Completion Notes

Added list_merge_ready_prs() function that queries open PRs targeting dev with [Complete]: in title. Returns PR number, title, author, linked issue, checks status, mergeable state, URL, and head branch.