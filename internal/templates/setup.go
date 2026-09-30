package templates

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SetupPath is where `mem init` puts a template's one-time setup. The file is the
// setup's only state: its checkboxes are the progress and deleting it ends the setup.
const SetupPath = ".mem/setup.md"

const setupFile = "setup.md"

var setupStep = regexp.MustCompile(`(?m)^## \[([ xX])\] `)

// Setup joins the setup files of the named templates, in order. It is empty when
// none of them has one.
func (l Library) Setup(names []string) (string, error) {
	var parts []string
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(l.Dir, name, setupFile))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		parts = append(parts, strings.TrimSpace(string(data)))
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, "\n\n") + "\n", nil
}

// SetupProgress counts the setup's step headings (`## [ ] …`) and how many are ticked.
func SetupProgress(text string) (done, total int) {
	for _, m := range setupStep.FindAllStringSubmatch(text, -1) {
		total++
		if m[1] != " " {
			done++
		}
	}
	return done, total
}
