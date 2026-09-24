package release

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
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
	if err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	pl, _ = Prepare(ctx, p, "staging", "")
	if err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}

	pl, err = Prepare(ctx, p, "production", "")
	if err != nil || !strings.HasPrefix(pl.Tag, "v") || !strings.HasSuffix(pl.Tag, ".1") || len(pl.Commits) != 2 {
		t.Fatalf("production plan = %+v, %v", pl, err)
	}
	if err := Execute(ctx, p, pl, ""); err == nil {
		t.Fatal("production release without notes accepted")
	}
	if err := Execute(ctx, p, pl, "Notes"); err != nil {
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
	if err := Execute(ctx, p, pl, ""); err != nil {
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
	if err := Execute(ctx, p, pl, ""); err == nil {
		t.Fatal("diverged promotion executed")
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
