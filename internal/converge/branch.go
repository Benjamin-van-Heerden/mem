package converge

import (
	"context"
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// feature reports whether r is on a branch that follows development: a named branch other than development,
// staging and production, with the remote development branch known.
func feature(r Report) bool {
	return r.Branch != "" && r.Branch != r.Dev && r.Branch != r.Staging && r.Branch != r.Production && r.HasRemote
}

// followDevelopment rebases a feature branch onto a development branch that has moved on, so the branch never
// drifts far and merging it back is a fast-forward. A branch already pushed is force-pushed with a lease, but only
// when every commit on its remote copy outside development is the user's; a branch others commit to is never
// rewritten. A rebase that would conflict is aborted and reported, leaving the branch as it was.
func followDevelopment(ctx context.Context, p project.Project, r Report) Report {
	r.devHandled = true
	behind := r.DevBehind
	if r.Dirty {
		r.Nudges = append(r.Nudges, fmt.Sprintf("%s is %d commit(s) behind %s, and uncommitted changes prevent rebasing it. Commit the work in progress, then run `mem sync` to bring %s in.", r.Branch, behind, r.Development, r.Dev))
		return r
	}
	published := r.Upstream != "" && countRange(ctx, p.Root, r.Development+".."+r.Upstream) > 0
	if published {
		if others := otherAuthors(ctx, p.Root, r.Development+".."+r.Upstream); len(others) > 0 {
			r.Nudges = append(r.Nudges, fmt.Sprintf("%s is %d commit(s) behind %s, but %s also has commits by %s, so mem does not rewrite it. Tell the user; bring %s in with them (`git merge %s` keeps everyone's history intact).", r.Branch, behind, r.Development, r.Branch, strings.Join(others, ", "), r.Dev, r.Development))
			return r
		}
	}
	lease, _ := git.Run(ctx, p.Root, "rev-parse", "--verify", "--quiet", r.Upstream)
	if _, err := git.Run(ctx, p.Root, "rebase", "--quiet", r.Development); err != nil {
		conflicts, _ := git.Run(ctx, p.Root, "diff", "--name-only", "--diff-filter=U")
		git.Run(ctx, p.Root, "rebase", "--abort")
		files := strings.Join(strings.Fields(conflicts), ", ")
		if files == "" {
			files = "files Git did not name"
		}
		r.Nudges = append(r.Nudges, fmt.Sprintf("%s is %d commit(s) behind %s, and rebasing it would conflict in %s, so mem left it unchanged. Tell the user; resolve it while the conflict is small: `git rebase %s`, fix the conflicts, `git rebase --continue`.", r.Branch, behind, r.Development, files, r.Development))
		return r
	}
	r.Done = append(r.Done, fmt.Sprintf("Rebased %s onto %s (%d new commit(s)). Run the tests before continuing: code can break without a Git conflict.", r.Branch, r.Development, behind))
	if published {
		if _, err := git.Run(ctx, p.Root, "push", "--quiet", "--force-with-lease=refs/heads/"+r.Branch+":"+lease, r.Remote, "HEAD:refs/heads/"+r.Branch); err != nil {
			r.Nudges = append(r.Nudges, fmt.Sprintf("Pushing the rebased %s was refused (%s): someone pushed to it in the meantime, so nothing was overwritten. Tell the user; run `mem sync` to bring their commits in.", r.Branch, firstLine(err.Error())))
		} else {
			r.Done = append(r.Done, fmt.Sprintf("Pushed the rebased %s to %s (with a lease, so nothing unseen was overwritten).", r.Branch, r.Remote))
		}
	}
	updated := inspect(ctx, p)
	updated.Fetched, updated.Done, updated.Nudges, updated.Before, updated.After, updated.devHandled = r.Fetched, r.Done, r.Nudges, r.Before, r.After, true
	return updated
}

// rewritten reports whether the upstream moved to a history that no longer contains its previous commit, as when
// a teammate's or another checkout's mem rebased and force-pushed the branch.
func rewritten(ctx context.Context, root string, r Report) bool {
	if r.Before == "" || r.After == "" || r.Before == r.After {
		return false
	}
	_, err := git.Run(ctx, root, "merge-base", "--is-ancestor", r.Before, r.After)
	return err != nil
}

// otherAuthors lists the author emails in rangeSpec other than the configured user's.
func otherAuthors(ctx context.Context, root, rangeSpec string) []string {
	email := git.UserEmail(ctx, root)
	out, err := git.Run(ctx, root, "log", "--format=%ae", rangeSpec)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var others []string
	for _, author := range strings.Fields(out) {
		if !strings.EqualFold(author, email) && !seen[author] {
			seen[author] = true
			others = append(others, author)
		}
	}
	return others
}

func countRange(ctx context.Context, root, rangeSpec string) int {
	out, err := git.Run(ctx, root, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0
	}
	var n int
	fmt.Sscanf(out, "%d", &n)
	return n
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
