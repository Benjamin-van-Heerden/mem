package cli

import (
	"strings"
	"testing"
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
	commit(t, teammate, "parser.go")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Complete the index task")
	run(t, teammate, "push", "--quiet")

	out := mem(t, mine, "sync")
	for _, want := range []string{"📥 INCOMING", "Test: parser.go", "Spec started by alice: Search", "Task completed in search: Index", "Todo opened: Flaky CI", "Review the incoming changes"} {
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
