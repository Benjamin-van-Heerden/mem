package work

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// TodoMeta describes a todo. A todo has no states: it is open until it is done, and then it is deleted.
type TodoMeta struct {
	Title   string `yaml:"title"`
	Created string `yaml:"created_at"`
}

type Todo struct {
	Slug string
	Path string
	Meta TodoMeta
	Body string
}

func (t Todo) slug() string  { return t.Slug }
func (t Todo) title() string { return t.Meta.Title }

func todosDir(p project.Project) string { return p.Path("todos") }

func Todos(p project.Project) ([]Todo, error) {
	entries, err := os.ReadDir(todosDir(p))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var todos []Todo
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || e.IsDir() {
			continue
		}
		t := Todo{Slug: slug, Path: filepath.Join(todosDir(p), e.Name())}
		if t.Body, err = ReadMarkdown(t.Path, &t.Meta); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	sort.Slice(todos, func(i, j int) bool { return todos[i].Meta.Created < todos[j].Meta.Created })
	return todos, nil
}

func FindTodo(p project.Project, ref string) (Todo, error) {
	todos, err := Todos(p)
	if err != nil {
		return Todo{}, err
	}
	return resolve("todo", ref, todos)
}

func NewTodo(p project.Project, title, description string) (Todo, error) {
	slug, err := uniqueSlug(recordSlug(title), func(s string) bool {
		return pathExists(filepath.Join(todosDir(p), s+".md"))
	})
	if err != nil {
		return Todo{}, err
	}
	t := Todo{Slug: slug, Path: filepath.Join(todosDir(p), slug+".md"), Body: description}
	t.Meta = TodoMeta{Title: title, Created: now()}
	return t, WriteMarkdown(t.Path, t.Meta, t.Body)
}
