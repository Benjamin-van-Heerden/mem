package cli

import (
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/importer"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) importCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "import", Short: "Convert a project from another harness"}
	cmd.AddCommand(&cobra.Command{
		Use:   "agent-core",
		Short: "Convert a project that uses the Python coding harness (.agent_core/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Toplevel(cmd.Context(), a.dir)
			if err != nil {
				return err
			}
			sum, err := importer.Import(root, buildinfo.Version)
			if err != nil {
				return err
			}
			if err := ensureIgnored(root, localIgnore); err != nil {
				return err
			}
			hookLines, err := hooks.Sync(cmd.Context(), sum.Project)
			if err != nil {
				return err
			}
			g := sum.Project.Config.Git
			out := cmd.OutOrStdout()
			output.Heading(out, "📦 IMPORTED FROM .agent_core")
			fmt.Fprintf(out, "Project: %s\nBranches: %s → %s → %s (remote %s)\n", sum.Project.Config.Name, g.Development, g.Staging, g.Production, g.Remote)
			output.Section(out, "📄 CONVERTED")
			fmt.Fprintf(out, "AGENTS.md: old harness block replaced by the mem block, %d memories added, other content kept\n", sum.Memories)
			fmt.Fprintf(out, "Docs: %d in .mem/docs/\n", sum.Docs)
			if sum.Structure {
				fmt.Fprintln(out, "Structure doc: codebase_and_structure.md moved to .mem/structure.md")
			}
			fmt.Fprintf(out, "Specs: %d open, %d archived\nTodos: %d\nWork logs: %d\n", sum.Specs, sum.Archived, sum.Todos, sum.Logs)
			if len(sum.Runnables) > 0 {
				fmt.Fprintf(out, "Runnables: %s in .mem/runnables/, from the old files, tree_dirs and runnables settings\n", strings.Join(sum.Runnables, ", "))
			}
			for _, line := range append(hookLines, sum.Notes...) {
				fmt.Fprintln(out, line)
			}
			output.Instruction(out,
				"Nothing was committed, and .agent_core/ is still in place.",
				"1. Show the user the result: `git status`, AGENTS.md and .mem/.",
				"2. Once the user is happy, remove the old harness with `git rm -r -q .agent_core`, and drop its entries from .gitignore.",
				fmt.Sprintf("3. Commit everything on %s and push it.", g.Development),
				"4. Run `mem onboard`.",
			)
			return nil
		},
	})
	return cmd
}
