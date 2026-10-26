package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

func (a *app) logCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "log", Short: "Write and read work logs"}
	cmd.AddCommand(a.logNew(), a.logCommit(), a.logList(), a.logShow())
	return cmd
}

func (a *app) logNew() *cobra.Command {
	var specRef string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a work log for the work since your last one",
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
			// A log is still worth writing when Git cannot say what it covers, such as in a repository without commits.
			cov, _ := checkpoint.Cover(cmd.Context(), p.Root, user, git.UserEmail(cmd.Context(), p.Root))
			var files []work.LogFile
			for _, f := range cov.Files {
				files = append(files, work.LogFile{Path: f.Path, Lines: f.Lines})
			}
			l, err := work.NewLog(p, user, spec, cov.Commits, files)
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
			output.Section(out, "📌 OPEN TODOS")
			if len(todos) == 0 {
				fmt.Fprintln(out, "No open todos.")
			}
			for _, t := range todos {
				fmt.Fprintf(out, "- %s (%s)\n", t.Meta.Title, t.Slug)
			}
			var lines []string
			step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines)+1, text)) }
			step(fmt.Sprintf("Read %s and replace every {placeholder} with facts from the work since your last log: what was done, decided and tried, as things stand. Leave out the sections marked as optional when they have nothing to say. Key Files Affected already lists the files your commits since your last log and the uncommitted changes touched, from Git. The log is not updated later, so do not list future work in it.", p.Rel(l.Path)))
			step("If this work completed any of the open todos above, delete them now (`mem todo delete <todo>`). Record anything still to be done, including blockers and decisions waiting on the user, as todos (`mem todo new \"<title>\" \"<description>\"`).")
			switch {
			case drift.Stale():
				step(fmt.Sprintf("The codebase structure doc is out of date (%d code files, %d lines changed since it was last updated). Update it now: run `mem structure` and follow its instructions.", len(drift.Changes), drift.Lines))
			case drift.Missing:
				step("There is no codebase structure doc yet. Offer the user to create one with `mem structure`.")
			}
			step(retitleSession())
			step("Run `mem log commit <path>...`, naming the .mem/ records you changed for this log, such as the todos you added or deleted and `.mem/structure.md` (`mem log commit` alone when there are none): it commits them with the log, syncs with the shared codebase and pushes. Other sessions' records stay out.")
			output.Instruction(out, lines...)
			driftNudges(cmd.Context(), out, p)
			return nil
		},
	}
	cmd.Flags().StringVar(&specRef, "spec", "", "Spec this work belongs to (defaults to your single active spec)")
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
		Use:   "commit [<log>] [<.mem/ path>...]",
		Short: "Commit the log and the records you name, sync with the shared codebase and push",
		Long:  "Commits the log together with the .mem/ records you name, such as the todos you added or deleted for it or the structure doc, brings the branch up to date with its upstream (fast-forward, or rebase when safe), and pushes. Other changed records stay uncommitted and are listed, since another session may be working in this checkout. Code outside .mem/ is never committed; uncommitted work is reported. Without a log argument it commits your newest log.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			ref, paths, err := logCommitArgs(p.Root, args)
			if err != nil {
				return err
			}
			var l work.Log
			if ref != "" {
				if l, err = work.FindLog(p, ref); err != nil {
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
			// An ignored log never shows in git status, which would read as "already committed".
			if rule := ignoreRule(ctx, p.Root, p.Rel(l.Path)); rule != "" {
				return fmt.Errorf("%s is ignored by Git (%s), so it cannot be committed. Change that rule so it no longer matches .mem/ (a bare `logs/`, for example, also matches .mem/logs/; `/logs/` matches only the top-level folder), then run `mem log commit` again", p.Rel(l.Path), rule)
			}
			if err := l.Finish(); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			output.Section(out, "📝 WORK LOG COMMITTED: "+l.Heading())
			fmt.Fprintf(out, "Log: %s\n", p.Rel(l.Path))
			var warnings []string
			records := append([]string{p.Rel(l.Path)}, paths...)
			if status, _ := git.Run(ctx, p.Root, append([]string{"status", "--porcelain", "--"}, records...)...); status != "" {
				message := "Work log: " + l.Heading()
				if err := git.Commit(ctx, p.Root, message, records...); err != nil {
					warnings = append(warnings, fmt.Sprintf("Could not commit the log and the records you named (%v). Tell the user, and commit them by hand.", err))
				} else {
					fmt.Fprintf(out, "Committed the log and the records you named: %s\n", message)
					if notice := branchNotice(ctx, p); notice != "" {
						fmt.Fprintln(out, notice)
					}
				}
			} else {
				fmt.Fprintln(out, "The log and the records you named were already committed.")
			}
			if left := uncommittedRecords(ctx, p.Root); len(left) > 0 {
				fmt.Fprintln(out, "\nRecords left uncommitted (not in the paths you named):")
				for _, f := range left {
					fmt.Fprintln(out, "  "+f)
				}
				warnings = append(warnings, "Changed .mem/ records outside the paths you named were left uncommitted (listed above). If any belong to this log, such as a todo you added or deleted for it, commit them now with `mem log commit <path>...`. Leave the others alone: another session may be working in this checkout.")
			}

			final := converge.Push(ctx, p, converge.Sync(ctx, p))
			final.Nudges = append(warnings, final.Nudges...)
			if n := uncommittedOutsideMem(ctx, p.Root); n > 0 {
				final.Dirty = true
				final.Nudges = append(final.Nudges, fmt.Sprintf("%d file(s) outside .mem/ have uncommitted changes (`git status`). Commit them now if they are finished, or tell the user why they stay uncommitted.", n))
			}
			renderReport(out, final, false)
			if len(final.Nudges) > 0 {
				output.Instruction(out, "Tell the user about each ⚠️ item above and resolve it with them.")
				return nil
			}
			target := final.Upstream
			if target == "" {
				target = "its remote"
			}
			output.Instruction(out, fmt.Sprintf("Tell the user the log is committed and pushed: %s matches %s. Writing a log does not end the session; carry on with the work.", final.Branch, target))
			return nil
		},
	}
}

