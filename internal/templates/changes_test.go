package templates

import (
	"os"
	"path/filepath"
	"reflect"
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
