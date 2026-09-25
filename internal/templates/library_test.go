package templates

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// library creates a bare library repository with the given files and a
// working clone to push further changes from, and isolates the user cache.
func library(t *testing.T, files map[string]string) (source, work string) {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("LocalAppData", filepath.Join(base, "home", "cache"))
	source = filepath.Join(base, "library.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", source)
	work = filepath.Join(base, "library")
	run(t, base, "clone", "--quiet", source, work)
	commitFiles(t, work, files)
	return source, work
}

func commitFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if content == "" {
			os.RemoveAll(path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run(t, dir, "add", "--all")
	run(t, dir, "commit", "--quiet", "--allow-empty", "-m", "change")
	run(t, dir, "push", "--quiet", "origin", "HEAD:main")
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

func TestLibraryClonesPullsAndLetsLaterTemplatesOverride(t *testing.T) {
	source, work := library(t, map[string]string{
		"base/template.toml":          "description = \"Shared\"\n",
		"base/memories/logging.md":    "Use the logger.\n",
		"base/skills/review/SKILL.md": "base review\n",
		"web/template.toml":           "description = \"Web\"\n",
		"web/skills/review/SKILL.md":  "web review\n",
		"web/docs/routing.md":         "# Routing\n",
		"notes/README.md":             "not a template\n",
	})
	ctx := context.Background()
	lib, warning, err := Open(ctx, source, true)
	if err != nil || warning != "" {
		t.Fatalf("open: %v %q", err, warning)
	}
	templates, err := lib.Templates()
	if err != nil || len(templates) != 2 || templates[1].Description != "Web" {
		t.Fatalf("templates = %+v, %v", templates, err)
	}
	items, notes, err := lib.Items([]string{"base", "web"})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, item := range items {
		got[item.Key()] = item.Template
	}
	want := map[string]string{"memory:logging": "base", "skill:review": "web", "doc:routing": "web"}
	if len(got) != len(want) || len(notes) != 1 {
		t.Fatalf("items = %v, notes = %v", got, notes)
	}
	for key, template := range want {
		if got[key] != template {
			t.Fatalf("%s from %q, want %q", key, got[key], template)
		}
	}
	if _, _, err := lib.Items([]string{"mobile"}); err == nil {
		t.Fatal("an unknown template was accepted")
	}

	commitFiles(t, work, map[string]string{"web/memories/rsc.md": "Prefer server components.\n"})
	if _, _, err := Open(ctx, source, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lib.ItemPath("web", Memory, "rsc")); err == nil {
		t.Fatal("open without pull updated the clone")
	}
	if _, _, err := Open(ctx, source, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lib.ItemPath("web", Memory, "rsc")); err != nil {
		t.Fatalf("pull did not bring in the new memory: %v", err)
	}
}
