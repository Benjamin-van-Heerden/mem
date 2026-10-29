package release

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

func setup(t *testing.T) project.Project {
	t.Helper()
	base := t.TempDir()
	run(t, base, "init", "--quiet", "--bare", "remote.git")
	root := filepath.Join(base, "repo")
	run(t, base, "init", "--quiet", "-b", "dev", root)
	run(t, root, "config", "user.name", "Test")
	run(t, root, "config", "user.email", "test@example.com")
	run(t, root, "remote", "add", "origin", filepath.Join(base, "remote.git"))
	commit(t, root, "one")
	run(t, root, "push", "--quiet", "-u", "origin", "dev")
	return project.Project{Root: root, Config: project.Config{Git: project.GitConfig{Remote: "origin", Development: "dev", Staging: "test", Production: "main"}}}
}

func TestPromotionMovesStagesAlongDevelopmentAndTagsReleases(t *testing.T) {
	ctx := context.Background()
	p := setup(t)
	first := run(t, p.Root, "rev-parse", "HEAD")
	commit(t, p.Root, "two")
	run(t, p.Root, "push", "--quiet")

	if _, err := Prepare(ctx, p, "staging", "not-a-commit"); err == nil {
		t.Fatal("unknown --to commit accepted")
	}
	pl, err := Prepare(ctx, p, "staging", first)
	if err != nil || pl.From != "" || pl.To != first || len(pl.Commits) != 1 {
		t.Fatalf("partial staging plan = %+v, %v", pl, err)
	}
	if _, err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	pl, _ = Prepare(ctx, p, "staging", "")
	if _, err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}

	pl, err = Prepare(ctx, p, "production", "")
	if err != nil || !strings.HasPrefix(pl.Tag, "v") || !strings.HasSuffix(pl.Tag, ".1") || len(pl.Commits) != 2 {
		t.Fatalf("production plan = %+v, %v", pl, err)
	}
	if _, err := Execute(ctx, p, pl, ""); err == nil {
		t.Fatal("production release without notes accepted")
	}
	if _, err := Execute(ctx, p, pl, "Notes"); err != nil {
		t.Fatal(err)
	}
	if tags := run(t, p.Root, "ls-remote", "--tags", "origin"); !strings.Contains(tags, pl.Tag) {
		t.Fatalf("tag %s not pushed: %q", pl.Tag, tags)
	}
	if st := CurrentStatus(ctx, p); st.Tag != pl.Tag || st.StagingAhead != 0 || st.DevAhead != 0 {
		t.Fatalf("status = %+v", st)
	}
	if next, _ := Prepare(ctx, p, "production", ""); !next.UpToDate {
		t.Fatal("promoting again was not a no-op")
	}
}

func TestPromotionStopsWhenStageHasCommitsOutsideDevelopment(t *testing.T) {
	ctx := context.Background()
	p := setup(t)
	pl, _ := Prepare(ctx, p, "staging", "")
	if _, err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	run(t, p.Root, "switch", "--quiet", "-c", "hotfix", "origin/test")
	commit(t, p.Root, "hotfix")
	run(t, p.Root, "push", "--quiet", "origin", "hotfix:test")
	run(t, p.Root, "switch", "--quiet", "dev")
	commit(t, p.Root, "three")
	run(t, p.Root, "push", "--quiet")

	pl, err := Prepare(ctx, p, "staging", "")
	if err != nil || len(pl.Diverged) != 1 || pl.Diverged[0].Subject != "hotfix" {
		t.Fatalf("diverged plan = %+v, %v", pl, err)
	}
	if _, err := Execute(ctx, p, pl, ""); err == nil {
		t.Fatal("diverged promotion executed")
	}
}

func TestPromotionStopsWhenStageHasOnlyAMergeOutsideDevelopment(t *testing.T) {
	ctx := context.Background()
	p := setup(t)
	pl, _ := Prepare(ctx, p, "staging", "")
	if _, err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	run(t, p.Root, "push", "--quiet", "origin", "origin/test:refs/heads/main")
	commit(t, p.Root, "merged")
	run(t, p.Root, "push", "--quiet")
	run(t, p.Root, "switch", "--quiet", "-c", "staging", "origin/test")
	run(t, p.Root, "merge", "--quiet", "--no-ff", "--no-edit", "dev")
	run(t, p.Root, "push", "--quiet", "origin", "staging:test")
	run(t, p.Root, "switch", "--quiet", "dev")
	commit(t, p.Root, "three")
	run(t, p.Root, "push", "--quiet")
	before := run(t, p.Root, "rev-parse", "origin/test")

	pl, err := Prepare(ctx, p, "staging", "")
	if err != nil || len(pl.Diverged) != 1 {
		t.Fatalf("diverged plan = %+v, %v", pl, err)
	}
	if st := CurrentStatus(ctx, p); st.StagingOutside != 1 || st.ProductionOutside != 0 {
		t.Fatalf("status = %+v", st)
	}
	if _, err := Execute(ctx, p, pl, ""); err == nil {
		t.Fatal("diverged promotion executed")
	}
	if after := run(t, p.Root, "rev-parse", "origin/test"); after != before {
		t.Fatalf("test moved from %s to %s", before, after)
	}
}

func commit(t *testing.T, root, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name+".txt"), []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, "add", ".")
	run(t, root, "commit", "--quiet", "-m", name)
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git.Run(context.Background(), dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestReleasesBringTheLocalStageBranchAlong(t *testing.T) {
	ctx := context.Background()
	p := setup(t)
	run(t, p.Root, "branch", "test")
	commit(t, p.Root, "two")
	run(t, p.Root, "push", "--quiet")
	release := func(stage, notes string) (Plan, string) {
		t.Helper()
		pl, err := Prepare(ctx, p, stage, "")
		if err != nil {
			t.Fatal(err)
		}
		line, err := Execute(ctx, p, pl, notes)
		if err != nil {
			t.Fatal(err)
		}
		return pl, line
	}

	pl, line := release("staging", "")
	if local := run(t, p.Root, "rev-parse", "test"); local != pl.To || !strings.Contains(line, "The local test branch is at") {
		t.Fatalf("local test = %s, want %s; line %q", local, pl.To, line)
	}
	if _, line := release("production", "Notes"); line != "" || run(t, p.Root, "branch", "--list", "main") != "" {
		t.Fatalf("a release created a local branch the project did not have (line %q)", line)
	}

	commit(t, p.Root, "local-only")
	run(t, p.Root, "branch", "--force", "test", "HEAD")
	run(t, p.Root, "reset", "--quiet", "--hard", "HEAD~1")
	commit(t, p.Root, "three")
	run(t, p.Root, "push", "--quiet")
	before := run(t, p.Root, "rev-parse", "test")
	if _, line := release("staging", ""); run(t, p.Root, "rev-parse", "test") != before || !strings.Contains(line, "has commits the release does not contain") {
		t.Fatalf("a local test with its own commits was moved or not reported (line %q)", line)
	}

	run(t, p.Root, "branch", "--force", "test", "origin/test~1")
	run(t, p.Root, "worktree", "add", "--quiet", filepath.Join(filepath.Dir(p.Root), "staging-checkout"), "test")
	before = run(t, p.Root, "rev-parse", "test")
	commit(t, p.Root, "four")
	run(t, p.Root, "push", "--quiet")
	if _, line := release("staging", ""); run(t, p.Root, "rev-parse", "test") != before || !strings.Contains(line, "is checked out") {
		t.Fatalf("a checked-out local test was moved or not reported (line %q)", line)
	}
}
