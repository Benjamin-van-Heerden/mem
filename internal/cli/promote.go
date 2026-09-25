package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

const maxListedCommits = 30

func (a *app) promoteCommand() *cobra.Command {
	var to, notesFile string
	cmd := &cobra.Command{
		Use:   "promote <staging|production>",
		Short: "Fast-forward staging (a preview release) or production (a release)",
		Long:  "Staging is fast-forwarded to the development branch, or to an earlier development commit with --to. Production is fast-forwarded to staging and tagged with release notes.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			stage := args[0]
			if stage == "production" && to != "" {
				return fmt.Errorf("production always moves to what staging previewed; to release less, promote staging --to <commit> first")
			}
			pl, err := release.Prepare(ctx, p, stage, to)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			remote := p.Config.Git.Remote
			switch {
			case pl.UpToDate && to != "":
				output.Section(out, "🚀 PROMOTION")
				fmt.Fprintf(out, "%s/%s is already at %s. Nothing to promote.\n", remote, pl.Branch, to)
				return nil
			case pl.UpToDate:
				output.Section(out, "🚀 PROMOTION")
				fmt.Fprintf(out, "%s/%s already matches %s/%s. Nothing to promote.\n", remote, pl.Branch, remote, pl.Source)
				return nil
			case pl.BehindStage:
				output.Section(out, "🚀 PROMOTION")
				fmt.Fprintf(out, "%s/%s is already past that commit. Nothing to promote.\n", remote, pl.Branch)
				return nil
			case len(pl.Diverged) > 0:
				return divergedError(p, pl)
			}

			var notes string
			if pl.Tag != "" {
				if notesFile == "" {
					renderPlan(out, p, pl)
					output.Instruction(out,
						"Nothing has been pushed yet.",
						fmt.Sprintf("1. Write release notes for %s to a file outside the repository: a short summary of what this release delivers, based on the commits and specs above.", pl.Tag),
						"2. Show the user the release notes.",
						"3. Run `mem promote production --notes <file>`.",
					)
					return nil
				}
				data, err := os.ReadFile(notesFile)
				if err != nil {
					return err
				}
				notes = string(data)
			}
			if err := release.Execute(ctx, p, pl, notes); err != nil {
				return err
			}
			renderPlan(out, p, pl)
			output.Section(out, "✅ PROMOTED")
			if pl.Tag != "" {
				fmt.Fprintf(out, "Released %s: %s/%s is now at %s, tagged with the release notes.\n", pl.Tag, remote, pl.Branch, pl.To[:7])
				output.Instruction(out, "Tell the user the release is out and CI will deploy it. Check the deployment before reporting it as live.")
			} else {
				fmt.Fprintf(out, "%s/%s is now at %s.\n", remote, pl.Branch, pl.To[:7])
				output.Instruction(out, "Tell the user the preview release is out and CI will deploy it. Once the preview checks out, release it with `mem promote production`.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "Promote staging only up to this development commit")
	cmd.Flags().StringVar(&notesFile, "notes", "", "File with the release notes for a production release")
	return cmd
}

func renderPlan(out io.Writer, p project.Project, pl release.Plan) {
	remote := p.Config.Git.Remote
	title := "🚀 PREVIEW RELEASE"
	if pl.Tag != "" {
		title = "🚀 RELEASE " + pl.Tag
	}
	output.Section(out, title)
	from := "(new branch)"
	if pl.From != "" {
		from = pl.From[:7]
	}
	fmt.Fprintf(out, "%s/%s: %s → %s (from %s/%s)\n", remote, pl.Branch, from, pl.To[:7], remote, pl.Source)
	fmt.Fprintf(out, "\nCommits (%d):\n", len(pl.Commits))
	for i, c := range pl.Commits {
		if i == maxListedCommits {
			fmt.Fprintf(out, "  ... and %d more\n", len(pl.Commits)-maxListedCommits)
			break
		}
		fmt.Fprintf(out, "  %s  %s  (%s)\n", c.Hash, c.Subject, c.Author)
	}
	if len(pl.Specs) > 0 {
		fmt.Fprintf(out, "\nSpecs completed: %s\n", strings.Join(pl.Specs, ", "))
	}
	if specs, err := work.Specs(p, false); err == nil {
		var active []string
		for _, s := range specs {
			if s.Meta.Status == work.SpecActive {
				active = append(active, s.Slug)
			}
		}
		if len(active) > 0 {
			fmt.Fprintf(out, "\n⚠️ Specs still in progress: %s. Their partial work may be included; check with the user if unsure.\n", strings.Join(active, ", "))
		}
	}
	if pl.Unpushed > 0 {
		fmt.Fprintf(out, "\n⚠️ %d local commit(s) on %s are not pushed and are not included.\n", pl.Unpushed, pl.Source)
	}
}

func divergedError(p project.Project, pl release.Plan) error {
	g := p.Config.Git
	var list []string
	for _, c := range pl.Diverged {
		list = append(list, fmt.Sprintf("  %s  %s  (%s)", c.Hash, c.Subject, c.Author))
	}
	steps := fmt.Sprintf("`git switch %s`, `mem sync`, `git merge --no-ff --no-edit %s/%s`, `git push`, then `mem promote staging`", g.Development, g.Remote, pl.Branch)
	if pl.Stage == "production" {
		steps += " and `mem promote production`"
	}
	return fmt.Errorf("%s/%s has %d commit(s) that are not on %s, so it cannot be fast-forwarded:\n%s\nTell the user. To bring them back into the shared history, run %s", g.Remote, pl.Branch, len(pl.Diverged), pl.Source, strings.Join(list, "\n"), steps)
}
