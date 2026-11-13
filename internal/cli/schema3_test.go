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
	writeFile(t, mine, ".mem/config.toml", strings.Replace(string(data), "schema = 3", "schema = 2", 1)+"\n[precommit]\ncommands = ['git rev-parse HEAD']\n")
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
	if !strings.Contains(out, "Upgraded the project format to schema 3.") {
		t.Fatalf("onboard did not upgrade:\n%s", out)
	}
	if runtime.GOOS != "windows" && !strings.Contains(out, "GET /orders") {
		t.Fatalf("the migrated runnable did not run:\n%s", out)
	}
	upgraded, _ := os.ReadFile(config)
	for _, want := range []string{"[checks]", "git rev-parse HEAD", "[runnables]\nroutes = '.mem/runnables/routes'"} {
		if !strings.Contains(string(upgraded), want) {
			t.Fatalf("config after the upgrade lacks %q:\n%s", want, upgraded)
		}
	}
	if strings.Contains(string(upgraded), "[precommit]") || strings.Contains(string(upgraded), "notes.txt") {
		t.Fatalf("config after the upgrade:\n%s", upgraded)
	}
	if !strings.Contains(out, "Review the runnables moved into the config") {
		t.Fatalf("onboard does not show the review todo:\n%s", out)
	}

	mem(t, mine, "sync")
	run(t, teammate, "pull", "--quiet")
	todo, err := os.ReadFile(filepath.Join(teammate, ".mem", "todos", "review_runnables_config.md"))
	if err != nil {
		t.Fatalf("the review todo was not committed and shared: %v", err)
	}
	if !strings.Contains(string(todo), "as before: `routes`.") || !strings.Contains(string(todo), "not listed: `notes.txt`.") {
		t.Fatalf("review todo:\n%s", todo)
	}
	if out := mem(t, teammate, "todo", "show", "review_runnables_config"); !strings.Contains(out, "Review the runnables moved into the config") {
		t.Fatalf("mem cannot read the review todo:\n%s", out)
	}
}
