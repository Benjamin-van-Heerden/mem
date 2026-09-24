package structure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

func TestMeasureCountsCodeChangesSinceTheDocWasLastCommitted(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	run(t, root, "init", "--quiet", "-b", "dev")
	run(t, root, "config", "user.name", "Test")
	run(t, root, "config", "user.email", "test@example.com")
	p := project.Project{Root: root, Config: project.Config{Structure: project.StructureConfig{Ignore: []string{"migrations/**"}}}}

	if d, _ := Measure(ctx, p); !d.Missing {
		t.Fatal("missing doc not reported")
	}
	write(t, root, RelPath, Template)
	if d, _ := Measure(ctx, p); !d.Editing || d.Stale() {
		t.Fatal("an uncommitted doc should count as up to date")
	}
	run(t, root, "add", ".")
	run(t, root, "commit", "--quiet", "-m", "structure")

	for i := range 5 {
		write(t, root, fmt.Sprintf("src/f%d.go", i), "package src\n")
	}
	write(t, root, "README.md", "docs\n")
	write(t, root, "go.sum", "checksums\n")
	write(t, root, "migrations/001.sql", "create table x;\n")
	d, err := Measure(ctx, p)
	if err != nil || len(d.Changes) != 5 || d.Stale() {
		t.Fatalf("changes=%v stale=%v err=%v", d.Changes, d.Stale(), err)
	}
	write(t, root, "src/f5.go", "package src\n")
	if d, _ := Measure(ctx, p); !d.Stale() {
		t.Fatal("six changed code files should make the doc stale")
	}

	run(t, root, "add", ".")
	run(t, root, "commit", "--quiet", "-m", "code")
	write(t, root, RelPath, Template+"\nUpdated.\n")
	run(t, root, "commit", "--quiet", "-am", "update structure")
	if d, _ := Measure(ctx, p); len(d.Changes) != 0 {
		t.Fatalf("committing the doc did not move the baseline: %v", d.Changes)
	}
}

func TestRenamedToResolvesNumstatRenames(t *testing.T) {
	cases := map[string]string{
		"src/{old => new}/f.go": "src/new/f.go",
		"a.go => b.go":          "b.go",
		"src/{ => api}/f.go":    "src/api/f.go",
	}
	for in, want := range cases {
		if got := renamedTo(in); got != want {
			t.Errorf("renamedTo(%q) = %q, want %q", in, got, want)
		}
	}
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := git.Run(context.Background(), dir, args...); err != nil {
		t.Fatal(err)
	}
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
