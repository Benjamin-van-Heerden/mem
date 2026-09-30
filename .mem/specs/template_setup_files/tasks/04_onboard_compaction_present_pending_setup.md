---
title: Onboard and compaction present the pending setup
status: completed
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T12:01:12+02:00"
completed_at: "2026-09-30T12:01:12+02:00"
---

New internal/cli/onboard_setup.go. writeContext starts with a '🏗️ SETUP' section holding .mem/setup.md in full when it exists; contextState records progress (done, total, present). renderOnboardInstruction puts a first step: 'Setup is pending (N of M steps done). Work through .mem/setup.md with the user, in order, before anything else. After each step, when its Done when holds, tick its box and commit. Stop at steps marked (you) and hand them to the user. When every step is ticked, delete .mem/setup.md, commit and push.' When all boxes are ticked: 'Every setup step is ticked: delete .mem/setup.md, commit and push, and tell the user setup is done.' compact.go writeDigest: line 'Setup pending: N of M steps done in .mem/setup.md.' and instruction sentence 'Continue the setup in .mem/setup.md.' Tests in onboard_test.go and compact_test.go, including the absent case printing nothing.

## Completion Notes

internal/cli/onboard_setup.go: readSetup (setupState present/done/total via templates.SetupProgress), the 🏗️ SETUP section (progress line plus the file) written first in writeContext, setupInstruction, setupDigestLine. renderOnboardInstruction puts the setup step after knowledge/nudge/template-warning steps and, while steps remain, ends there (no summary or 'ask how to proceed'); with every box ticked it tells the agent to delete, commit and push and then continues normally. The compaction digest prints 'Setup pending: N of M steps done in .mem/setup.md.' and adds 'Continue the setup in .mem/setup.md.'. onboard_setup_test.go covers pending, finished and deleted for onboard and the digest; onboard/compact tests and go vet pass; the output was checked by eye with dist/mem-dev.
