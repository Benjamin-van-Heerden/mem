package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/memr/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/memr/internal/work"
)

func TestImportConvertsHarnessStateAndKeepsUserContent(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":                                            "# Team notes\n\n<core_instructions>\nold harness text\n</core_instructions>\n\nKeep this footer.\n",
		".agent_core/config.toml":                              "[project]\nname = \"demo\"\ndescription = \"\"\"\nA demo\nproject.\n\"\"\"\n\n[[files]]\npath = \"docs/arch.md\"\ndescription = \"Architecture\"\n\n[branches]\ndev = \"development\"\ntest = \"test\"\nmain = \"production\"\n",
		".agent_core/user_mappings.toml":                       "[octocat]\nname = \"Octo Cat\"\nemail = \"o@x\"\n",
		".agent_core/memories/style.md":                        "---\ntitle: Style\n---\nUse tabs.\n\n## Detail\nAlways.\n",
		".agent_core/docs/codebase_and_structure.md":           "# Codebase\n",
		".agent_core/docs/idea.md":                             "# Idea\n",
		".agent_core/specs/login/spec.md":                      "---\ntitle: Login\nstatus: todo\nassigned_to: octocat\ncreated_at: '2026-05-27T09:36:56.756450'\nupdated_at: '2026-05-27T09:36:56.756450'\n---\n## Overview\nLogin.\n",
		".agent_core/specs/login/tasks/01_form.md":             "---\ntitle: Form\nstatus: completed\ncreated_at: '2026-05-27T09:36:56'\nupdated_at: '2026-05-27T09:36:56'\ncompleted_at: '2026-05-27T10:00:00'\n---\nBuild it.\n",
		".agent_core/specs/completed/auth/spec.md":             "---\ntitle: Auth\nstatus: completed\ncreated_at: '2026-05-01T09:00:00'\nupdated_at: '2026-05-02T09:00:00'\ncompleted_at: '2026-05-02T09:00:00'\n---\nDone.\n",
		".agent_core/specs/completed/auth/handoff.md":          "Extra file.\n",
		".agent_core/todos/email.md":                           "---\ntitle: Email\nstatus: claimed\ncreated_at: '2026-07-06T13:16:26.321585'\nclaimed_by: octocat\nclaimed_at: '2026-07-07T10:00:00'\n---\nSend email.\n",
		".agent_core/logs/octo_cat_20260910_094413_session.md": "---\ncreated_at: '2026-09-10T09:44:13.513279'\nusername: octo_cat\nspec_slug: login\n---\nWork Log - Test\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	sum, err := Import(root, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	p := sum.Project
	if p.Config.Description != "A demo project." || p.Config.Git.Development != "development" || p.Config.Git.Production != "production" {
		t.Fatalf("config = %+v", p.Config)
	}

	agents, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	text := string(agents)
	if strings.Contains(text, "old harness text") || !strings.HasPrefix(text, "# Team notes\n\n<memr>") || !strings.Contains(text, "</memr>\n\nKeep this footer.") {
		t.Fatalf("AGENTS.md not converted in place:\n%s", text)
	}
	if memories, _ := agentsmd.Memories(text); len(memories) != 1 || memories[0].Body != "Use tabs.\n\n### Detail\nAlways." {
		t.Fatalf("memories = %#v", memories)
	}

	login, err := work.FindSpec(p, "login")
	if err != nil || login.Meta.Status != work.SpecActive || login.Meta.AssignedTo != "octo_cat" || !strings.HasPrefix(login.Meta.Created, "2026-05-27T09:36:56") {
		t.Fatalf("login spec = %+v, %v", login.Meta, err)
	}
	if tasks, _ := work.Tasks(login); len(tasks) != 1 || !tasks[0].Done() {
		t.Fatalf("login tasks = %+v", tasks)
	}
	auth, err := work.FindSpec(p, "auth")
	if err != nil || !auth.Archived() {
		t.Fatalf("auth spec = %+v, %v", auth.Meta, err)
	}
	if _, err := os.Stat(filepath.Join(auth.Dir, "handoff.md")); err != nil {
		t.Fatal("extra spec file not copied")
	}
	if todo, err := work.FindTodo(p, "email"); err != nil || todo.Meta.ClaimedBy != "octo_cat" {
		t.Fatalf("todo = %+v, %v", todo.Meta, err)
	}
	if log, err := work.FindLog(p, "octo_cat_20260910_094413"); err != nil || log.Meta.Spec != "login" {
		t.Fatalf("log = %+v, %v", log.Meta, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".memr", "structure.md")); err != nil {
		t.Fatal("structure doc not moved")
	}
	if len(sum.Runnables) != 1 {
		t.Fatalf("runnables = %v", sum.Runnables)
	}
	if _, err := Import(root, "1.0.0"); err == nil {
		t.Fatal("second import was allowed")
	}
}
