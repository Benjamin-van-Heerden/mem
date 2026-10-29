// Package release fast-forwards the staging and production branches along the
// development history, and reports how far each stage lags behind.
package release

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// PromoteEnv marks pushes made by mem promote so the pre-push hook allows them.
const PromoteEnv = "MEM_PROMOTE"

const networkTimeout = 60 * time.Second

type Commit struct {
	Hash    string
	Author  string
	Subject string
}

type Plan struct {
	Stage       string
	Branch      string
	Source      string
	From        string
	To          string
	Commits     []Commit
	Specs       []string
	Unpushed    int
	Tag         string
	UpToDate    bool
	Diverged    []Commit
	BehindStage bool
}

func remoteRef(p project.Project, branch string) string {
	return p.Config.Git.Remote + "/" + branch
}

// SingleBranchError refuses a release in a project without staging and production branches.
func SingleBranchError(p project.Project) error {
	return fmt.Errorf("this project is single-branch ([git] single_branch = true), so mem makes no releases and pushes no tags here. Work on %s reaches everyone through `mem sync`", p.Config.Git.Development)
}

// Prepare fetches and works out what promoting stage would ship. to optionally
// selects an earlier commit of the source branch.
func Prepare(ctx context.Context, p project.Project, stage, to string) (Plan, error) {
	if p.Config.Git.SingleBranch {
		return Plan{}, SingleBranchError(p)
	}
	branch, source, ok := p.Config.Git.Stage(stage)
	if !ok {
		return Plan{}, fmt.Errorf("unknown stage %q; use staging or production", stage)
	}
	pl := Plan{Stage: stage, Branch: branch, Source: source}
	remote := p.Config.Git.Remote
	fetchCtx, cancel := context.WithTimeout(ctx, networkTimeout)
	defer cancel()
	if _, err := git.Run(fetchCtx, p.Root, "fetch", "--prune", "--tags", "--quiet", remote); err != nil {
		return pl, fmt.Errorf("could not fetch from %s, so promotion cannot check what is published: %w", remote, err)
	}

	sourceTip, err := commitOf(ctx, p.Root, remoteRef(p, source))
	if err != nil {
		return pl, fmt.Errorf("%s/%s does not exist; push %s before promoting", remote, source, source)
	}
	pl.To = sourceTip
	if to != "" {
		if pl.To, err = commitOf(ctx, p.Root, to); err != nil {
			return pl, fmt.Errorf("unknown commit %q", to)
		}
		if !isAncestor(ctx, p.Root, pl.To, sourceTip) {
			return pl, fmt.Errorf("%s is not on %s/%s; staging and production only move along the %s history", to, remote, source, source)
		}
	}
	if pl.From, err = commitOf(ctx, p.Root, remoteRef(p, branch)); err == nil {
		if pl.From == pl.To {
			pl.UpToDate = true
			return pl, nil
		}
		if !isAncestor(ctx, p.Root, pl.From, pl.To) {
			if isAncestor(ctx, p.Root, pl.To, pl.From) {
				pl.BehindStage = true
				return pl, nil
			}
			// Merges count here: a stage branch whose only extra commit is a merge has still diverged.
			pl.Diverged, err = commits(ctx, p.Root, sourceTip+".."+pl.From, true)
			return pl, err
		}
	}
	rangeSpec := pl.To
	if pl.From != "" {
		rangeSpec = pl.From + ".." + pl.To
	}
	if pl.Commits, err = commits(ctx, p.Root, rangeSpec, false); err != nil {
		return pl, err
	}
	pl.Specs = completedSpecs(ctx, p.Root, pl.From, pl.To)
	if current := git.CurrentBranch(ctx, p.Root); current == source {
		if out, err := git.Run(ctx, p.Root, "rev-list", "--count", remoteRef(p, source)+"..HEAD"); err == nil {
			pl.Unpushed, _ = strconv.Atoi(out)
		}
	}
	if stage == "production" {
		pl.Tag, err = nextTag(ctx, p.Root)
	}
	return pl, err
}

// Execute pushes the stage branch, and for production an annotated tag with the notes, in one atomic push. It then
// brings the local copy of the stage branch along, and returns a line about it for the user, empty when there is none.
func Execute(ctx context.Context, p project.Project, pl Plan, notes string) (string, error) {
	if err := push(ctx, p, pl, notes); err != nil {
		return "", err
	}
	return followLocal(ctx, p.Root, pl.Branch, pl.To), nil
}

func push(ctx context.Context, p project.Project, pl Plan, notes string) error {
	if pl.UpToDate || pl.BehindStage || len(pl.Diverged) > 0 {
		return errors.New("nothing to promote")
	}
	refs := []string{pl.To + ":refs/heads/" + pl.Branch}
	if pl.Tag != "" {
		if strings.TrimSpace(notes) == "" {
			return errors.New("production releases need release notes")
		}
		// Release notes are Markdown: the default cleanup would strip their "#" headings as comments.
		if _, err := git.Run(ctx, p.Root, "tag", "--annotate", "--cleanup=whitespace", pl.Tag, pl.To, "--message", notes); err != nil {
			return err
		}
		refs = append(refs, "refs/tags/"+pl.Tag)
	}
	pushCtx, cancel := context.WithTimeout(ctx, networkTimeout)
	defer cancel()
	// Never forced: a stage branch only fast-forwards, and a branch someone else moved meanwhile rejects the push.
	args := append([]string{"push", "--atomic", "--quiet", p.Config.Git.Remote}, refs...)
	if _, err := git.RunEnv(pushCtx, p.Root, []string{PromoteEnv + "=1"}, args...); err != nil {
		if pl.Tag != "" {
			git.Run(ctx, p.Root, "tag", "--delete", pl.Tag)
		}
		return fmt.Errorf("push failed and nothing was promoted; if someone else just promoted, run the command again: %w", err)
	}
	return nil
}

