package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/shell"
	"github.com/spf13/cobra"
)

func (a *app) checkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run the project's checks, such as its linter and tests",
		Long:  "Runs each command in [checks] commands of .mem/config.toml from the repository root, in order, and stops at the first failure. `mem promote staging` and `mem deploy` run the same checks and refuse when they fail; in a single-branch project, which has no promotion, `mem task complete` and `mem spec complete` do.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(p.Config.Checks.Commands) == 0 {
				output.Section(out, "✅ CHECKS")
				fmt.Fprintln(out, "No checks are set for this project.")
				output.Instruction(out,
					"Add the project's linter, type checker and tests to .mem/config.toml, then run `mem check` again:",
					"",
					"[checks]",
					"commands = ['<lint command>', '<test command>']",
				)
				return nil
			}
			return runChecks(cmd.Context(), out, p, "run `mem check` again")
		},
	}
}

// runChecks runs the project's checks and prints a line for each; retry says what to run again once a failure is fixed.
// A project without any passes.
func runChecks(ctx context.Context, out io.Writer, p project.Project, retry string) error {
	commands := p.Config.Checks.Commands
	if len(commands) == 0 {
		return nil
	}
	output.Section(out, "✅ CHECKS")
	for _, command := range commands {
		start := time.Now()
		cmd := shell.Command(ctx, command)
		cmd.Dir = p.Root
		result, err := cmd.CombinedOutput()
		elapsed := time.Since(start).Round(100 * time.Millisecond)
		if err != nil {
			fmt.Fprintf(out, "✘ %s (%s)\n\n%s\n", command, elapsed, result)
			output.Instruction(out, fmt.Sprintf("Fix what failed, then %s. `mem check` runs the checks on their own.", retry))
			return fmt.Errorf("checks failed: %s", command)
		}
		fmt.Fprintf(out, "✔ %s (%s)\n", command, elapsed)
	}
	return nil
}
