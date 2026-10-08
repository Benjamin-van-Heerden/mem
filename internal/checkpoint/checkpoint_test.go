package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

const (
	user  = "test_user"
	email = "test@example.com"
)

func repo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	run(t, root, "init", "--quiet", "-b", "dev")
	run(t, root, "config", "user.name", "Test User")
	run(t, root, "config", "user.email", email)
	write(t, root, "README.md", "readme")
	commitAll(t, root, "Initial commit")
	write(t, root, ".mem/config.toml", "schema = 1\n")
	commitAll(t, root, "Set up mem")
	return root
}

func TestWorkSinceLogEscalatesAndResetsAtCheckpoints(t *testing.T) {
	ctx := context.Background()
	root := repo(t)
	count := func() (int, bool) {
		t.Helper()
		n, latest, err := WorkSinceLog(ctx, root, user, email)
		if err != nil {
			t.Fatal(err)
		}
		return n, latest
	}

	for i, want := range []string{"1 commit since", "2 commits since", "3 commits since"} {
		write(t, root, "main.go", strings.Repeat("x", i+1))
		commitAll(t, root, "Work")
		if n, latest := count(); !latest || !strings.HasPrefix(LogLine(n), want) {
			t.Fatalf("after %d work commits: %q, latest %v", i+1, LogLine(n), latest)
		}
	}
	if n, _ := count(); !strings.Contains(LogLine(n), "consider writing one now") {
		t.Fatalf("3 commits: %q", LogLine(n))
	}

	// Record commits, another person's work and mem's project files do not count.
	write(t, root, ".mem/todos/a.md", "todo")
	commitAll(t, root, "Add todo a")
	write(t, root, "AGENTS.md", "instructions")
	commitAll(t, root, ProjectFilesCommit)
	write(t, root, "other.go", "theirs")
	run(t, root, "add", ".")
	run(t, root, "-c", "user.name=Ann", "-c", "user.email=ann@example.com", "commit", "--quiet", "-m", "Their work")
	if n, latest := count(); n != 3 || latest {
		t.Fatalf("after non-work commits: %d, latest %v", n, latest)
	}

	write(t, root, "main.go", "four")
	commitAll(t, root, "Work")
	write(t, root, "main.go", "five")
	commitAll(t, root, "Work")
	if n, _ := count(); !strings.Contains(LogLine(n), "you should write one now, before continuing") {
		t.Fatalf("5 commits: %q", LogLine(n))
	}

	write(t, root, ".mem/logs/"+user+"_20261008_120000.md", "log")
	commitAll(t, root, "Work log: checkpoint")
	if n, latest := count(); n != 0 || latest {
		t.Fatalf("after a work log: %d, latest %v", n, latest)
	}

	write(t, root, "main.go", "six")
	commitAll(t, root, "Work")
	write(t, root, "main.go", "task work")
	write(t, root, ".mem/specs/s/tasks/01_t.md", "done")
	commitAll(t, root, "Tokenize\n\nSplit the input; tests pass.\n\n"+TaskTrailer+" s/t")
	write(t, root, "main.go", "seven")
	commitAll(t, root, "Work")
	if n, _ := count(); n != 1 {
		t.Fatalf("after a completed task: %d", n)
	}
	write(t, root, ".mem/specs/s/tasks/02_u.md", "done")
	commitAll(t, root, "Complete task u")
	write(t, root, "main.go", "eight")
	commitAll(t, root, "Work")
	if n, _ := count(); n != 1 {
		t.Fatalf("after a record-only task completion from an earlier mem: %d", n)
	}
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := git.Run(context.Background(), dir, args...); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, root, message string) {
	t.Helper()
	run(t, root, "add", ".")
	run(t, root, "commit", "--quiet", "-m", message)
}
