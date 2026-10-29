package cli

import (
	"fmt"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
)

const changeCheck = "The project was built from the template before this change. Apply it to this project's existing code as the note describes. Where the project has diverged from the template so that the note does not fit, tell the user and ask how to proceed. Where the change needs something the project has not adopted, tell the user and ask whether to adopt it, as its own todo or spec. Delete this todo once the project matches the note, or once the user decides it does not apply here."

// changeTodos records a todo for each change note a sync saw for the first time. A note is how a template says
// that projects must change their own files, such as copied starter files, which no sync touches. The todo quotes
// the note in full, so it can be acted on without the template library.
func changeTodos(p project.Project, changes []templates.Change) (lines, paths []string, err error) {
	if len(changes) == 0 {
		return nil, nil, nil
	}
	_, setup, err := readSetup(p)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range changes {
		note := c.Body
		if note == "" {
			note = c.Title
		}
		body := fmt.Sprintf("The %s template changed. Its change note %s:\n\n%s\n\n", c.Template, c.Name, quote(note))
		check := "asks for the change to be applied to this project"
		if setup.present {
			body += fmt.Sprintf("It arrived while %s was pending. The setup was copied from the template before this change, so its steps may not include it. Apply the change in the remaining setup steps, or to what earlier steps built, and check it as described below once the setup is done.\n\n", templates.SetupPath)
			check = "asks for it to be applied in the remaining setup steps and checked once the setup is done"
		}
		todo, err := upsertTodo(p, fmt.Sprintf("Apply the %s template change: %s", c.Template, c.Title), body+changeCheck)
		if err != nil {
			return nil, nil, err
		}
		paths = append(paths, p.Rel(todo.Path))
		lines = append(lines, fmt.Sprintf("%s The %s template changed (%s), so todo %s %s.", templateTodoMark, c.Template, c.Title, todo.Slug, check))
	}
	return lines, paths, nil
}
