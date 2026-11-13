package work

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// TodoMeta describes a todo. A todo has no states: it is open until it is done, and then it is deleted.
type TodoMeta struct {
	Title string `yaml:"title"`
	// Priority is one of Priorities; a todo without one is medium.
	Priority string `yaml:"priority,omitempty"`
	Created  string `yaml:"created_at"`
}

// Priorities are a todo's priorities, most urgent first. A critical todo is taken up right after onboard.
var Priorities = []string{"critical", "high", "medium", "low"}

const DefaultPriority = "medium"

// CheckPriority rejects anything but one of Priorities.
func CheckPriority(priority string) error {
	if !slices.Contains(Priorities, priority) {
		return fmt.Errorf("unknown priority %q; use one of %v", priority, Priorities)
	}
	return nil
}

// Priority is the todo's priority, medium when none is recorded.
func (t Todo) Priority() string {
	if t.Meta.Priority == "" {
		return DefaultPriority
	}
	return t.Meta.Priority
}

func priorityRank(priority string) int {
	if i := slices.Index(Priorities, priority); i >= 0 {
		return i
	}
	return slices.Index(Priorities, DefaultPriority)
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
	// Most urgent first, oldest first within a priority.
	sort.SliceStable(todos, func(i, j int) bool {
		if a, b := priorityRank(todos[i].Priority()), priorityRank(todos[j].Priority()); a != b {
			return a < b
		}
		return todos[i].Meta.Created < todos[j].Meta.Created
	})
	return todos, nil
}

func FindTodo(p project.Project, ref string) (Todo, error) {
	todos, err := Todos(p)
	if err != nil {
		return Todo{}, err
	}
	return resolve("todo", ref, todos)
}

// NewTodo records a todo; an empty priority leaves it at the default.
func NewTodo(p project.Project, title, description, priority string) (Todo, error) {
	if priority != "" {
		if err := CheckPriority(priority); err != nil {
			return Todo{}, err
		}
	}
	slug, err := uniqueSlug(recordSlug(title), func(s string) bool {
		return pathExists(filepath.Join(todosDir(p), s+".md"))
	})
	if err != nil {
		return Todo{}, err
	}
	t := Todo{Slug: slug, Path: filepath.Join(todosDir(p), slug+".md"), Body: description}
	t.Meta = TodoMeta{Title: title, Priority: priority, Created: now()}
	return t, WriteMarkdown(t.Path, t.Meta, t.Body)
}

// SetTodoPriority records a new priority for the todo.
func SetTodoPriority(t Todo, priority string) (Todo, error) {
	if err := CheckPriority(priority); err != nil {
		return t, err
	}
	t.Meta.Priority = priority
	return t, WriteMarkdown(t.Path, t.Meta, t.Body)
}
