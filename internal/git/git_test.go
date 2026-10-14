package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitPathsCommitsOnlyGivenPathsAndPushes(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	remote, repo := filepath.Join(base, "remote.git"), filepath.Join(base, "repo")
	mustRun(t, base, "init", "--quiet", "--bare", remote)
	mustRun(t, base, "init", "--quiet", "-b", "dev", repo)
	mustRun(t, repo, "config", "user.name", "Test")
	mustRun(t, repo, "config", "user.email", "test@example.com")
	write(t, repo, "code.go", "package main\n")
	mustRun(t, repo, "add", ".")
	mustRun(t, repo, "commit", "--quiet", "-m", "init")
	mustRun(t, repo, "remote", "add", "origin", remote)
	mustRun(t, repo, "push", "--quiet", "-u", "origin", "dev")

	write(t, repo, "code.go", "package main // edited\n")
	mustRun(t, repo, "add", "code.go")
	write(t, repo, ".mem/todos/x.md", "todo\n")

	pushed, err := CommitPaths(ctx, repo, "Claim todo x", ".mem/todos/x.md")
	if err != nil || !pushed {
		t.Fatalf("pushed=%v err=%v", pushed, err)
	}
	files := mustRun(t, repo, "show", "--name-only", "--format=", "origin/dev")
	if files != ".mem/todos/x.md" {
		t.Fatalf("pushed commit contains %q", files)
	}
	if status := mustRun(t, repo, "status", "--porcelain"); !strings.Contains(status, "M  code.go") {
		t.Fatalf("unrelated staged change was disturbed: %q", status)
	}
}

func TestCommitTakesPathsAlreadyRemovedOrMovedInTheIndex(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	mustRun(t, repo, "init", "--quiet", "-b", "dev")
	mustRun(t, repo, "config", "user.name", "Test")
	mustRun(t, repo, "config", "user.email", "test@example.com")
	for _, file := range []string{"requirements.txt", "HANDOVER.md", "old.txt", "other.txt"} {
		write(t, repo, file, file+"\n")
	}
	mustRun(t, repo, "add", ".")
	mustRun(t, repo, "commit", "--quiet", "-m", "init")

	mustRun(t, repo, "rm", "--quiet", "requirements.txt")
	write(t, repo, "docs/history/.keep", "")
	mustRun(t, repo, "mv", "HANDOVER.md", "docs/history/HANDOVER.md")
	if err := os.Remove(filepath.Join(repo, "old.txt")); err != nil {
		t.Fatal(err)
	}
	write(t, repo, "other.txt", "changed\n")

	if err := Commit(ctx, repo, "Move docs", "requirements.txt", "HANDOVER.md", "docs", "old.txt"); err != nil {
		t.Fatal(err)
	}
	files := mustRun(t, repo, "show", "--name-status", "--format=", "HEAD")
	for _, want := range []string{"D\trequirements.txt", "D\told.txt", "docs/history/HANDOVER.md", "docs/history/.keep"} {
		if !strings.Contains(files, want) {
			t.Fatalf("commit lacks %q:\n%s", want, files)
		}
	}
	if status := mustRun(t, repo, "status", "--porcelain"); status != " M other.txt" {
		t.Fatalf("a path that was not named was committed or staged: %q", status)
	}
	if err := Commit(ctx, repo, "Typo", "no-such-file.txt"); err == nil {
		t.Fatal("a path that matches nothing was accepted")
	}
}

func TestPushReportsARemoteThatMovedOnInOneLine(t *testing.T) {
	base := t.TempDir()
	remote, repo, other := filepath.Join(base, "remote.git"), filepath.Join(base, "repo"), filepath.Join(base, "other")
	mustRun(t, base, "init", "--quiet", "--bare", "-b", "dev", remote)
	for _, dir := range []string{repo, other} {
		mustRun(t, base, "clone", "--quiet", remote, dir)
		mustRun(t, dir, "config", "user.name", "Test")
		mustRun(t, dir, "config", "user.email", "test@example.com")
		mustRun(t, dir, "commit", "--quiet", "--allow-empty", "-m", "from "+filepath.Base(dir))
	}
	mustRun(t, other, "push", "--quiet", "-u", "origin", "dev")
	mustRun(t, repo, "fetch", "--quiet")
	mustRun(t, repo, "branch", "--quiet", "--set-upstream-to", "origin/dev")

	err := Push(context.Background(), repo)
	if !errors.Is(err, ErrPush) || err.Error() != "the remote has commits this checkout does not have yet" {
		t.Fatalf("Push error = %v", err)
	}
}

func mustRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := Run(context.Background(), dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
