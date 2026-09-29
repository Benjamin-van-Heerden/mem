---
title: Report a newer mem release in mem sync
status: todo
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:05:30+02:00"
---

Call selfupdate.Latest with a short timeout (skip for dev builds and when MEM_NO_UPDATE=1, as autoUpdate does). When newer, print a line naming the version and 'mem update'. Do not replace the binary mid-session. Test with an httptest release server via MEM_RELEASES_URL, as the selfupdate tests do.
