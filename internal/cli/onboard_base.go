package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// baseLine describes a development branch built on released code from [git] base: how much of the base it lacks
// and the merge-only rule, since rebasing such a published, deployed branch would need a force-push. It returns ""
// when no base is set or another branch is checked out, and the count of base commits not merged in yet.
func baseLine(ctx context.Context, p project.Project, branch string) (string, int) {
	g := p.Config.Git
	if g.Base == "" || branch != g.Development {
		return "", 0
	}
	ref := g.Remote + "/" + g.Base
	rule := fmt.Sprintf("Released commits come in only when the user asks, by a merge commit: `git fetch %s && git merge --no-ff %s`, then `git push`. Never rebase %s, never force-push it, and never merge unreleased work into it.", g.Remote, ref, g.Development)
	out, err := git.Run(ctx, p.Root, "rev-list", "--count", "HEAD.."+ref)
	if err != nil {
		return fmt.Sprintf("Base: %s is built on released code from %s (%s is not in this checkout yet). %s", g.Development, g.Base, ref, rule), 0
	}
	missing, _ := strconv.Atoi(out)
	if missing == 0 {
		return fmt.Sprintf("Base: %s is built on released code from %s and has all of it. %s", g.Development, g.Base, rule), 0
	}
	return fmt.Sprintf("Base: %s is built on released code from %s; %d commit(s) of %s are not merged in yet. %s", g.Development, g.Base, missing, ref, rule), missing
}
