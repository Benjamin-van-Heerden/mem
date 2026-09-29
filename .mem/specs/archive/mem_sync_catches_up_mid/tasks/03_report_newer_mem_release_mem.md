---
title: Report a newer mem release in mem sync
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:11:25+02:00"
completed_at: "2026-09-29T14:11:25+02:00"
---

Call selfupdate.Latest with a short timeout (skip for dev builds and when MEM_NO_UPDATE=1, as autoUpdate does). When newer, print a line naming the version and 'mem update'. Do not replace the binary mid-session. Test with an httptest release server via MEM_RELEASES_URL, as the selfupdate tests do.

## Completion Notes

newerRelease (internal/cli/update.go) checks selfupdate.Latest (5 s limit) unless this is a dev build or MEM_NO_UPDATE=1, and catchUp adds a ⚠️ nudge naming the newer version and `mem update`; the binary is not replaced mid-session. Verified with an httptest release server: sync with version v0.8.0 reports v0.9.0 and requests only /latest, never a download.
