package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) todoCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "todo", Short: "Track standalone matters that need attention"}
	cmd.AddCommand(a.todoNew(), a.todoList(), a.todoShow(), a.todoPriority(), a.todoDelete())
	return cmd
}

var priorityHelp = "Priority: critical (taken up right after onboard), high, medium (the default) or low"

func (a *app) todoNew() *cobra.Command {
	var priority string
	cmd := &cobra.Command{
		Use:   "new <title> [description]",
		Short: "Record a todo",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			description := ""
			if len(args) == 2 {
				description = args[1]
			}
			t, err := work.NewTodo(p, args[0], description, priority)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO CREATED")
			fmt.Fprintf(out, "Todo: %s (%s)\nPriority: %s\nFile: %s\n", t.Meta.Title, t.Slug, t.Priority(), p.Rel(t.Path))
			return nil
		},
	}
	cmd.Flags().StringVar(&priority, "priority", "", priorityHelp)
	return cmd
}

func (a *app) todoPriority() *cobra.Command {
	return &cobra.Command{
		Use:   "priority <todo> <critical|high|medium|low>",
		Short: "Change a todo's priority",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			t, err := work.FindTodo(p, args[0])
			if err != nil {
				return err
			}
			previous := t.Priority()
			if t, err = work.SetTodoPriority(t, args[1]); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO PRIORITY")
			fmt.Fprintf(out, "%s (%s): %s → %s.", t.Meta.Title, t.Slug, previous, t.Priority())
			if !p.Folder() {
				fmt.Fprint(out, " Commit the change with your next commit.")
			}
			fmt.Fprintln(out)
			return nil
		},
	}
}

func (a *app) todoList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List open todos",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			todos, err := work.Todos(p)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODOS")
			if len(todos) == 0 {
				fmt.Fprintln(out, "No open todos.")
				return nil
			}
			rows := [][]string{{"SLUG", "PRIORITY", "AGE", "TITLE"}}
			for _, t := range todos {
				rows = append(rows, []string{t.Slug, t.Priority(), age(t.Meta.Created), t.Meta.Title})
			}
			table(out, rows)
			return nil
		},
	}
}

func (a *app) todoShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show <todo>",
		Short: "Show a todo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			t, err := work.FindTodo(p, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO: "+t.Meta.Title)
			fmt.Fprintf(out, "Slug: %s\nPriority: %s\nOpened: %s ago\n", t.Slug, t.Priority(), age(t.Meta.Created))
			if body := strings.TrimSpace(t.Body); body != "" {
				fmt.Fprintf(out, "\n%s\n", body)
			}
			return nil
		},
	}
}

func (a *app) todoDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <todo>",
		Short: "Delete a todo once it is done or no longer relevant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			t, err := work.FindTodo(p, args[0])
			if err != nil {
				return err
			}
			if err := os.Remove(t.Path); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO DELETED")
			fmt.Fprintf(out, "Deleted %s (%s).", t.Meta.Title, t.Slug)
			if !p.Folder() {
				fmt.Fprint(out, " Commit the removal with your next commit.")
			}
			fmt.Fprintln(out)
			return nil
		},
	}
}
