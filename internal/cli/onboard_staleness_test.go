package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

// oldCommit commits everything staged with author and committer dates daysAgo in the past.
func oldCommit(t *testing.T, dir string, daysAgo int, message string) {
	t.Helper()
	date := time.Now().AddDate(0, 0, -daysAgo).Format(time.RFC3339)
	env := []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date}
	if _, err := git.RunEnv(context.Background(), dir, env, "commit", "--quiet", "-m", message); err != nil {
		t.Fatal(err)
	}
}

func TestOnboardFlagsSpecsAndBranchesThatStoppedMoving(t *testing.T) {
	mine, _ := sharedProject(t)

	// A spec started 20 days ago that has not changed since.
	mem(t, mine, "spec", "new", "Parser")
	mem(t, mine, "task", "new", "Tokenize", "Split input into tokens.", "--spec", "parser")
	spec := filepath.Join(mine, ".mem", "specs", "parser", "spec.md")
	data, _ := os.ReadFile(spec)
	os.WriteFile(spec, []byte(strings.Replace(string(data), "status: draft", "status: active\nassigned_to: test_user", 1)), 0o644)
	run(t, mine, "add", "--all")
	oldCommit(t, mine, 20, "Start spec parser")

	run(t, mine, "push", "--quiet")

	// A branch nobody has touched for 30 days.
	run(t, mine, "switch", "--quiet", "-c", "spike")
	commit(t, mine, "unused.txt")
	os.WriteFile(filepath.Join(mine, "unused.txt"), []byte("old\n"), 0o644)
	run(t, mine, "add", "unused.txt")
	oldCommit(t, mine, 30, "Old spike")
	run(t, mine, "push", "--quiet", "-u", "origin", "spike")
	run(t, mine, "switch", "--quiet", "dev")

	out := mem(t, mine, "onboard", "--offline")
	if data, err := os.ReadFile(filepath.Join(mine, ".mem", "local", "onboard.md")); err == nil {
		out += string(data)
	}
	for _, want := range []string{
		"⏳ CHECK THESE",
		"Active spec parser (test_user) has not changed for 20 days",
		"Branch origin/spike is not merged into dev and has had no commit for 30 days (last by Test User)",
		"commit(s) ahead of staging, the oldest from ",
		"Go through each item under ⏳ CHECK THESE with the user",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("onboard lacks %q:\n%s", want, out)
		}
	}
}
