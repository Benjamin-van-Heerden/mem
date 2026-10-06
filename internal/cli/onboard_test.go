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
	writeFile(t, teammate, ".mem/docs/glossary.md", strings.Repeat("A long project doc that pushes the context into a file.\n", 300))
	writeFile(t, teammate, ".agents/skills/deploy/SKILL.md", "---\nname: deploy\ndescription: Deploy the app to staging.\n---\n\nRun the deploy script.\n")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Add a memory and a skill")
	run(t, teammate, "push", "--quiet")

	contextFile := filepath.Join(mine, ".mem/local/onboard.md")
	out := mem(t, mine, "onboard")
	context, _ := os.ReadFile(contextFile)
	for _, want := range []string{"🧠 CHANGED MEMORIES\n---", "New: logging\nUse the structured logger.", "🛠️ CHANGED SKILLS\n---", "New: deploy (.agents/skills/deploy/SKILL.md)\n  Deploy the app to staging."} {
		if !strings.Contains(string(context), want) || strings.Contains(out, want) {
			t.Fatalf("%q belongs in onboard.md only.\nstdout:\n%s\nonboard.md:\n%s", want, out, context)
		}
	}
	if !strings.Contains(out, "Follow the memories under 🧠 CHANGED MEMORIES") || strings.Contains(out, "Read AGENTS.md again") {
		t.Fatalf("onboard instruction:\n%s", out)
	}

	out = mem(t, mine, "onboard")
	context, _ = os.ReadFile(contextFile)
	if strings.Contains(string(context), "CHANGED MEMORIES") || strings.Contains(string(context), "CHANGED SKILLS") || strings.Contains(out, "Follow the memories") {
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

func TestOnboardGivesTeammatesTheInstallGuide(t *testing.T) {
	mine, teammate := sharedProject(t)
	if _, err := os.Stat(filepath.Join(mine, ".mem", "install.md")); err != nil {
		t.Fatalf("init did not write the install guide: %v", err)
	}
	run(t, mine, "rm", "--quiet", ".mem/install.md")
	run(t, mine, "commit", "--quiet", "-m", "A project from before the install guide")
	run(t, mine, "push", "--quiet")

	if out := mem(t, mine, "onboard"); !strings.Contains(out, "Updated .mem/install.md") {
		t.Fatalf("onboard output:\n%s", out)
	}
	run(t, teammate, "pull", "--quiet")
	guide, err := os.ReadFile(filepath.Join(teammate, ".mem", "install.md"))
	if err != nil || !strings.Contains(string(guide), "releases/download/$tag") {
		t.Fatalf("teammate's install guide = %q, %v", guide, err)
	}
	agents, _ := os.ReadFile(filepath.Join(teammate, "AGENTS.md"))
	if !strings.Contains(string(agents), "read `.mem/install.md`") {
		t.Fatal("AGENTS.md does not point to the install guide")
	}
}

func TestOnboardInstallsTheCompactionHookAndHonoursTheOptOut(t *testing.T) {
	mine, _ := sharedProject(t)
	settings := filepath.Join(mine, ".claude/settings.json")
	run(t, mine, "rm", "--quiet", ".claude/settings.json")
	run(t, mine, "commit", "--quiet", "-m", "Drop the Claude settings")
	run(t, mine, "push", "--quiet")

	out := mem(t, mine, "onboard")
	if data, _ := os.ReadFile(settings); !strings.Contains(out, "Installed the Claude Code compaction hook") || !strings.Contains(string(data), "mem hook compact") {
		t.Fatalf("onboard output:\n%s\nsettings:\n%s", out, data)
	}
	if status := run(t, mine, "status", "--porcelain"); status != "" {
		t.Fatalf("uncommitted after onboard:\n%s", status)
	}

	config := filepath.Join(mine, ".mem/config.toml")
	data, _ := os.ReadFile(config)
	os.WriteFile(config, append(data, []byte("\n[claude]\ncompact_hook = false\n")...), 0o644)
	run(t, mine, "commit", "--quiet", "-am", "Turn the compaction hook off")
	out = mem(t, mine, "onboard")
	if data, _ := os.ReadFile(settings); !strings.Contains(out, "Removed the Claude Code compaction hook") || strings.Contains(string(data), "mem hook compact") {
		t.Fatalf("onboard output:\n%s\nsettings:\n%s", out, data)
	}
}

func TestOnboardInstallsTheCompactionHookWhereClaudeSettingsAreIgnored(t *testing.T) {
	mine, _ := sharedProject(t)
	run(t, mine, "rm", "--quiet", ".claude/settings.json")
	writeFile(t, mine, ".gitignore", "/.mem/local/\n.claude/\n")
	run(t, mine, "commit", "--quiet", "-am", "Ignore Claude settings")
	run(t, mine, "push", "--quiet")

	out := mem(t, mine, "onboard")
	data, _ := os.ReadFile(filepath.Join(mine, ".claude/settings.json"))
	if !strings.Contains(string(data), "mem hook compact") || !strings.Contains(out, "it stays on this machine") || strings.Contains(out, "could not commit") {
		t.Fatalf("onboard output:\n%s\nsettings:\n%s", out, data)
	}
}

func TestInitRemovesACLAUDELinkButKeepsItsOwnFile(t *testing.T) {
	for _, own := range []bool{false, true} {
		base := t.TempDir()
		t.Setenv("HOME", filepath.Join(base, "home"))
		repo := filepath.Join(base, "repo")
		run(t, base, "init", "--quiet", "-b", "main", repo)
		writeFile(t, repo, "AGENTS.md", "# Notes\n")
		if own {
			writeFile(t, repo, "CLAUDE.md", "# Claude notes\n")
		} else if err := os.Symlink("AGENTS.md", filepath.Join(repo, "CLAUDE.md")); err != nil {
			t.Skip("symlinks unavailable:", err)
		}
		commit(t, repo, "README.md")
		mem(t, repo, "init", "--protect=false")
		_, err := os.Lstat(filepath.Join(repo, "CLAUDE.md"))
		if own == (err != nil) {
			t.Fatalf("own CLAUDE.md=%v: CLAUDE.md exists=%v", own, err == nil)
		}
	}
}
