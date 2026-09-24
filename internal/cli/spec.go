package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Benjamin-van-Heerden/memr/internal/output"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
	"github.com/Benjamin-van-Heerden/memr/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) specCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "spec", Short: "Plan and track larger pieces of work"}
	cmd.AddCommand(a.specNew(), a.specList(), a.specShow(), a.specStart(), a.specComplete(), a.specAbandon())
	return cmd
}

func (a *app) specNew() *cobra.Command {
	return &cobra.Command{
		Use:   "new <title>",
		Short: "Create a draft spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := work.NewSpec(p, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📋 SPEC CREATED")
			fmt.Fprintf(out, "Spec: %s\nSlug: %s\nFile: %s\n", s.Meta.Title, s.Slug, p.Rel(s.Path()))
			output.Instruction(out,
				"1. Research the codebase and clarify requirements with the user until the work can be described precisely.",
				fmt.Sprintf("2. Replace every {placeholder} in %s. Success Criteria must be concrete enough to check against the code.", p.Rel(s.Path())),
				fmt.Sprintf("3. Break the work into ordered tasks: `memr task new \"title\" \"detailed description\" --spec %s`", s.Slug),
				fmt.Sprintf("4. When the user wants implementation to begin: `memr spec start %s`", s.Slug),
				"",
				"Write the spec and tasks so that a fresh session can implement them without this conversation.",
			)
			return nil
		},
	}
}

func (a *app) specList() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List open specs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			specs, err := work.Specs(p, all)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📋 SPECS")
			if len(specs) == 0 {
				fmt.Fprintln(out, "No open specs.")
				return nil
			}
			rows := [][]string{{"SLUG", "STATUS", "ASSIGNED", "TASKS", "TITLE"}}
			for _, s := range specs {
				rows = append(rows, []string{s.Slug, s.Meta.Status, orDash(s.Meta.AssignedTo), taskProgress(s), s.Meta.Title})
			}
			table(out, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Include completed and abandoned specs")
	return cmd
}

func (a *app) specShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show <spec>",
		Short: "Show a spec and its tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := work.FindSpec(p, args[0])
			if err != nil {
				return err
			}
			return renderSpec(cmd.OutOrStdout(), p, s)
		},
	}
}

func (a *app) specStart() *cobra.Command {
	return &cobra.Command{
		Use:   "start <spec>",
		Short: "Assign a spec to yourself, mark it active and publish that",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			user, err := a.user(cmd, p)
			if err != nil {
				return err
			}
			s, err := work.FindSpec(p, args[0])
			if err != nil {
				return err
			}
			if s.Archived() {
				return fmt.Errorf("spec %s is %s", s.Slug, s.Meta.Status)
			}
			previous := s.Meta.AssignedTo
			s.Meta.Status = work.SpecActive
			s.Meta.AssignedTo = user
			if err := work.SaveSpec(s); err != nil {
				return err
			}
			tasks, err := work.Tasks(s)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "🚀 SPEC STARTED")
			fmt.Fprintf(out, "Spec: %s (%s)\nAssigned to: %s\n", s.Meta.Title, s.Slug, user)
			if previous != "" && previous != user {
				fmt.Fprintf(out, "Previously assigned to %s. Tell the user so they can let %s know.\n", previous, previous)
			}
			fmt.Fprintln(out, publish(cmd.Context(), p, "Start spec "+s.Slug, p.Rel(s.Dir)))
			pending := work.PendingTasks(tasks)
			if len(pending) == 0 {
				output.Instruction(out, fmt.Sprintf("This spec has no pending tasks. Add them before implementing: `memr task new \"title\" \"detailed description\" --spec %s`", s.Slug))
				return nil
			}
			output.Instruction(out,
				fmt.Sprintf("Implement the pending tasks in order, starting with %s (%s).", pending[0].Meta.Title, pending[0].Slug),
				"After each task, record it with `memr task complete <task> \"what was done and how it was verified\"` and continue with the next one.",
			)
			return nil
		},
	}
}

func (a *app) specComplete() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <spec>",
		Short: "Mark a spec completed and archive it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := work.FindSpec(p, args[0])
			if err != nil {
				return err
			}
			if s.Archived() {
				return fmt.Errorf("spec %s is already %s", s.Slug, s.Meta.Status)
			}
			tasks, err := work.Tasks(s)
			if err != nil {
				return err
			}
			if pending := work.PendingTasks(tasks); len(pending) > 0 {
				names := make([]string, len(pending))
				for i, t := range pending {
					names[i] = t.Slug
				}
				return fmt.Errorf("spec %s still has pending tasks: %s; complete them first", s.Slug, strings.Join(names, ", "))
			}
			s, err = work.ArchiveSpec(p, s, work.SpecCompleted, "")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "✅ SPEC COMPLETED")
			fmt.Fprintf(out, "Spec: %s\nArchived to: %s\n", s.Meta.Title, p.Rel(s.Dir))
			output.Instruction(out,
				"1. Summarize for the user what the spec delivered.",
				fmt.Sprintf("2. Offer to write a session log: `memr log new --spec %s`.", s.Slug),
				"3. Commit and push the work together with the .memr/ changes.",
			)
			driftNudges(cmd.Context(), out, p)
			return nil
		},
	}
}

func (a *app) specAbandon() *cobra.Command {
	var reason string
	cmd := &cobra.Command{
		Use:   "abandon <spec>",
		Short: "Abandon a spec and archive it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := work.FindSpec(p, args[0])
			if err != nil {
				return err
			}
			if s.Archived() {
				return fmt.Errorf("spec %s is already %s", s.Slug, s.Meta.Status)
			}
			s, err = work.ArchiveSpec(p, s, work.SpecAbandoned, reason)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "🗑️ SPEC ABANDONED")
			fmt.Fprintf(out, "Spec: %s\nReason: %s\nArchived to: %s\n", s.Meta.Title, reason, p.Rel(s.Dir))
			fmt.Fprintln(out, "Commit and push the .memr/ changes with your next commit.")
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Why the spec is being abandoned")
	cmd.MarkFlagRequired("reason")
	return cmd
}

func renderSpec(out io.Writer, p project.Project, s work.Spec) error {
	tasks, err := work.Tasks(s)
	if err != nil {
		return err
	}
	output.Section(out, "📋 SPEC: "+s.Meta.Title)
	fmt.Fprintf(out, "Slug: %s\nStatus: %s\nAssigned to: %s\nFile: %s\n\n", s.Slug, s.Meta.Status, orDash(s.Meta.AssignedTo), p.Rel(s.Path()))
	fmt.Fprintln(out, strings.TrimSpace(s.Body))
	if len(tasks) == 0 {
		fmt.Fprintln(out, "\n### Tasks\n\nNo tasks yet.")
		return nil
	}
	fmt.Fprintln(out, "\n### Tasks")
	for _, t := range tasks {
		if t.Done() {
			fmt.Fprintf(out, "\n- [x] %s (%s)\n", t.Meta.Title, t.Slug)
		}
	}
	for _, t := range work.PendingTasks(tasks) {
		fmt.Fprintf(out, "\n- [ ] %s (%s)\n", t.Meta.Title, t.Slug)
		if body := strings.TrimSpace(t.Body); body != "" {
			fmt.Fprintf(out, "\n%s\n", indent(body, "  "))
		}
	}
	return nil
}

func taskProgress(s work.Spec) string {
	tasks, err := work.Tasks(s)
	if err != nil || len(tasks) == 0 {
		return "-"
	}
	return fmt.Sprintf("%d/%d", len(tasks)-len(work.PendingTasks(tasks)), len(tasks))
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
