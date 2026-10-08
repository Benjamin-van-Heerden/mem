package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

const incomingCommitLimit = 10

func (a *app) syncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Bring this checkout up to date with the shared codebase and push",
		Long:  "Fetches and fast-forwards or safely rebases the current branch, pushes local commits once it is up to date, syncs template items, and reports what others pushed since the last fetch: commits, work record changes, and changed memories and skills.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, err := a.share(cmd, nil)
			if err != nil {
				return err
			}
			if len(lines) > 0 {
				output.Instruction(cmd.OutOrStdout(), lines...)
			}
			return nil
		},
	}
}

// share runs the mem sync catch-up, pushes when that is safe and prints the report, with nudges from the calling
// command added to it. It returns the instruction lines the report calls for.
func (a *app) share(cmd *cobra.Command, nudges []string) ([]string, error) {
	c, err := a.catchUp(cmd)
	if err != nil {
		return nil, err
	}
	c.report = converge.Push(cmd.Context(), c.p, c.report)
	c.report.Nudges = append(nudges, c.report.Nudges...)
	out := cmd.OutOrStdout()
	renderReport(out, c.report, false)
	if len(c.templates) > 0 {
		output.Section(out, "🧩 TEMPLATES")
		for _, line := range c.templates {
			fmt.Fprintln(out, line)
		}
	}
	renderIncoming(out, c.incoming)
	changed := renderKnowledgeChanges(out, c.before, c.after)
	var lines []string
	if len(c.incoming.commits) > 0 {
		lines = append(lines, "Review the incoming changes above that bear on your current work before continuing.")
	}
	if changed {
		lines = append(lines, "Follow the memories under 🧠 CHANGED MEMORIES for the rest of this session, and use the skills under 🛠️ CHANGED SKILLS where they apply.")
	}
	if len(c.report.Nudges) > 0 {
		lines = append(lines, "Tell the user about each ⚠️ item under 🌿 SHARED CODEBASE.")
	}
	if hasWarning(c.templates) {
		lines = append(lines, "Tell the user about each ⚠️ item under 🧩 TEMPLATES and settle it with them.")
	}
	return lines, nil
}

// commitRecord commits a work record on its own, after the agent has committed the work it describes. It prints the
// outcome and returns nudges for a failed commit or for changes left uncommitted.
func commitRecord(ctx context.Context, out io.Writer, p project.Project, message string, paths ...string) []string {
	if err := git.Commit(ctx, p.Root, message, paths...); err != nil {
		return []string{fmt.Sprintf("Could not commit %s (%v). Tell the user, and commit it with the next commit.", strings.Join(paths, ", "), err)}
	}
	fmt.Fprintln(out, "Committed: "+message)
	if notice := branchNotice(ctx, p); notice != "" {
		fmt.Fprintln(out, notice)
	}
	if n := len(uncommittedFiles(ctx, p.Root)); n > 0 {
		return []string{fmt.Sprintf("%d file(s) remain uncommitted (`git status`); only the record was committed. If they belong to the finished work, commit them now and run `mem sync` to push them; otherwise tell the user why they stay uncommitted.", n)}
	}
	return nil
}

// catchUp is the result of bringing a checkout up to date mid-session.
type catchUp struct {
	p             project.Project
	report        converge.Report
	templates     []string
	incoming      incoming
	before, after knowledge
}

// catchUp converges Git, syncs template items and collects what changed, for `mem sync` and the compaction hook.
func (a *app) catchUp(cmd *cobra.Command) (catchUp, error) {
	ctx := cmd.Context()
	var c catchUp
	p, err := a.project(cmd)
	if err != nil {
		return c, err
	}
	c.before = readKnowledge(p)
	c.report = converge.Sync(ctx, p)
	if c.p, err = a.project(cmd); err != nil {
		return c, err
	}
	if c.templates, err = syncTemplates(ctx, &c.p, true); err != nil {
		return c, err
	}
	c.report = afterUpdates(ctx, c.p, c.report)
	if line := newerRelease(ctx); line != "" {
		c.report.Nudges = append(c.report.Nudges, line)
	}
	c.incoming = readIncoming(ctx, c.p, c.report)
	c.after = readKnowledge(c.p)
	return c, nil
}

// incoming is what others pushed since this checkout last fetched.
type incoming struct {
	upstream string
	commits  []string
	more     int
	records  []string
}

func readIncoming(ctx context.Context, p project.Project, r converge.Report) incoming {
	in := incoming{upstream: r.Upstream}
	if r.Before == "" || r.After == "" || r.Before == r.After {
		return in
	}
	span := r.Before + ".." + r.After
	if log, err := git.Run(ctx, p.Root, "log", "--format=%h  %an: %s", "-n", strconv.Itoa(incomingCommitLimit), span); err == nil && log != "" {
		in.commits = strings.Split(log, "\n")
	}
	if count, err := git.Run(ctx, p.Root, "rev-list", "--count", span); err == nil {
		n, _ := strconv.Atoi(count)
		in.more = n - len(in.commits)
	}
	changed, err := git.Run(ctx, p.Root, "diff", "--name-only", "--no-renames", r.Before, r.After, "--", ".mem/specs", ".mem/todos")
	if err != nil || changed == "" {
		return in
	}
	at := func(rev string) work.Revision {
		return func(path string) ([]byte, bool) {
			data, err := git.Run(ctx, p.Root, "show", rev+":"+path)
			return []byte(data), err == nil
		}
	}
	in.records = work.RecordChanges(strings.Split(changed, "\n"), at(r.Before), at(r.After))
	return in
}

func renderIncoming(out io.Writer, in incoming) {
	if len(in.commits) == 0 {
		return
	}
	output.Section(out, "📥 INCOMING")
	fmt.Fprintf(out, "New on %s since the last fetch:\n", in.upstream)
	for _, c := range in.commits {
		fmt.Fprintln(out, "  "+c)
	}
	if in.more > 0 {
		fmt.Fprintf(out, "  … and %d more (`git log`)\n", in.more)
	}
	if len(in.records) > 0 {
		fmt.Fprintln(out, "\nWork records:")
		for _, line := range in.records {
			fmt.Fprintln(out, "  "+line)
		}
	}
}
