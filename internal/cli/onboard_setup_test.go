package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboardAndCompactionLeadWithAPendingSetupUntilItIsDeleted(t *testing.T) {
	mine, _ := sharedProject(t)
	setup := filepath.Join(mine, ".mem", "setup.md")
	write := func(text string) {
		t.Helper()
		if err := os.WriteFile(setup, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("# Setup: web\n\n## [x] 1. Scaffold\n\nDone when: it builds.\n\n## [ ] 2. Link Vercel (you)\n")
	out := mem(t, mine, "onboard", "--offline")
	for _, want := range []string{"🏗️ SETUP", "From .mem/setup.md, 1 of 2 steps done.", "## [ ] 2. Link Vercel (you)", "1. Setup is pending (1 of 2 steps done)."} {
		if !strings.Contains(out, want) {
			t.Fatalf("onboard lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Ask the user how they would like to proceed") {
		t.Fatalf("onboard asks what to do next while setup is pending:\n%s", out)
	}
	if digest := mem(t, mine, "hook", "compact"); !strings.Contains(digest, "Setup pending: 1 of 2 steps done in .mem/setup.md.") || !strings.Contains(digest, "Continue the setup in .mem/setup.md.") {
		t.Fatalf("digest lacks the setup:\n%s", digest)
	}

	write("# Setup: web\n\n## [x] 1. Scaffold\n\n## [x] 2. Link Vercel (you)\n")
	if out := mem(t, mine, "onboard", "--offline"); !strings.Contains(out, "Every step in .mem/setup.md is ticked: delete the file") || !strings.Contains(out, "Ask the user how they would like to proceed") {
		t.Fatalf("onboard does not close a finished setup:\n%s", out)
	}

	if err := os.Remove(setup); err != nil {
		t.Fatal(err)
	}
	if out := mem(t, mine, "onboard", "--offline"); strings.Contains(out, "SETUP") || strings.Contains(out, "setup.md") {
		t.Fatalf("onboard mentions a deleted setup:\n%s", out)
	}
	if digest := mem(t, mine, "hook", "compact"); strings.Contains(digest, "setup") {
		t.Fatalf("digest mentions a deleted setup:\n%s", digest)
	}
}
