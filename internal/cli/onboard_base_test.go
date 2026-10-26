package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboardReportsReleasedCommitsALongLivedBranchHasNotMergedIn(t *testing.T) {
	mine, teammate := sharedProject(t)
	if out := mem(t, mine, "onboard"); strings.Contains(out, "Base:") {
		t.Fatalf("onboard reports a base the project does not set:\n%s", out)
	}

	config := filepath.Join(mine, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(strings.Replace(string(data), "[git]\n", "[git]\nbase = 'main'\n", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, mine, "commit", "--quiet", "-am", "Build dev on released main")
	run(t, mine, "push", "--quiet")
	out := mem(t, mine, "onboard")
	if !strings.Contains(out, "Base: dev is built on released code from main and has all of it.") || strings.Contains(out, "Mention the released commits") {
		t.Fatalf("onboard with an up-to-date base:\n%s", out)
	}

	run(t, teammate, "fetch", "--quiet")
	run(t, teammate, "switch", "--quiet", "-c", "hotfix", "origin/main")
	writeFile(t, teammate, "hotfix.txt", "fix\n")
	run(t, teammate, "add", ".")
	run(t, teammate, "commit", "--quiet", "-m", "Hotfix")
	run(t, teammate, "push", "--quiet", "origin", "HEAD:main")
	out = mem(t, mine, "onboard")
	for _, want := range []string{"1 commit(s) of origin/main are not merged in yet.", "`git fetch origin && git merge --no-ff origin/main`", "Never rebase dev, never force-push it", "Mention the released commits not merged in yet"} {
		if !strings.Contains(out, want) {
			t.Fatalf("onboard lacks %q:\n%s", want, out)
		}
	}
}
