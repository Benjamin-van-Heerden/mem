---
title: Setup files in the template library
status: todo
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:56:30+02:00"
---

New internal/templates/setup.go: SetupPath = ".mem/setup.md"; Library.Setup(names) reads <template>/setup.md for each named template that has one, in order, joined by a blank line (empty when none); SetupProgress(text) (done, total) counts headings matching ^## \[( |x|X)\] . Template gains HasSetup, set in Library.Templates(). Tests in internal/templates for ordering, templates without setup, and progress counting.
