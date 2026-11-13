package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/runnables"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// writeContext renders the structure doc, docs, runnable output and work state.
func writeContext(ctx context.Context, out io.Writer, p project.Project, user string) (contextState, error) {
	var state contextState
	setup, setupNow, err := readSetup(p)
	if err != nil {
		return state, err
	}
	if state.setup = setupNow; setupNow.present {
		writeSetupSection(out, setup, setupNow)
	}
	if state.drift, err = writeStructure(ctx, out, p); err != nil {
		return state, err
	}
	if err := writeDocs(out, p); err != nil {
		return state, err
	}
	if err := writeRunnables(ctx, out, p); err != nil {
		return state, err
	}

	var specs []work.Spec
	if !p.Folder() {
		if specs, err = writeSpecs(out, p, user, &state); err != nil {
			return state, err
		}
	}

	todos, err := work.Todos(p)
	if err != nil {
		return state, err
	}
	output.Section(out, "📌 OPEN TODOS")
	if len(todos) == 0 {
		fmt.Fprintln(out, "No open todos.")
	} else {
		rows := [][]string{{"SLUG", "PRIORITY", "TITLE", "AGE"}}
		for _, t := range todos {
			rows = append(rows, []string{t.Slug, t.Priority(), t.Meta.Title, age(t.Meta.Created)})
			if t.Priority() == "critical" {
				state.critical = append(state.critical, fmt.Sprintf("%s (%s)", t.Meta.Title, t.Slug))
			}
		}
		table(out, rows)
	}
	if !p.Folder() {
		state.stale = writeStaleness(ctx, out, p, specs, time.Now())
	}

	logs, err := work.Logs(p)
	if err != nil {
		return state, err
	}
	output.Section(out, "🧾 WORK LOGS")
	full, earlier := recentLogs(logs, user, time.Now())
	if len(full) == 0 {
		fmt.Fprintln(out, "No work logs yet.")
		return state, nil
	}
	openWork := "the specs and todos"
	if p.Folder() {
		openWork = "the todos"
	}
	fmt.Fprintf(out, "Work logs record what past sessions did. They are background: open work is in %s above.\n", openWork)
	if len(earlier) > 0 {
		fmt.Fprintln(out, "\nOther recent logs (read one with `mem log show <log>`):")
		rows := [][]string{{"DATE", "USER", "TITLE", "LOG"}}
		if p.Folder() {
			rows = [][]string{{"DATE", "TITLE", "LOG"}}
		}
		for _, l := range earlier {
			date := strings.SplitN(l.Meta.Created, "T", 2)[0]
			if p.Folder() {
				rows = append(rows, []string{date, l.Heading(), l.Name})
				continue
			}
			rows = append(rows, []string{date, l.Meta.User, l.Heading(), l.Name})
		}
		table(out, rows)
	}
	for i, l := range full {
		output.File(out, "🧾 "+l.Name)
		switch {
		case !l.By(user):
			fmt.Fprintf(out, "The latest log, from %s.\n", l.Meta.User)
		case i == 0:
			fmt.Fprintln(out, "Your latest log.")
		default:
			fmt.Fprintln(out, "An earlier log of yours.")
		}
		if p.Folder() {
			fmt.Fprintf(out, "Date: %s\n\n%s\n", l.Meta.Created, strings.TrimSpace(l.Body))
			continue
		}
		fmt.Fprintf(out, "Date: %s\nSpec: %s\n\n%s\n", l.Meta.Created, orDash(l.Meta.Spec), strings.TrimSpace(l.Body))
	}
	return state, nil
}

