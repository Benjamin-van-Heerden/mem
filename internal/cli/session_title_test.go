package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/claude"
)

// promptHook runs `mem hook prompt` as Claude Code's UserPromptSubmit hook does, with the hook input on stdin.
func promptHook(t *testing.T, root, session string) string {
	t.Helper()
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(`{"session_id": "` + session + `", "prompt": "Add retries to the importer"}`))
	cmd.SetArgs([]string{"hook", "prompt", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mem hook prompt: %v\n%s", err, out.String())
	}
	return out.String()
}

func TestTheFirstPromptAfterOnboardAsksForAProjectSessionTitleOnce(t *testing.T) {
	mine, _ := sharedProject(t)
	const title = `title it "mine · <the work>"`

	if out := promptHook(t, mine, "s1"); out != "" {
		t.Fatalf("hook spoke before any onboard:\n%s", out)
	}

	t.Setenv(claude.SessionEnv, "s1")
	if out := mem(t, mine, "onboard"); !strings.Contains(out, title) {
		t.Fatalf("onboard instruction lacks the title format:\n%s", out)
	}
	if out := promptHook(t, mine, "s2"); out != "" {
		t.Fatalf("hook spoke for a session that did not onboard:\n%s", out)
	}
	if out := promptHook(t, mine, "s1"); !strings.Contains(out, "first request since onboarding") || !strings.Contains(out, title) {
		t.Fatalf("first prompt after onboard:\n%s", out)
	}
	if out := promptHook(t, mine, "s1"); out != "" {
		t.Fatalf("hook spoke again on a later prompt:\n%s", out)
	}

	if out := mem(t, mine, "log", "new"); !strings.Contains(out, title+" if its title no longer describes the work") {
		t.Fatalf("log new lacks the retitle step:\n%s", out)
	}
}

func TestPromptHookStaysSilentOnInputItCannotRead(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("not json"))
	cmd.SetArgs([]string{"hook", "prompt", "--dir", t.TempDir()})
	if err := cmd.Execute(); err != nil || out.Len() != 0 {
		t.Fatalf("err = %v, output:\n%s", err, out.String())
	}
}
