package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupJoinsTemplateSetupsInOrderAndSkipsTemplatesWithout(t *testing.T) {
	dir := t.TempDir()
	for rel, content := range map[string]string{
		"web/template.toml":  "description = \"Web\"\n",
		"web/setup.md":       "# Setup: web\n\n## [ ] 1. Scaffold\n",
		"base/template.toml": "description = \"Base\"\n",
		"auth/template.toml": "description = \"Auth\"\n",
		"auth/setup.md":      "# Setup: auth\n\n## [ ] 1. Add auth\n\n",
	} {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	lib := Library{Dir: dir}

	got, err := lib.Setup([]string{"web", "base", "auth"})
	if err != nil {
		t.Fatal(err)
	}
	want := "# Setup: web\n\n## [ ] 1. Scaffold\n\n# Setup: auth\n\n## [ ] 1. Add auth\n"
	if got != want {
		t.Fatalf("setup =\n%q\nwant\n%q", got, want)
	}
	if got, err := lib.Setup([]string{"base"}); err != nil || got != "" {
		t.Fatalf("setup without setup files = %q, %v", got, err)
	}

	templates, err := lib.Templates()
	if err != nil {
		t.Fatal(err)
	}
	for _, tpl := range templates {
		if tpl.HasSetup != (tpl.Name != "base") {
			t.Fatalf("%s HasSetup = %v", tpl.Name, tpl.HasSetup)
		}
	}
}

func TestSetupProgressCountsTickedStepHeadings(t *testing.T) {
	text := "# Setup\n\n## [x] 1. Scaffold\n\n- [ ] not a step\n\n## [X] 2. Install\n\n## [ ] 3. Verify\n\n### [ ] a sub-heading\n"
	if done, total := SetupProgress(text); done != 2 || total != 3 {
		t.Fatalf("progress = %d of %d, want 2 of 3", done, total)
	}
}
