package work

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

const (
	TaskTodo      = "todo"
	TaskCompleted = "completed"
)

var taskFile = regexp.MustCompile(`^(\d+)_(.+)\.md$`)

type TaskMeta struct {
	Title     string `yaml:"title"`
	Status    string `yaml:"status"`
	Created   string `yaml:"created_at"`
	Updated   string `yaml:"updated_at"`
	Completed string `yaml:"completed_at,omitempty"`
}

type Task struct {
	Slug  string
	Order int
	Path  string
	Meta  TaskMeta
	Body  string
}

func (t Task) slug() string  { return t.Slug }
func (t Task) title() string { return t.Meta.Title }
func (t Task) Done() bool    { return t.Meta.Status == TaskCompleted }

func Tasks(s Spec) ([]Task, error) {
	dir := filepath.Join(s.Dir, "tasks")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, e := range entries {
		m := taskFile.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		order, _ := strconv.Atoi(m[1])
		t := Task{Slug: m[2], Order: order, Path: filepath.Join(dir, e.Name())}
		if t.Body, err = ReadMarkdown(t.Path, &t.Meta); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })
	return tasks, nil
}

func PendingTasks(tasks []Task) []Task {
	var pending []Task
	for _, t := range tasks {
		if !t.Done() {
			pending = append(pending, t)
		}
	}
	return pending
}

func FindTask(s Spec, ref string) (Task, error) {
	tasks, err := Tasks(s)
	if err != nil {
		return Task{}, err
	}
	return resolve("task", ref, tasks)
}

func NewTask(s Spec, title, description string) (Task, error) {
	tasks, err := Tasks(s)
	if err != nil {
		return Task{}, err
	}
	slug, err := uniqueSlug(project.Slugify(title), func(slug string) bool {
		for _, t := range tasks {
			if t.Slug == slug {
				return true
			}
		}
		return false
	})
	if err != nil {
		return Task{}, err
	}
	order := 1
	if len(tasks) > 0 {
		order = tasks[len(tasks)-1].Order + 1
	}
	t := Task{Slug: slug, Order: order, Path: filepath.Join(s.Dir, "tasks", fmt.Sprintf("%02d_%s.md", order, slug)), Body: description}
	t.Meta = TaskMeta{Title: title, Status: TaskTodo, Created: now(), Updated: now()}
	return t, WriteMarkdown(t.Path, t.Meta, t.Body)
}

func CompleteTask(t Task, notes string) (Task, error) {
	t.Meta.Status = TaskCompleted
	t.Meta.Updated = now()
	t.Meta.Completed = t.Meta.Updated
	if notes = strings.TrimSpace(notes); notes != "" {
		t.Body = strings.TrimSpace(t.Body) + "\n\n## Completion Notes\n\n" + notes
	}
	return t, WriteMarkdown(t.Path, t.Meta, t.Body)
}
