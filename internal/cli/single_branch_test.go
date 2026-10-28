package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// singleBranchProject returns a checkout on dev, initialised with --single-branch, whose remote also has a main.
func singleBranchProject(t *testing.T) (root, initOut string) {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	remote := filepath.Join(base, "remote.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	root = filepath.Join(base, "mine")
	run(t, base, "clone", "--quiet", remote, root)
	run(t, root, "config", "user.name", "Test User")
	run(t, root, "config", "user.email", "test@example.com")
	commit(t, root, "README.md")
	run(t, root, "push", "--quiet", "origin", "HEAD:main")
	run(t, root, "switch", "--quiet", "-c", "dev")
	initOut = mem(t, root, "init", "--single-branch", "--protect=false", "--no-template")
	run(t, root, "add", "--all")
	run(t, root, "commit", "--quiet", "-m", "mem setup")
	run(t, root, "push", "--quiet")
	return root, initOut
}

func TestSingleBranchProjectHasNoStagesReleasesOrReleaseInstructions(t *testing.T) {
	root, out := singleBranchProject(t)
	if !strings.Contains(out, "Branches: dev (single branch; remote origin)") || !strings.Contains(out, "Published dev to origin") {
		t.Fatalf("init output:\n%s", out)
	}
	config, _ := os.ReadFile(filepath.Join(root, ".mem", "config.toml"))
	if !strings.Contains(string(config), "single_branch = true") || strings.Contains(string(config), "staging") || strings.Contains(string(config), "production") {
		t.Fatalf("config:\n%s", config)
	}
	if heads := run(t, root, "ls-remote", "--heads", "origin"); strings.Count(heads, "\n") != 1 || !strings.Contains(heads, "refs/heads/dev") || !strings.Contains(heads, "refs/heads/main") {
		t.Fatalf("remote branches:\n%s", heads)
	}
	agents, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(string(agents), "This project has a single branch") || strings.Contains(string(agents), "mem promote") {
		t.Fatalf("AGENTS.md has the staged instructions:\n%s", agents)
	}

	writeFile(t, root, ".mem/config.toml", strings.Replace(string(config), "[git]\n", "[git]\nbase = 'main'\n", 1))
	run(t, root, "commit", "--quiet", "-am", "Build dev on main")
	run(t, root, "push", "--quiet")
	out = mem(t, root, "onboard")
	if strings.Contains(out, "RELEASES") || strings.Contains(out, "release status") || !strings.Contains(out, "Base: dev is built on released code from main") || !strings.Contains(out, "from the open specs and open todos.") {
		t.Fatalf("onboard:\n%s", out)
	}

	commit(t, root, "feature.txt")
	heads := run(t, root, "ls-remote", "--heads", "origin")
	for _, args := range [][]string{{"promote", "staging"}, {"promote", "production"}, {"deploy"}} {
		if err := memErr(t, root, args...); err == nil || !strings.Contains(err.Error(), "single-branch ([git] single_branch = true)") {
			t.Fatalf("mem %s: %v", strings.Join(args, " "), err)
		}
	}
	if after := run(t, root, "ls-remote", "--heads", "origin"); after != heads {
		t.Fatalf("a refused release moved a branch:\n%s\n->\n%s", heads, after)
	}
	if tags := run(t, root, "ls-remote", "--tags", "origin"); tags != "" {
		t.Fatalf("a refused release pushed tags:\n%s", tags)
	}
}

func TestSingleBranchInitRefusesOtherBranchFlagsAndADetachedHead(t *testing.T) {
	root := t.TempDir()
	run(t, root, "init", "--quiet", "-b", "dev")
	commit(t, root, "README.md")
	if err := memErr(t, root, "init", "--single-branch", "--staging", "test", "--no-template"); err == nil || !strings.Contains(err.Error(), "--staging") {
		t.Fatalf("init --single-branch --staging: %v", err)
	}
	run(t, root, "switch", "--quiet", "--detach")
	if err := memErr(t, root, "init", "--single-branch", "--no-template"); err == nil || !strings.Contains(err.Error(), "HEAD is detached") {
		t.Fatalf("init --single-branch on a detached HEAD: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".mem")); !os.IsNotExist(err) {
		t.Fatalf("a refused init left .mem behind: %v", err)
	}
}
