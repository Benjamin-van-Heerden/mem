package templates

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

func newProject(t *testing.T, use ...string) *project.Project {
	t.Helper()
	root := t.TempDir()
	agents, err := agentsmd.Install("# Project\n", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "AGENTS.md", agents)
	p := &project.Project{Root: root, Config: project.Config{Schema: project.Schema, Name: "demo", Templates: project.TemplatesConfig{Use: use}}}
	if err := project.WriteConfig(root, p.Config); err != nil {
		t.Fatal(err)
	}
	return p
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r", "")
}

func memory(t *testing.T, p *project.Project, name string) string {
	t.Helper()
	memories, err := agentsmd.Memories(read(t, p.Root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range memories {
		if m.Name == name {
			return m.Body
		}
	}
	return ""
}

func sync(t *testing.T, p *project.Project, source string) Result {
	t.Helper()
	lib, _, err := Open(context.Background(), source, true)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Sync(p, lib)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func mentions(res Result, text string) bool {
	return strings.Contains(strings.Join(res.Lines, "\n"), text)
}

func TestSyncInstallsUpdatesAndRespectsLocalDecisions(t *testing.T) {
	source, work := library(t, map[string]string{
		"web/template.toml":          "description = \"Web\"\n",
		"web/memories/logging.md":    "Use the logger.\n",
		"web/skills/review/SKILL.md": "v1\n",
		"web/skills/review/check.sh": "echo check\n",
		"web/docs/routing.md":        "# Routing v1\n",
		"web/docs/deploy.md":         "# Deploy v1\n",
		"web/memories/retired.md":    "Old convention.\n",
	})
	p := newProject(t, "web")

	res := sync(t, p, source)
	if memory(t, p, "logging") != "Use the logger." || read(t, p.Root, ".agents/skills/review/check.sh") != "echo check\n" || read(t, p.Root, ".mem/docs/routing.md") != "# Routing v1\n" {
		t.Fatalf("items not installed: %v", res.Lines)
	}
	if target, err := os.Readlink(filepath.Join(p.Root, ".claude/skills/review")); err != nil || filepath.ToSlash(target) != "../../.agents/skills/review" {
		t.Fatalf("skill link = %q, %v", target, err)
	}
	for _, path := range []string{"AGENTS.md", ".agents/skills/review", ".claude/skills/review", ".mem/docs/routing.md", LockPath} {
		if !slices.Contains(res.Paths, path) {
			t.Fatalf("changed paths %v miss %s", res.Paths, path)
		}
	}
	if again := sync(t, p, source); len(again.Lines) != 0 || len(again.Paths) != 0 {
		t.Fatalf("a second sync changed something: %+v", again)
	}

	write(t, p.Root, ".mem/docs/routing.md", "# Routing, edited here\n")
	write(t, p.Root, ".agents/skills/review/SKILL.md", "edited here\n")
	os.Remove(filepath.Join(p.Root, ".mem/docs/deploy.md"))
	commitFiles(t, work, map[string]string{
		"web/memories/logging.md":    "Use the structured logger.\n",
		"web/skills/review/SKILL.md": "v2\n",
		"web/memories/retired.md":    "",
	})
	res = sync(t, p, source)

	if memory(t, p, "logging") != "Use the structured logger." || !mentions(res, "Updated memory logging") {
		t.Fatalf("an unedited memory was not updated: %v", res.Lines)
	}
	if !mentions(res, "Doc routing has local edits") || read(t, p.Root, ".mem/docs/routing.md") != "# Routing, edited here\n" {
		t.Fatalf("a local-only edit was not reported or was overwritten: %v", res.Lines)
	}
	if !mentions(res, "⚠️ Skill review was edited both") || read(t, p.Root, ".agents/skills/review/SKILL.md") != "edited here\n" {
		t.Fatalf("a two-sided edit was not flagged or was overwritten: %v", res.Lines)
	}
	if !slices.Contains(p.Config.Templates.Exclude, "doc:deploy") || !mentions(res, "now excluded") {
		t.Fatalf("a deleted item was not excluded: %v %v", p.Config.Templates.Exclude, res.Lines)
	}
	if memory(t, p, "retired") != "Old convention." || !mentions(res, "memory retired is no longer provided") {
		t.Fatalf("an item removed from the template was not kept and reported: %v", res.Lines)
	}
	if strings.Contains(read(t, p.Root, LockPath), "retired") || !strings.Contains(read(t, p.Root, ".mem/config.toml"), "doc:deploy") {
		t.Fatal("the lock or config was not saved")
	}

	sync(t, p, source)
	if _, err := os.Stat(filepath.Join(p.Root, ".mem/docs/deploy.md")); err == nil {
		t.Fatal("an excluded doc came back")
	}
}

func TestSyncLeavesAnExistingDifferentItemAndFlagsIt(t *testing.T) {
	source, _ := library(t, map[string]string{
		"web/template.toml":          "description = \"Web\"\n",
		"web/skills/review/SKILL.md": "template\n",
	})
	p := newProject(t, "web")
	write(t, p.Root, ".agents/skills/review/SKILL.md", "mine\n")

	res := sync(t, p, source)
	if !mentions(res, "⚠️ This project already has a skill named review") || read(t, p.Root, ".agents/skills/review/SKILL.md") != "mine\n" {
		t.Fatalf("an existing skill was overwritten or not flagged: %v", res.Lines)
	}
}

func TestResetTakesTheTemplateCopyAndLiftsAnExclusion(t *testing.T) {
	source, _ := library(t, map[string]string{
		"web/template.toml":        "description = \"Web\"\n",
		"base/template.toml":       "description = \"Base\"\n",
		"web/docs/routing.md":      "# Routing\n",
		"web/skills/lint/SKILL.md": "template lint\n",
	})
	p := newProject(t, "base", "web")
	sync(t, p, source)
	os.Remove(filepath.Join(p.Root, ".mem/docs/routing.md"))
	write(t, p.Root, ".agents/skills/lint/SKILL.md", "local lint\n")
	sync(t, p, source)
	lib, _, _ := Open(context.Background(), source, false)

	if _, err := Reset(p, lib, Doc, "routing"); err != nil {
		t.Fatal(err)
	}
	if read(t, p.Root, ".mem/docs/routing.md") != "# Routing\n" || slices.Contains(p.Config.Templates.Exclude, "doc:routing") {
		t.Fatal("reset did not restore an excluded doc")
	}
	if _, err := Reset(p, lib, Skill, "lint"); err != nil || read(t, p.Root, ".agents/skills/lint/SKILL.md") != "template lint\n" {
		t.Fatalf("reset did not replace local edits: %v", err)
	}
	if res := sync(t, p, source); len(res.Lines) != 0 {
		t.Fatalf("items were not in sync after reset: %v", res.Lines)
	}

	write(t, p.Root, ".mem/docs/untracked.md", "mine\n")
	if _, err := Promote(context.Background(), p, lib, Doc, "untracked", ""); err == nil || !strings.Contains(err.Error(), "--to") {
		t.Fatalf("promoting an untracked item with two templates did not ask for --to: %v", err)
	}
}
