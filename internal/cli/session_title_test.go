package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboardAndWorkLogsAskForASessionTitle(t *testing.T) {
	mine, _ := sharedProject(t)
	if out := mem(t, mine, "onboard"); !strings.Contains(out, `give it a short title for the work, a few words such as "Tugela charger startup issue"`) || !strings.Contains(out, "Name only the work, not the project") || !strings.Contains(out, "as soon as the user has said what to work on") {
		t.Fatalf("onboard instruction lacks the session title:\n%s", out)
	}
	if out := mem(t, mine, "log", "new"); !strings.Contains(out, `look up its current title`) || !strings.Contains(out, `rename it to a short title for the current work now`) {
		t.Fatalf("log new lacks the retitle step:\n%s", out)
	}
	for _, args := range [][]string{{"onboard"}, {"log", "new"}} {
		if out := mem(t, mine, args...); !strings.Contains(out, `session id "self"`) || !strings.Contains(out, "If the rename fails, correct the call and retry.") {
			t.Fatalf("%v does not say how to address this session or to retry a failed rename:\n%s", args, out)
		}
	}
}

func TestALongOnboardContextEndsWithTheSessionTitleReminder(t *testing.T) {
	mine, _ := sharedProject(t)
	writeFile(t, mine, ".mem/docs/glossary.md", strings.Repeat("A long project doc that pushes the context into a file.\n", 300))

	out := mem(t, mine, "onboard")
	context, err := os.ReadFile(filepath.Join(mine, ".mem/local/onboard.md"))
	if err != nil {
		t.Fatalf("the context was not written to a file: %v\n%s", err, out)
	}
	if !strings.HasSuffix(string(context), "🏷️ SESSION TITLE\n"+strings.Repeat("-", 70)+"\n\nRemember to name this session. "+nameSession()+"\n") {
		t.Fatalf("onboard.md does not end with the session title reminder:\n%s", context[max(0, len(context)-600):])
	}
	if strings.Contains(out, "Remember to name this session.") {
		t.Fatalf("the reminder belongs at the end of onboard.md, not in stdout:\n%s", out)
	}
}
