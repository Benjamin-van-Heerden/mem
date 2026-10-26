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
			lines := []string{"Apply this convention for the rest of this session. AGENTS.md is updated for future sessions; commit it with your work."}
			if warning := agentsSizeWarning(updated); warning != "" {
				fmt.Fprintf(out, "\n%s\n", warning)
				lines = append(lines, "Stop here: tell the user about the ⚠️ above and settle it with them before you continue with other work.")
			}
			output.Instruction(out, lines...)
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

// agentsSizeLimit is the most of AGENTS.md that Codex reads by default (project_doc_max_bytes).
const agentsSizeLimit = 32 << 10

// agentsSizeWarning flags an AGENTS.md that agents may read only in part.
// Memories and the project's own instructions follow mem's block, so a cut drops them first.
func agentsSizeWarning(text string) string {
	if len(text) <= agentsSizeLimit {
		return ""
	}
	return fmt.Sprintf("⚠️ AGENTS.md is %.1f KiB, over 32 KiB. Some agent runtimes, such as Codex, read only the first 32 KiB of it and silently drop the rest: the project memories and instructions at the end of the file. Propose to the user which memories to merge, shorten or retire (`mem memory list` shows them all), and change them only as the user decides. For Codex alone, the project can instead raise `project_doc_max_bytes` in `.codex/config.toml`.", float64(len(text))/1024)
}
