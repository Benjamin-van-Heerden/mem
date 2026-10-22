package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// conventionTodoMark starts the lines that report a recorded convention todo, so instructions can refer to them.
const conventionTodoMark = "📌"

const conventionTodoInstruction = "Tell the user about each 📌 item under 🧩 TEMPLATES: a template convention arrived, and its todo asks for the existing code to be checked against it. Ask whether to do that now or later; until then, follow the convention in new work."

const conventionCheck = "Code written before it may not follow it. Check the existing code against it: fix small deviations directly, and record larger ones as their own todo or spec. Delete this todo once the code follows the convention, or once the user decides it does not apply here."

// conventionTodos records a todo for each memory a template sync added or changed, because a memory only
// guides new work: the code already written needs checking against it. A memory with an open todo gets its
// todo rewritten rather than a second one. A project still in its template setup is skipped, since the setup
// builds the code from the current template.
func conventionTodos(p project.Project, changes []templates.MemoryChange) (lines, paths []string, err error) {
	if len(changes) == 0 {
		return nil, nil, nil
	}
	if _, err := os.Stat(filepath.Join(p.Root, templates.SetupPath)); err == nil {
		return nil, nil, nil
	}
	for _, change := range changes {
		verb := "changed"
		if change.Previous == "" {
			verb = "added"
		}
		body := fmt.Sprintf("The %s template %s the convention `%s`, now in the memories section of AGENTS.md:\n\n%s\n\n", change.Template, verb, change.Name, quote(change.Body))
		if change.Previous != "" {
			body += fmt.Sprintf("It used to read:\n\n%s\n\n", quote(change.Previous))
		}
		todo, err := upsertTodo(p, "Apply the template convention "+change.Name, body+conventionCheck)
		if err != nil {
			return nil, nil, err
		}
		paths = append(paths, p.Rel(todo.Path))
		lines = append(lines, fmt.Sprintf("%s The %s template %s the convention %s, so todo %s asks for the existing code to be checked against it.", conventionTodoMark, change.Template, verb, change.Name, todo.Slug))
	}
	return lines, paths, nil
}

// templateUseTodo records one todo for the memories a newly used template brings into an existing project.
func templateUseTodo(p project.Project, name string, changes []templates.MemoryChange) (string, string, error) {
	if len(changes) == 0 {
		return "", "", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "This project started using the %s template, which added these conventions to the memories section of AGENTS.md:\n\n", name)
	for _, change := range changes {
		fmt.Fprintf(&b, "- `%s`: %s\n", change.Name, change.Body)
	}
	b.WriteString("\nCode written before them may not follow them. Check the existing code against each: fix small deviations directly, and record larger ones as their own todo or spec. Remove a convention that does not fit this project with `mem memory remove <name>` after asking the user. Delete this todo once the code follows the conventions that stay.")
	todo, err := upsertTodo(p, fmt.Sprintf("Apply the %s template's conventions", name), b.String())
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s Todo %s asks for the existing code to be checked against the %d conventions %s added.", conventionTodoMark, todo.Slug, len(changes), name), p.Rel(todo.Path), nil
}

func upsertTodo(p project.Project, title, body string) (work.Todo, error) {
	todos, err := work.Todos(p)
	if err != nil {
		return work.Todo{}, err
	}
	for _, t := range todos {
		if t.Meta.Title == title {
			t.Body = body
			return t, work.WriteMarkdown(t.Path, t.Meta, t.Body)
		}
	}
	return work.NewTodo(p, title, body)
}

func quote(text string) string {
	return "> " + strings.ReplaceAll(strings.TrimSpace(text), "\n", "\n> ")
}

func hasConventionTodo(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, conventionTodoMark) {
			return true
		}
	}
	return false
}