// followLocal fast-forwards the local copy of a released branch to what was pushed, so it does not fall behind the
// remote. It leaves a branch checked out in a worktree, whose files would change under someone, and one with commits
// the release does not contain. A project without a local copy keeps having none.
func followLocal(ctx context.Context, root, branch, to string) string {
	ref := "refs/heads/" + branch
	local, err := commitOf(ctx, root, ref)
	if err != nil || local == to {
		return ""
	}
	if checkedOut(ctx, root, ref) {
		return fmt.Sprintf("⚠️ The local %s branch is checked out, so it was left at %s. Switch back to the development branch; releases never need %s checked out.", branch, local[:7], branch)
	}
	if !isAncestor(ctx, root, local, to) {
		return fmt.Sprintf("⚠️ The local %s branch has commits the release does not contain, so it was left at %s. Tell the user: %s only moves through releases.", branch, local[:7], branch)
	}
	if _, err := git.Run(ctx, root, "update-ref", ref, to, local); err != nil {
		return fmt.Sprintf("⚠️ Could not move the local %s branch to the release: %v", branch, err)
	}
	return fmt.Sprintf("The local %s branch is at %s too.", branch, to[:7])
}

func checkedOut(ctx context.Context, root, ref string) bool {
	out, err := git.Run(ctx, root, "worktree", "list", "--porcelain")
	return err == nil && slices.Contains(strings.Split(out, "\n"), "branch "+ref)
}

// nextTag returns today's next date tag, e.g. v2026.09.24.2.
func nextTag(ctx context.Context, root string) (string, error) {
	prefix := "v" + time.Now().Format("2006.01.02") + "."
	out, err := git.Run(ctx, root, "tag", "--list", prefix+"*")
	if err != nil {
		return "", err
	}
	next := 1
	for _, tag := range strings.Fields(out) {
		if n, err := strconv.Atoi(strings.TrimPrefix(tag, prefix)); err == nil && n >= next {
			next = n + 1
		}
	}
	return prefix + strconv.Itoa(next), nil
}

// completedSpecs lists the slugs of specs archived as completed within the promoted range.
func completedSpecs(ctx context.Context, root, from, to string) []string {
	var slugs []string
	for _, s := range specSummaries(ctx, root, from, to) {
		slugs = append(slugs, s.Slug)
	}
	return slugs
}

func commits(ctx context.Context, root, rangeSpec string, merges bool) ([]Commit, error) {
	args := []string{"log", "--format=%h%x09%an%x09%s", rangeSpec}
	if !merges {
		args = append(args, "--no-merges")
	}
	out, err := git.Run(ctx, root, args...)
	if err != nil {
		return nil, err
	}
	var list []Commit
	for _, line := range strings.Split(out, "\n") {
		if fields := strings.SplitN(line, "\t", 3); len(fields) == 3 {
			list = append(list, Commit{Hash: fields[0], Author: fields[1], Subject: fields[2]})
		}
	}
	return list, nil
}

func commitOf(ctx context.Context, root, ref string) (string, error) {
	return git.Run(ctx, root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
}

func isAncestor(ctx context.Context, root, ancestor, descendant string) bool {
	_, err := git.Run(ctx, root, "merge-base", "--is-ancestor", ancestor, descendant)
	return err == nil
}

type Status struct {
	Tag          string
	TagAge       string
	StagingAhead int
	DevAhead     int
	// OldestUnreleased is how long ago the oldest commit on development not yet on staging was made, e.g. "3 weeks ago".
	OldestUnreleased string
	// StagingOutside and ProductionOutside count commits on those branches that development lacks, such as a
	// change merged into production directly. Promotion cannot fast-forward past them.
	StagingOutside    int
	ProductionOutside int
	Missing           []string
}

// CurrentStatus reports the latest production tag, how far staging and development are ahead, and commits on
// staging or production outside development, from the remote-tracking refs of the last fetch.
func CurrentStatus(ctx context.Context, p project.Project) Status {
	g := p.Config.Git
	var st Status
	for _, b := range []string{g.Development, g.Staging, g.Production} {
		if _, err := commitOf(ctx, p.Root, remoteRef(p, b)); err != nil {
			st.Missing = append(st.Missing, b)
		}
	}
	if len(st.Missing) > 0 {
		return st
	}
	st.StagingAhead = count(ctx, p.Root, remoteRef(p, g.Production)+".."+remoteRef(p, g.Staging))
	st.DevAhead = count(ctx, p.Root, remoteRef(p, g.Staging)+".."+remoteRef(p, g.Development))
	if st.DevAhead > 0 {
		if out, err := git.Run(ctx, p.Root, "log", "--reverse", "--format=%cr", remoteRef(p, g.Staging)+".."+remoteRef(p, g.Development)); err == nil {
			st.OldestUnreleased, _, _ = strings.Cut(out, "\n")
		}
	}
	st.StagingOutside = count(ctx, p.Root, remoteRef(p, g.Development)+".."+remoteRef(p, g.Staging))
	st.ProductionOutside = count(ctx, p.Root, remoteRef(p, g.Development)+".."+remoteRef(p, g.Production))
	if tag, err := git.Run(ctx, p.Root, "describe", "--tags", "--abbrev=0", "--match", "v*", remoteRef(p, g.Production)); err == nil {
		st.Tag = tag
		st.TagAge, _ = git.Run(ctx, p.Root, "for-each-ref", "--format=%(creatordate:relative)", "refs/tags/"+tag)
	}
	return st
}

func count(ctx context.Context, root, rangeSpec string) int {
	out, err := git.Run(ctx, root, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(out)
	return n
}
