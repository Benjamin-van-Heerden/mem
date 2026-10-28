package project

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

func TestSlugifySeparatesWordsJoinedByPunctuation(t *testing.T) {
	cases := map[string]string{
		"init --template and mem template use/list": "init_template_and_mem_template_use_list",
		"Benjamin van Heerden":                      "benjamin_van_heerden",
		"Upgrade to v1.2: faster sync":              "upgrade_to_v1_2_faster_sync",
		"Fix `mem` crash (again)!":                  "fix_mem_crash_again",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadRejectsASingleBranchConfigThatKeepsReleaseSettings(t *testing.T) {
	for name, config := range map[string]string{
		"staging":       "[git]\ndevelopment = 'dev'\nsingle_branch = true\nstaging = 'test'\n",
		"production_pr": "[git]\ndevelopment = 'dev'\nsingle_branch = true\n\n[release]\nproduction_pr = true\n",
	} {
		root := t.TempDir()
		if _, err := git.Run(context.Background(), root, "init", "--quiet"); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, ".mem"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ConfigPath(root), []byte(config), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(context.Background(), root); err == nil || !strings.Contains(err.Error(), "single_branch = true") {
			t.Errorf("%s: Load error = %v", name, err)
		}
	}
}

func TestGuardedBranchesFollowTheMode(t *testing.T) {
	staged := GitConfig{Development: "dev", Staging: "test", Production: "main", Base: "upstream"}
	single := GitConfig{Development: "dev", SingleBranch: true, Base: "main"}
	for _, c := range []struct {
		g      GitConfig
		branch string
		want   bool
	}{
		{staged, "test", true}, {staged, "main", true}, {staged, "upstream", false}, {staged, "dev", false},
		{single, "main", true}, {single, "dev", false},
		// A detached HEAD has no branch name; it must not match the empty staging and production names.
		{single, "", false}, {GitConfig{Development: "dev", SingleBranch: true}, "", false},
	} {
		if got := c.g.Guarded(c.branch); got != c.want {
			t.Errorf("Guarded(%q) with %+v = %v, want %v", c.branch, c.g, got, c.want)
		}
	}
}
