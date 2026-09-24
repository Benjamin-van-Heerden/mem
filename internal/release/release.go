// Package release fast-forwards the staging and production branches along the
// development history, and reports how far each stage lags behind.
package release

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

// PromoteEnv marks pushes made by memr promote so the pre-push hook allows them.
const PromoteEnv = "MEMR_PROMOTE"

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

// Prepare fetches and works out what promoting stage would ship. to optionally
// selects an earlier commit of the source branch.
func Prepare(ctx context.Context, p project.Project, stage, to string) (Plan, error) {
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
			pl.Diverged, err = commits(ctx, p.Root, sourceTip+".."+pl.From)
			return pl, err
		}
	}
	rangeSpec := pl.To
	if pl.From != "" {
		rangeSpec = pl.From + ".." + pl.To
	}
	if pl.Commits, err = commits(ctx, p.Root, rangeSpec); err != nil {
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

// Execute pushes the stage branch, and for production an annotated tag with the notes, in one atomic push.
func Execute(ctx context.Context, p project.Project, pl Plan, notes string) error {
	if pl.UpToDate || pl.BehindStage || len(pl.Diverged) > 0 {
		return errors.New("nothing to promote")
	}
	refs := []string{pl.To + ":refs/heads/" + pl.Branch}
	if pl.Tag != "" {
		if strings.TrimSpace(notes) == "" {
			return errors.New("production releases need release notes")
		}
		if _, err := git.Run(ctx, p.Root, "tag", "--annotate", pl.Tag, pl.To, "--message", notes); err != nil {
			return err
		}
		refs = append(refs, "refs/tags/"+pl.Tag)
	}
	pushCtx, cancel := context.WithTimeout(ctx, networkTimeout)
	defer cancel()
	args := append([]string{"push", "--atomic", "--quiet", "--force-with-lease=refs/heads/" + pl.Branch + ":" + pl.From, p.Config.Git.Remote}, refs...)
	if _, err := git.RunEnv(pushCtx, p.Root, []string{PromoteEnv + "=1"}, args...); err != nil {
		if pl.Tag != "" {
			git.Run(ctx, p.Root, "tag", "--delete", pl.Tag)
		}
		return fmt.Errorf("push failed and nothing was promoted; if someone else just promoted, run the command again: %w", err)
	}
	return nil
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

// completedSpecs finds specs archived as completed within the promoted range.
func completedSpecs(ctx context.Context, root, from, to string) []string {
	args := []string{"ls-tree", "-r", "--name-only", to, "--", ".memr/specs/archive"}
	if from != "" {
		args = []string{"diff", "--name-only", "--diff-filter=A", from, to, "--", ".memr/specs/archive"}
	}
	out, err := git.Run(ctx, root, args...)
	if err != nil {
		return nil
	}
	var specs []string
	for _, file := range strings.Split(out, "\n") {
		if path.Base(file) != "spec.md" {
			continue
		}
		content, err := git.Run(ctx, root, "show", to+":"+file)
		if err != nil || !strings.Contains(content, "\nstatus: completed\n") {
			continue
		}
		specs = append(specs, path.Base(path.Dir(file)))
	}
	return specs
}

func commits(ctx context.Context, root, rangeSpec string) ([]Commit, error) {
	out, err := git.Run(ctx, root, "log", "--no-merges", "--format=%h%x09%an%x09%s", rangeSpec)
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
	Missing      []string
}

// CurrentStatus reports the latest production tag and how far staging and
// development are ahead, from the remote-tracking refs of the last fetch.
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
