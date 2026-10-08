package cli

import (
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

func TestPostCommitHookNudgesAfterWorkCommitsAndStaysQuietForMem(t *testing.T) {
	mine, _ := sharedProject(t)
	for _, file := range []string{"a.go", "b.go", "c.go"} {
		commit(t, mine, file)
	}

	out := mem(t, mine, "hook", "post-commit")
	if !strings.Contains(out, "mem: 3 commits since your last work log or completed task; consider writing one now") || !strings.Contains(out, "mem: 3 commits on dev are not pushed") {
		t.Fatalf("post-commit output:\n%s", out)
	}

	t.Setenv(git.InternalEnv, "1")
	if out := mem(t, mine, "hook", "post-commit"); out != "" {
		t.Fatalf("post-commit spoke during one of mem's own Git commands:\n%s", out)
	}
}
