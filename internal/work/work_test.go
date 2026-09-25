package work

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

func TestResolveBySlugTitleAndPrefix(t *testing.T) {
	p := project.Project{Root: t.TempDir()}
	for _, title := range []string{"Add login", "Add login page", "Fix footer"} {
		if _, err := NewTodo(p, title, ""); err != nil {
			t.Fatal(err)
		}
	}
	cases := map[string]string{
		"add_login":      "add_login",
		"ADD LOGIN PAGE": "add_login_page",
		"fix":            "fix_footer",
	}
	for ref, want := range cases {
		got, err := FindTodo(p, ref)
		if err != nil || got.Slug != want {
			t.Errorf("FindTodo(%q) = %q, %v; want %q", ref, got.Slug, err, want)
		}
	}
	if _, err := FindTodo(p, "add"); err == nil || !strings.Contains(err.Error(), "add_login_page") {
		t.Errorf("ambiguous prefix error = %v", err)
	}
}

func TestSpecLifecycleKeepsTaskOrderAndArchivesTasks(t *testing.T) {
	p := project.Project{Root: t.TempDir()}
	s, err := NewSpec(p, "Login")
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"Form", "Auth", "Form"} {
		if _, err := NewTask(s, title, "details"); err != nil {
			t.Fatal(err)
		}
	}
	tasks, _ := Tasks(s)
	var slugs []string
	for _, task := range tasks {
		slugs = append(slugs, task.Slug)
	}
	if strings.Join(slugs, ",") != "form,auth,form_2" {
		t.Fatalf("task order = %v", slugs)
	}
	if _, err := CompleteTask(tasks[0], "done"); err != nil {
		t.Fatal(err)
	}
	if pending, _ := Tasks(s); len(PendingTasks(pending)) != 2 {
		t.Fatal("completion was not persisted")
	}
	archived, err := ArchiveSpec(p, s, SpecCompleted, "")
	if err != nil {
		t.Fatal(err)
	}
	if archived.Dir != filepath.Join(p.Root, ".mem", "specs", "archive", "login") {
		t.Fatalf("archived to %s", archived.Dir)
	}
	if open, _ := Specs(p, false); len(open) != 0 {
		t.Fatal("archived spec still listed as open")
	}
	found, err := FindSpec(p, "login")
	if err != nil || found.Meta.Status != SpecCompleted {
		t.Fatalf("archived spec lookup = %+v, %v", found.Meta, err)
	}
	if moved, _ := Tasks(found); len(moved) != 3 {
		t.Fatal("tasks did not move with the spec")
	}
}
