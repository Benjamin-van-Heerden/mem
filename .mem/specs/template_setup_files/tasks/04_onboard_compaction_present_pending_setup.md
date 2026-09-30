---
title: Onboard and compaction present the pending setup
status: todo
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:56:30+02:00"
---

New internal/cli/onboard_setup.go. writeContext starts with a '🏗️ SETUP' section holding .mem/setup.md in full when it exists; contextState records progress (done, total, present). renderOnboardInstruction puts a first step: 'Setup is pending (N of M steps done). Work through .mem/setup.md with the user, in order, before anything else. After each step, when its Done when holds, tick its box and commit. Stop at steps marked (you) and hand them to the user. When every step is ticked, delete .mem/setup.md, commit and push.' When all boxes are ticked: 'Every setup step is ticked: delete .mem/setup.md, commit and push, and tell the user setup is done.' compact.go writeDigest: line 'Setup pending: N of M steps done in .mem/setup.md.' and instruction sentence 'Continue the setup in .mem/setup.md.' Tests in onboard_test.go and compact_test.go, including the absent case printing nothing.
