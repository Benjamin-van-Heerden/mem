---
title: Add GitHub issue creation for migrated specs
status: todo
subtasks: []
created_at: '2026-01-10T15:12:40.457341'
updated_at: '2026-01-10T15:12:40.457341'
completed_at: null
---
After creating each spec file, create a GitHub issue and immediately close it. Use src/utils/github/client.py for get_github_client(), src/utils/github/repo.py for get_repo_from_git(), and src/utils/github/api.py for create_github_issue() and close_issue_with_comment(). Labels should be ['mem-spec', 'mem-status:completed']. Close with comment 'Migrated from legacy agent_rules system'. Update spec frontmatter with issue_id and issue_url after creation.