package cli

import (
	"fmt"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) memoryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "memory", Short: "Manage project conventions in AGENTS.md"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List project memories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			text, err := readAgents(p)
			if err != nil {
				return err
			}
			memories, err := agentsmd.Memories(text)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "💾 PROJECT MEMORIES")
			if len(memories) == 0 {
				fmt.Fprintln(out, "No memories recorded.")
			}
			for _, m := range memories {
				fmt.Fprintf(out, "## %s\n%s\n\n", m.Name, m.Body)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "set <name> <content>",
		Short: "Add or replace a project memory",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			text, err := readAgents(p)
			if err != nil {
				return err
			}
			updated, err := agentsmd.SetMemory(text, args[0], args[1])
			if err != nil {
				return err
			}
			if err := writeAgents(p, updated); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "💾 MEMORY SAVED")
			fmt.Fprintf(out, "## %s\n%s\n", args[0], args[1])
			output.Instruction(out, "Apply this convention for the rest of this session. AGENTS.md is updated for future sessions; commit it with your work.")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a project memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			text, err := readAgents(p)
			if err != nil {
				return err
			}
			updated, err := agentsmd.RemoveMemory(text, args[0])
			if err != nil {
				return err
			}
			if err := writeAgents(p, updated); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "💾 MEMORY REMOVED")
			fmt.Fprintf(out, "Removed %s. It no longer applies, including for the rest of this session.\n", args[0])
			return nil
		},
	})
	return cmd
}
