package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

// hookCommand is called by the installed Git hooks and mem's Claude Code and Cursor hooks.
func (a *app) hookCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "hook", Short: "Commands run by mem's Git, Claude Code and Cursor hooks", Hidden: true}
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
	cmd.AddCommand(&cobra.Command{
		Use:  "stop",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cursor runs this when an agent ends its turn and submits a non-empty followup_message as the next user
			// message. It always answers with a JSON object and never fails the turn.
			reply := map[string]string{}
			if line := a.stopFollowUp(cmd); line != "" {
				reply["followup_message"] = "mem: " + line
			}
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			if err := enc.Encode(reply); err != nil {
				buf.Reset()
				buf.WriteString("{}\n")
			}
			_, err := cmd.OutOrStdout().Write(buf.Bytes())
			return err
		},
	})
	cmd.AddCommand(a.compactHookCommand())
	return cmd
}

// stopNudgePath records the work summary mem last sent a stop follow-up for, so an agent that put the log off is not
// asked again until there is more work.
const stopNudgePath = "local/stop_nudge"

// stopFollowUp is the work-log follow-up for a Cursor agent that completed its turn with a log due, or "". It never
// answers its own follow-up (loop_count above 0) or a turn the user stopped.
func (a *app) stopFollowUp(cmd *cobra.Command) string {
	var in struct {
		Status    string `json:"status"`
		LoopCount int    `json:"loop_count"`
	}
	if json.NewDecoder(cmd.InOrStdin()).Decode(&in) != nil || in.Status != "completed" || in.LoopCount != 0 {
		return ""
	}
	p, err := a.project(cmd)
	if err != nil {
		return ""
	}
	user, err := project.User(cmd.Context(), p.Root)
	if err != nil {
		return ""
	}
	w, err := checkpoint.Since(cmd.Context(), p, user, git.UserEmail(cmd.Context(), p.Root))
	if err != nil {
		return ""
	}
	marker := p.Path(filepath.FromSlash(stopNudgePath))
	line := checkpoint.TurnEndLine(w)
	if line == "" {
		os.Remove(marker)
		return ""
	}
	if last, err := os.ReadFile(marker); err == nil && string(last) == w.Summary() {
		return ""
	}
	// Losing the record only means the next completed turn is asked again.
	if os.MkdirAll(filepath.Dir(marker), 0o755) == nil {
		os.WriteFile(marker, []byte(w.Summary()), 0o644)
	}
	return line
}