// ignoreRule names the .gitignore rule that keeps an untracked path out of Git ("<file>:<line>:<pattern>"), or ""
// when none does.
func ignoreRule(ctx context.Context, root, path string) string {
	out, err := git.Run(ctx, root, "check-ignore", "--verbose", "--", path)
	if err != nil {
		return ""
	}
	rule, _, _ := strings.Cut(out, "\t")
	return rule
}

// logCommitArgs splits log commit's arguments into the log reference, which is optional and names no file outside
// .mem/, and the .mem/ records to commit with the log. Code is committed with the work, never with a log.
func logCommitArgs(root string, args []string) (ref string, paths []string, err error) {
	for _, arg := range args {
		rel := path.Clean(filepath.ToSlash(arg))
		_, statErr := os.Stat(filepath.Join(root, arg))
		switch {
		case strings.HasPrefix(rel, ".mem/"):
			paths = append(paths, rel)
		case ref == "" && len(paths) == 0 && !strings.Contains(rel, "/") && statErr != nil:
			ref = arg
		default:
			return "", nil, fmt.Errorf("%s is not a .mem/ record: log commit commits the log and .mem/ records only. Commit code with the work it belongs to", arg)
		}
	}
	return ref, paths, nil
}

// uncommittedRecords lists changed and untracked files under .mem/, ignoring .DS_Store.
func uncommittedRecords(ctx context.Context, root string) []string {
	var files []string
	for _, file := range uncommittedFiles(ctx, root) {
		if strings.HasPrefix(file, ".mem/") {
			files = append(files, file)
		}
	}
	return files
}

// uncommittedOutsideMem counts changed and untracked files outside .mem/, ignoring .DS_Store.
func uncommittedOutsideMem(ctx context.Context, root string) int {
	n := 0
	for _, file := range uncommittedFiles(ctx, root) {
		if !strings.HasPrefix(file, ".mem/") {
			n++
		}
	}
	return n
}

// uncommittedFiles lists changed and untracked files, ignoring .DS_Store.
func uncommittedFiles(ctx context.Context, root string) []string {
	status, _ := git.Run(ctx, root, "status", "--porcelain", "--untracked-files=all")
	var files []string
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		file := strings.Trim(line[3:], "\"")
		if path.Base(file) != ".DS_Store" {
			files = append(files, file)
		}
	}
	return files
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
