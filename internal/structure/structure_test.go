package structure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
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
		if got := RenamedTo(in); got != want {
			t.Errorf("RenamedTo(%q) = %q, want %q", in, got, want)
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

func TestRelevantAppliesIncludeAndIgnorePatterns(t *testing.T) {
	c := project.StructureConfig{Include: []string{"*.md", "assets/**/*.svg"}, Ignore: []string{"docs/archive/**", "content/**/draft-*.md"}}
	for file, want := range map[string]bool{
		"main.go":                       true,
		"docs/guide.md":                 true,
		"README.md":                     true,
		"assets/icons/ui/logo.svg":      true,
		"logo.svg":                      false,
		"docs/archive/2025/old.md":      false,
		"content/posts/draft-intro.md":  false,
		"content/posts/intro.md":        true,
		".mem/logs/someone_20261012.md": false,
		"AGENTS.md":                     false,
		"go.sum":                        false,
	} {
		if got := Relevant(file, c); got != want {
			t.Errorf("Relevant(%q) = %v, want %v", file, got, want)
		}
	}
	if Relevant("docs/guide.md", project.StructureConfig{}) {
		t.Error("Markdown counts as work without an include pattern")
	}
}

func TestChangedSinceCountsDocumentsAndSkipsWhatIsNotWork(t *testing.T) {
	root := t.TempDir()
	p := project.Project{Root: root, Config: project.Config{Mode: project.ModeFolder, Structure: project.StructureConfig{Ignore: []string{"_generated/**", "*.tmp"}}}}
	baseline := time.Now().Add(-time.Hour)
	for file, modified := range map[string]time.Time{
		"clients/acme/invoice.pdf": baseline.Add(time.Minute),
		"notes.md":                 baseline.Add(2 * time.Minute),
		"old.md":                   baseline.Add(-time.Minute),
		"AGENTS.md":                baseline.Add(time.Minute),
		"_generated/invoice.pdf":   baseline.Add(time.Minute),
		"scratch.tmp":              baseline.Add(time.Minute),
		".mem/logs/log.md":         baseline.Add(time.Minute),
		".claude/settings.json":    baseline.Add(time.Minute),
		".DS_Store":                baseline.Add(time.Minute),
	} {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, modified, modified); err != nil {
			t.Fatal(err)
		}
	}
	changes, err := ChangedSince(p, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0].Path != "notes.md" || changes[1].Path != "clients/acme/invoice.pdf" {
		t.Fatalf("ChangedSince = %+v", changes)
	}
}

func TestTreeOfAFolderProjectNeedsNoGit(t *testing.T) {
	root := t.TempDir()
	p := project.Project{Root: root, Config: project.Config{Mode: project.ModeFolder}}
	for _, file := range []string{"clients/acme/invoice.pdf", ".mem/config.toml", "notes.md"} {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", t.TempDir())
	tree, err := Tree(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if want := "clients/\n  acme/\n    invoice.pdf\nnotes.md"; tree != want {
		t.Fatalf("Tree =\n%s", tree)
	}
}
