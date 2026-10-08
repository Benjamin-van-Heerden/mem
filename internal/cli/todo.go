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
	cmd.AddCommand(a.todoNew(), a.todoList(), a.todoShow(), a.todoDelete())
	return cmd
}

func (a *app) todoNew() *cobra.Command {
	return &cobra.Command{
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
			t, err := work.NewTodo(p, args[0], description)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO CREATED")
			fmt.Fprintf(out, "Todo: %s (%s)\nFile: %s\n", t.Meta.Title, t.Slug, p.Rel(t.Path))
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
			rows := [][]string{{"SLUG", "AGE", "TITLE"}}
			for _, t := range todos {
				rows = append(rows, []string{t.Slug, age(t.Meta.Created), t.Meta.Title})
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
			fmt.Fprintf(out, "Slug: %s\nOpened: %s ago\n", t.Slug, age(t.Meta.Created))
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
			fmt.Fprintf(out, "Deleted %s (%s). Commit the removal with your next commit.\n", t.Meta.Title, t.Slug)
			return nil
		},
	}
}
