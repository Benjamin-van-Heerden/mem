package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// Records and branches that have not moved for this long have probably stopped being true: a claim nobody works
// on, a spec that stalled, a branch someone forgot.
const (
	staleClaim  = 30 * 24 * time.Hour
	staleSpec   = 14 * 24 * time.Hour
	staleBranch = 14 * 24 * time.Hour
)

// writeStaleness lists records and branches that have probably stopped being true, and reports whether it listed
// anything.
func writeStaleness(ctx context.Context, out io.Writer, p project.Project, specs []work.Spec, todos []work.Todo, now time.Time) bool {
	var lines []string
	for _, t := range todos {
		if claimed, err := time.Parse(time.RFC3339, t.Meta.ClaimedAt); err == nil && t.Meta.ClaimedBy != "" && now.Sub(claimed) > staleClaim {
			lines = append(lines, fmt.Sprintf("Todo %s was claimed by %s %s ago. Is it still being worked on, already done (`mem todo delete %s`), or abandoned?", t.Slug, t.Meta.ClaimedBy, days(now.Sub(claimed)), t.Slug))
		}
	}
	for _, s := range specs {
		if s.Meta.Status != work.SpecActive {
			continue
		}
		last, err := git.Run(ctx, p.Root, "log", "-1", "--format=%ct", "--", p.Rel(s.Dir))
		if unix, convErr := strconv.ParseInt(last, 10, 64); err == nil && convErr == nil && now.Sub(time.Unix(unix, 0)) > staleSpec {
			lines = append(lines, fmt.Sprintf("Active spec %s (%s) has not changed for %s: no task completed and nothing added. Is it still being worked on, or should it be completed or abandoned?", s.Slug, orDash(s.Meta.AssignedTo), days(now.Sub(time.Unix(unix, 0)))))
		}
	}
	lines = append(lines, staleBranches(ctx, p, now)...)
	if len(lines) == 0 {
		return false
	}
	output.Section(out, "⏳ CHECK THESE")
	for _, line := range lines {
		fmt.Fprintln(out, "- "+line)
	}
	return true
}

// staleBranches lists remote branches that are not merged into development and have had no commit for a while.
func staleBranches(ctx context.Context, p project.Project, now time.Time) []string {
	g := p.Config.Git
	out, err := git.Run(ctx, p.Root, "for-each-ref", "--no-merged="+g.Remote+"/"+g.Development, "--format=%(refname:short)%09%(committerdate:unix)%09%(authorname)", "refs/remotes/"+g.Remote+"/")
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			continue
		}
		branch := strings.TrimPrefix(fields[0], g.Remote+"/")
		if branch == "HEAD" || branch == g.Remote || branch == g.Development || branch == g.Staging || branch == g.Production || strings.HasPrefix(branch, "promotion/") {
			continue
		}
		unix, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || now.Sub(time.Unix(unix, 0)) <= staleBranch {
			continue
		}
		lines = append(lines, fmt.Sprintf("Branch %s is not merged into %s and has had no commit for %s (last by %s). Merge it, delete it, or keep it?", fields[0], g.Development, days(now.Sub(time.Unix(unix, 0))), fields[2]))
	}
	return lines
}

func days(d time.Duration) string {
	if n := int(d.Hours() / 24); n != 1 {
		return fmt.Sprintf("%d days", n)
	}
	return "1 day"
}