// writeSpecs renders the user's active specs in full and lists the other open ones.
func writeSpecs(out io.Writer, p project.Project, user string, state *contextState) ([]work.Spec, error) {
	specs, err := work.Specs(p, false)
	if err != nil {
		return nil, err
	}
	var others [][]string
	for i, s := range specs {
		if s.Meta.Status == work.SpecActive && s.Meta.AssignedTo == user {
			if state.active == nil {
				state.active = &specs[i]
			}
			if err := renderSpec(out, p, s); err != nil {
				return nil, err
			}
			continue
		}
		others = append(others, []string{s.Slug, s.Meta.Status, orDash(s.Meta.AssignedTo), taskProgress(s), s.Meta.Title})
	}
	output.Section(out, "📋 OPEN SPECS")
	if len(others) == 0 {
		fmt.Fprintln(out, "No other open specs.")
	} else {
		table(out, append([][]string{{"SLUG", "STATUS", "ASSIGNED", "TASKS", "TITLE"}}, others...))
	}
	return specs, nil
}

func writeStructure(ctx context.Context, out io.Writer, p project.Project) (structure.Drift, error) {
	d, err := structure.Measure(ctx, p)
	if err != nil || d.Missing {
		return d, err
	}
	data, err := os.ReadFile(structure.Path(p))
	if err != nil {
		return d, err
	}
	heading := "🗺️ CODEBASE AND STRUCTURE"
	if p.Folder() {
		heading = "🗺️ FOLDER STRUCTURE"
	}
	output.Heading(out, heading)
	if d.Stale() {
		fmt.Fprintf(out, "⚠️ Out of date: %d code file(s) and %d line(s) changed since this doc was last updated. Refresh it with `mem structure`.\n\n", len(d.Changes), d.Lines)
	}
	fmt.Fprintln(out, strings.TrimSpace(string(data)))
	return d, nil
}

func writeRunnables(ctx context.Context, out io.Writer, p project.Project) error {
	results := runnables.Run(ctx, p)
	if len(results) == 0 {
		return nil
	}
	output.Heading(out, "🏃 RUNNABLES")
	for _, r := range results {
		output.File(out, r.Name)
		if r.Output != "" {
			fmt.Fprintln(out, r.Output)
		}
		if r.Error != "" {
			fmt.Fprintf(out, "⚠️ %s failed: %s\n", r.Name, r.Error)
		}
	}
	return nil
}

func writeDocs(out io.Writer, p project.Project) error {
	paths, err := filepath.Glob(p.Path("docs", "*.md"))
	if err != nil || len(paths) == 0 {
		return err
	}
	output.Heading(out, "📚 PROJECT DOCS")
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		output.File(out, p.Rel(path))
		fmt.Fprintln(out, strings.TrimSpace(string(data)))
	}
	return nil
}

const (
	recentLogDays = 14
	recentLogMax  = 10
	fullLogs      = 3
	fullLogBytes  = 12000
)

// recentLogs picks the logs to show in full, newest first: the user's latest and, since logs are checkpoints that
// each cover a stretch of work, up to two more of theirs from the recent days while they fit fullLogBytes; or, when
// the user has none, the latest log of anyone. It also lists the other logs of the recent days.
func recentLogs(logs []work.Log, user string, now time.Time) (full, others []work.Log) {
	if len(logs) == 0 {
		return nil, nil
	}
	cutoff := now.AddDate(0, 0, -recentLogDays).Format(time.RFC3339)
	shown := map[string]bool{}
	size := 0
	for _, l := range logs {
		if !l.By(user) {
			continue
		}
		if len(full) > 0 && (len(full) == fullLogs || l.Meta.Created < cutoff || size+len(l.Body) > fullLogBytes) {
			break
		}
		full = append(full, l)
		shown[l.Name] = true
		size += len(l.Body)
	}
	if len(full) == 0 {
		full = logs[:1]
		shown[logs[0].Name] = true
	}
	for _, l := range logs {
		if !shown[l.Name] && l.Meta.Created >= cutoff && len(others) < recentLogMax {
			others = append(others, l)
		}
	}
	return full, others
}

// age describes how long ago an RFC 3339 timestamp was, in days.
func age(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "-"
	}
	switch days := int(time.Since(t).Hours() / 24); days {
	case 0:
		return "today"
	case 1:
		return "1 day"
	default:
		return fmt.Sprintf("%d days", days)
	}
}
