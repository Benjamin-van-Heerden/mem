package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/Benjamin-van-Heerden/mem/internal/runnables"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

const inlineContextLimit = 14000

func (a *app) onboardCommand() *cobra.Command {
	var offline bool
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Sync with the shared codebase, apply updates and build project context",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var updateLine string
			if !offline {
				var takeOver bool
				if updateLine, takeOver = autoUpdate(ctx); takeOver {
					fmt.Fprintln(cmd.OutOrStdout(), updateLine)
					rerun(cmd.OutOrStdout())
				}
			}
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			user, err := a.user(cmd, p)
			if err != nil {
				return err
			}
			report := converge.Local(ctx, p)
			if !offline {
				report = converge.Sync(ctx, p)
				if p, err = a.project(cmd); err != nil {
					return err
				}
			}
			updates, err := applyUpdates(ctx, p)
			if err != nil {
				return err
			}
			if updateLine != "" {
				updates = append([]string{updateLine}, updates...)
			}
			templateLines, err := syncTemplates(ctx, &p, !offline)
			if err != nil {
				return err
			}
			report = afterUpdates(ctx, p, report)

			var buf bytes.Buffer
			state, err := writeContext(ctx, &buf, p, user)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			output.Heading(out, "📄 MEM ONBOARD: "+p.Config.Name)
			if p.Config.Description != "" {
				fmt.Fprintln(out, p.Config.Description)
			}
			fmt.Fprintf(out, "Developer: %s\n", user)
			renderReport(out, report, offline)
			renderReleases(out, p, release.CurrentStatus(ctx, p))
			if len(updates) > 0 {
				output.Section(out, "🔄 UPDATES")
				for _, line := range updates {
					fmt.Fprintln(out, line)
				}
			}
			if len(templateLines) > 0 {
				output.Section(out, "🧩 TEMPLATES")
				for _, line := range templateLines {
					fmt.Fprintln(out, line)
				}
			}
			if buf.Len() <= inlineContextLimit {
				out.Write(buf.Bytes())
			} else {
				path := p.Path("local", "onboard.md")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					return err
				}
				output.Section(out, "📚 PROJECT CONTEXT")
				fmt.Fprintf(out, "The project context is %d lines long and was written to %s.\n", bytes.Count(buf.Bytes(), []byte("\n")), p.Rel(path))
				fmt.Fprintln(out, "You must read that file in full, every line, before doing anything else. A partial read is not enough.")
			}
			state.templateWarnings = hasWarning(templateLines)
			state.instructionsChanged = slices.ContainsFunc(append(updates, templateLines...), func(line string) bool {
				return strings.HasPrefix(line, "Refreshed the mem instructions") || strings.HasPrefix(line, "Added memory") || strings.HasPrefix(line, "Updated memory")
			})
			renderOnboardInstruction(out, report, state)
			return nil
		},
	}
	cmd.Flags().BoolVar(&offline, "offline", false, "Skip fetching and syncing with the remote")
	return cmd
}

// applyUpdates brings the project and its managed instructions up to date with this executable.
func applyUpdates(ctx context.Context, p project.Project) ([]string, error) {
	var lines, publishPaths []string
	applied, err := project.Upgrade(p)
	if err != nil {
		return nil, err
	}
	if len(applied) > 0 {
		lines = append(lines, fmt.Sprintf("Upgraded the project format to schema %d.", project.Schema))
		publishPaths = append(publishPaths, p.Rel(project.ConfigPath(p.Root)))
	}

	text, err := readAgents(p)
	if err != nil {
		return nil, err
	}
	refreshed, newer, err := agentsmd.Refresh(text, buildinfo.Version)
	if err != nil {
		return nil, err
	}
	switch {
	case newer != "":
		lines = append(lines, fmt.Sprintf("AGENTS.md was written by mem %s, which is newer than this mem (%s). Tell the user to update mem.", newer, buildinfo.Version))
	case refreshed != text:
		status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", "AGENTS.md")
		if err := writeAgents(p, refreshed); err != nil {
			return nil, err
		}
		if status != "" {
			lines = append(lines, "Refreshed the mem instructions in AGENTS.md. It already had uncommitted edits, so commit it together with them.")
		} else {
			lines = append(lines, "Refreshed the mem instructions in AGENTS.md.")
			publishPaths = append(publishPaths, "AGENTS.md")
		}
	}
	if len(publishPaths) > 0 {
		lines = append(lines, publish(ctx, p, "Update mem project files", publishPaths...))
	}
	hookLines, err := hooks.Sync(ctx, p)
	return append(lines, hookLines...), err
}

var templatePaths = []string{"AGENTS.md", ".agents/skills", ".claude/skills", ".mem/docs", templates.LockPath, ".mem/config.toml"}

