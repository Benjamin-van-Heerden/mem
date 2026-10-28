package cli

import (
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// conventionTodoMark starts the lines that report a recorded convention todo, so instructions can refer to them.
const conventionTodoMark = "📌"

// conventionTodoInstruction is the step that follows 📌 lines. A pending setup was copied from the template before
// these conventions arrived, so they belong in its remaining steps.
func conventionTodoInstruction(setupPending bool) string {
	text := "Tell the user about each 📌 item under 🧩 TEMPLATES: a template convention arrived, and its todo asks for the existing code to be checked against it. "
	if setupPending {
		text += fmt.Sprintf("The setup in %s was copied from the template before the convention arrived, so its steps may not build what the convention describes: follow the convention in the remaining setup steps, and check the code against it once the setup is done.", templates.SetupPath)
	} else {
		text += "Ask whether to do that now or later; until then, follow the convention in new work."
	}
	return text + " If it relies on tooling or setup this project has not adopted, say so and ask the user whether to adopt it or drop the convention, as its todo describes."
}

// conventionFit covers a convention that cannot simply be checked against the code, because the project never
// adopted what it builds on.
func conventionFit(name string) string {
	return fmt.Sprintf("If it relies on tooling or setup this project has not adopted (a package, service, script or workflow the code does not use), do not apply it piecemeal: tell the user, and ask whether to adopt it, as its own todo or spec, or to drop the convention with `mem memory remove %s`. Where it contradicts a project memory, update or remove whichever no longer holds once the user decides, with `mem memory set` or `mem memory remove`, rather than leave both.", name)
}

func conventionCheck(name string) string {
	return "Code written before it may not follow it. Check the existing code against it: fix small deviations directly, and record larger ones as their own todo or spec. " + conventionFit(name) + " Delete this todo once the code follows the convention, or once the user decides it does not apply here."
}

// conventionTodos records a todo for each memory a template sync added or changed, because a memory only
// guides new work: the code already written needs checking against it. A memory with an open todo gets its
// todo rewritten rather than a second one. A pending setup gets the todos too: init copies the template's
// setup once and never updates it, so it may not build a convention that arrived later. Memories installed
// with the setup came from the same template revision and never reach here, so they record nothing.
func conventionTodos(p project.Project, changes []templates.MemoryChange) (lines, paths []string, err error) {
	if len(changes) == 0 {
		return nil, nil, nil
	}
	_, setup, err := readSetup(p)
	if err != nil {
		return nil, nil, err
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
		check := "asks for the existing code to be checked against it"
		if setup.present {
			body += fmt.Sprintf("It arrived while %s was pending. The setup was copied from the template before this change, so its steps may not build what the convention describes. Follow the convention in the remaining setup steps, and once the setup is done, check the code it built as described below.\n\n", templates.SetupPath)
			check = "asks for it to be followed in the remaining setup steps and checked once the setup is done"
		}
		todo, err := upsertTodo(p, "Apply the template convention "+change.Name, body+conventionCheck(change.Name))
		if err != nil {
			return nil, nil, err
		}
		paths = append(paths, p.Rel(todo.Path))
		lines = append(lines, fmt.Sprintf("%s The %s template %s the convention %s, so todo %s %s.", conventionTodoMark, change.Template, verb, change.Name, todo.Slug, check))
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
	b.WriteString("\nCode written before them may not follow them. Check the existing code against each: fix small deviations directly, and record larger ones as their own todo or spec. " + conventionFit("<name>") + " Remove a convention that does not fit this project for another reason the same way, after asking the user. Delete this todo once the code follows the conventions that stay.")
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
