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
	cmd.AddCommand(a.todoNew(), a.todoList(), a.todoShow(), a.todoClaim(), a.todoDelete())
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
	var all bool
	cmd := &cobra.Command{
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
			if !all {
				todos = work.OpenTodos(todos)
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODOS")
			if len(todos) == 0 {
				fmt.Fprintln(out, "No open todos.")
				return nil
			}
			rows := [][]string{{"SLUG", "STATUS", "CLAIMED BY", "TITLE"}}
			for _, t := range todos {
				rows = append(rows, []string{t.Slug, t.Meta.Status, orDash(t.Meta.ClaimedBy), t.Meta.Title})
			}
			table(out, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Include claimed todos")
	return cmd
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
			fmt.Fprintf(out, "Slug: %s\nStatus: %s\n", t.Slug, t.Meta.Status)
			if t.Meta.ClaimedBy != "" {
				fmt.Fprintf(out, "Claimed by: %s\n", t.Meta.ClaimedBy)
			}
			if body := strings.TrimSpace(t.Body); body != "" {
				fmt.Fprintf(out, "\n%s\n", body)
			}
			return nil
		},
	}
}

func (a *app) todoClaim() *cobra.Command {
	return &cobra.Command{
		Use:   "claim <todo>",
		Short: "Claim a todo you are working on and publish that",
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
			t, err := work.FindTodo(p, args[0])
			if err != nil {
				return err
			}
			if t.Meta.Status == work.TodoClaimed {
				return fmt.Errorf("todo %s is already claimed by %s", t.Slug, t.Meta.ClaimedBy)
			}
			if t, err = work.ClaimTodo(t, user); err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📌 TODO CLAIMED")
			fmt.Fprintf(out, "Todo: %s (%s)\nClaimed by: %s\n", t.Meta.Title, t.Slug, user)
			fmt.Fprintln(out, publish(cmd.Context(), p, "Claim todo "+t.Slug, p.Rel(t.Path)))
			if body := strings.TrimSpace(t.Body); body != "" {
				fmt.Fprintf(out, "\n%s\n", body)
			}
			output.Instruction(out, "Work on this todo now. If it turns out to be substantial, propose turning it into a spec with the user.")
			return nil
		},
	}
}

func (a *app) todoDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <todo>",
		Short: "Delete a todo that is no longer relevant",
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
