package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

func TestEnsureBranchesTracksRemoteCreatesMissingPublishesAndSwitches(t *testing.T) {
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	run(t, base, "init", "--quiet", "--bare", remote)
	other := filepath.Join(base, "other")
	run(t, base, "init", "--quiet", "-b", "main", other)
	commit(t, other, "base.txt")
	run(t, other, "remote", "add", "origin", remote)
	run(t, other, "push", "--quiet", "origin", "main")
	run(t, other, "switch", "--quiet", "-c", "dev")
	commit(t, other, "teammate.txt")
	run(t, other, "push", "--quiet", "origin", "dev")

	mine := filepath.Join(base, "mine")
	run(t, base, "clone", "--quiet", "-b", "main", remote, mine)
	if err := os.WriteFile(filepath.Join(mine, "wip.txt"), []byte("uncommitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	g := project.GitConfig{Remote: "origin", Development: "dev", Staging: "test", Production: "main"}
	if _, err := ensureBranches(context.Background(), mine, g); err != nil {
		t.Fatal(err)
	}

	if got := run(t, mine, "branch", "--show-current"); got != "dev" {
		t.Fatalf("current branch = %s, want dev", got)
	}
	if got, want := run(t, mine, "rev-parse", "dev"), run(t, other, "rev-parse", "dev"); got != want {
		t.Fatalf("dev = %s, want the teammate's %s", got, want)
	}
	if got, want := run(t, mine, "rev-parse", "test"), run(t, mine, "rev-parse", "main"); got != want {
		t.Fatalf("test = %s, want main %s", got, want)
	}
	if got := run(t, mine, "rev-parse", "--abbrev-ref", "test@{upstream}"); got != "origin/test" {
		t.Fatalf("test upstream = %s", got)
	}
	if _, err := os.Stat(filepath.Join(mine, "wip.txt")); err != nil {
		t.Fatalf("uncommitted file lost: %v", err)
	}
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.com"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func commit(t *testing.T, dir, file string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(file+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "add", file)
	run(t, dir, "commit", "--quiet", "-m", file)
}

func TestEnsureBranchesStartsAnEmptyCloneWithAnEmptyCommit(t *testing.T) {
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	run(t, base, "init", "--quiet", "--bare", remote)
	mine := filepath.Join(base, "mine")
	run(t, base, "clone", "--quiet", remote, mine)
	if err := os.WriteFile(filepath.Join(mine, "staged.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, mine, "add", "staged.txt")

	g := project.GitConfig{Remote: "origin", Development: "dev", Staging: "test", Production: "main"}
	lines, err := ensureBranches(context.Background(), mine, g)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(strings.Join(lines, "\n"), "empty first commit on main") {
		t.Fatalf("lines do not report the first commit:\n%s", strings.Join(lines, "\n"))
	}
	if got := run(t, mine, "branch", "--show-current"); got != "dev" {
		t.Fatalf("current branch = %s, want dev", got)
	}
	if got := run(t, mine, "rev-list", "--count", "main"); got != "1" {
		t.Fatalf("main has %s commits, want 1", got)
	}
	if got := run(t, mine, "ls-tree", "main"); got != "" {
		t.Fatalf("first commit is not empty: %s", got)
	}
	for _, branch := range []string{"main", "test", "dev"} {
		if got, want := run(t, remote, "rev-parse", branch), run(t, mine, "rev-parse", "main"); got != want {
			t.Fatalf("remote %s = %s, want %s", branch, got, want)
		}
	}
	if got := run(t, mine, "diff", "--cached", "--name-only"); got != "staged.txt" {
		t.Fatalf("staged files = %q, want staged.txt still staged", got)
	}
}
