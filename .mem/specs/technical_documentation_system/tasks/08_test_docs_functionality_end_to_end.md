---
title: Test docs functionality end-to-end
status: todo
created_at: '2026-01-11T12:42:19.108167'
updated_at: '2026-01-11T12:42:19.108167'
completed_at: null
---
Manual testing checklist:
1. Run mem init - verify env var warnings appear if missing
2. Create .mem/docs/test_doc.md with some markdown content
3. Run mem docs list - verify doc shows as unindexed
4. Run mem docs index - verify indexing completes, summary generated
5. Run mem docs list - verify doc shows as indexed with summary preview
6. Run mem docs read test_doc - verify full content displayed
7. Run mem docs search 'some query' - verify results returned with source info
8. Run mem onboard - verify docs section appears with summary
9. Modify test_doc.md, run mem docs index - verify re-indexing
10. Run mem docs delete test_doc - verify cleanup
11. Test worktree symlinks by creating a spec and checking .mem/docs/data/ is linked