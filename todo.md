- mem onboard should show work logs + more descriptive messages on what mem is and how it works
- README.md is no longer accurate for what mem is
- mem spec complete does not work as expected
❯ mem spec complete test_spec "it is done"
Completing spec: Test Spec...
Committing and pushing changes...
Creating Pull Request...
Created Pull Request: https://github.com/Benjamin-van-Heerden/mem/pull/2
Warning: Could not switch to 'dev' branch: Failed to switch to branch 'dev': Cmd('git') failed due to: exit code(1)
  cmdline: git switch dev
  stderr: 'error: Your local changes to the following files would be overwritten by checkout:
	.mem/specs/test_spec/spec.md
Please commit your changes or stash them before you switch branches.
Aborting'

Spec 'test_spec' marked as MERGE READY.
PR: https://github.com/Benjamin-van-Heerden/mem/pull/2

Next steps: Merge the PR on GitHub.
(should happen with less friction, i.e. commit + push, I think maybe the order of things is wrong)
- mem merge command, which shows a list of pr's and lets you merge them in editor.
- sanitize tests (one would hope that our large test suite would have caught an issue like this)
