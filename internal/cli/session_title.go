package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Benjamin-van-Heerden/mem/internal/claude"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

// renameSession asks the agent to title its session after the project and the work, so sessions listed
// together across projects can be told apart. It names the desktop app because that is where an agent can
// rename its own session.
func renameSession(name, when string) string {
	return fmt.Sprintf("If you can rename this session (as in the Claude desktop app), title it \"%s · <the work>\" %s.", name, when)
}

func sessionsDir(p project.Project) string { return p.Path("local", "sessions") }

// markOnboarded lets the session's next prompt carry the title hint. Only Claude Code sessions are marked,
// and only where mem's prompt hook is installed to deliver it.
func markOnboarded(p project.Project) {
	session := os.Getenv(claude.SessionEnv)
	if session == "" || !p.Config.Claude.CompactHookEnabled() {
		return
	}
	_ = claude.MarkOnboarded(sessionsDir(p), session)
}

// promptHookCommand runs on every prompt through Claude Code's UserPromptSubmit hook; its output joins the
// agent's context. It speaks once, on the first prompt after an onboard, and never fails the prompt.
func (a *app) promptHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "prompt",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var input struct {
				SessionID string `json:"session_id"`
			}
			if json.NewDecoder(cmd.InOrStdin()).Decode(&input) != nil {
				return nil
			}
			p, err := a.project(cmd)
			if err != nil || !claude.TakeTitleHint(sessionsDir(p), input.SessionID) {
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "mem: this is the first request since onboarding. "+renameSession(p.Config.Name, "for the work it asks for, unless the title already says so"))
			return nil
		},
	}
}
