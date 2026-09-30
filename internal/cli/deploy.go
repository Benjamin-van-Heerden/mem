package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/spf13/cobra"
)

// deployCommand is the user's one-step release: development → staging → production, no notes, no review.
// Agents are told never to run it unless the user explicitly asks for `mem deploy`.
func (a *app) deployCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Release everything: push development, then fast-forward staging and production to it",
		Long:  "Pushes the development branch, fast-forwards staging to it and production to staging, and tags production with a generated summary. No release notes, no review: use `mem promote` for a documented release.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			if p.Config.Release.ProductionPR && !force {
				return errors.New("this repository releases production through a pull request, so `mem deploy` cannot release it; use `mem promote staging` and `mem promote production`")
			}
			out := cmd.OutOrStdout()
			output.Heading(out, "🚀 DEPLOY: "+p.Config.Name)
			if err := pushDevelopment(ctx, out, p); err != nil {
				return err
			}
			for _, stage := range []string{"staging", "production"} {
				if err := deployStage(ctx, out, p, stage); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "")
	cmd.Flags().MarkHidden("force")
	return cmd
}

// pushDevelopment publishes unpushed development commits, so "deploy" ships what the user sees locally.
func pushDevelopment(ctx context.Context, out io.Writer, p project.Project) error {
	g := p.Config.Git
	if git.CurrentBranch(ctx, p.Root) != g.Development {
		return nil
	}
	if dirty, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--untracked-files=no"); dirty != "" {
		fmt.Fprintln(out, "⚠️ Uncommitted changes are not part of this deployment.")
	}
	ahead, err := git.Run(ctx, p.Root, "rev-list", "--count", "@{upstream}..HEAD")
	if err != nil || ahead == "0" {
		return nil
	}
	if _, err := git.Run(ctx, p.Root, "push", "--quiet", g.Remote, g.Development); err != nil {
		return fmt.Errorf("could not push %s (run `mem sync`, then deploy again): %w", g.Development, err)
	}
	fmt.Fprintf(out, "Pushed %s commit(s) on %s.\n", ahead, g.Development)
	return nil
}

func deployStage(ctx context.Context, out io.Writer, p project.Project, stage string) error {
	pl, err := release.Prepare(ctx, p, stage, "")
	if err != nil {
		return err
	}
	remote := p.Config.Git.Remote
	switch {
	case pl.UpToDate || pl.BehindStage:
		output.Section(out, "✔ "+stage)
		fmt.Fprintf(out, "%s/%s is already current.\n", remote, pl.Branch)
		return nil
	case len(pl.Diverged) > 0:
		return divergedError(p, pl)
	}
	notes := ""
	if pl.Tag != "" {
		notes = release.GeneratedMessage(ctx, p, pl)
	}
	if err := release.Execute(ctx, p, pl, notes); err != nil {
		return err
	}
	renderPlan(out, p, pl)
	if pl.Tag != "" {
		fmt.Fprintf(out, "\nReleased %s: %s/%s is now at %s.\n", pl.Tag, remote, pl.Branch, pl.To[:7])
	} else {
		fmt.Fprintf(out, "\n%s/%s is now at %s.\n", remote, pl.Branch, pl.To[:7])
	}
	return nil
}
