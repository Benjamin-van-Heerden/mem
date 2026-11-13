package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// checksConfig adds a [checks] section to the project's config and publishes it.
func checksConfig(t *testing.T, root, commands string) {
	t.Helper()
	config := filepath.Join(root, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, append(data, []byte("\n[checks]\ncommands = "+commands+"\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, "commit", "--quiet", "-am", "Project checks")
	run(t, root, "push", "--quiet")
}

func TestCheckRunsChecksInOrderAndStopsAtTheFirstFailure(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	if out := mem(t, mine, "check"); !strings.Contains(out, "No checks are set") || !strings.Contains(out, "[checks]") {
		t.Fatalf("no checks set:\n%s", out)
	}

	checksConfig(t, mine, `['git status --short', 'git rev-parse HEAD']`)
	if out := mem(t, mine, "check"); !strings.Contains(out, "✔ git status --short") || !strings.Contains(out, "✔ git rev-parse HEAD") {
		t.Fatalf("passing checks:\n%s", out)
	}

	config, err := os.ReadFile(filepath.Join(mine, ".mem", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	before, _, _ := strings.Cut(string(config), "\n[checks]")
	writeFile(t, mine, ".mem/config.toml", before+"\n[checks]\ncommands = ['git nosuchcommand', 'git rev-parse HEAD']\n")
	cmd := New()
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"check", "--dir", mine})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "checks failed: git nosuchcommand") {
		t.Fatalf("a failing check returned %v", err)
	}
	if !strings.Contains(out.String(), "not a git command") || strings.Contains(out.String(), "rev-parse") {
		t.Fatalf("a failure should show its output and stop:\n%s", out.String())
	}
}

func TestPromotingToStagingRefusesWhenTheChecksFail(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	checksConfig(t, mine, `['git nosuchcommand']`)
	before := run(t, mine, "ls-remote", "origin", "refs/heads/test")
	if err := memErr(t, mine, "promote", "staging"); err == nil || !strings.Contains(err.Error(), "checks failed") {
		t.Fatalf("promote staging returned %v", err)
	}
	if after := run(t, mine, "ls-remote", "origin", "refs/heads/test"); after != before {
		t.Fatalf("staging moved from %s to %s", before, after)
	}
}

func TestOnboardRenamesTheSchema2PrecommitTableToChecks(t *testing.T) {
	t.Parallel()
	mine, _ := sharedProject(t)
	config := filepath.Join(mine, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	old := strings.Replace(string(data), "schema = 3", "schema = 2", 1) + "\n[precommit]\ncommands = ['git rev-parse HEAD']\n"
	writeFile(t, mine, ".mem/config.toml", old)
	run(t, mine, "commit", "--quiet", "-am", "A project from schema 2")
	run(t, mine, "push", "--quiet")

	if out := mem(t, mine, "check"); !strings.Contains(out, "✔ git rev-parse HEAD") {
		t.Fatalf("a project not yet upgraded should still run its checks:\n%s", out)
	}
	if out := mem(t, mine, "onboard", "--offline"); !strings.Contains(out, "Upgraded the project format to schema 3.") {
		t.Fatalf("onboard did not upgrade:\n%s", out)
	}
	upgraded, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(upgraded), "[precommit]") || !strings.Contains(string(upgraded), "[checks]") || !strings.Contains(string(upgraded), "git rev-parse HEAD") {
		t.Fatalf("config after the upgrade:\n%s", upgraded)
	}
	if status := run(t, mine, "status", "--porcelain", "--", ".mem/config.toml"); status != "" {
		t.Fatalf("the upgraded config was not committed: %s", status)
	}
}
