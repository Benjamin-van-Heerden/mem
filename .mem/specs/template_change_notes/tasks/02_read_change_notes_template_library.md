---
title: Read change notes from the template library
status: todo
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:27:22+02:00"
---

In internal/templates: add a Change type (Template, Name, Title, Body; Key() = template/name) and Library.Changes(names []string) ([]Change, error), reading <template>/changes/*.md for each used template in template order, then name order. Skip templates without changes/, ignore dotfiles and non-.md files, normalise CRLF, take the title from the first-line '# ' heading (the body is the rest, trimmed) or use the name when there is none, and error on an unknown template as Items does. Put it in a focused file (e.g. internal/templates/changes.go) rather than growing library.go. Colocated tests: order, title fallback, CRLF, ignored files, a template without changes/. See the spec's Note format and Library sections.
