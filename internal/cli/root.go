package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

type app struct {
	dir string
}

func New() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:           "mem",
		Short:         "Context building and work records for coding agents",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&a.dir, "dir", ".", "Run against the repository containing this directory")
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(
		versionCommand(),
		a.initCommand(),
		a.memoryCommand(),
		a.specCommand(),
		a.taskCommand(),
		a.todoCommand(),
		a.logCommand(),
		a.onboardCommand(),
		a.syncCommand(),
		a.structureCommand(),
		a.promoteCommand(),
		a.hookCommand(),
		a.importCommand(),
	)
	return root
}

func (a *app) project(cmd *cobra.Command) (project.Project, error) {
	return project.Load(cmd.Context(), a.dir)
}

func (a *app) user(cmd *cobra.Command, p project.Project) (string, error) {
	return project.User(cmd.Context(), p.Root)
}

// publish commits and pushes mem-owned paths so teammates see the change, and describes the outcome.
func publish(ctx context.Context, p project.Project, message string, paths ...string) string {
	pushed, err := git.CommitPaths(ctx, p.Root, message, paths...)
	switch {
	case err != nil:
		return fmt.Sprintf("Saved locally but could not commit and push (%v). Tell the user that teammates will not see this until it is committed and pushed.", err)
	case pushed:
		return "Committed and pushed: " + message
	default:
		return "Committed: " + message + " (this branch has no upstream, so nothing was pushed)"
	}
}

func table(w io.Writer, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func readAgents(p project.Project) (string, error) {
	data, err := os.ReadFile(filepath.Join(p.Root, "AGENTS.md"))
	return string(data), err
}

func writeAgents(p project.Project, text string) error {
	return os.WriteFile(filepath.Join(p.Root, "AGENTS.md"), []byte(text), 0o644)
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the mem version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "mem %s (%s, schema %d)\n", buildinfo.Version, buildinfo.Commit, project.Schema)
		},
	}
}
