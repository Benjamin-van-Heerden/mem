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

func TestRecordChangesOnAFeatureBranchStayThereAndSaySo(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.")
	mem(t, mine, "todo", "new", "Flaky CI", "The cache step fails sometimes.")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Record todos")
	if out := mem(t, mine, "todo", "claim", "flaky_ci"); strings.Contains(out, "teammates see this") {
		t.Fatalf("claim on dev mentioned a merge:\n%s", out)
	}

	run(t, mine, "switch", "--quiet", "-c", "refunds")
	out := mem(t, mine, "todo", "claim", "parser_docs")
	if !strings.Contains(out, "You are on refunds, not dev: teammates see this record change once refunds merges into dev.") {
		t.Fatalf("claim on a feature branch:\n%s", out)
	}
	if subject := run(t, mine, "log", "-1", "--format=%s", "refunds"); subject != "Claim todo parser_docs" {
		t.Fatalf("the claim was not committed on the feature branch: %q", subject)
	}
	if subject := run(t, mine, "log", "-1", "--format=%s", "dev"); subject == "Claim todo parser_docs" {
		t.Fatal("the claim was committed on dev")
	}
}