// syncTemplates brings the project's template items up to date and publishes
// the result, unless those paths already had uncommitted edits.
func syncTemplates(ctx context.Context, p *project.Project, pull bool) ([]string, error) {
	if len(p.Config.Templates.Use) == 0 {
		return nil, nil
	}
	lib, warning, err := openTemplates(ctx, *p, pull)
	if err != nil {
		return []string{fmt.Sprintf("⚠️ Could not open the template library: %v. Template items were not synced.", err)}, nil
	}
	var lines []string
	if warning != "" {
		lines = append(lines, "⚠️ "+warning)
	}
	dirty, _ := git.Run(ctx, p.Root, append([]string{"status", "--porcelain", "--"}, templatePaths...)...)
	res, err := templates.Sync(p, lib)
	if err != nil {
		return nil, err
	}
	lines = append(lines, res.Lines...)
	switch {
	case len(res.Paths) == 0:
	case dirty != "":
		lines = append(lines, fmt.Sprintf("These paths already had uncommitted edits, so nothing was committed: commit the template changes (%s) together with them.", strings.Join(res.Paths, ", ")))
	default:
		lines = append(lines, publish(ctx, *p, "Sync template items", res.Paths...))
	}
	return lines, nil
}

func hasWarning(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, "⚠️") {
			return true
		}
	}
	return false
}

// afterUpdates refreshes the ahead/behind counts and the unpushed nudge, since
// publishing updates may have pushed commits that were local during the sync.
func afterUpdates(ctx context.Context, p project.Project, r converge.Report) converge.Report {
	stale := converge.Unpushed(r)
	fresh := converge.Local(ctx, p)
	r.Ahead, r.Behind = fresh.Ahead, fresh.Behind
	nudges := r.Nudges[:0:0]
	for _, line := range r.Nudges {
		if line != stale {
			nudges = append(nudges, line)
		}
	}
	if line := converge.Unpushed(r); line != "" {
		nudges = append(nudges, line)
	}
	r.Nudges = nudges
	return r
}

func renderReport(out io.Writer, r converge.Report, offline bool) {
	output.Section(out, "🌿 SHARED CODEBASE")
	switch {
	case offline:
		fmt.Fprintln(out, "Remote: not fetched (--offline)")
	case r.Fetched:
		fmt.Fprintf(out, "Remote: fetched from %s\n", r.Remote)
	}
	branch := r.Branch
	if branch == "" {
		branch = "(detached HEAD)"
	}
	fmt.Fprintf(out, "Branch: %s", branch)
	if r.Upstream != "" {
		fmt.Fprintf(out, " (tracking %s: %d ahead, %d behind)", r.Upstream, r.Ahead, r.Behind)
	}
	fmt.Fprintln(out)
	if r.Dirty {
		fmt.Fprintln(out, "Uncommitted changes: yes")
	} else {
		fmt.Fprintln(out, "Uncommitted changes: no")
	}
	for _, line := range r.Done {
		fmt.Fprintln(out, "✔ "+line)
	}
	for _, line := range r.Nudges {
		fmt.Fprintln(out, "⚠️ "+line)
	}
}

type contextState struct {
	active              *work.Spec
	drift               structure.Drift
	templateWarnings    bool
	instructionsChanged bool
}

func renderReleases(out io.Writer, p project.Project, st release.Status) {
	g := p.Config.Git
	output.Section(out, "🚀 RELEASES")
	if len(st.Missing) > 0 {
		fmt.Fprintf(out, "Not on %s yet: %s. The first `mem promote staging` creates %s and `mem promote production` creates %s.\n", g.Remote, strings.Join(st.Missing, ", "), g.Staging, g.Production)
		return
	}
	switch {
	case st.Tag != "":
		fmt.Fprintf(out, "Production (%s): %s, released %s\n", g.Production, st.Tag, st.TagAge)
	default:
		fmt.Fprintf(out, "Production (%s): no release tag yet\n", g.Production)
	}
	fmt.Fprintf(out, "Staging (%s): %d commit(s) ahead of production\n", g.Staging, st.StagingAhead)
	fmt.Fprintf(out, "Development (%s): %d commit(s) ahead of staging\n", g.Development, st.DevAhead)
}

