package work

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Revision reads a repository file at some revision; ok is false when the file does not exist there.
type Revision func(path string) (data []byte, ok bool)

// record is the part of a work record's state that changes are described by.
type record struct {
	kind, spec, title, status, owner string
}

// RecordChanges describes, in plain terms, how the specs, tasks and todos at
// the given repository paths differ between two revisions. Records are matched
// by kind and slug, so a spec moving into the archive is one change.
func RecordChanges(paths []string, before, after Revision) []string {
	old, current := records(paths, before), records(paths, after)
	var keys []string
	for key := range old {
		keys = append(keys, key)
	}
	for key := range current {
		if _, ok := old[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var lines []string
	for _, key := range keys {
		a, hadA := old[key]
		b, hasB := current[key]
		if line := describe(a, hadA, b, hasB, old); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func describe(a record, hadA bool, b record, hasB bool, old map[string]record) string {
	r := b
	if !hasB {
		r = a
	}
	switch r.kind {
	case "todo":
		switch {
		case !hadA:
			return fmt.Sprintf("Todo opened: %s", r.title)
		case !hasB:
			return fmt.Sprintf("Todo closed: %s", r.title)
		}
	case "spec":
		switch {
		case !hadA:
			return fmt.Sprintf("Spec drafted: %s", r.title)
		case !hasB:
			return fmt.Sprintf("Spec removed: %s", r.title)
		case a.status == b.status:
		case b.status == SpecActive:
			return fmt.Sprintf("Spec started by %s: %s", b.owner, r.title)
		case b.status == SpecCompleted:
			return fmt.Sprintf("Spec completed: %s", r.title)
		case b.status == SpecAbandoned:
			return fmt.Sprintf("Spec abandoned: %s", r.title)
		}
	case "task":
		_, specExisted := old["spec:"+r.spec]
		switch {
		case !hadA && hasB && specExisted:
			return fmt.Sprintf("Task added to %s: %s", r.spec, r.title)
		case hadA && hasB && a.status != b.status && b.status == TaskCompleted:
			return fmt.Sprintf("Task completed in %s: %s", r.spec, r.title)
		}
	}
	return ""
}

// records reads the records among paths at one revision, keyed by kind and slug.
func records(paths []string, rev Revision) map[string]record {
	out := map[string]record{}
	for _, p := range paths {
		kind, spec, key := classify(p)
		if kind == "" {
			continue
		}
		data, ok := rev(p)
		if !ok {
			continue
		}
		r := record{kind: kind, spec: spec}
		switch kind {
		case "todo":
			var m TodoMeta
			if _, err := parseMarkdown(p, data, &m); err != nil {
				continue
			}
			r.title = m.Title
		case "spec":
			var m SpecMeta
			if _, err := parseMarkdown(p, data, &m); err != nil {
				continue
			}
			r.title, r.status, r.owner = m.Title, m.Status, m.AssignedTo
		case "task":
			var m TaskMeta
			if _, err := parseMarkdown(p, data, &m); err != nil {
				continue
			}
			r.title, r.status = m.Title, m.Status
		}
		out[key] = r
	}
	return out
}

// classify maps a repository path to its record kind, owning spec and key.
func classify(p string) (kind, spec, key string) {
	rest, ok := strings.CutPrefix(path.Clean(strings.ReplaceAll(p, "\\", "/")), ".mem/")
	if !ok || path.Ext(rest) != ".md" {
		return "", "", ""
	}
	if name, ok := strings.CutPrefix(rest, "todos/"); ok && !strings.Contains(name, "/") {
		return "todo", "", "todo:" + strings.TrimSuffix(name, ".md")
	}
	specPath, ok := strings.CutPrefix(rest, "specs/")
	if !ok {
		return "", "", ""
	}
	parts := strings.Split(strings.TrimPrefix(specPath, "archive/"), "/")
	switch {
	case len(parts) == 2 && parts[1] == "spec.md":
		return "spec", parts[0], "spec:" + parts[0]
	case len(parts) == 3 && parts[1] == "tasks":
		return "task", parts[0], "task:" + parts[0] + "/" + strings.TrimSuffix(parts[2], ".md")
	}
	return "", "", ""
}
