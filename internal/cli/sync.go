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
		Short: "Fetch and bring this checkout up to date with the shared codebase",
		Long:  "Fetches and fast-forwards or safely rebases the current branch, syncs template items, and reports what others pushed since the last fetch: commits, work record changes, and changed memories and skills.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := a.catchUp(cmd)
			if err != nil {
				return err
			}
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
			if len(lines) > 0 {
				output.Instruction(out, lines...)
			}
			return nil
		},
	}
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
