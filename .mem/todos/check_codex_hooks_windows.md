---
title: Check the Codex hooks on Windows
created_at: "2026-11-11T08:49:23+02:00"
---

The Codex hooks in .codex/hooks.json were verified with Codex 0.164 on macOS only (2026-11-11). On Windows each entry runs its commandWindows (plain mem hook compact / mem hook prompt). Check in a real Codex on Windows that both run and reach the model, which shell runs them, and that a missing mem fails the hook without breaking the session.
