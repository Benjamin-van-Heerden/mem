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

// hookCommand is called by the installed Git hooks and the Claude Code compaction hook.
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
	cmd.AddCommand(a.compactHookCommand())
	return cmd
}
