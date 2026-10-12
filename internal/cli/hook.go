package cli

import (
	"fmt"
	"os"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

// hookCommand is called by the installed Git hooks and mem's Claude Code hooks.
func (a *app) hookCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "hook", Short: "Commands run by mem's Git and Claude Code hooks", Hidden: true}
	cmd.AddCommand(&cobra.Command{
		Use:  "pre-push <remote> <url>",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil || !p.Config.Git.Protect {
				return nil
			}
			return hooks.PrePush(p, args[0], cmd.InOrStdin())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:  "pre-commit",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil || !p.Config.Git.Protect {
				return nil
			}
			return hooks.PreCommit(cmd.Context(), p)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:  "post-commit",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// mem's own commits and rebases are not checkpoints to report on, and a nudge never fails a commit.
			if os.Getenv(git.InternalEnv) != "" {
				return nil
			}
			p, err := a.project(cmd)
			if err != nil {
				return nil
			}
			user, err := project.User(cmd.Context(), p.Root)
			if err != nil {
				return nil
			}
			for _, line := range checkpoint.CommitLines(cmd.Context(), p, user, git.UserEmail(cmd.Context(), p.Root)) {
				fmt.Fprintln(cmd.OutOrStdout(), "mem: "+line)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:  "prompt",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Claude Code adds this output to the agent's context with the user's message; it stays silent until a
			// work log is called for and never fails the prompt.
			p, err := a.project(cmd)
			if err != nil {
				return nil
			}
			user, err := project.User(cmd.Context(), p.Root)
			if err != nil {
				return nil
			}
			w, err := checkpoint.Since(cmd.Context(), p, user, git.UserEmail(cmd.Context(), p.Root))
			if line := checkpoint.TurnStartLine(w); err == nil && line != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "mem: "+line)
			}
			return nil
		},
	})
	cmd.AddCommand(a.compactHookCommand())
	return cmd
}
