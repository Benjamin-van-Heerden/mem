package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/github"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/spf13/cobra"
)

const maxListedCommits = 30

func (a *app) promoteCommand() *cobra.Command {
	var to string
	var confirm bool
	cmd := &cobra.Command{
		Use:   "promote <staging|production>",
		Short: "Fast-forward staging (a preview release) or production (a release)",
		Long:  "Staging is fast-forwarded to the development branch, or to an earlier development commit with --to. Production is fast-forwarded to staging and tagged with release notes: the first run drafts the notes in .mem/local/release-notes.md, and --confirm releases with them.",
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
			out := cmd.OutOrStdout()
			var gh *github.Client
			if stage == "production" && p.Config.Release.ProductionPR {
				if gh, err = githubClient(ctx, p); err != nil {
					return err
				}
				if handled, err := promotionPR(ctx, out, p, gh, confirm); handled || err != nil {
					return err
				}
			}
			pl, err := release.Prepare(ctx, p, stage, to)
			if err != nil {
				return err
			}
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
				if !confirm {
					status, err := writeDraft(ctx, p, pl)
					if err != nil {
						return err
					}
					renderPlan(out, p, pl)
					output.Section(out, "📝 RELEASE NOTES")
					fmt.Fprintf(out, "Draft: %s (%s)\n", release.NotesPath, status)
					output.Instruction(out,
						"Nothing has been pushed yet.",
						fmt.Sprintf("1. Turn %s into a short summary of what %s delivers for its users. Keep the first line; drop the commit list unless it helps.", release.NotesPath, pl.Tag),
						"2. Show the user the notes and ask them to confirm the release.",
						confirmStep(gh != nil),
					)
					return nil
				}
				if notes, err = confirmedNotes(p, pl); err != nil {
					return err
				}
				if gh != nil {
					return openPromotionPR(ctx, out, p, gh, pl, notes)
				}
			}
			if err := release.Execute(ctx, p, pl, notes); err != nil {
				return err
			}
			if pl.Tag != "" {
				os.Remove(filepath.Join(p.Root, release.NotesPath))
			}
			renderPlan(out, p, pl)
			output.Section(out, "✅ PROMOTED")
			if pl.Tag != "" {
				fmt.Fprintf(out, "Released %s: %s/%s is now at %s, tagged with the release notes.\n", pl.Tag, remote, pl.Branch, pl.To[:7])
				output.Instruction(out, "Tell the user the release is out. Where CI deploys production, check the deployment before reporting it as live.")
			} else {
				fmt.Fprintf(out, "%s/%s is now at %s.\n", remote, pl.Branch, pl.To[:7])
				output.Instruction(out, "Tell the user staging is updated. If this project deploys staging, check the preview before releasing; release with `mem promote production` when the user asks for it.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "Promote staging only up to this development commit")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Release production with the reviewed notes in "+release.NotesPath)
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

// writeDraft writes the release notes draft for pl, keeping a draft already written for the same commit so
// edits survive a re-run. It returns what it did, for the output.
func writeDraft(ctx context.Context, p project.Project, pl release.Plan) (string, error) {
	path := filepath.Join(p.Root, release.NotesPath)
	status := "drafted"
	if data, err := os.ReadFile(path); err == nil {
		commit, _ := release.DraftCommit(string(data))
		if commit == pl.To {
			return "kept: it was written for this release", nil
		}
		status = "drafted again: the previous draft was for another commit"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return status, os.WriteFile(path, []byte(release.DraftNotes(ctx, p, pl)), 0o644)
}

// confirmedNotes reads the reviewed draft, refusing one that is missing, empty or written for another commit.
func confirmedNotes(p project.Project, pl release.Plan) (string, error) {
	data, err := os.ReadFile(filepath.Join(p.Root, release.NotesPath))
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("there are no release notes to confirm; run `mem promote production` first to draft them")
	}
	if err != nil {
		return "", err
	}
	commit, notes := release.DraftCommit(string(data))
	if commit != "" && commit != pl.To {
		return "", fmt.Errorf("the release notes were drafted for %s, but production would now move to %s; run `mem promote production` to draft them again", commit[:7], pl.To[:7])
	}
	if notes == "" {
		return "", fmt.Errorf("%s is empty; write the release notes, or run `mem promote production` to draft them again", release.NotesPath)
	}
	return notes, nil
}

func confirmStep(pullRequest bool) string {
	if pullRequest {
		return "3. When they confirm, run `mem promote production --confirm`. This project releases through a pull request: it opens one with these notes for review."
	}
	return "3. When they confirm, run `mem promote production --confirm`."
}
