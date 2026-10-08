package cli

import (
	"strings"
	"testing"
)

func TestTodoListShowsEveryOpenTodoWithItsAge(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.")
	mem(t, mine, "todo", "new", "Flaky CI", "The cache step fails sometimes.")

	out := mem(t, mine, "todo", "list")
	if !strings.Contains(out, "flaky_ci") || !strings.Contains(out, "parser_docs") || !strings.Contains(out, "today") {
		t.Fatalf("todo list:\n%s", out)
	}
}
