package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
)

// productionClone returns a project's development checkout and a fresh clone of it on production, the remote's
// default branch, the way a teammate gets it from GitHub. The project protects its branches, and the clone's onboard
// has a project file to commit: the install guide is missing.
func productionClone(t *testing.T) (mine, clone string) {
	t.Helper()
	mine, _ = sharedProject(t)
	config, err := os.ReadFile(filepath.Join(mine, ".mem", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, mine, ".mem/config.toml", strings.Replace(string(config), "protect = false", "protect = true", 1))
	run(t, mine, "rm", "--quiet", ".mem/install.md")
	run(t, mine, "commit", "--quiet", "--all", "-m", "protect, no guide")
	run(t, mine, "push", "--quiet", "origin", "dev", "dev:test", "dev:main")
	base := filepath.Dir(mine)
	clone = filepath.Join(base, "fresh")
	run(t, base, "clone", "--quiet", filepath.Join(base, "remote.git"), clone)
	run(t, clone, "config", "user.name", "Teammate")
	run(t, clone, "config", "user.email", "teammate@example.com")
	return mine, clone
}

func TestOnboardSwitchesAFreshCloneOffProductionBeforeCommitting(t *testing.T) {
	_, clone := productionClone(t)
	published := run(t, clone, "rev-parse", "origin/main")
	writeFile(t, clone, "hotfix.txt", "local fix\n")
	run(t, clone, "add", "hotfix.txt")
	run(t, clone, "commit", "--quiet", "-m", "local fix on production")
	local := run(t, clone, "rev-parse", "main")

	out := mem(t, clone, "onboard")

	if branch := run(t, clone, "branch", "--show-current"); branch != "dev" || !strings.Contains(out, "Switched from main to dev") {
		t.Fatalf("onboard left the clone on %s:\n%s", branch, out)
	}
	if subject := run(t, clone, "log", "-1", "--format=%s", "dev"); subject != checkpoint.ProjectFilesCommit || run(t, clone, "rev-parse", "dev") != run(t, clone, "rev-parse", "origin/dev") {
		t.Fatalf("the project files commit is not on dev and pushed (dev's last commit %q):\n%s", subject, out)
	}
	if remote := run(t, clone, "ls-remote", "origin", "refs/heads/main"); !strings.HasPrefix(remote, published) || run(t, clone, "rev-parse", "main") != local {
		t.Fatalf("onboard changed production (remote %s, local %s):\n%s", remote, run(t, clone, "rev-parse", "main"), out)
	}
	if !strings.Contains(out, "⚠️ The local main branch has 1 commit(s) that are not on origin") {
		t.Fatalf("onboard did not warn about the commit left on main:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(clone, ".git", "hooks", "pre-commit")); err != nil {
		t.Fatalf("the protection hooks are not installed: %v", err)
	}
}

func TestOnboardCommitsNothingWhenItCannotLeaveProduction(t *testing.T) {
	mine, clone := productionClone(t)
	writeFile(t, mine, "README.md", "changed on dev\n")
	run(t, mine, "commit", "--quiet", "--all", "-m", "readme on dev")
	run(t, mine, "push", "--quiet", "origin", "dev")
	run(t, clone, "fetch", "--quiet")
	writeFile(t, clone, "README.md", "edited in the clone\n")
	before := run(t, clone, "rev-parse", "main")

	out := mem(t, clone, "onboard")

	if branch := run(t, clone, "branch", "--show-current"); branch != "main" || !strings.Contains(out, "⚠️ This checkout is on main, which only moves through `mem promote`, and could not switch to dev") {
		t.Fatalf("onboard on %s did not report the failed switch:\n%s", branch, out)
	}
	if run(t, clone, "rev-parse", "main") != before || run(t, clone, "status", "--porcelain") != "M README.md" {
		t.Fatalf("onboard committed or wrote project files on production:\n%s\n%s", run(t, clone, "status", "--porcelain"), out)
	}
	if _, err := os.Stat(filepath.Join(clone, ".git", "hooks", "pre-commit")); err != nil {
		t.Fatalf("the protection hooks are not installed: %v", err)
	}
}
