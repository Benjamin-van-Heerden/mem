package cli

import (
	"bytes"
	"encoding/json"
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

// memWithInput runs mem with stdin, as an agent's hook does.
func memWithInput(t *testing.T, root, stdin string, args ...string) string {
	t.Helper()
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(append(args, "--dir", root))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mem %s: %v\n%s", strings.Join(args, " "), err, out.String())
	}
	return out.String()
}

func TestStopHookAsksForADueWorkLogOncePerStateOfWork(t *testing.T) {
	mine, _ := sharedProject(t)
	const completed = `{"status": "completed", "loop_count": 0}`
	followUp := func(stdin string) string {
		t.Helper()
		out := memWithInput(t, mine, stdin, "hook", "stop")
		var reply map[string]string
		if err := json.Unmarshal([]byte(out), &reply); err != nil {
			t.Fatalf("stop hook output is not a JSON object: %q", out)
		}
		return reply["followup_message"]
	}

	commit(t, mine, "a.go")
	if msg := followUp(completed); msg != "" {
		t.Fatalf("follow-up below the due level: %q", msg)
	}
	writeFile(t, mine, "parser.go", strings.Repeat("x\n", 1200))
	for _, stdin := range []string{`{"status": "aborted", "loop_count": 0}`, `{"status": "error", "loop_count": 0}`, `{"status": "completed", "loop_count": 1}`, "not json"} {
		if msg := followUp(stdin); msg != "" {
			t.Fatalf("follow-up for %s: %q", stdin, msg)
		}
	}
	msg := followUp(completed)
	if !strings.HasPrefix(msg, "mem: This is an automatic follow-up from mem, not a message from the user. 1 commit, 2 files and 1201 lines of work") || !strings.Contains(msg, "`mem log new`, then `mem log commit`") {
		t.Fatalf("follow-up with a log due: %q", msg)
	}
	if msg := followUp(completed); msg != "" {
		t.Fatalf("asked again with no new work: %q", msg)
	}
	writeFile(t, mine, "lexer.go", "x\n")
	if msg := followUp(completed); !strings.Contains(msg, "3 files and 1202 lines") {
		t.Fatalf("not asked again after more work: %q", msg)
	}
	if msg := memWithInput(t, t.TempDir(), completed, "hook", "stop"); msg != "{}\n" {
		t.Fatalf("stop hook outside a mem project: %q", msg)
	}
}
