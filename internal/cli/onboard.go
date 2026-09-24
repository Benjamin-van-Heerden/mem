package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Benjamin-van-Heerden/memr/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/memr/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/memr/internal/converge"
	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/hooks"
	"github.com/Benjamin-van-Heerden/memr/internal/output"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
	"github.com/Benjamin-van-Heerden/memr/internal/release"
	"github.com/Benjamin-van-Heerden/memr/internal/runnables"
	"github.com/Benjamin-van-Heerden/memr/internal/structure"
	"github.com/Benjamin-van-Heerden/memr/internal/work"
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

			var buf bytes.Buffer
			state, err := writeContext(ctx, &buf, p, user)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			output.Heading(out, "📄 MEMR ONBOARD: "+p.Config.Name)
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
		lines = append(lines, fmt.Sprintf("AGENTS.md was written by memr %s, which is newer than this memr (%s). Tell the user to update memr.", newer, buildinfo.Version))
	case refreshed != text:
		status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", "AGENTS.md")
		if err := writeAgents(p, refreshed); err != nil {
			return nil, err
		}
		if status != "" {
			lines = append(lines, "Refreshed the memr instructions in AGENTS.md. It already had uncommitted edits, so commit it together with them.")
		} else {
			lines = append(lines, "Refreshed the memr instructions in AGENTS.md.")
			publishPaths = append(publishPaths, "AGENTS.md")
		}
	}
	if len(publishPaths) > 0 {
		lines = append(lines, publish(ctx, p, "Update memr project files", publishPaths...))
	}
	hookLines, err := hooks.Sync(ctx, p)
	return append(lines, hookLines...), err
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
	active *work.Spec
	drift  structure.Drift
}

func renderReleases(out io.Writer, p project.Project, st release.Status) {
	g := p.Config.Git
	output.Section(out, "🚀 RELEASES")
	if len(st.Missing) > 0 {
		fmt.Fprintf(out, "Not on %s yet: %s. The first `memr promote staging` creates %s and `memr promote production` creates %s.\n", g.Remote, strings.Join(st.Missing, ", "), g.Staging, g.Production)
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
		rows := [][]string{{"SLUG", "TITLE"}}
		for _, t := range open {
			rows = append(rows, []string{t.Slug, t.Meta.Title})
		}
		table(out, rows)
	}

	logs, err := work.Logs(p)
	if err != nil {
		return state, err
	}
	output.Section(out, "🧾 RECENT WORK LOGS")
	recent := recentLogs(logs, user)
	if len(recent) == 0 {
		fmt.Fprintln(out, "No work logs yet.")
	}
	for _, l := range recent {
		output.File(out, "🧾 "+l.Name)
		if l.Meta.User == user {
			fmt.Fprintln(out, "Your log.")
		} else {
			fmt.Fprintf(out, "From %s.\n", l.Meta.User)
		}
		fmt.Fprintf(out, "Date: %s\nSpec: %s\n\n%s\n", l.Meta.Created, orDash(l.Meta.Spec), strings.TrimSpace(l.Body))
	}
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
		fmt.Fprintf(out, "⚠️ Out of date: %d code file(s) and %d line(s) changed since this doc was last updated. Refresh it with `memr structure`.\n\n", len(d.Changes), d.Lines)
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

// recentLogs picks up to three of the user's latest logs and fills to five with
// the latest from anyone, oldest first so they read as a timeline.
func recentLogs(logs []work.Log, user string) []work.Log {
	var picked []work.Log
	seen := map[string]bool{}
	for _, l := range logs {
		if l.Meta.User == user && len(picked) < 3 {
			picked = append(picked, l)
			seen[l.Name] = true
		}
	}
	for _, l := range logs {
		if len(picked) == 5 {
			break
		}
		if !seen[l.Name] {
			picked = append(picked, l)
		}
	}
	sort.Slice(picked, func(i, j int) bool { return picked[i].Meta.Created < picked[j].Meta.Created })
	return picked
}

func renderOnboardInstruction(out io.Writer, r converge.Report, state contextState) {
	lines := []string{"Your next response must:"}
	step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines), text)) }
	if len(r.Nudges) > 0 {
		step("Tell the user about each ⚠️ item under 🌿 SHARED CODEBASE.")
	}
	if state.active != nil {
		step(fmt.Sprintf("Summarize where your active spec %s stands and name its next pending task.", state.active.Slug))
	}
	switch {
	case state.drift.Missing:
		step("Mention that there is no codebase structure doc yet, and offer to create one with `memr structure`.")
	case state.drift.Stale():
		step("Mention that the codebase structure doc is out of date, and offer to refresh it with `memr structure`.")
	}
	step("Summarize the project state: open specs, open todos and what the recent work logs say comes next. Use tables where they help.")
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
