package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
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
		Short: "Mark a task completed, commit the work with it, sync and push",
		Long:  "Marks the task completed with notes on what was done and how it was verified, then commits every change in the working tree (ignored files excepted) together with the task record: the task title is the commit subject, the notes its body. The branch is then brought up to date with its upstream and pushed, as `mem sync` does.",
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
			nudges := commitTask(cmd.Context(), out, p, s, t, args[1])
			if len(pending) > 0 {
				fmt.Fprintln(out, "\nRemaining:")
				for _, r := range pending {
					fmt.Fprintf(out, "  - %s (%s)\n", r.Meta.Title, r.Slug)
				}
				next := pending[0]
				output.Section(out, "➡️ NEXT TASK: "+next.Meta.Title)
				fmt.Fprintln(out, strings.TrimSpace(next.Body))
			}
			lines, err := a.share(cmd, nudges)
			if err != nil {
				return err
			}
			if len(pending) > 0 {
				next := pending[0]
				output.Instruction(out, append(lines, fmt.Sprintf("Continue with the next task now, without waiting for approval: %s (%s), described above. The spec with its full context: `mem spec show %s`.", next.Meta.Title, next.Slug, s.Slug))...)
				return nil
			}
			output.Instruction(out, append(lines,
				"All tasks are done.",
				fmt.Sprintf("1. Check every Success Criterion in %s against the actual code and fix any gaps.", p.Rel(s.Path())),
				fmt.Sprintf("2. Run `mem spec complete %s`.", s.Slug),
			)...)
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec the task belongs to (defaults to your single active spec)")
	return cmd
}

// commitTask commits the task's work together with its record: everything changed in the working tree, ignored
// files excepted. The Mem-Task trailer marks the commit as a checkpoint for the work-log count.
func commitTask(ctx context.Context, out io.Writer, p project.Project, s work.Spec, t work.Task, notes string) []string {
	message := fmt.Sprintf("%s\n\n%s\n\n%s %s/%s", t.Meta.Title, strings.TrimSpace(notes), checkpoint.TaskTrailer, s.Slug, t.Slug)
	if err := git.CommitAll(ctx, p.Root, message); err != nil {
		return []string{fmt.Sprintf("Could not commit the task's work (%v). Tell the user, and commit it by hand with the task record.", err)}
	}
	fmt.Fprintln(out, "Committed: "+t.Meta.Title+" (every change in the working tree, with the task record)")
	if notice := branchNotice(ctx, p); notice != "" {
		fmt.Fprintln(out, notice)
	}
	return nil
}
