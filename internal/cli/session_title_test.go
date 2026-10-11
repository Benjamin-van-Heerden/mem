package cli

import (
	"strings"
	"testing"
)

func TestOnboardAndWorkLogsAskForASessionTitle(t *testing.T) {
	mine, _ := sharedProject(t)
	if out := mem(t, mine, "onboard"); !strings.Contains(out, `give it a short title for the work, a few words such as "Tugela charger startup issue", as soon as the user has said what to work on`) {
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
