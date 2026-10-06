package cli

import (
	"strings"
	"testing"
)

func TestTodoListIncludesClaimedTodos(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.")
	mem(t, mine, "todo", "new", "Flaky CI", "The cache step fails sometimes.")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Record todos")
	mem(t, mine, "todo", "claim", "parser_docs")

	out := mem(t, mine, "todo", "list")
	if !strings.Contains(out, "flaky_ci") || !strings.Contains(out, "parser_docs") || !strings.Contains(out, "test_user") {
		t.Fatalf("todo list:\n%s", out)
	}
}
