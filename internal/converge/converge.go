// Package converge keeps a checkout close to the shared codebase: it applies
// safe Git updates automatically and produces nudges for everything else.
package converge

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
)

const (
	fetchTimeout = 20 * time.Second
	largeFiles   = 15
	largeLines   = 800
)

type Report struct {
	Remote      string
	HasRemote   bool
	Dev         string
	Staging     string
	Production  string
	Branch      string
	Upstream    string
	Development string
	Fetched     bool
	FetchError  string
	Dirty       bool
	Uncommitted structure.Drift
	Ahead       int
	Behind      int
	DevAhead    int
	DevBehind   int
	Done        []string
	Nudges      []string
}

// Sync fetches, fast-forwards or rebases where that is safe, and reports what remains.
func Sync(ctx context.Context, p project.Project) Report {
	remote := p.Config.Git.Remote
	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	_, fetchErr := git.Run(fetchCtx, p.Root, "fetch", "--prune", "--quiet", remote)

	r := inspect(ctx, p)
	if fetchErr != nil {
		r.FetchError, _, _ = strings.Cut(fetchErr.Error(), "\n")
		if r.HasRemote {
			r.Nudges = append(r.Nudges, fmt.Sprintf("Could not fetch from %s, so the state below may be stale (%s). Tell the user, and run `mem sync` once the remote is reachable.", remote, r.FetchError))
		}
		r.Nudges = append(r.Nudges, standingNudges(r, "run `mem sync`")...)
		return r
	}
	r.Fetched = true
	if r.Branch != "" && r.Upstream != "" && r.Behind > 0 {
		r = update(ctx, p, r)
	}
	r.Nudges = append(r.Nudges, standingNudges(r, "")...)
	return r
}

// Local inspects cached refs without fetching or changing anything.
func Local(ctx context.Context, p project.Project) Report {
	r := inspect(ctx, p)
	r.Nudges = standingNudges(r, "run `mem sync`")
	return r
}

func update(ctx context.Context, p project.Project, r Report) Report {
	behind, ahead := r.Behind, r.Ahead
	switch {
	case ahead == 0:
		if _, err := git.Run(ctx, p.Root, "merge", "--ff-only", "--quiet", r.Upstream); err != nil {
			r.Nudges = append(r.Nudges, fmt.Sprintf("%s is %d commit(s) behind %s but could not be fast-forwarded, most likely because uncommitted changes touch the same files. Tell the user; commit the work in progress, then run `mem sync`.", r.Branch, behind, r.Upstream))
			return r
		}
		r.Done = append(r.Done, fmt.Sprintf("Fast-forwarded %s by %d commit(s) from %s.", r.Branch, behind, r.Upstream))
	case r.Dirty:
		r.Nudges = append(r.Nudges, fmt.Sprintf("%s has diverged from %s (%d local, %d incoming commit(s)) and uncommitted changes prevent rebasing. Tell the user; commit the work in progress, then run `mem sync`.", r.Branch, r.Upstream, ahead, behind))
		return r
	default:
		if _, err := git.Run(ctx, p.Root, "rebase", "--quiet", r.Upstream); err != nil {
			git.Run(ctx, p.Root, "rebase", "--abort")
			r.Nudges = append(r.Nudges, fmt.Sprintf("%s has diverged from %s (%d local, %d incoming commit(s)). Rebasing hit conflicts, so it was aborted and nothing changed. Tell the user and resolve it together: `git rebase %s`, fix the conflicts, `git rebase --continue`.", r.Branch, r.Upstream, ahead, behind, r.Upstream))
			return r
		}
		r.Done = append(r.Done, fmt.Sprintf("Rebased %d local commit(s) onto %s, which had %d new commit(s).", ahead, r.Upstream, behind))
	}
	updated := inspect(ctx, p)
	updated.Fetched, updated.Done, updated.Nudges = r.Fetched, r.Done, r.Nudges
	return updated
}

