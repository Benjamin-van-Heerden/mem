package cli

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) logCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "log", Short: "Write and read session work logs"}
	cmd.AddCommand(a.logNew(), a.logCommit(), a.logList(), a.logShow())
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
			todos, err := work.Todos(p)
			if err != nil {
				return err
			}
			open := work.OpenTodos(todos)
			output.Section(out, "📌 OPEN TODOS")
			if len(open) == 0 {
				fmt.Fprintln(out, "No open todos.")
			}
			for _, t := range open {
				fmt.Fprintf(out, "- %s (%s)\n", t.Meta.Title, t.Slug)
			}
			var lines []string
			step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines)+1, text)) }
			step(fmt.Sprintf("Read %s and replace every {placeholder} with facts from this session: what was done, decided and tried. The log is not updated later, so do not list future work in it.", p.Rel(l.Path)))
			step("If this session completed any of the open todos above, delete them now (`mem todo delete <todo>`). Record anything still to be done, including blockers and decisions waiting on the user, as todos (`mem todo new \"<title>\" \"<description>\"`).")
			switch {
			case drift.Stale():
				step(fmt.Sprintf("The codebase structure doc is out of date (%d code files, %d lines changed since it was last updated). Update it now: run `mem structure` and follow its instructions.", len(drift.Changes), drift.Lines))
			case drift.Missing:
				step("There is no codebase structure doc yet. Offer the user to create one with `mem structure`.")
			}
			step("Run `mem log commit` to close the session: it commits the log and the other .mem/ records, syncs with the shared codebase and pushes.")
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

func (a *app) logCommit() *cobra.Command {
	return &cobra.Command{
		Use:   "commit [<log>]",
		Short: "Close the session: commit the log and .mem/ records, sync with the shared codebase and push",
		Long:  "Commits the changed .mem/ records (the log, todos, specs, structure doc) with the log, brings the branch up to date with its upstream (fast-forward, or rebase when safe), and pushes. Code outside .mem/ is never committed; uncommitted work is reported. Without an argument it closes your newest log.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			var l work.Log
			if len(args) == 1 {
				if l, err = work.FindLog(p, args[0]); err != nil {
					return err
				}
			} else {
				user, err := a.user(cmd, p)
				if err != nil {
					return err
				}
				var ok bool
				if l, ok, err = work.LatestLog(p, user); err != nil {
					return err
				} else if !ok {
					return errors.New("you have no work log yet; start one with `mem log new`")
				}
			}
			if left := l.Unfilled(); len(left) > 0 {
				return fmt.Errorf("%s still has %d unfilled placeholder(s): %s. Fill them in, then run `mem log commit` again", p.Rel(l.Path), len(left), strings.Join(left, " / "))
			}

			out := cmd.OutOrStdout()
			output.Section(out, "📝 WORK LOG COMMITTED: "+l.Heading())
			fmt.Fprintf(out, "Log: %s\n", p.Rel(l.Path))
			var warnings []string
			if status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", ".mem"); status != "" {
				message := "Work log: " + l.Heading()
				if err := git.Commit(ctx, p.Root, message, ".mem"); err != nil {
					warnings = append(warnings, fmt.Sprintf("Could not commit the .mem/ records (%v). Tell the user, and commit them by hand.", err))
				} else {
					fmt.Fprintf(out, "Committed .mem/ records: %s\n", message)
				}
			} else {
				fmt.Fprintln(out, "The .mem/ records were already committed.")
			}

			r := converge.Sync(ctx, p)
			if r.Upstream != "" && r.Ahead > 0 && r.Behind == 0 {
				if err := git.Push(ctx, p.Root); err != nil {
					warnings = append(warnings, fmt.Sprintf("Pushing %s failed (%v). Tell the user; teammates will not see this session's work until `git push` succeeds.", r.Branch, err))
				} else {
					r.Done = append(r.Done, fmt.Sprintf("Pushed %d commit(s) to %s.", r.Ahead, r.Upstream))
				}
			}
			final := converge.Local(ctx, p)
			final.Fetched, final.FetchError, final.Done = r.Fetched, r.FetchError, r.Done
			final.Nudges = warnings
			for _, line := range r.Nudges {
				if line != converge.Unpushed(r) {
					final.Nudges = append(final.Nudges, line)
				}
			}
			if line := converge.Unpushed(final); line != "" {
				final.Nudges = append(final.Nudges, line)
			}
			if n := uncommittedOutsideMem(ctx, p.Root); n > 0 {
				final.Dirty = true
				final.Nudges = append(final.Nudges, fmt.Sprintf("The session ends with uncommitted work in %d file(s) outside .mem/ (`git status`). Commit it now if it is finished, or tell the user why it stays uncommitted.", n))
			}
			renderReport(out, final, false)
			if len(final.Nudges) > 0 {
				output.Instruction(out, "Tell the user about each ⚠️ item above and resolve it with them before the session ends.")
				return nil
			}
			target := final.Upstream
			if target == "" {
				target = "its remote"
			}
			output.Instruction(out, fmt.Sprintf("Tell the user the session is closed: the log is committed and %s matches %s.", final.Branch, target))
			return nil
		},
	}
}

// uncommittedOutsideMem counts changed and untracked files outside .mem/, ignoring .DS_Store.
func uncommittedOutsideMem(ctx context.Context, root string) int {
	status, _ := git.Run(ctx, root, "status", "--porcelain", "--untracked-files=all")
	n := 0
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		file := strings.Trim(line[3:], "\"")
		if strings.HasPrefix(file, ".mem/") || path.Base(file) == ".DS_Store" {
			continue
		}
		n++
	}
	return n
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