// writeContext renders the structure doc, docs, runnable output and work state.
func writeContext(ctx context.Context, out io.Writer, p project.Project, user string) (contextState, error) {
	var state contextState
	var err error
	if state.drift, err = writeStructure(ctx, out, p); err != nil {
		return state, err
	}
	if err := writeDocs(out, p); err != nil {
		return state, err
	}
	if err := writeRunnables(ctx, out, p); err != nil {
		return state, err
	}

	specs, err := work.Specs(p, false)
	if err != nil {
		return state, err
	}
	var others [][]string
	for i, s := range specs {
		if s.Meta.Status == work.SpecActive && s.Meta.AssignedTo == user {
			if state.active == nil {
				state.active = &specs[i]
			}
			if err := renderSpec(out, p, s); err != nil {
				return state, err
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

	todos, err := work.Todos(p)
	if err != nil {
		return state, err
	}
	output.Section(out, "📌 OPEN TODOS")
	if open := work.OpenTodos(todos); len(open) == 0 {
		fmt.Fprintln(out, "No open todos.")
	} else {
		rows := [][]string{{"SLUG", "TITLE", "AGE", "CLAIMED BY"}}
		for _, t := range open {
			rows = append(rows, []string{t.Slug, t.Meta.Title, age(t.Meta.Created), orDash(t.Meta.ClaimedBy)})
		}
		table(out, rows)
	}

	logs, err := work.Logs(p)
	if err != nil {
		return state, err
	}
	output.Section(out, "🧾 WORK LOGS")
	latest, earlier := recentLogs(logs, user, time.Now())
	if latest == nil {
		fmt.Fprintln(out, "No work logs yet.")
		return state, nil
	}
	fmt.Fprintln(out, "Work logs record what past sessions did. They are background: open work is in the specs and todos above.")
	if len(earlier) > 0 {
		fmt.Fprintln(out, "\nOther recent logs (read one with `mem log show <log>`):")
		rows := [][]string{{"DATE", "USER", "TITLE", "LOG"}}
		for _, l := range earlier {
			rows = append(rows, []string{strings.SplitN(l.Meta.Created, "T", 2)[0], l.Meta.User, l.Heading(), l.Name})
		}
		table(out, rows)
	}
	output.File(out, "🧾 "+latest.Name)
	if latest.Meta.User == user {
		fmt.Fprintln(out, "Your latest log.")
	} else {
		fmt.Fprintf(out, "The latest log, from %s.\n", latest.Meta.User)
	}
	fmt.Fprintf(out, "Date: %s\nSpec: %s\n\n%s\n", latest.Meta.Created, orDash(latest.Meta.Spec), strings.TrimSpace(latest.Body))
	return state, nil
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
	output.Heading(out, "🗺️ CODEBASE AND STRUCTURE")
	if d.Stale() {
		fmt.Fprintf(out, "⚠️ Out of date: %d code file(s) and %d line(s) changed since this doc was last updated. Refresh it with `mem structure`.\n\n", len(d.Changes), d.Lines)
	}
	fmt.Fprintln(out, strings.TrimSpace(string(data)))
	return d, nil
}

func writeRunnables(ctx context.Context, out io.Writer, p project.Project) error {
	results, err := runnables.Run(ctx, p)
	if err != nil || len(results) == 0 {
		return err
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
)

// recentLogs picks the log to show in full (the user's latest, or else the
// latest from anyone) and the other logs of the last two weeks to list by title.
func recentLogs(logs []work.Log, user string, now time.Time) (*work.Log, []work.Log) {
	if len(logs) == 0 {
		return nil, nil
	}
	latest := &logs[0]
	for i := range logs {
		if logs[i].Meta.User == user {
			latest = &logs[i]
			break
		}
	}
	cutoff := now.AddDate(0, 0, -recentLogDays).Format(time.RFC3339)
	var others []work.Log
	for _, l := range logs {
		if l.Name != latest.Name && l.Meta.Created >= cutoff && len(others) < recentLogMax {
			others = append(others, l)
		}
	}
	return latest, others
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

func renderOnboardInstruction(out io.Writer, r converge.Report, state contextState) {
	lines := []string{"Your next response must:"}
	step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines), text)) }
	if state.instructionsChanged {
		step("Read AGENTS.md again before anything else: this onboard changed its mem instructions or memories, so the copy you started with is out of date.")
	}
	if len(r.Nudges) > 0 {
		step("Tell the user about each ⚠️ item under 🌿 SHARED CODEBASE.")
	}
	if state.templateWarnings {
		step("Tell the user about each ⚠️ item under 🧩 TEMPLATES and settle it with them.")
	}
	if state.active != nil {
		step(fmt.Sprintf("Summarize where your active spec %s stands and name its next pending task.", state.active.Slug))
	}
	switch {
	case state.drift.Missing:
		step("Mention that there is no codebase structure doc yet, and offer to create one with `mem structure`.")
	case state.drift.Stale():
		step("Mention that the codebase structure doc is out of date, and offer to refresh it with `mem structure`.")
	}
	step("Summarize the project state from the open specs, open todos and release status. Use tables where they help. Work logs are background: do not present what an old log planned as open work unless a spec or todo still holds it.")
	step("Ask the user how they would like to proceed.")
	lines = append(lines, "", "If the user has already said what to work on, confirm it briefly and continue with that instead of asking.")
	output.Instruction(out, lines...)
}

func driftNudges(ctx context.Context, out io.Writer, p project.Project) {
	r := converge.Local(ctx, p)
	if len(r.Nudges) == 0 {
		return
	}
	output.Section(out, "🌿 SHARED CODEBASE")
	for _, line := range r.Nudges {
		fmt.Fprintln(out, "⚠️ "+line)
	}
}
