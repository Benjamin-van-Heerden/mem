package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// leaveGuardedBranch switches a checkout from staging or production to the development branch before onboard
// commits anything. A fresh clone checks out the remote's default branch, which may be production, and onboard's
// own commits would otherwise land and be pushed there, ahead of development. It returns lines for 🔄 UPDATES, and
// stuck when the checkout is still on the guarded branch, so onboard must not commit or push. Detached checkouts
// and single-branch projects are left to the usual nudges.
func leaveGuardedBranch(ctx context.Context, p project.Project) (lines []string, stuck bool) {
	g := p.Config.Git
	branch := git.CurrentBranch(ctx, p.Root)
	if g.SingleBranch || !g.Guarded(branch) {
		return nil, false
	}
	// Commits no remote branch has would be easy to forget once the checkout is elsewhere.
	if n := unpublishedCommits(ctx, p); n > 0 {
		lines = append(lines, fmt.Sprintf("⚠️ The local %s branch has %d commit(s) that are not on %s. They stay on %s, which only moves through `mem promote`. Tell the user; if they belong in the shared history, bring them into %s.", branch, n, g.Remote, branch, g.Development))
	}
	args := []string{"switch", "--quiet", g.Development}
	if !refExists(ctx, p.Root, "refs/heads/"+g.Development) && refExists(ctx, p.Root, "refs/remotes/"+g.Remote+"/"+g.Development) {
		args = []string{"switch", "--quiet", "--create", g.Development, "--track", g.Remote + "/" + g.Development}
	}
	if _, err := git.Run(ctx, p.Root, args...); err != nil {
		return append(lines, fmt.Sprintf("⚠️ This checkout is on %s, which only moves through `mem promote`, and could not switch to %s (%v). Onboard committed and pushed nothing. Tell the user, and resolve it with them: commit or stash the conflicting changes, run `git switch %s`, then run `mem onboard` again.", branch, g.Development, err, g.Development)), true
	}
	return append(lines, fmt.Sprintf("Switched from %s to %s: %s only moves through `mem promote`, and work happens on %s.", branch, g.Development, branch, g.Development)), false
}

// leaveDeletedBranch follows the development branch after the one this checkout is on was deleted from the remote,
// as when a project becomes single-branch on its production branch. The checkout's own config still names the deleted
// branch, so the new one is read from the config on the remote's default branch. It switches only when every local
// commit is on the remote. Otherwise, or with nowhere to go, it reports stuck: onboard's commits could then only be
// shared by publishing the deleted branch again. switched means the checkout is now on the new development branch.
func leaveDeletedBranch(ctx context.Context, p project.Project, r converge.Report) (lines []string, switched, stuck bool) {
	g := p.Config.Git
	if !r.Fetched || r.Branch != g.Development || r.Gone != g.Remote+"/"+g.Development {
		return nil, false, false
	}
	dev := g.Development
	target := movedDevelopment(ctx, p)
	if target == "" {
		return []string{fmt.Sprintf("⚠️ %s, the development branch, has been deleted from %s, and the config on %s's default branch names no other. Onboard committed and pushed nothing. Tell the user, and ask whether %s should be published again or where work continues.", dev, g.Remote, g.Remote, dev)}, false, true
	}
	if n := unpublishedCommits(ctx, p); n > 0 {
		return []string{fmt.Sprintf("⚠️ %s has been deleted from %s, which now develops on %s, but the local %s branch has %d commit(s) that are not on %s, so onboard stayed on %s and committed and pushed nothing. Tell the user, and bring those commits into %s with them (`git log %s --not --remotes=%s` lists them; cherry-pick them onto %s), then run `mem onboard` again.", dev, g.Remote, target, dev, n, g.Remote, dev, target, dev, g.Remote, target)}, false, true
	}
	args := []string{"switch", "--quiet", target}
	if !refExists(ctx, p.Root, "refs/heads/"+target) {
		args = []string{"switch", "--quiet", "--create", target, "--track", g.Remote + "/" + target}
	}
	if _, err := git.Run(ctx, p.Root, args...); err != nil {
		return []string{fmt.Sprintf("⚠️ %s has been deleted from %s, which now develops on %s, and switching to %s failed (%v). Onboard committed and pushed nothing. Tell the user, and resolve it with them: commit or stash the conflicting changes, run `git switch %s`, then run `mem onboard` again.", dev, g.Remote, target, target, err, target)}, false, true
	}
	return []string{fmt.Sprintf("Switched from %s to %s: %s was deleted from %s, whose default branch names %s as the development branch. Every commit on the local %s is on %s; delete it with `git branch -D %s`.", dev, target, dev, g.Remote, target, dev, g.Remote, dev)}, true, false
}

// movedDevelopment is the development branch named by the config on the remote's default branch, when that differs
// from this checkout's and exists on the remote, or "".
func movedDevelopment(ctx context.Context, p project.Project) string {
	g := p.Config.Git
	head, err := git.Run(ctx, p.Root, "symbolic-ref", "--short", "refs/remotes/"+g.Remote+"/HEAD")
	if err != nil {
		return ""
	}
	data, err := git.Run(ctx, p.Root, "show", head+":.mem/config.toml")
	if err != nil {
		return ""
	}
	config, err := project.ParseConfig([]byte(data))
	if err != nil || config.Git.Development == "" || config.Git.Development == g.Development || !refExists(ctx, p.Root, "refs/remotes/"+g.Remote+"/"+config.Git.Development) {
		return ""
	}
	return config.Git.Development
}

// unpublishedCommits counts the commits on HEAD that no branch on the remote has.
func unpublishedCommits(ctx context.Context, p project.Project) int {
	out, err := git.Run(ctx, p.Root, "rev-list", "--count", "HEAD", "--not", "--remotes="+p.Config.Git.Remote)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(out)
	return n
}
