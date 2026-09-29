package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
)

func TestSyncReportsWhatTeammatesPushed(t *testing.T) {
	mine, teammate := sharedProject(t)
	run(t, teammate, "config", "user.name", "Alice")
	run(t, teammate, "config", "user.email", "alice@example.com")
	mem(t, teammate, "spec", "new", "Search")
	mem(t, teammate, "task", "new", "Index", "Build the index.", "--spec", "search")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Draft the search spec")
	run(t, teammate, "push", "--quiet")
	mem(t, mine, "sync")

	mem(t, teammate, "spec", "start", "search")
	mem(t, teammate, "task", "complete", "index", "Built and tested.")
	mem(t, teammate, "todo", "new", "Flaky CI", "The cache step fails sometimes.")
	mem(t, teammate, "memory", "set", "logging", "Use the structured logger.")
	commit(t, teammate, "parser.go")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Complete the index task")
	run(t, teammate, "push", "--quiet")

	out := mem(t, mine, "sync")
	for _, want := range []string{"📥 INCOMING", "Test: parser.go", "Spec started by alice: Search", "Task completed in search: Index", "Todo opened: Flaky CI", "Review the incoming changes", "🧠 CHANGED MEMORIES", "New: logging\nUse the structured logger.", "Follow the memories under 🧠 CHANGED MEMORIES"} {
		if !strings.Contains(out, want) {
			t.Fatalf("sync output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "re-read AGENTS.md") {
		t.Fatalf("sync still asks to re-read AGENTS.md:\n%s", out)
	}
	if out := mem(t, mine, "sync"); strings.Contains(out, "INCOMING") {
		t.Fatalf("a second sync reports incoming changes again:\n%s", out)
	}
}

func TestSyncInstallsAndReportsPromotedTemplateItems(t *testing.T) {
	base, library, _ := templateWorld(t)
	a := templateProject(t, base, "a", library)
	b := templateProject(t, base, "b", library)
	mem(t, b, "memory", "set", "testing", "Run focused tests only.")
	mem(t, b, "template", "promote", "memory", "testing")

	out := mem(t, a, "sync")
	for _, want := range []string{"🧩 TEMPLATES", "Added memory testing from nextjs-web.", "New: testing\nRun focused tests only."} {
		if !strings.Contains(out, want) {
			t.Fatalf("sync output lacks %q:\n%s", want, out)
		}
	}
}

func TestSyncReportsANewerReleaseWithoutInstallingIt(t *testing.T) {
	mine, _ := sharedProject(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest" {
			t.Errorf("sync requested %s; it must not download a release", r.URL.Path)
		}
		http.Redirect(w, r, "/tag/v0.9.0", http.StatusFound)
	}))
	t.Cleanup(server.Close)
	t.Setenv("MEM_RELEASES_URL", server.URL)
	version := buildinfo.Version
	buildinfo.Version = "v0.8.0"
	t.Cleanup(func() { buildinfo.Version = version })

	out := mem(t, mine, "sync")
	if !strings.Contains(out, "⚠️ mem v0.9.0 is available (this is v0.8.0)") || !strings.Contains(out, "`mem update`") {
		t.Fatalf("sync output:\n%s", out)
	}
}
