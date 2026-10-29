package cli

import (
	"context"
	"fmt"
	"strconv"

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
	if out, err := git.Run(ctx, p.Root, "rev-list", "--count", "HEAD", "--not", "--remotes="+g.Remote); err == nil {
		if n, _ := strconv.Atoi(out); n > 0 {
			lines = append(lines, fmt.Sprintf("⚠️ The local %s branch has %d commit(s) that are not on %s. They stay on %s, which only moves through `mem promote`. Tell the user; if they belong in the shared history, bring them into %s.", branch, n, g.Remote, branch, g.Development))
		}
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
