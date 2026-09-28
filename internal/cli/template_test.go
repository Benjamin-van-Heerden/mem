package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// templateWorld creates a template library with a nextjs-web template, pushed to
// a bare repository, and isolates the user cache and config.
func templateWorld(t *testing.T) (base, library, libraryWork string) {
	t.Helper()
	base = t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LocalAppData", filepath.Join(base, "home", "cache"))
	library = filepath.Join(base, "library.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", library)
	libraryWork = filepath.Join(base, "library")
	run(t, base, "clone", "--quiet", library, libraryWork)
	writeFile(t, libraryWork, "nextjs-web/template.toml", "description = \"Next.js web app\"\n")
	writeFile(t, libraryWork, "nextjs-web/memories/package-manager.md", "Use pnpm.\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "seed")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
	return base, library, libraryWork
}

// templateProject initializes a mem project with a bare remote that uses nextjs-web.
func templateProject(t *testing.T, base, name, library string) string {
	t.Helper()
	remote := filepath.Join(base, name+".git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	root := filepath.Join(base, name)
	run(t, base, "clone", "--quiet", remote, root)
	run(t, root, "config", "user.name", "Test User")
	run(t, root, "config", "user.email", "test@example.com")
	commit(t, root, "README.md")
	run(t, root, "push", "--quiet", "origin", "HEAD:main")
	mem(t, root, "init", "--protect=false", "--template", "nextjs-web", "--template-source", library)
	run(t, root, "add", "--all")
	run(t, root, "commit", "--quiet", "-m", "mem setup")
	run(t, root, "push", "--quiet")
	return root
}

func mem(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(append(args, "--dir", root))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mem %s: %v\n%s", strings.Join(args, " "), err, out.String())
	}
	return out.String()
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOnboardDrawsInNewTemplateItemsAndPublishesThem(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	root := templateProject(t, base, "app", library)
	if agents, _ := os.ReadFile(filepath.Join(root, "AGENTS.md")); !strings.Contains(string(agents), "## package-manager\nUse pnpm.") {
		t.Fatal("init did not install the template memory")
	}

	writeFile(t, libraryWork, "nextjs-web/skills/next-routing/SKILL.md", "Await params.\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "add skill")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")

	out := mem(t, root, "onboard")
	if !strings.Contains(out, "🧩 TEMPLATES") || !strings.Contains(out, "Added skill next-routing from nextjs-web.") || !strings.Contains(out, "Committed and pushed: Sync template items") {
		t.Fatalf("onboard output:\n%s", out)
	}
	if data, err := os.ReadFile(filepath.Join(root, ".claude/skills/next-routing/SKILL.md")); err != nil || strings.TrimSpace(string(data)) != "Await params." {
		t.Fatalf("skill not reachable through .claude/skills: %v", err)
	}
	if status := run(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("onboard left uncommitted template changes:\n%s", status)
	}
	if strings.Contains(out, "not pushed") {
		t.Fatalf("onboard reported unpushed commits after publishing them:\n%s", out)
	}

	writeFile(t, libraryWork, "nextjs-web/docs/deploy.md", "# Deploy\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "add doc")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
	if out := mem(t, root, "onboard", "--offline"); strings.Contains(out, "Added doc deploy") {
		t.Fatalf("offline onboard pulled the library:\n%s", out)
	}
	if out := mem(t, root, "onboard"); !strings.Contains(out, "Added doc deploy from nextjs-web.") {
		t.Fatalf("onboard did not pull the new doc:\n%s", out)
	}
}

func TestPromotedItemsReachOtherProjectsAtOnboard(t *testing.T) {
	base, library, _ := templateWorld(t)
	a := templateProject(t, base, "a", library)
	b := templateProject(t, base, "b", library)

	mem(t, b, "memory", "set", "testing", "Run focused tests only.")
	writeFile(t, b, ".agents/skills/deploy/SKILL.md", "Deploy with care.\n")
	out := mem(t, b, "template", "promote", "memory", "testing")
	if !strings.Contains(out, "Pushed \"Promote memory testing from b\"") {
		t.Fatalf("promote output:\n%s", out)
	}
	out = mem(t, b, "template", "promote", "skill", "deploy", "--to", "ops")
	if !strings.Contains(out, "Created the template ops") || !strings.Contains(out, "This project now uses ops.") || !strings.Contains(out, ".agents/skills/deploy") {
		t.Fatalf("promote to a new template:\n%s", out)
	}

	if _, err := os.Readlink(filepath.Join(b, ".claude/skills/deploy")); err != nil {
		t.Fatalf("promoting a hand-made skill did not link it for Claude Code: %v", err)
	}

	out = mem(t, a, "onboard")
	if !strings.Contains(out, "Added memory testing from nextjs-web.") || !strings.Contains(out, "New: testing\nRun focused tests only.") {
		t.Fatalf("the promoted memory did not reach project a:\n%s", out)
	}
	if agents, _ := os.ReadFile(filepath.Join(a, "AGENTS.md")); !strings.Contains(string(agents), "## testing\nRun focused tests only.") {
		t.Fatal("project a's AGENTS.md lacks the promoted memory")
	}
	if strings.Contains(out, "deploy") {
		t.Fatalf("project a received an item from a template it does not use:\n%s", out)
	}

	if list := mem(t, a, "template", "list"); !strings.Contains(list, "ops") || !strings.Contains(list, "testing           nextjs-web   in sync") {
		t.Fatalf("template list:\n%s", list)
	}
}
