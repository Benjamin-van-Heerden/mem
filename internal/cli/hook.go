package cli

import (
	"github.com/Benjamin-van-Heerden/memr/internal/hooks"
	"github.com/spf13/cobra"
)

// hookCommand is called by the installed Git hooks.
func (a *app) hookCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "hook", Short: "Checks run by memr's Git hooks", Hidden: true}
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
	return cmd
}
