package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
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
		".agent_core/specs/abandoned/sso/spec.md":              "---\ntitle: SSO\nstatus: abandoned\n---\nDropped.\n",
		".agent_core/todos/docs.md":                            "---\ntitle: Docs\nstatus: open\ncreated_at: '2026-07-06T13:16:26'\n---\nWrite docs.\n",
		".agent_core/todos/claimed/email.md":                   "---\ntitle: Email\nstatus: claimed\nissue_id: 30\nissue_url: https://github.com/o/r/issues/30\ncreated_at: '2026-07-06T13:16:26.321585'\nclaimed_by: octocat\nclaimed_at: '2026-07-07T10:00:00'\n---\nSend email.\n",
		".agent_core/logs/octo_cat_20260910_094413_session.md": "---\ncreated_at: '2026-09-10T09:44:13.513279'\nusername: octo_cat\nspec_slug: login\n---\nWork Log - Test\n\n## Overarching Goals\n\nTest.\n\n## What Comes Next\n\n- Ship it.\n\n## Errors and Barriers\n\nNone.\n",
		".agent_core/logs/octo_cat_20250926_133200_session.md": "---\ncreated_at: '2025-09-26T13:32:00'\nusername: octo_cat\n---\n# Overarching Goals\nGoals.\n\n# What Was Accomplished\n\n## Queue\n```sh\n# a shell comment\n```\n\n# What Comes Next\n```sh\n# run this next\n```\n- Later.\n",
		".agent_core/logs/octo_cat_20251106_110500_session.md": "---\ncreated_at: '2025-11-06T11:05:00'\nusername: octo_cat\n---\n# Work Log - Credentials\n\n## Overarching Goals\n",
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
	if strings.Contains(text, "old harness text") || !strings.HasPrefix(text, "# Team notes\n\n<mem>") || !strings.Contains(text, "</mem>\n\nKeep this footer.") {
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
	if sso, err := work.FindSpec(p, "sso"); err != nil || sso.Meta.Status != work.SpecAbandoned || !sso.Archived() {
		t.Fatalf("sso spec = %+v, %v", sso.Meta, err)
	}
	if todo, err := work.FindTodo(p, "email"); err != nil || todo.Meta.ClaimedBy != "octo_cat" || strings.TrimSpace(todo.Body) != "Send email.\n\nGitHub issue: https://github.com/o/r/issues/30" {
		t.Fatalf("todo = %+v, %v", todo.Meta, err)
	}
	if todo, err := work.FindTodo(p, "docs"); err != nil || todo.Meta.Status != work.TodoOpen {
		t.Fatalf("open todo = %+v, %v", todo.Meta, err)
	}
	latest, err := work.FindLog(p, "octo_cat_20260910_094413")
	if err != nil || latest.Meta.Spec != "login" || latest.Heading() != "Test" {
		t.Fatalf("log = %+v, %q, %v", latest.Meta, latest.Heading(), err)
	}
	if strings.Contains(latest.Body, "What Comes Next") || !strings.Contains(latest.Body, "## Errors and Barriers\n\nNone.") {
		t.Fatalf("next steps not removed from the latest log:\n%s", latest.Body)
	}
	if sum.LatestLog != "octo_cat_20260910_094413" || sum.NextSteps != "- Ship it." {
		t.Fatalf("latest next steps = %q from %q", sum.NextSteps, sum.LatestLog)
	}
	if log, _ := work.FindLog(p, "octo_cat_20251106_110500"); log.Heading() != "Credentials" {
		t.Fatalf("titled log heading = %q", log.Heading())
	}
	early, _ := work.FindLog(p, "octo_cat_20250926_133200")
	if early.Heading() != "Session of 2025-09-26" || !strings.HasSuffix(strings.TrimSpace(early.Body), "\n## What Was Accomplished\n\n### Queue\n```sh\n# a shell comment\n```") {
		t.Fatalf("untitled log not normalised:\n%s", early.Body)
	}
	if _, err := os.Stat(filepath.Join(root, ".mem", "structure.md")); err != nil {
		t.Fatal("structure doc not moved")
	}
	if len(sum.Runnables) != 1 {
		t.Fatalf("runnables = %v", sum.Runnables)
	}
	if _, err := Import(root, "1.0.0"); err == nil {
		t.Fatal("second import was allowed")
	}
}
