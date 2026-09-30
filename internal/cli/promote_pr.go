package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/github"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
)

// Projects with [release] production_pr release production through a pull request from a snapshot branch.
// mem completes it itself by fast-forwarding production to the pull request's head, so history stays linear;
// GitHub's merge button would add or rewrite commits.

var releaseHeading = regexp.MustCompile(`^# Release v[0-9.]+`)

func promotionPrefix(p project.Project) string { return "promotion/" + p.Config.Git.Production + "/" }

func githubClient(ctx context.Context, p project.Project) (*github.Client, error) {
	// The raw configured URL, before any insteadOf rewriting, names the GitHub repository.
	remoteURL, err := git.Run(ctx, p.Root, "config", "--get", "remote."+p.Config.Git.Remote+".url")
	if err != nil {
		return nil, fmt.Errorf("remote %s has no URL", p.Config.Git.Remote)
	}
	return github.New(ctx, remoteURL)
}

// openPromotionPR pushes the release commit to a snapshot branch and opens the pull request with the notes.
func openPromotionPR(ctx context.Context, out io.Writer, p project.Project, gh *github.Client, pl release.Plan, notes string) error {
	g := p.Config.Git
	branch := promotionPrefix(p) + time.Now().UTC().Format("20060102-150405")
	if _, err := git.RunEnv(ctx, p.Root, []string{release.PromoteEnv + "=1"}, "push", "--quiet", g.Remote, pl.To+":refs/heads/"+branch); err != nil {
		return fmt.Errorf("could not push the snapshot branch %s: %w", branch, err)
	}
	title := fmt.Sprintf("Release: %s → %s (%s)", g.Staging, g.Production, time.Now().Format("2006-01-02"))
	pr, err := gh.CreatePR(ctx, title, notes, branch, g.Production)
	if err != nil {
		git.RunEnv(ctx, p.Root, []string{release.PromoteEnv + "=1"}, "push", "--quiet", "--delete", g.Remote, branch)
		return err
	}
	os.Remove(filepath.Join(p.Root, release.NotesPath))
	output.Section(out, "📬 RELEASE PULL REQUEST")
	fmt.Fprintf(out, "%s\nSnapshot branch: %s at %s. %s/%s has not moved.\n", pr.URL, branch, pl.To[:7], g.Remote, g.Production)
	output.Instruction(out, "Give the user the pull request link for review. When they say it is approved, run `mem promote production --confirm` to release it. Never use GitHub's merge button: it would break the linear history.")
	return nil
}

// promotionPR handles a project whose release pull request is already open: it reports it, and with confirm
// completes it. It returns false when no promotion pull request is open.
func promotionPR(ctx context.Context, out io.Writer, p project.Project, gh *github.Client, confirm bool) (bool, error) {
	g := p.Config.Git
	open, err := gh.OpenPRs(ctx, g.Production, promotionPrefix(p))
	if err != nil || len(open) == 0 {
		return false, err
	}
	pr := open[0]
	review, err := gh.Reviews(ctx, pr.Number)
	if err != nil {
		return true, err
	}
	output.Section(out, "📬 RELEASE PULL REQUEST")
	fmt.Fprintf(out, "%s (#%d)\nHead: %s at %s. Approvals: %d.", pr.URL, pr.Number, pr.HeadRef, pr.HeadSHA[:7], review.Approvals)
	if review.ChangesRequested {
		fmt.Fprint(out, " Changes requested.")
	}
	fmt.Fprintln(out)
	if !confirm {
		output.Instruction(out, "A release pull request is open. When the user says it is approved, run `mem promote production --confirm` to release it by fast-forward. To release something else, close it on GitHub first.")
		return true, nil
	}
	if review.ChangesRequested {
		return true, fmt.Errorf("changes were requested on #%d; address them and get it approved, or close it and start a new release", pr.Number)
	}
	if head, err := git.Run(ctx, p.Root, "ls-remote", g.Remote, "refs/heads/"+pr.HeadRef); err != nil || len(head) < 40 || head[:40] != pr.HeadSHA {
		return true, fmt.Errorf("the snapshot branch %s no longer matches #%d's head; close the pull request and start a new release", pr.HeadRef, pr.Number)
	}
	pl, err := release.Prepare(ctx, p, "production", pr.HeadSHA)
	if err != nil {
		return true, err
	}
	switch {
	case pl.UpToDate || pl.BehindStage:
		return true, fmt.Errorf("%s/%s already contains #%d; close it on GitHub", g.Remote, g.Production, pr.Number)
	case len(pl.Diverged) > 0:
		return true, divergedError(p, pl)
	}
	notes := releaseHeading.ReplaceAllString(pr.Body, "# Release "+pl.Tag)
	if notes == "" {
		notes = release.GeneratedMessage(ctx, p, pl)
	}
	if err := release.Execute(ctx, p, pl, notes); err != nil {
		return true, err
	}
	renderPlan(out, p, pl)
	output.Section(out, "✅ PROMOTED")
	fmt.Fprintf(out, "Released %s: %s/%s is now at %s, the head of #%d.\n", pl.Tag, g.Remote, g.Production, pl.To[:7], pr.Number)
	// GitHub marks the pull request merged once it sees its head on the base branch. Delete the snapshot branch
	// only after that, or GitHub may close the pull request as unmerged.
	if !waitMerged(ctx, gh, pr.Number) {
		gh.Comment(ctx, pr.Number, fmt.Sprintf("Released as %s: `%s` was fast-forwarded to this pull request's head by `mem promote production`.", pl.Tag, g.Production))
		gh.Close(ctx, pr.Number)
		fmt.Fprintf(out, "GitHub did not mark #%d as merged, so mem closed it with a comment naming %s.\n", pr.Number, pl.Tag)
	}
	git.RunEnv(ctx, p.Root, []string{release.PromoteEnv + "=1"}, "push", "--quiet", "--delete", g.Remote, pr.HeadRef)
	output.Instruction(out, "Tell the user the release is out. Where CI deploys production, check the deployment before reporting it as live.")
	return true, nil
}

// mergeWait is how long mem waits for GitHub to mark a fast-forwarded pull request as merged.
var mergeWait = 15 * time.Second

func waitMerged(ctx context.Context, gh *github.Client, number int) bool {
	deadline := time.Now().Add(mergeWait)
	for {
		if pr, err := gh.PR(ctx, number); err == nil && pr.Merged {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Second)
	}
}
