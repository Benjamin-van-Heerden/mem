---
title: Library access and project configuration
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-25T14:30:48+02:00"
completed_at: "2026-09-25T14:30:48+02:00"
---

Add [templates] (source, use, exclude) to project.Config and a user-level config read from os.UserConfigDir()/mem/config.toml (template_source). Create internal/templates with: cloning a library URL into os.UserCacheDir()/mem/templates/<slug>, updating it with git pull --ff-only under the converge fetch time limit, discovering templates by <name>/template.toml, and reading each template's memories/, skills/ and docs/. Tests use a local bare repository as the library. See the spec's Template library and Project configuration sections.

## Completion Notes

Added [templates] (source, use, exclude) to project.Config and internal/templates: Open clones a library into the user cache and fast-forwards it (a failed pull becomes a warning), Templates discovers <name>/template.toml, Items lists memories/skills/docs with later templates overriding earlier ones, DefaultSource reads template_source from ~/.config/mem/config.toml (or XDG_CONFIG_HOME). Verified by TestLibraryClonesPullsAndLetsLaterTemplatesOverride against a bare library repo; go vet clean.
