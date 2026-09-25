package cli

import (
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) logCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "log", Short: "Write and read session work logs"}
	cmd.AddCommand(a.logNew(), a.logList(), a.logShow())
	return cmd
}

func (a *app) logNew() *cobra.Command {
	var specRef string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a work log for this session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			user, err := a.user(cmd, p)
			if err != nil {
				return err
			}
			spec := ""
			if specRef != "" {
				s, err := work.FindSpec(p, specRef)
				if err != nil {
					return err
				}
				spec = s.Slug
			} else if s, ok, err := work.ActiveSpecFor(p, user); err != nil {
				return err
			} else if ok {
				spec = s.Slug
			}
			l, err := work.NewLog(p, user, spec)
			if err != nil {
				return err
			}
			drift, err := structure.Measure(cmd.Context(), p)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📝 WORK LOG CREATED")
			fmt.Fprintf(out, "File: %s\nSpec: %s\n", p.Rel(l.Path), orDash(spec))
			lines := []string{fmt.Sprintf("1. Read %s and replace every {placeholder} with details from this session. The next session starts with only this log and the records, so be specific.", p.Rel(l.Path))}
			switch {
			case drift.Stale():
				lines = append(lines, fmt.Sprintf("2. The codebase structure doc is out of date (%d code files, %d lines changed since it was last updated). Update it now: run `mem structure` and follow its instructions.", len(drift.Changes), drift.Lines))
			case drift.Missing:
				lines = append(lines, "2. There is no codebase structure doc yet. Offer the user to create one with `mem structure`.")
			}
			lines = append(lines, fmt.Sprintf("%d. Commit and push the log together with the session's work.", len(lines)+1))
			output.Instruction(out, lines...)
			driftNudges(cmd.Context(), out, p)
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec this session worked on (defaults to your single active spec)")
	return cmd
}

func (a *app) logList() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent work logs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			logs, err := work.Logs(p)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "📝 WORK LOGS")
			if len(logs) == 0 {
				fmt.Fprintln(out, "No work logs yet.")
				return nil
			}
			rows := [][]string{{"LOG", "USER", "SPEC", "CREATED"}}
			for i, l := range logs {
				if i == limit {
					break
				}
				rows = append(rows, []string{l.Name, l.Meta.User, orDash(l.Meta.Spec), l.Meta.Created})
			}
			table(out, rows)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of logs to list")
	return cmd
}

func (a *app) logShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show <log>",
		Short: "Show a work log",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			l, err := work.FindLog(p, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.File(out, "🧾 "+l.Name)
			fmt.Fprintf(out, "User: %s\nDate: %s\nSpec: %s\n\n%s\n", l.Meta.User, l.Meta.Created, orDash(l.Meta.Spec), strings.TrimSpace(l.Body))
			return nil
		},
	}
}
