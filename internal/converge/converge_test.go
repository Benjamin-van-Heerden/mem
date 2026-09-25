package converge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// clones returns two checkouts of dev sharing a bare remote; the second is the project under test.
func clones(t *testing.T) (other string, p project.Project) {
	t.Helper()
	base := t.TempDir()
	run(t, base, "init", "--quiet", "--bare", "remote.git")
	other = filepath.Join(base, "other")
	run(t, base, "init", "--quiet", "-b", "dev", other)
	identify(t, other)
	write(t, other, "shared.txt", "base\n")
	run(t, other, "add", ".")
	run(t, other, "commit", "--quiet", "-m", "init")
	run(t, other, "remote", "add", "origin", filepath.Join(base, "remote.git"))
	run(t, other, "push", "--quiet", "-u", "origin", "dev")
	mine := filepath.Join(base, "mine")
	run(t, base, "clone", "--quiet", "-b", "dev", filepath.Join(base, "remote.git"), mine)
	identify(t, mine)
	return other, project.Project{Root: mine, Config: project.Config{Git: project.GitConfig{Remote: "origin", Development: "dev"}}}
}

func pushChange(t *testing.T, dir, file, content string) {
	t.Helper()
	write(t, dir, file, content)
	run(t, dir, "add", ".")
	run(t, dir, "commit", "--quiet", "-m", "change "+file)
	run(t, dir, "push", "--quiet")
}

func TestSyncFastForwardsCleanCheckout(t *testing.T) {
	other, p := clones(t)
	pushChange(t, other, "shared.txt", "theirs\n")
	r := Sync(context.Background(), p)
	if r.Behind != 0 || len(r.Done) != 1 || len(r.Nudges) != 0 {
		t.Fatalf("report = %+v", r)
	}
}

func TestSyncRebasesUnpushedCommitsOntoIncomingWork(t *testing.T) {
	other, p := clones(t)
	pushChange(t, other, "theirs.txt", "theirs\n")
	write(t, p.Root, "mine.txt", "mine\n")
	run(t, p.Root, "add", ".")
	run(t, p.Root, "commit", "--quiet", "-m", "mine")
	r := Sync(context.Background(), p)
	if r.Ahead != 1 || r.Behind != 0 {
		t.Fatalf("after rebase ahead=%d behind=%d; report %+v", r.Ahead, r.Behind, r)
	}
	if !strings.Contains(strings.Join(r.Nudges, " "), "not pushed") {
		t.Fatalf("no push nudge: %v", r.Nudges)
	}
}

func TestSyncAbortsConflictingRebaseAndLeavesCheckoutUnchanged(t *testing.T) {
	other, p := clones(t)
	pushChange(t, other, "shared.txt", "theirs\n")
	write(t, p.Root, "shared.txt", "mine\n")
	run(t, p.Root, "commit", "--quiet", "-am", "mine")
	head := run(t, p.Root, "rev-parse", "HEAD")
	r := Sync(context.Background(), p)
	if got := run(t, p.Root, "rev-parse", "HEAD"); got != head {
		t.Fatal("HEAD moved after an aborted rebase")
	}
	if _, err := os.Stat(filepath.Join(p.Root, ".git", "rebase-merge")); err == nil {
		t.Fatal("rebase left in progress")
	}
	if !strings.Contains(strings.Join(r.Nudges, " "), "conflicts") {
		t.Fatalf("no conflict nudge: %v", r.Nudges)
	}
}

func TestSyncKeepsOverlappingUncommittedWork(t *testing.T) {
	other, p := clones(t)
	pushChange(t, other, "shared.txt", "theirs\n")
	write(t, p.Root, "shared.txt", "work in progress\n")
	r := Sync(context.Background(), p)
	data, _ := os.ReadFile(filepath.Join(p.Root, "shared.txt"))
	if string(data) != "work in progress\n" || r.Behind != 1 || len(r.Nudges) == 0 {
		t.Fatalf("content %q, report %+v", data, r)
	}
}

func identify(t *testing.T, dir string) {
	run(t, dir, "config", "user.name", "Test")
	run(t, dir, "config", "user.email", "test@example.com")
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git.Run(context.Background(), dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLocalNudgesWhenUncommittedWorkGrowsLarge(t *testing.T) {
	_, p := clones(t)
	for i := range largeFiles {
		write(t, p.Root, fmt.Sprintf("f%d.go", i), "package f\n")
	}
	write(t, p.Root, "notes.md", strings.Repeat("docs\n", 2000))
	if r := Local(context.Background(), p); len(r.Nudges) != 0 {
		t.Fatalf("nudged below the threshold: %v", r.Nudges)
	}
	write(t, p.Root, "one_more.go", "package f\n")
	if r := Local(context.Background(), p); !strings.Contains(strings.Join(r.Nudges, " "), "Commit the finished") {
		t.Fatalf("no nudge for large uncommitted work: %v", r.Nudges)
	}
}
