package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOnboardUpgradesASchema2ProjectsChecksAndRunnables(t *testing.T) {
	t.Parallel()
	mine, teammate := sharedProject(t)
	config := filepath.Join(mine, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, mine, ".mem/config.toml", strings.Replace(string(data), "schema = 4", "schema = 2", 1)+"\n[precommit]\ncommands = ['git rev-parse HEAD']\n")
	writeFile(t, mine, ".mem/runnables/routes", "#!/bin/sh\necho 'GET /orders'\n")
	if err := os.Chmod(filepath.Join(mine, ".mem", "runnables", "routes"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, mine, ".mem/runnables/notes.txt", "never ran\n")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "A project from schema 2")
	run(t, mine, "push", "--quiet")

	if out := mem(t, mine, "check"); !strings.Contains(out, "✔ git rev-parse HEAD") {
		t.Fatalf("a project not yet upgraded should still run its checks:\n%s", out)
	}
	out := mem(t, mine, "onboard", "--offline")
	if !strings.Contains(out, "Upgraded the project format to schema 4.") {
		t.Fatalf("onboard did not upgrade:\n%s", out)
	}
	if runtime.GOOS != "windows" && !strings.Contains(out, "GET /orders") {
		t.Fatalf("the migrated runnable did not run:\n%s", out)
	}
	// Windows keeps no executable bit, so there the upgrade lists every file, including the one that never ran.
	listed, skipped := "routes = '.mem/runnables/routes'", "not listed: `notes.txt`."
	if runtime.GOOS == "windows" {
		listed, skipped = "'notes.txt' = '.mem/runnables/notes.txt'\nroutes = '.mem/runnables/routes'", "as before: `notes.txt`, `routes`."
	}
	upgraded, _ := os.ReadFile(config)
	for _, want := range []string{"[checks]", "git rev-parse HEAD", "[runnables]\n" + listed} {
		if !strings.Contains(string(upgraded), want) {
			t.Fatalf("config after the upgrade lacks %q:\n%s", want, upgraded)
		}
	}
	if strings.Contains(string(upgraded), "[precommit]") || runtime.GOOS != "windows" && strings.Contains(string(upgraded), "notes.txt") {
		t.Fatalf("config after the upgrade:\n%s", upgraded)
	}
	if !strings.Contains(out, "Critical todos are open: Review the runnables moved into the config (review_runnables_config).") {
		t.Fatalf("onboard does not lead with the critical review todo:\n%s", out)
	}

	mem(t, mine, "sync")
	run(t, teammate, "pull", "--quiet")
	todo, err := os.ReadFile(filepath.Join(teammate, ".mem", "todos", "review_runnables_config.md"))
	if err != nil {
		t.Fatalf("the review todo was not committed and shared: %v", err)
	}
	if !strings.Contains(string(todo), "`routes`") || !strings.Contains(string(todo), skipped) {
		t.Fatalf("review todo:\n%s", todo)
	}
	if out := mem(t, teammate, "todo", "show", "review_runnables_config"); !strings.Contains(out, "Review the runnables moved into the config") {
		t.Fatalf("mem cannot read the review todo:\n%s", out)
	}
}

func TestOnboardMakesARunnablesReviewFromSchema3Critical(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	config := filepath.Join(mine, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	// mem v0.15.0 wrote this todo without a priority.
	writeFile(t, mine, ".mem/config.toml", strings.Replace(string(data), "schema = 4", "schema = 3", 1))
	writeFile(t, mine, ".mem/todos/review_runnables_config.md", "---\ntitle: Review the runnables moved into the config\ncreated_at: \"2026-11-14T09:00:00+02:00\"\n---\n\nGo through them.\n")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "A project upgraded by v0.15.0")
	run(t, mine, "push", "--quiet")

	if out := mem(t, mine, "onboard", "--offline"); !strings.Contains(out, "Critical todos are open: Review the runnables moved into the config") {
		t.Fatalf("onboard does not lead with the review todo:\n%s", out)
	}
	todo, _ := os.ReadFile(filepath.Join(mine, ".mem", "todos", "review_runnables_config.md"))
	if string(todo) != "---\ntitle: Review the runnables moved into the config\npriority: critical\ncreated_at: \"2026-11-14T09:00:00+02:00\"\n---\n\nGo through them.\n" {
		t.Fatalf("todo after the upgrade:\n%s", todo)
	}
	if status := run(t, mine, "status", "--porcelain"); status != "" {
		t.Fatalf("the upgrade was not committed:\n%s", status)
	}
}
