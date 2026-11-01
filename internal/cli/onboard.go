package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
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
			if p.Folder() {
				return a.onboardFolder(cmd, p, user, updateLine)
			}
			before := readKnowledge(p)
			branchLines, stuck := leaveGuardedBranch(ctx, p)
			if branchLines != nil && !stuck {
				// Development may hold a different config than the branch the checkout was on.
				if p, err = a.project(cmd); err != nil {
					return err
				}
			}
			report := converge.Local(ctx, p)
			if !offline && !stuck {
				report = converge.Sync(ctx, p)
				if p, err = a.project(cmd); err != nil {
					return err
				}
			}
			var updates, templateLines []string
			if stuck {
				// Still on staging or production: no commits, so nothing lands there. The guards still go in.
				if updates, err = hooks.Sync(ctx, p); err != nil {
					return err
				}
			} else {
				if updates, err = applyUpdates(ctx, p); err != nil {
					return err
				}
				if templateLines, err = syncTemplates(ctx, &p, !offline); err != nil {
					return err
				}
			}
			updates = append(branchLines, updates...)
			if updateLine != "" {
				updates = append([]string{updateLine}, updates...)
			}
			agents, err := readAgents(p)
			if err != nil {
				return err
			}
			if warning := agentsSizeWarning(agents); warning != "" {
				updates = append(updates, warning)
			}
			report = afterUpdates(ctx, p, report)
			if line := converge.Worktree(report); line != "" {
				report.Nudges = append([]string{line}, report.Nudges...)
			}

			var buf bytes.Buffer
			knowledgeChanged := renderKnowledgeChanges(&buf, before, readKnowledge(p))
			state, err := writeContext(ctx, &buf, p, user)
			if err != nil {
				return err
			}
			state.knowledgeChanged = knowledgeChanged

			out := cmd.OutOrStdout()
			output.Heading(out, "📄 MEM ONBOARD: "+p.Config.Name)
			if p.Config.Description != "" {
				fmt.Fprintln(out, p.Config.Description)
			}
			fmt.Fprintf(out, "Developer: %s\n", user)
			renderReport(out, report, offline)
			line, baseMissing := baseLine(ctx, p, report.Branch)
			if line != "" {
				fmt.Fprintln(out, line)
			}
			if !p.Config.Git.SingleBranch {
				renderReleases(out, p, release.CurrentStatus(ctx, p))
			}
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
			if err := emitContext(out, &buf, p); err != nil {
				return err
			}
			state.baseMissing = baseMissing
			state.updateWarnings = hasWarning(updates)
			state.templateWarnings = hasWarning(templateLines)
			state.templateTodos = hasTemplateTodo(templateLines)
			renderOnboardInstruction(out, report, state)
			return nil
		},
	}
	cmd.Flags().BoolVar(&offline, "offline", false, "Skip fetching and syncing with the remote")
	return cmd
}

// emitContext prints the context, or writes it to a file and points the agent there when it is too long to print.
func emitContext(out io.Writer, buf *bytes.Buffer, p project.Project) error {
	if buf.Len() <= inlineContextLimit {
		_, err := out.Write(buf.Bytes())
		return err
	}
	output.Section(buf, "🏷️ SESSION TITLE")
	fmt.Fprintln(buf, sessionTitleReminder())
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
	return nil
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
	active           *work.Spec
	drift            structure.Drift
	updateWarnings   bool
	templateWarnings bool
	templateTodos    bool
	knowledgeChanged bool
	stale            bool
	setup            setupState
	baseMissing      int
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
	fmt.Fprintf(out, "Development (%s): %d commit(s) ahead of staging", g.Development, st.DevAhead)
	if st.OldestUnreleased != "" {
		fmt.Fprintf(out, ", the oldest from %s", st.OldestUnreleased)
	}
	fmt.Fprintln(out)
	for _, b := range []struct {
		name    string
		outside int
	}{{g.Staging, st.StagingOutside}, {g.Production, st.ProductionOutside}} {
		if b.outside > 0 {
			fmt.Fprintf(out, "⚠️ %s/%s has %d commit(s) that are not on %s, so promotion cannot fast-forward it. Tell the user; bring them into %s with `git merge --no-ff --no-edit %s/%s` on %s, then push.\n", g.Remote, b.name, b.outside, g.Development, g.Development, g.Remote, b.name, g.Development)
		}
	}
}

func renderOnboardInstruction(out io.Writer, r converge.Report, state contextState) {
	lines := []string{"Your next response must:"}
	step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines), text)) }
	if state.knowledgeChanged {
		step("Follow the memories under 🧠 CHANGED MEMORIES for the rest of this session, and use the skills under 🛠️ CHANGED SKILLS where they apply: they changed after this session started.")
	}
	if len(r.Nudges) > 0 {
		step("Tell the user about each ⚠️ item under 🌿 SHARED CODEBASE.")
	}
	if state.baseMissing > 0 {
		step("Mention the released commits not merged in yet, under Base in 🌿 SHARED CODEBASE. Merge them only when the user asks, as described there.")
	}
	if state.updateWarnings {
		step("Tell the user about each ⚠️ item under 🔄 UPDATES and settle it with them.")
	}
	if state.templateWarnings {
		step("Tell the user about each ⚠️ item under 🧩 TEMPLATES and settle it with them.")
	}
	if state.templateTodos {
		step(templateTodoInstruction(state.setup.present))
	}
	if state.setup.present {
		step(setupInstruction(state.setup))
		if !state.setup.finished() {
			output.Instruction(out, lines...)
			return
		}
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
	if state.stale {
		step("Go through each item under ⏳ CHECK THESE with the user: these records and branches have probably stopped being true. Update, complete, delete or keep each as they decide.")
	}
	sources := "the open specs, open todos and release status"
	if r.SingleBranch {
		sources = "the open specs and open todos"
	}
	step("Summarize the project state from " + sources + ". Use tables where they help. Work logs are background: do not present what an old log planned as open work unless a spec or todo still holds it.")
	step("Ask the user how they would like to proceed.")
	lines = append(lines, "", "If the user has already said what to work on, confirm it briefly and continue with that instead of asking.", nameSession())
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
