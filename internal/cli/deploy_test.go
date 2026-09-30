package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployReleasesEverythingInOneRun(t *testing.T) {
	mine, _ := sharedProject(t)
	commit(t, mine, "feature.txt")

	out := mem(t, mine, "deploy")
	head := run(t, mine, "rev-parse", "HEAD")
	for _, branch := range []string{"dev", "test", "main"} {
		if got := run(t, mine, "rev-parse", "origin/"+branch); got != head {
			t.Fatalf("origin/%s = %s, want the development head %s\n%s", branch, got, head, out)
		}
	}
	tag := run(t, mine, "describe", "--tags", "--exact-match", "origin/main")
	if notes := run(t, mine, "for-each-ref", "--format=%(contents)", "refs/tags/"+tag); !strings.Contains(notes, "Release "+tag) || !strings.Contains(notes, "- feature.txt") {
		t.Fatalf("generated tag message = %q", notes)
	}
	if !strings.Contains(out, "Pushed 1 commit(s) on dev.") || strings.Contains(out, "AGENT INSTRUCTION") {
		t.Fatalf("deploy output:\n%s", out)
	}
	if again := mem(t, mine, "deploy"); strings.Count(again, "is already current.") != 2 {
		t.Fatalf("second deploy:\n%s", again)
	}
}

func TestDeployRefusesRepositoriesThatRequireProductionPRs(t *testing.T) {
	mine, _ := sharedProject(t)
	config := filepath.Join(mine, ".mem", "config.toml")
	data, _ := os.ReadFile(config)
	if err := os.WriteFile(config, append(data, []byte("\n[release]\nproduction_pr = true\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, mine, "commit", "--quiet", "-am", "Require production PRs")
	run(t, mine, "push", "--quiet")

	stagingBefore := run(t, mine, "ls-remote", "origin", "refs/heads/test")
	if err := memErr(t, mine, "deploy"); err == nil || !strings.Contains(err.Error(), "through a pull request") {
		t.Fatalf("deploy in a PR repository: %v", err)
	}
	if after := run(t, mine, "ls-remote", "origin", "refs/heads/test"); after != stagingBefore {
		t.Fatalf("a refused deploy moved staging: %q -> %q", stagingBefore, after)
	}
	if help := mem(t, mine, "deploy", "--help"); strings.Contains(help, "force") {
		t.Fatalf("--force is visible in help:\n%s", help)
	}
	mem(t, mine, "deploy", "--force")
	if got, want := run(t, mine, "rev-parse", "origin/main"), run(t, mine, "rev-parse", "HEAD"); got != want {
		t.Fatalf("forced deploy did not release: main %s, head %s", got, want)
	}
}
