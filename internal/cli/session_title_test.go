package cli

import (
	"strings"
	"testing"
)

func TestOnboardAndWorkLogsAskForAProjectSessionTitle(t *testing.T) {
	mine, _ := sharedProject(t)
	const title = `title it "mine · <the work>"`

	if out := mem(t, mine, "onboard"); !strings.Contains(out, title+" as soon as the user has said what to work on") {
		t.Fatalf("onboard instruction lacks the session title:\n%s", out)
	}
	if out := mem(t, mine, "log", "new"); !strings.Contains(out, title+" if its title no longer describes the work") {
		t.Fatalf("log new lacks the retitle step:\n%s", out)
	}
}
