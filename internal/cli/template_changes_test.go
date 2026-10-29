package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pushLibrary(t *testing.T, libraryWork string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		writeFile(t, libraryWork, rel, content)
	}
	run(t, libraryWork, "add", "--all")
	run(t, libraryWork, "commit", "--quiet", "-m", "library change")
	run(t, libraryWork, "push", "--quiet", "origin", "HEAD:main")
}

func changeTodo(t *testing.T, root string) (path, body string) {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(root, ".mem", "todos", "apply_nextjs_web_template_change*.md"))
	if len(paths) != 1 {
		t.Fatalf("change todos = %v, want one", paths)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	return paths[0], string(data)
}

func TestTemplateChangeNotesBecomeTodosOnceForProjectsBuiltBeforeThem(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	pushLibrary(t, libraryWork, map[string]string{"nextjs-web/changes/2026-10-01-env-schema.md": "# Env schema split\n\nSplit the schema.\n"})
	root := templateProject(t, base, "app", library)

	if out := mem(t, root, "onboard"); strings.Contains(out, "📌 The nextjs-web template") || !strings.Contains(out, "No open todos.") {
		t.Fatalf("a note that existed at init recorded a todo:\n%s", out)
	}

	pushLibrary(t, libraryWork, map[string]string{"nextjs-web/changes/2026-10-29-react-compiler.md": "# React Compiler on\n\nSet reactCompiler in next.config.ts.\n"})
	out := mem(t, root, "onboard")
	path, body := changeTodo(t, root)
	slug := strings.TrimSuffix(filepath.Base(path), ".md")
	if !strings.Contains(out, "📌 The nextjs-web template changed (React Compiler on), so todo "+slug+" asks for the change to be applied to this project.") || !strings.Contains(out, "Tell the user about each 📌 item") {
		t.Fatalf("onboard does not report the change todo:\n%s", out)
	}
	if !strings.Contains(body, "Apply the nextjs-web template change: React Compiler on") {
		t.Fatalf("todo title:\n%s", body)
	}
	if !strings.Contains(body, "Its change note 2026-10-29-react-compiler:\n\n> Set reactCompiler in next.config.ts.") || !strings.Contains(body, "Apply it to this project's existing code as the note describes.") {
		t.Fatalf("todo body does not quote the note and ask for it to be applied:\n%s", body)
	}
	if status := run(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("the sync left the todo or lock uncommitted:\n%s", status)
	}
	if committed := run(t, root, "log", "-1", "--name-only", "--format="); !strings.Contains(committed, ".mem/todos/"+slug+".md") || !strings.Contains(committed, ".mem/templates.lock") {
		t.Fatalf("the todo and the lock were not committed together:\n%s", committed)
	}

	if out := mem(t, root, "onboard"); strings.Contains(out, "📌 The nextjs-web template") {
		t.Fatalf("a seen note recorded a todo again:\n%s", out)
	}
}

// template use takes the new template as it is now, but the same sync also sees the notes the project's other
// templates gained since the last one.
func TestTemplateUseRecordsTheNewTemplatesNotesButNotTheOthers(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	root := templateProject(t, base, "app", library)
	pushLibrary(t, libraryWork, map[string]string{
		"nextjs-web/changes/2026-10-29-react-compiler.md": "# React Compiler on\n\nSet reactCompiler.\n",
		"auth/template.toml":                              "description = \"Auth\"\n",
		"auth/changes/2026-10-01-cookie-name.md":          "# Session cookie renamed\n\nRename it.\n",
	})

	out := mem(t, root, "template", "use", "auth", "--template-source", library)
	changeTodo(t, root)
	if strings.Contains(out, "Session cookie renamed") || !strings.Contains(out, "📌 The nextjs-web template changed (React Compiler on)") || !strings.Contains(out, "apply a template change to this project") {
		t.Fatalf("template use did not turn only the other template's note into a todo:\n%s", out)
	}
	if lock, _ := os.ReadFile(filepath.Join(root, ".mem", "templates.lock")); !strings.Contains(string(lock), "2026-10-01-cookie-name") {
		t.Fatalf("the new template's note was not recorded as seen:\n%s", lock)
	}
}

func TestTemplateChangesArrivingDuringTheSetupAreLedIntoIt(t *testing.T) {
	base, library, libraryWork := templateWorld(t)
	pushLibrary(t, libraryWork, map[string]string{"nextjs-web/setup.md": "# Setup: nextjs-web\n\n## [ ] 1. Scaffold\n\nDone when: it builds.\n"})
	root := templateProject(t, base, "app", library)

	pushLibrary(t, libraryWork, map[string]string{"nextjs-web/changes/2026-10-29-react-compiler.md": "# React Compiler on\n\nSet reactCompiler.\n"})
	out := mem(t, root, "onboard")
	_, body := changeTodo(t, root)
	if !strings.Contains(out, "asks for it to be applied in the remaining setup steps and checked once the setup is done") || !strings.Contains(out, "follow them in the remaining setup steps, and check the code once the setup is done") {
		t.Fatalf("onboard does not lead the change into the pending setup:\n%s", out)
	}
	if !strings.Contains(body, "It arrived while .mem/setup.md was pending") {
		t.Fatalf("todo for a change that arrived during the setup:\n%s", body)
	}
}
