package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
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

func TestSinceCountsTheUsersWorkAndResetsAtCheckpoints(t *testing.T) {
	ctx := context.Background()
	root := repo(t)
	p := project.Project{Root: root}
	since := func() Work {
		t.Helper()
		w, err := Since(ctx, p, user, email)
		if err != nil {
			t.Fatal(err)
		}
		return w
	}

	write(t, root, "main.go", lines(30))
	commitAll(t, root, "Work")
	if w := since(); w != (Work{Commits: 1, Files: 1, Lines: 30, LatestIsWork: true}) || w.Level() != Quiet || MidTurnLine(w) != "1 commit, 1 file and 30 lines of work since your last work log or completed task." {
		t.Fatalf("after one work commit: %+v %q", w, MidTurnLine(w))
	}

	// Record commits, another person's work, mem's project files and documentation do not count.
	write(t, root, ".mem/todos/a.md", "todo")
	commitAll(t, root, "Add todo a")
	write(t, root, "AGENTS.md", "instructions")
	commitAll(t, root, ProjectFilesCommit)
	write(t, root, "docs/guide.md", lines(400))
	commitAll(t, root, "Write the guide")
	write(t, root, "other.go", lines(900))
	run(t, root, "add", ".")
	run(t, root, "-c", "user.name=Ann", "-c", "user.email=ann@example.com", "commit", "--quiet", "-m", "Their work")
	if w := since(); w.Commits != 1 || w.Files != 1 || w.Lines != 30 || w.LatestIsWork {
		t.Fatalf("after commits that are not the user's work: %+v", w)
	}

	// Uncommitted changes count, and lines alone can make a log due.
	write(t, root, "parser.go", lines(600))
	if w := since(); w.Lines != 630 || w.Level() != Suggested || !strings.Contains(MidTurnLine(w), "write a work log at the next natural stopping point") || !strings.Contains(TurnStartLine(w), "at the next natural stopping point in this turn") || TurnEndLine(w) != "" {
		t.Fatalf("with uncommitted work: %+v %q", w, TurnStartLine(w))
	}
	write(t, root, "lexer.go", lines(600))
	if w := since(); w.Level() != Due || !strings.Contains(MidTurnLine(w), "a work log is due") || strings.Contains(MidTurnLine(w), "now") || !strings.Contains(TurnStartLine(w), "before taking up substantial new work") || !strings.Contains(TurnEndLine(w), "not a message from the user") {
		t.Fatalf("past the line threshold: %+v %q", w, MidTurnLine(w))
	}
	commitAll(t, root, "Parse")

	// An include pattern makes Markdown work.
	p.Config.Structure.Include = []string{"*.md"}
	if w := since(); w.Commits != 3 || w.Files != 4 {
		t.Fatalf("with Markdown included: %+v", w)
	}
	p.Config.Structure.Include = nil

	write(t, root, ".mem/logs/"+user+"_20261008_120000.md", "log")
	commitAll(t, root, "Work log: checkpoint")
	if w := since(); w != (Work{}) || TurnStartLine(w) != "" {
		t.Fatalf("after a work log: %+v", w)
	}

	for i := range 5 {
		write(t, root, "main.go", lines(i+1))
		commitAll(t, root, "Work")
	}
	if w := since(); w.Commits != 5 || w.Level() != Suggested {
		t.Fatalf("five small commits: %+v", w)
	}
	write(t, root, "main.go", "task work")
	write(t, root, ".mem/specs/s/tasks/01_t.md", "done")
	commitAll(t, root, "Tokenize\n\nSplit the input; tests pass.\n\n"+TaskTrailer+" s/t")
	write(t, root, "main.go", "seven")
	commitAll(t, root, "Work")
	if w := since(); w.Commits != 1 {
		t.Fatalf("after a completed task: %+v", w)
	}
	write(t, root, ".mem/specs/s/tasks/02_u.md", "done")
	commitAll(t, root, "Complete task u")
	write(t, root, "main.go", "eight")
	commitAll(t, root, "Work")
	if w := since(); w.Commits != 1 {
		t.Fatalf("after a record-only task completion from an earlier mem: %+v", w)
	}
}

func lines(n int) string { return strings.Repeat("x\n", n) }

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
