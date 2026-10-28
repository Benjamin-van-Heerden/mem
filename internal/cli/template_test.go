package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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

func TestInitInstallsTheTemplateSetupAndTemplateUseDoesNot(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	setup := "# Setup: nextjs-web\n\n## [ ] 1. Scaffold\n\nDone when: it builds.\n"
	writeFile(t, libraryWork, "nextjs-web/setup.md", setup)
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "setup")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")

	root := templateProject(t, base, "app", library)
	if got, err := os.ReadFile(filepath.Join(root, ".mem", "setup.md")); err != nil || string(got) != setup {
		t.Fatalf(".mem/setup.md = %q, %v", got, err)
	}
	if listed := mem(t, root, "template", "list"); !strings.Contains(listed, "SETUP") {
		t.Fatalf("template list does not show setups:\n%s", listed)
	}

	remote := filepath.Join(base, "later.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	later := filepath.Join(base, "later")
	run(t, base, "clone", "--quiet", remote, later)
	run(t, later, "config", "user.name", "Test User")
	run(t, later, "config", "user.email", "test@example.com")
	commit(t, later, "README.md")
	run(t, later, "push", "--quiet", "origin", "HEAD:main")
	mem(t, later, "init", "--protect=false", "--no-template")
	out := mem(t, later, "template", "use", "nextjs-web", "--template-source", library)
	if _, err := os.Stat(filepath.Join(later, ".mem", "setup.md")); !os.IsNotExist(err) {
		t.Fatalf("template use installed the setup: %v", err)
	}
	if !strings.Contains(out, "nextjs-web has a setup for new projects") {
		t.Fatalf("template use does not mention the setup:\n%s", out)
	}
	todos, _ := filepath.Glob(filepath.Join(later, ".mem", "todos", "*.md"))
	if len(todos) != 1 {
		t.Fatalf("template use recorded %d todos, want one for the template's conventions:\n%s", len(todos), out)
	}
	todo, err := os.ReadFile(todos[0])
	if err != nil || !strings.Contains(string(todo), "- `package-manager`: Use pnpm.") || !strings.Contains(out, "📌 Todo apply_") {
		t.Fatalf("template use did not record a todo for the template's conventions (%v):\n%s\n%s", err, todo, out)
	}
}

func TestTemplateMemoryChangesRecordTodosToCheckExistingCode(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	root := templateProject(t, base, "app", library)
	push := func(message string) {
		run(t, libraryWork, "add", "--all")
		run(t, libraryWork, "commit", "--quiet", "-m", message)
		run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
	}
	todoFiles := func() []string {
		entries, _ := os.ReadDir(filepath.Join(root, ".mem", "todos"))
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return names
	}

	writeFile(t, libraryWork, "nextjs-web/memories/package-manager.md", "Use bun.\n")
	writeFile(t, libraryWork, "nextjs-web/memories/repositories.md", "Queries live in src/db/repositories.\n")
	push("conventions")
	out := mem(t, root, "onboard")

	if !strings.Contains(out, "📌 The nextjs-web template changed the convention package-manager") || !strings.Contains(out, "📌 The nextjs-web template added the convention repositories") || !strings.Contains(out, "Tell the user about each 📌 item") {
		t.Fatalf("onboard does not report the convention todos:\n%s", out)
	}
	changed, err := os.ReadFile(filepath.Join(root, ".mem", "todos", "apply_template_convention_package_manager.md"))
	if err != nil || !strings.Contains(string(changed), "> Use bun.") || !strings.Contains(string(changed), "It used to read:\n\n> Use pnpm.") {
		t.Fatalf("todo for the changed memory (%v):\n%s", err, changed)
	}
	if status := run(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("the todos were not committed with the template sync:\n%s", status)
	}

	writeFile(t, libraryWork, "nextjs-web/memories/package-manager.md", "Use bun, never npm.\n")
	push("reword")
	mem(t, root, "onboard")
	if got := todoFiles(); len(got) != 2 {
		t.Fatalf("a second change to a memory with an open todo added another todo: %v", got)
	}
	if changed, _ := os.ReadFile(filepath.Join(root, ".mem", "todos", "apply_template_convention_package_manager.md")); !strings.Contains(string(changed), "> Use bun, never npm.") {
		t.Fatalf("the open todo was not rewritten with the latest text:\n%s", changed)
	}

	if !strings.Contains(string(changed), "ask whether to adopt it, as its own todo or spec, or to drop the convention with `mem memory remove package-manager`") {
		t.Fatalf("the todo does not cover a convention the project has not adopted:\n%s", changed)
	}
}

