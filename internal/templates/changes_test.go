package templates

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestChangesListsNotesInTemplateThenNameOrder(t *testing.T) {
	dir := t.TempDir()
	for rel, content := range map[string]string{
		"web/template.toml":                      "description = \"Web\"\n",
		"web/changes/2026-11-02-tailwind.md":     "# Tailwind through Turbopack\n\nAdd the loader rules.\n",
		"web/changes/2026-10-29-compiler.md":     "# React Compiler on\r\n\r\nSet reactCompiler.\r\n",
		"web/changes/.2026-12-01-draft.md":       "# Draft\n",
		"web/changes/notes.txt":                  "not a note\n",
		"base/template.toml":                     "description = \"Base\"\n",
		"auth/template.toml":                     "description = \"Auth\"\n",
		"auth/changes/2026-10-01-cookie-name.md": "Rename the session cookie.\n",
	} {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Library{Dir: dir}.Changes([]string{"web", "base", "auth"})
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Template: "web", Name: "2026-10-29-compiler", Title: "React Compiler on", Body: "Set reactCompiler."},
		{Template: "web", Name: "2026-11-02-tailwind", Title: "Tailwind through Turbopack", Body: "Add the loader rules."},
		{Template: "auth", Name: "2026-10-01-cookie-name", Title: "2026-10-01-cookie-name", Body: "Rename the session cookie."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("changes =\n%#v\nwant\n%#v", got, want)
	}
}

func TestSyncReturnsEachChangeNoteOnceExceptForAdoptedTemplates(t *testing.T) {
	source, work := library(t, map[string]string{
		"web/template.toml":                      "description = \"Web\"\n",
		"web/changes/2026-10-29-compiler.md":     "# React Compiler on\n\nSet reactCompiler.\n",
		"auth/template.toml":                     "description = \"Auth\"\n",
		"auth/changes/2026-10-01-cookie-name.md": "# Session cookie renamed\n\nRename it.\n",
	})
	p := newProject(t, "web", "auth")
	syncWith := func(adopted ...string) Result {
		t.Helper()
		lib, _, err := Open(context.Background(), source, true)
		if err != nil {
			t.Fatal(err)
		}
		res, err := Sync(p, lib, adopted...)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	names := func(res Result) []string {
		var keys []string
		for _, c := range res.Changes {
			keys = append(keys, c.Key())
		}
		return keys
	}

	if got := names(syncWith("web")); !reflect.DeepEqual(got, []string{"auth/2026-10-01-cookie-name"}) {
		t.Fatalf("first sync returned %v; want only the note of the template that was not adopted", got)
	}
	if lock := read(t, p.Root, LockPath); !strings.Contains(lock, "name = '2026-10-29-compiler'") || !strings.Contains(lock, "name = '2026-10-01-cookie-name'") {
		t.Fatalf("lock does not record both notes as seen:\n%s", lock)
	}
	if res := syncWith(); len(res.Changes) != 0 || slices.Contains(res.Paths, LockPath) {
		t.Fatalf("second sync returned %v and changed %v", names(res), res.Paths)
	}

	commitFiles(t, work, map[string]string{
		"web/changes/2026-11-02-tailwind.md":     "# Tailwind through Turbopack\n",
		"auth/changes/2026-10-01-cookie-name.md": "",
	})
	res := syncWith()
	if got := names(res); !reflect.DeepEqual(got, []string{"web/2026-11-02-tailwind"}) {
		t.Fatalf("sync after a new note returned %v", got)
	}
	if lock := read(t, p.Root, LockPath); strings.Contains(lock, "cookie-name") || !strings.Contains(lock, "2026-11-02-tailwind") {
		t.Fatalf("lock keeps a deleted note or misses the new one:\n%s", lock)
	}
}
