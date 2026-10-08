package converge

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

// featureBranch starts a branch named feature in p with one commit of its own.
func featureBranch(t *testing.T, p string) {
	t.Helper()
	run(t, p, "switch", "--quiet", "-c", "feature")
	write(t, p, "feature.txt", "feature\n")
	run(t, p, "add", ".")
	run(t, p, "commit", "--quiet", "-m", "feature work")
}

func TestSyncRebasesAnUnpushedFeatureBranchOntoNewDevelopment(t *testing.T) {
	other, p := clones(t)
	featureBranch(t, p.Root)
	pushChange(t, other, "theirs.txt", "theirs\n")

	r := Sync(context.Background(), p)
	if !strings.Contains(strings.Join(r.Done, "\n"), "Rebased feature onto origin/dev (1 new commit(s)). Run the tests") || nudged(r, "behind") {
		t.Fatalf("report = %+v", r)
	}
	if _, err := runErr(p.Root, "merge-base", "--is-ancestor", "origin/dev", "HEAD"); err != nil {
		t.Fatal("feature does not contain the new development commit")
	}
}

func TestSyncForcePushesTheUsersOwnFeatureBranchAndOtherCheckoutsCatchUp(t *testing.T) {
	ctx := context.Background()
	other, p := clones(t)
	featureBranch(t, p.Root)
	run(t, p.Root, "push", "--quiet", "-u", "origin", "feature")

	// A second checkout of the same person, holding the branch and a commit of its own.
	laptop := filepath.Join(filepath.Dir(p.Root), "laptop")
	run(t, filepath.Dir(p.Root), "clone", "--quiet", "-b", "feature", filepath.Join(filepath.Dir(p.Root), "remote.git"), laptop)
	identify(t, laptop)
	write(t, laptop, "laptop.txt", "laptop\n")
	run(t, laptop, "add", ".")
	run(t, laptop, "commit", "--quiet", "-m", "laptop work")

	pushChange(t, other, "theirs.txt", "theirs\n")
	r := Sync(ctx, p)
	if !strings.Contains(strings.Join(r.Done, "\n"), "Pushed the rebased feature to origin") || nudged(r, "behind") || nudged(r, "refused") {
		t.Fatalf("report = %+v", r)
	}
	if run(t, p.Root, "rev-parse", "origin/feature") != run(t, p.Root, "rev-parse", "HEAD") {
		t.Fatal("origin/feature is not the rebased branch")
	}

	lp := p
	lp.Root = laptop
	lr := Sync(ctx, lp)
	log := run(t, laptop, "log", "--format=%s", "origin/dev..HEAD")
	if log != "laptop work\nfeature work" {
		t.Fatalf("laptop history after catching up:\n%s\nreport = %+v", log, lr)
	}
}

func TestSyncLeavesAFeatureBranchUnchangedWhenTheRebaseWouldConflict(t *testing.T) {
	other, p := clones(t)
	run(t, p.Root, "switch", "--quiet", "-c", "feature")
	write(t, p.Root, "shared.txt", "mine\n")
	run(t, p.Root, "commit", "--quiet", "-am", "mine")
	before := run(t, p.Root, "rev-parse", "HEAD")
	pushChange(t, other, "shared.txt", "theirs\n")

	r := Sync(context.Background(), p)
	if !nudged(r, "would conflict in shared.txt, so mem left it unchanged") || len(r.Done) != 0 {
		t.Fatalf("report = %+v", r)
	}
	if after := run(t, p.Root, "rev-parse", "HEAD"); after != before {
		t.Fatal("the conflicting rebase changed the branch")
	}
}

func TestSyncDoesNotRewriteAFeatureBranchOthersCommittedTo(t *testing.T) {
	other, p := clones(t)
	featureBranch(t, p.Root)
	run(t, p.Root, "push", "--quiet", "-u", "origin", "feature")
	run(t, other, "fetch", "--quiet")
	run(t, other, "switch", "--quiet", "feature")
	write(t, other, "ann.txt", "ann\n")
	run(t, other, "add", ".")
	run(t, other, "-c", "user.email=ann@example.com", "commit", "--quiet", "-m", "ann's work")
	run(t, other, "push", "--quiet")
	run(t, other, "switch", "--quiet", "dev")
	pushChange(t, other, "theirs.txt", "theirs\n")

	r := Sync(context.Background(), p)
	if !nudged(r, "also has commits by ann@example.com, so mem does not rewrite it") || strings.Contains(strings.Join(r.Done, "\n"), "Rebased") {
		t.Fatalf("report = %+v", r)
	}
	if run(t, p.Root, "rev-parse", "origin/feature") != run(t, p.Root, "rev-parse", "HEAD") {
		t.Fatal("the shared branch was rewritten")
	}
}

func runErr(dir string, args ...string) (string, error) {
	return git.Run(context.Background(), dir, args...)
}

func nudged(r Report, text string) bool {
	return strings.Contains(strings.Join(r.Nudges, "\n"), text)
}

func TestSyncDoesNotRestoreACommitDroppedFromARewrittenBranch(t *testing.T) {
	ctx := context.Background()
	_, p := clones(t)
	featureBranch(t, p.Root)
	write(t, p.Root, "dropped.txt", "dropped\n")
	run(t, p.Root, "add", ".")
	run(t, p.Root, "commit", "--quiet", "-m", "dropped work")
	run(t, p.Root, "push", "--quiet", "-u", "origin", "feature")

	laptop := filepath.Join(filepath.Dir(p.Root), "laptop")
	run(t, filepath.Dir(p.Root), "clone", "--quiet", "-b", "feature", filepath.Join(filepath.Dir(p.Root), "remote.git"), laptop)
	identify(t, laptop)
	write(t, laptop, "laptop.txt", "laptop\n")
	run(t, laptop, "add", ".")
	run(t, laptop, "commit", "--quiet", "-m", "laptop work")

	// Rewritten elsewhere: the last commit is dropped and the branch force-pushed.
	run(t, p.Root, "reset", "--quiet", "--hard", "HEAD~1")
	run(t, p.Root, "push", "--quiet", "--force", "origin", "feature")

	lp := p
	lp.Root = laptop
	Sync(ctx, lp)
	if log := run(t, laptop, "log", "--format=%s", "origin/dev..HEAD"); log != "laptop work\nfeature work" {
		t.Fatalf("laptop history after catching up with the rewritten branch:\n%s", log)
	}
}
