package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboardShowsMemoriesAndSkillsThatArriveDuringTheSync(t *testing.T) {
	mine, teammate := sharedProject(t)
	mem(t, teammate, "memory", "set", "logging", "Use the structured logger.")
	writeFile(t, teammate, ".agents/skills/deploy/SKILL.md", "---\nname: deploy\ndescription: Deploy the app to staging.\n---\n\nRun the deploy script.\n")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Add a memory and a skill")
	run(t, teammate, "push", "--quiet")

	out := mem(t, mine, "onboard")
	for _, want := range []string{"🧠 CHANGED MEMORIES", "New: logging\nUse the structured logger.", "🛠️ CHANGED SKILLS", "New: deploy (.agents/skills/deploy/SKILL.md)\n  Deploy the app to staging.", "Follow the memories under 🧠 CHANGED MEMORIES"} {
		if !strings.Contains(out, want) {
			t.Fatalf("onboard output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Read AGENTS.md again") {
		t.Fatalf("onboard still asks to re-read AGENTS.md:\n%s", out)
	}

	out = mem(t, mine, "onboard")
	if strings.Contains(out, "CHANGED MEMORIES") || strings.Contains(out, "CHANGED SKILLS") || strings.Contains(out, "Follow the memories") {
		t.Fatalf("a second onboard reports changes again:\n%s", out)
	}
}

func TestOnboardRestoresTheLocalIgnoreAndPublishesIt(t *testing.T) {
	mine, _ := sharedProject(t)
	if err := os.WriteFile(filepath.Join(mine, ".gitignore"), []byte("node_modules/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, mine, "commit", "--quiet", "-am", "Drop the mem ignore")
	run(t, mine, "push", "--quiet")

	out := mem(t, mine, "onboard")
	if !strings.Contains(out, "Added .mem/local/ to .gitignore.") {
		t.Fatalf("onboard output:\n%s", out)
	}
	data, _ := os.ReadFile(filepath.Join(mine, ".gitignore"))
	if string(data) != "node_modules/\n/.mem/local/\n" {
		t.Fatalf(".gitignore = %q", data)
	}
	if status := run(t, mine, "status", "--porcelain"); status != "" {
		t.Fatalf("uncommitted after onboard:\n%s", status)
	}
	if local, remote := run(t, mine, "rev-parse", "HEAD"), run(t, mine, "rev-parse", "origin/dev"); local != remote {
		t.Fatal("the .gitignore change was not pushed")
	}
}
