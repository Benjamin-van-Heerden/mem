package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
)

// setupState describes the template setup left in .mem/setup.md, if any.
type setupState struct {
	present     bool
	done, total int
}

func (s setupState) finished() bool { return s.present && s.total > 0 && s.done == s.total }

func readSetup(p project.Project) (string, setupState, error) {
	data, err := os.ReadFile(filepath.Join(p.Root, templates.SetupPath))
	if errors.Is(err, os.ErrNotExist) {
		return "", setupState{}, nil
	}
	if err != nil {
		return "", setupState{}, err
	}
	done, total := templates.SetupProgress(string(data))
	return string(data), setupState{present: true, done: done, total: total}, nil
}

func writeSetupSection(out io.Writer, text string, s setupState) {
	output.Section(out, "🏗️ SETUP")
	fmt.Fprintf(out, "From %s, %d of %d steps done.\n\n", templates.SetupPath, s.done, s.total)
	fmt.Fprint(out, text)
}

// setupInstruction is the onboard step for a pending or finished setup.
func setupInstruction(s setupState) string {
	if s.finished() {
		return fmt.Sprintf("Every step in %s is ticked: delete the file, commit and push, and tell the user the setup is done.", templates.SetupPath)
	}
	return fmt.Sprintf("Setup is pending (%d of %d steps done). Work through %s under 🏗️ SETUP with the user, in order, before anything else: start with the first unticked step now unless the user says otherwise. After each step, when its \"Done when\" holds, tick its box and commit. Steps marked (you) need the user: ask for what the step names, then continue. When every step is ticked, delete %s, commit and push.", s.done, s.total, templates.SetupPath, templates.SetupPath)
}

// setupDigestLine is the compaction digest's reminder of a pending setup.
func setupDigestLine(s setupState) string {
	return fmt.Sprintf("Setup pending: %d of %d steps done in %s.", s.done, s.total, templates.SetupPath)
}
