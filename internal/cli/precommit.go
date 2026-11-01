package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

func (a *app) precommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "precommit",
		Short: "Run the project's checks, such as its linter and tests",
		Long:  "Runs each command in [precommit] commands of .mem/config.toml from the repository root, in order, and stops at the first failure. `mem promote staging` and `mem deploy` run the same checks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(p.Config.Precommit.Commands) == 0 {
				output.Section(out, "✅ PRECOMMIT")
				fmt.Fprintln(out, "No checks are set for this project.")
				output.Instruction(out,
					"Add the project's linter, type checker and tests to .mem/config.toml, then run `mem precommit` again:",
					"",
					"[precommit]",
					"commands = ['<lint command>', '<test command>']",
				)
				return nil
			}
			return runChecks(cmd.Context(), out, p)
		},
	}
}

// runChecks runs the project's precommit commands and prints a line for each. A project without any passes.
func runChecks(ctx context.Context, out io.Writer, p project.Project) error {
	commands := p.Config.Precommit.Commands
	if len(commands) == 0 {
		return nil
	}
	output.Section(out, "✅ PRECOMMIT")
	for _, command := range commands {
		start := time.Now()
		cmd := shell(ctx, command)
		cmd.Dir = p.Root
		result, err := cmd.CombinedOutput()
		elapsed := time.Since(start).Round(100 * time.Millisecond)
		if err != nil {
			fmt.Fprintf(out, "✘ %s (%s)\n\n%s\n", command, elapsed, result)
			output.Instruction(out, "Fix what failed and run `mem precommit` again. Do not commit or promote until it passes.")
			return fmt.Errorf("precommit failed: %s", command)
		}
		fmt.Fprintf(out, "✔ %s (%s)\n", command, elapsed)
	}
	return nil
}

func shell(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