// The setup is copied from the template once, so a convention that arrives while it is pending may describe
// something its steps never build.
func TestConventionsArrivingDuringTheSetupRecordTodosForIt(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	writeFile(t, libraryWork, "nextjs-web/setup.md", "# Setup: nextjs-web\n\n## [ ] 1. Scaffold\n\nDone when: it builds.\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "setup")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
	root := templateProject(t, base, "app", library)

	if out := mem(t, root, "onboard"); strings.Contains(out, "📌 The nextjs-web template") || !strings.Contains(out, "No open todos.") {
		t.Fatalf("conventions installed with the setup recorded a todo:\n%s", out)
	}

	writeFile(t, libraryWork, "nextjs-web/memories/local-dev.md", "Serve the dev app through portless.\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "portless")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
	out := mem(t, root, "onboard")

	if !strings.Contains(out, "📌 The nextjs-web template added the convention local-dev, so todo apply_template_convention_local_dev asks for it to be followed in the remaining setup steps") || !strings.Contains(out, "follow the convention in the remaining setup steps, and check the code against it once the setup is done") {
		t.Fatalf("onboard does not lead the convention into the pending setup:\n%s", out)
	}
	todo, err := os.ReadFile(filepath.Join(root, ".mem", "todos", "apply_template_convention_local_dev.md"))
	if err != nil || !strings.Contains(string(todo), "It arrived while .mem/setup.md was pending") {
		t.Fatalf("todo for the convention that arrived during the setup (%v):\n%s", err, todo)
	}
}

func TestInitWithoutTemplateChoiceListsTemplatesAndChangesNothing(t *testing.T) {
	base, library, _ := templateWorld(t)
	remote := filepath.Join(base, "app.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	root := filepath.Join(base, "app")
	run(t, base, "clone", "--quiet", remote, root)
	commit(t, root, "README.md")
	run(t, root, "push", "--quiet", "origin", "HEAD:main")

	out := mem(t, root, "init", "--template-source", library)

	if !strings.Contains(out, "nextjs-web") || !strings.Contains(out, "--no-template") {
		t.Fatalf("init does not offer the templates:\n%s", out)
	}
	if status := run(t, root, "status", "--porcelain", "--ignored"); status != "" {
		t.Fatalf("init changed the working tree:\n%s", status)
	}
	if branches := run(t, remote, "branch", "--list"); strings.Contains(branches, "dev") {
		t.Fatalf("init published branches before a template was chosen:\n%s", branches)
	}
}

func TestOnboardRunsTemplateRunnablesAndPromoteSendsThemBack(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	root := templateProject(t, base, "app", library)

	writeFile(t, libraryWork, "nextjs-web/runnables/checks", "#!/bin/sh\necho checks from the template\n")
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "add runnable")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")

	out := mem(t, root, "onboard")
	// Windows cannot execute shell scripts directly, so there onboard only installs it.
	if !strings.Contains(out, "Added runnable checks from nextjs-web.") || runtime.GOOS != "windows" && !strings.Contains(out, "checks from the template") {
		t.Fatalf("onboard did not install and run the template runnable:\n%s", out)
	}
	if status := run(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("onboard left the runnable uncommitted:\n%s", status)
	}
	if mode := run(t, root, "ls-files", "--stage", ".mem/runnables/checks"); runtime.GOOS != "windows" && !strings.HasPrefix(mode, "100755") {
		t.Fatalf("runnable committed without the executable bit: %s", mode)
	}

	writeFile(t, root, ".mem/runnables/checks", "#!/bin/sh\necho improved checks\n")
	out = mem(t, root, "template", "promote", "runnable", "checks")
	if !strings.Contains(out, "Pushed \"Promote runnable checks from app\"") {
		t.Fatalf("promote output:\n%s", out)
	}
	run(t, libraryWork, "pull", "--quiet", "origin", "main")
	if data, err := os.ReadFile(filepath.Join(libraryWork, "nextjs-web/runnables/checks")); err != nil || !strings.Contains(string(data), "improved checks") {
		t.Fatalf("the promoted runnable did not reach the library: %v", err)
	}
	if list := mem(t, root, "template", "list"); !strings.Contains(list, "runnable   checks            nextjs-web   in sync") {
		t.Fatalf("template list:\n%s", list)
	}
}