func standingNudges(r Report, behindAction string) []string {
	var nudges []string
	dev := r.Dev
	if n := len(r.Uncommitted.Changes); n > largeFiles || r.Uncommitted.Lines > largeLines {
		nudges = append(nudges, fmt.Sprintf("Uncommitted changes span %d code file(s) and %d line(s). Commit the finished, coherent parts before continuing, so each commit stays reviewable.", n, r.Uncommitted.Lines))
	}
	switch {
	case r.Branch == "":
		return append(nudges, "HEAD is detached. Tell the user, and switch to a branch before making changes.")
	case !r.HasRemote:
		return append(nudges, fmt.Sprintf("This repository has no remote named %s, so nobody else can see its work. Tell the user; add the shared remote (`git remote add %s <url>`) or set the remote in .mem/config.toml.", r.Remote, r.Remote))
	case r.Upstream == "" && r.Branch == dev:
		nudges = append(nudges, fmt.Sprintf("%s is not on %s yet, so nobody else can see its commits. Tell the user, and publish it with `git push -u %s %s`.", dev, r.Remote, r.Remote, dev))
	case r.Upstream == "":
		nudges = append(nudges, fmt.Sprintf("%s has no remote counterpart, so nobody else can see its commits. Tell the user; publish it (`git push -u %s %s`) or integrate it into %s.", r.Branch, r.Remote, r.Branch, dev))
	case behindAction != "" && r.Behind > 0:
		nudges = append(nudges, fmt.Sprintf("%s is %d commit(s) behind %s (as of the last fetch). %s to update it before continuing.", r.Branch, r.Behind, r.Upstream, capitalize(behindAction)))
	case r.Ahead > 0 && r.Behind == 0:
		nudges = append(nudges, fmt.Sprintf("%d local commit(s) on %s are not pushed to %s. Tell the user, and push them at the next sensible point (`git push`) so everyone works on the same code.", r.Ahead, r.Branch, r.Upstream))
	}
	if r.Branch == r.Staging || r.Branch == r.Production {
		return append(nudges, fmt.Sprintf("You are on %s, which only moves by `mem promote`. Tell the user, and switch to %s (`git switch %s`) before making changes.", r.Branch, dev, dev))
	}
	if r.Branch != dev && r.DevBehind > 0 {
		nudges = append(nudges, fmt.Sprintf("You are on %s, not the development branch %s, and it is %d commit(s) behind %s. Tell the user this branch is drifting from the shared codebase; bring %s into it soon and merge it back into %s as early as possible.", r.Branch, dev, r.DevBehind, r.Development, r.Development, dev))
	} else if r.Branch != dev && r.DevAhead > 0 {
		nudges = append(nudges, fmt.Sprintf("You are on %s, which has %d commit(s) not yet in %s. Tell the user; merge it into %s soon so everyone works on the same code.", r.Branch, r.DevAhead, dev, dev))
	}
	return nudges
}

func inspect(ctx context.Context, p project.Project) Report {
	remote, dev := p.Config.Git.Remote, p.Config.Git.Development
	r := Report{Remote: remote, Dev: dev, Staging: p.Config.Git.Staging, Production: p.Config.Git.Production, Branch: git.CurrentBranch(ctx, p.Root), Development: remote + "/" + dev}
	status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--untracked-files=no")
	r.Dirty = status != ""
	r.Uncommitted, _ = structure.ChangesSince(ctx, p, "HEAD")
	_, err := git.Run(ctx, p.Root, "remote", "get-url", remote)
	r.HasRemote = err == nil
	if r.Branch == "" {
		return r
	}
	if upstream, err := git.Run(ctx, p.Root, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
		r.Upstream = upstream
	} else if refExists(ctx, p.Root, "refs/remotes/"+remote+"/"+r.Branch) {
		r.Upstream = remote + "/" + r.Branch
	}
	if r.Upstream != "" {
		r.Ahead, r.Behind = counts(ctx, p.Root, r.Upstream)
	}
	if r.Branch != dev && refExists(ctx, p.Root, "refs/remotes/"+r.Development) {
		r.DevAhead, r.DevBehind = counts(ctx, p.Root, r.Development)
	}
	return r
}

func counts(ctx context.Context, root, ref string) (ahead, behind int) {
	out, err := git.Run(ctx, root, "rev-list", "--left-right", "--count", "HEAD..."+ref)
	if err != nil {
		return 0, 0
	}
	fields := strings.Fields(out)
	if len(fields) == 2 {
		ahead, _ = strconv.Atoi(fields[0])
		behind, _ = strconv.Atoi(fields[1])
	}
	return ahead, behind
}

func refExists(ctx context.Context, root, ref string) bool {
	_, err := git.Run(ctx, root, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
