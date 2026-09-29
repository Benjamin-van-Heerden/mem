package cli

import (
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
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
	cmd.AddCommand(a.compactHookCommand())
	return cmd
}
