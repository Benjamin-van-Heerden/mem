package cli

import (
	"strings"
	"testing"
)

func TestTodoListShowsEveryOpenTodoWithItsAge(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.")
	mem(t, mine, "todo", "new", "Flaky CI", "The cache step fails sometimes.")

	out := mem(t, mine, "todo", "list")
	if !strings.Contains(out, "flaky_ci") || !strings.Contains(out, "parser_docs") || !strings.Contains(out, "today") {
		t.Fatalf("todo list:\n%s", out)
	}
}

func TestTodoPrioritiesOrderTheTodosAndCriticalOnesLeadOnboard(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.", "--priority", "low")
	mem(t, mine, "todo", "new", "Flaky CI", "The cache step fails sometimes.")
	mem(t, mine, "todo", "new", "Refunds broken", "Refunds fail for every order.", "--priority", "high")
	if err := memErr(t, mine, "todo", "new", "Typo", "", "--priority", "urgent"); err == nil || !strings.Contains(err.Error(), `unknown priority "urgent"`) {
		t.Fatalf("an unknown priority was accepted: %v", err)
	}

	order := func(out string, slugs ...string) bool {
		last := -1
		for _, s := range slugs {
			i := strings.Index(out, s)
			if i <= last {
				return false
			}
			last = i
		}
		return true
	}
	if out := mem(t, mine, "todo", "list"); !order(out, "refunds_broken", "flaky_ci", "parser_docs") || !strings.Contains(out, "medium") {
		t.Fatalf("todos are not ordered by priority:\n%s", out)
	}
	if out := mem(t, mine, "onboard", "--offline"); strings.Contains(out, "Critical todos are open") {
		t.Fatalf("onboard named critical todos when there are none:\n%s", out)
	}

	if out := mem(t, mine, "todo", "priority", "refunds_broken", "critical"); !strings.Contains(out, "high → critical") {
		t.Fatalf("todo priority:\n%s", out)
	}
	if out := mem(t, mine, "todo", "show", "refunds_broken"); !strings.Contains(out, "Priority: critical") {
		t.Fatalf("todo show:\n%s", out)
	}
	if out := mem(t, mine, "onboard", "--offline"); !strings.Contains(out, "Critical todos are open: Refunds broken (refunds_broken).") {
		t.Fatalf("onboard does not lead with the critical todo:\n%s", out)
	}
}
