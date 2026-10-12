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
	if !strings.Contains(out, "mem: 3 commits, 3 files and 3 lines of work since your last work log or completed task.") || !strings.Contains(out, "mem: 3 commits on dev are not pushed") {
		t.Fatalf("post-commit output:\n%s", out)
	}

	t.Setenv(git.InternalEnv, "1")
	if out := mem(t, mine, "hook", "post-commit"); out != "" {
		t.Fatalf("post-commit spoke during one of mem's own Git commands:\n%s", out)
	}
}

func TestPromptHookStaysQuietUntilAWorkLogIsCalledFor(t *testing.T) {
	mine, _ := sharedProject(t)
	commit(t, mine, "a.go")
	if out := mem(t, mine, "hook", "prompt"); out != "" {
		t.Fatalf("prompt hook spoke below the thresholds:\n%s", out)
	}
	writeFile(t, mine, "parser.go", strings.Repeat("x\n", 1200))
	out := mem(t, mine, "hook", "prompt")
	if !strings.HasPrefix(out, "mem: 1 commit, 2 files and 1201 lines of work since your last work log or completed task. A work log is due") || !strings.Contains(out, "if you are in the middle of a change, finish it first") {
		t.Fatalf("prompt hook output:\n%s", out)
	}
	if out := mem(t, t.TempDir(), "hook", "prompt"); out != "" {
		t.Fatalf("prompt hook spoke outside a mem project:\n%s", out)
	}
}
