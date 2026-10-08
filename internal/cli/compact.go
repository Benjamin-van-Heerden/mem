package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

// digestCommitLimit and digestRecordLimit keep the compaction digest small; `mem sync` has the full report.
const (
	digestCommitLimit = 5
	digestRecordLimit = 10
)

// compactHookCommand runs after Claude Code compacts a conversation. Its stdout
// joins the agent's context, next to the compaction summary: it brings the
// checkout up to date, pushes committed work when that is safe, and adds the
// shared state the summary cannot know.
func (a *app) compactHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:  "compact",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			c, err := a.catchUp(cmd)
			if err != nil {
				// Never disturb compaction; a failed catch-up is one line of context.
				fmt.Fprintf(out, "mem could not sync this checkout after compaction (%v). Run `mem sync` when convenient.\n", err)
				return nil
			}
			// A compaction is a natural sync point in a long session; Push only acts when mem sync would.
			c.report = converge.Push(cmd.Context(), c.p, c.report)
			user, err := a.user(cmd, c.p)
			if err != nil {
				user = ""
			}
			var sinceLog int
			if user != "" {
				sinceLog, _, _ = checkpoint.WorkSinceLog(cmd.Context(), c.p.Root, user, git.UserEmail(cmd.Context(), c.p.Root))
			}
			writeDigest(out, c, user, sinceLog)
			return nil
		},
	}
}

// writeDigest prints the shared state after a compaction; sinceLog is the user's work commits since their last
// work log or completed task.
func writeDigest(out io.Writer, c catchUp, user string, sinceLog int) {
	output.Section(out, "🔄 MEM AFTER COMPACTION")
	fmt.Fprintln(out, "mem synced this checkout with the shared codebase. The compaction summary holds this session's progress; this is the shared state.")
	fmt.Fprintln(out)
	r := c.report
	branch := r.Branch
	if branch == "" {
		branch = "(detached HEAD)"
	}
	if r.Upstream != "" {
		branch += fmt.Sprintf(" (tracking %s: %d ahead, %d behind)", r.Upstream, r.Ahead, r.Behind)
	}
	if r.Dirty {
		branch += ", with uncommitted changes"
	}
	fmt.Fprintln(out, "Branch: "+branch)
	for _, line := range r.Done {
		fmt.Fprintln(out, "✔ "+line)
	}
	writeWorkDigest(out, c.p, user)
	if sinceLog > 0 {
		fmt.Fprintln(out, "Work log: "+checkpoint.LogLine(sinceLog))
	}
	// The digest never fails, so an unreadable setup file only loses the reminder.
	_, setup, _ := readSetup(c.p)
	if setup.present {
		fmt.Fprintln(out, setupDigestLine(setup))
	}

	in := c.incoming
	if len(in.commits) > 0 {
		fmt.Fprintf(out, "\nNew on %s since the last fetch:\n", in.upstream)
		shown := in.commits[:min(len(in.commits), digestCommitLimit)]
		for _, commit := range shown {
			fmt.Fprintln(out, "  "+commit)
		}
		if more := len(in.commits) - len(shown) + in.more; more > 0 {
			fmt.Fprintf(out, "  … and %d more\n", more)
		}
		if len(in.records) > 0 {
			fmt.Fprintln(out, "Work records:")
		}
		for i, line := range in.records {
			if i == digestRecordLimit {
				fmt.Fprintf(out, "  … and %d more record change(s) (`mem sync` lists them)\n", len(in.records)-i)
				break
			}
			fmt.Fprintln(out, "  "+line)
		}
	}
	for _, line := range c.templates {
		fmt.Fprintln(out, line)
	}
	changed := renderKnowledgeChanges(out, c.before, c.after)
	for _, line := range r.Nudges {
		fmt.Fprintln(out, "⚠️ "+line)
	}

	lines := []string{
		"Continue with the work in progress from the compaction summary.",
		"If you can rename this session (as in the Claude desktop app), give it a short title for that work now, unless its title already describes it.",
	}
	if setup.present {
		lines = append(lines, "Continue the setup in "+templates.SetupPath+".")
	}
	if len(in.commits) > 0 {
		lines = append(lines, "Check whether the incoming changes above bear on it.")
	}
	if changed {
		lines = append(lines, "Follow the changed memories above for the rest of this session.")
	}
	if checkpoint.LogRequired(sinceLog) {
		lines = append(lines, "Write a work log for the work since your last one now (`mem log new`), while the compaction summary still holds it.")
	}
	if len(r.Nudges) > 0 || hasWarning(c.templates) {
		lines = append(lines, "Tell the user about each ⚠️ item above.")
	}
	output.Instruction(out, strings.Join(lines, " "))
}

// writeWorkDigest names the user's active spec with its next task and the todos they claimed.
func writeWorkDigest(out io.Writer, p project.Project, user string) {
	if user == "" {
		return
	}
	specs, _ := work.Specs(p, false)
	for _, s := range specs {
		if s.Meta.Status != work.SpecActive || s.Meta.AssignedTo != user {
			continue
		}
		tasks, _ := work.Tasks(s)
		pending := work.PendingTasks(tasks)
		line := fmt.Sprintf("Active spec: %s (%s), %d of %d task(s) done", s.Meta.Title, s.Slug, len(tasks)-len(pending), len(tasks))
		if len(pending) > 0 {
			line += fmt.Sprintf("; next: %s (%s)", pending[0].Meta.Title, pending[0].Slug)
		}
		fmt.Fprintln(out, line)
	}
	todos, _ := work.Todos(p)
	var claimed []string
	for _, t := range todos {
		if t.Meta.ClaimedBy == user {
			claimed = append(claimed, t.Slug)
		}
	}
	if len(claimed) > 0 {
		fmt.Fprintln(out, "Your claimed todos: "+strings.Join(claimed, ", "))
	}
}
