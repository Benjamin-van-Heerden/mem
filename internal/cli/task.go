package cli

import (
	"errors"
	"fmt"

	"github.com/Benjamin-van-Heerden/memr/internal/output"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
	"github.com/Benjamin-van-Heerden/memr/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) taskCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "Manage the ordered tasks of a spec"}
	cmd.AddCommand(a.taskNew(), a.taskList(), a.taskComplete())
	return cmd
}

// taskSpec resolves --spec, defaulting to the user's single active spec.
func (a *app) taskSpec(cmd *cobra.Command, p project.Project, ref string) (work.Spec, error) {
	if ref != "" {
		return work.FindSpec(p, ref)
	}
	user, err := a.user(cmd, p)
	if err != nil {
		return work.Spec{}, err
	}
	s, ok, err := work.ActiveSpecFor(p, user)
	if err != nil {
		return work.Spec{}, err
	}
	if !ok {
		return work.Spec{}, errors.New("pass --spec <spec>; it can only be omitted when exactly one active spec is assigned to you")
	}
	return s, nil
}

func (a *app) taskNew() *cobra.Command {
	var specRef string
	cmd := &cobra.Command{
		Use:   "new <title> [description]",
		Short: "Add a task to a spec",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := a.taskSpec(cmd, p, specRef)
			if err != nil {
				return err
			}
			if s.Archived() {
				return fmt.Errorf("spec %s is %s", s.Slug, s.Meta.Status)
			}
			description := ""
			if len(args) == 2 {
				description = args[1]
			}
			t, err := work.NewTask(s, args[0], description)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TASK CREATED")
			fmt.Fprintf(out, "Task: %s (%s)\nSpec: %s\nFile: %s\n", t.Meta.Title, t.Slug, s.Slug, p.Rel(t.Path))
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec to add the task to (defaults to your single active spec)")
	return cmd
}

func (a *app) taskList() *cobra.Command {
	var specRef string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the tasks of a spec",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := a.taskSpec(cmd, p, specRef)
			if err != nil {
				return err
			}
			tasks, err := work.Tasks(s)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TASKS: "+s.Meta.Title)
			if len(tasks) == 0 {
				fmt.Fprintln(out, "No tasks yet.")
				return nil
			}
			rows := [][]string{{"SLUG", "STATUS", "TITLE"}}
			for _, t := range tasks {
				rows = append(rows, []string{t.Slug, t.Meta.Status, t.Meta.Title})
			}
			table(out, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec to list (defaults to your single active spec)")
	return cmd
}

func (a *app) taskComplete() *cobra.Command {
	var specRef string
	cmd := &cobra.Command{
		Use:   "complete <task> <notes>",
		Short: "Mark a task completed with notes on what was done",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			s, err := a.taskSpec(cmd, p, specRef)
			if err != nil {
				return err
			}
			t, err := work.FindTask(s, args[0])
			if err != nil {
				return err
			}
			if t.Done() {
				return fmt.Errorf("task %s is already completed", t.Slug)
			}
			if t, err = work.CompleteTask(t, args[1]); err != nil {
				return err
			}
			tasks, err := work.Tasks(s)
			if err != nil {
				return err
			}
			pending := work.PendingTasks(tasks)
			out := cmd.OutOrStdout()
			output.Section(out, "✅ TASK COMPLETED")
			fmt.Fprintf(out, "Task: %s (%s)\nSpec: %s — %d of %d tasks done\n", t.Meta.Title, t.Slug, s.Slug, len(tasks)-len(pending), len(tasks))
			if len(pending) > 0 {
				fmt.Fprintln(out, "\nRemaining:")
				for _, r := range pending {
					fmt.Fprintf(out, "  - %s (%s)\n", r.Meta.Title, r.Slug)
				}
				output.Instruction(out,
					"1. Commit this task's changes now with a descriptive message.",
					fmt.Sprintf("2. Continue with the next task: %s (%s). Its details are in `memr spec show %s`.", pending[0].Meta.Title, pending[0].Slug, s.Slug),
				)
				driftNudges(cmd.Context(), out, p)
				return nil
			}
			output.Instruction(out,
				"All tasks are done.",
				"1. Commit this task's changes now with a descriptive message.",
				fmt.Sprintf("2. Check every Success Criterion in %s against the actual code and fix any gaps.", p.Rel(s.Path())),
				fmt.Sprintf("3. Run `memr spec complete %s`.", s.Slug),
			)
			driftNudges(cmd.Context(), out, p)
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec the task belongs to (defaults to your single active spec)")
	return cmd
}
