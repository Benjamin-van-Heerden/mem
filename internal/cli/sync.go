package cli

import (
	"github.com/Benjamin-van-Heerden/memr/internal/converge"
	"github.com/Benjamin-van-Heerden/memr/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) syncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Fetch and bring this checkout up to date with the shared codebase",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			r := converge.Sync(cmd.Context(), p)
			out := cmd.OutOrStdout()
			renderReport(out, r, false)
			var lines []string
			if len(r.Done) > 0 {
				lines = append(lines, "Incoming commits were applied. Review the ones relevant to your current work before continuing, and re-read AGENTS.md if it changed.")
			}
			if len(r.Nudges) > 0 {
				lines = append(lines, "Tell the user about each ⚠️ item above.")
			}
			if len(lines) > 0 {
				output.Instruction(out, lines...)
			}
			return nil
		},
	}
}
