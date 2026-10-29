package templates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Change is a template's change note, <template>/changes/<name>.md: a change to the template that projects
// built from an earlier state of it need to act on. It is not installed; a sync turns it into a todo once.
type Change struct {
	Template string
	Name     string
	Title    string
	Body     string
}

func (c Change) Key() string { return c.Template + "/" + c.Name }

const changesDir = "changes"

// Changes lists the change notes of the named templates, in template order and then by name. Names are
// dated, so name order is the order they were written in. Unknown templates are rejected by Items.
func (l Library) Changes(names []string) ([]Change, error) {
	var changes []Change
	for _, template := range names {
		dir := filepath.Join(l.Dir, template, changesDir)
		entries, err := os.ReadDir(dir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			name, ok := itemName(Doc, e)
			if !ok {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			title, body := parseChange(name, data)
			changes = append(changes, Change{Template: template, Name: name, Title: title, Body: body})
		}
	}
	return changes, nil
}

// parseChange takes the title from a first-line `# ` heading, or falls back to the note's name.
func parseChange(name string, data []byte) (title, body string) {
	text := strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n"))
	first, rest, _ := strings.Cut(text, "\n")
	if heading, ok := strings.CutPrefix(first, "# "); ok {
		return strings.TrimSpace(heading), strings.TrimSpace(rest)
	}
	return name, text
}
