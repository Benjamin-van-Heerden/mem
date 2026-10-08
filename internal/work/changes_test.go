package work

import (
	"strings"
	"testing"
)

func revision(files map[string]string) Revision {
	return func(path string) ([]byte, bool) {
		data, ok := files[path]
		return []byte(data), ok
	}
}

func front(fields string) string { return "---\n" + fields + "\n---\n\nBody.\n" }

func TestRecordChangesDescribesTheLifecycleOfRecords(t *testing.T) {
	before := map[string]string{
		".mem/todos/flaky_ci.md":              front("title: Flaky CI\nstatus: open"),
		".mem/todos/old_docs.md":              front("title: Old docs\nstatus: open"),
		".mem/specs/login/spec.md":            front("title: Login\nstatus: draft"),
		".mem/specs/login/tasks/01_form.md":   front("title: Form\nstatus: todo"),
		".mem/specs/search/spec.md":           front("title: Search\nstatus: active\nassigned_to: alice"),
		".mem/specs/search/tasks/01_index.md": front("title: Index\nstatus: completed"),
	}
	after := map[string]string{
		".mem/todos/flaky_ci.md":                      front("title: Flaky CI\nstatus: open"),
		".mem/todos/new_logo.md":                      front("title: New logo\nstatus: open"),
		".mem/specs/login/spec.md":                    front("title: Login\nstatus: active\nassigned_to: bob"),
		".mem/specs/login/tasks/01_form.md":           front("title: Form\nstatus: completed"),
		".mem/specs/login/tasks/02_auth.md":           front("title: Auth\nstatus: todo"),
		".mem/specs/archive/search/spec.md":           front("title: Search\nstatus: completed\nassigned_to: alice"),
		".mem/specs/archive/search/tasks/01_index.md": front("title: Index\nstatus: completed"),
		".mem/specs/billing/spec.md":                  front("title: Billing\nstatus: draft"),
		".mem/specs/billing/tasks/01_plans.md":        front("title: Plans\nstatus: todo"),
	}
	var paths []string
	for _, files := range []map[string]string{before, after} {
		for path := range files {
			paths = append(paths, path)
		}
	}
	got := strings.Join(RecordChanges(paths, revision(before), revision(after)), "\n")
	want := strings.Join([]string{
		"Spec drafted: Billing",
		"Spec started by bob: Login",
		"Spec completed: Search",
		"Task completed in login: Form",
		"Task added to login: Auth",
		"Todo opened: New logo",
		"Todo closed: Old docs",
	}, "\n")
	if got != want {
		t.Fatalf("changes:\n%s\nwant:\n%s", got, want)
	}
}
