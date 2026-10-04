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
