package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/spf13/cobra"
)

const maxListedChanges = 40

func (a *app) structureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "structure",
		Short: "Create the codebase structure doc, or show what changed since it was last updated",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			d, err := structure.Measure(ctx, p)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case d.Missing:
				if err := os.MkdirAll(filepath.Dir(structure.Path(p)), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(structure.Path(p), []byte(structure.Template), 0o644); err != nil {
					return err
				}
				tree, err := structure.Tree(ctx, p)
				if err != nil {
					return err
				}
				output.Section(out, "🗺️ STRUCTURE DOC CREATED")
				fmt.Fprintf(out, "File: %s\n", structure.RelPath)
				output.Section(out, "🌲 FILE TREE")
				fmt.Fprintln(out, tree)
				output.Instruction(out,
					fmt.Sprintf("Research the codebase and replace every placeholder in %s. It is a factual map of the repository that every future session reads at onboard.", structure.RelPath),
					"",
					"1. Orient yourself with the file tree above. Read the manifests and configuration that exist (package.json, pyproject.toml, go.mod, Cargo.toml, Makefile, docker-compose.yml, README files, env examples) and the primary entry points.",
					"2. Explore the architecture methodically: routers or controllers, core logic, data models and schemas, storage access, configuration and widely imported shared code. Follow imports and data flow; do not read every file.",
					"3. Write the doc. Use actual paths, module, command and function names. Only state facts confirmed from the code or durable project docs. Leave out goals, motivation and future direction. Omit sections that do not apply, and keep it well under 5000 words.",
					"4. Show the user what you wrote and ask whether it is accurate, then commit it.",
				)
			case d.Editing:
				output.Section(out, "🗺️ STRUCTURE DOC")
				fmt.Fprintf(out, "%s has uncommitted edits, so it counts as up to date.\n", structure.RelPath)
				output.Instruction(out, "Finish the update and commit it with your work. That commit becomes the baseline for measuring future changes.")
			default:
				output.Section(out, "🗺️ STRUCTURE DOC")
				fmt.Fprintf(out, "Last updated in commit %s.\nCode changed since then: %d file(s), %d line(s).\n", d.Baseline[:7], len(d.Changes), d.Lines)
				if len(d.Changes) == 0 {
					fmt.Fprintln(out, "Nothing to update.")
					return nil
				}
				fmt.Fprintln(out)
				for i, c := range d.Changes {
					if i == maxListedChanges {
						fmt.Fprintf(out, "  ... and %d more\n", len(d.Changes)-maxListedChanges)
						break
					}
					fmt.Fprintf(out, "  %s (%d lines)\n", c.Path, c.Lines)
				}
				lines := []string{
					fmt.Sprintf("1. Read %s.", structure.RelPath),
					fmt.Sprintf("2. Inspect the changes listed above, e.g. `git diff %s -- <file>`, and read new files.", d.Baseline[:7]),
					"3. Update only the sections these changes affect. Keep it factual and concise; leave unaffected sections alone.",
					"4. Commit it with your work. That commit becomes the new baseline.",
				}
				if !d.Stale() {
					lines = append([]string{"The doc is still broadly current. Update it only where these changes affect what it describes:", ""}, lines...)
				}
				output.Instruction(out, lines...)
			}
			return nil
		},
	}
}
