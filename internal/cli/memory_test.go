package cli

import (
	"strings"
	"testing"
)

func TestAnAgentsMdOverCodexsLimitIsFlaggedAtMemorySetAndOnboard(t *testing.T) {
	mine, teammate := sharedProject(t)
	out := mem(t, teammate, "memory", "set", "logging", "Use the structured logger.")
	if strings.Contains(out, "⚠️ AGENTS.md is") || strings.Contains(out, "Stop here") {
		t.Fatalf("a small AGENTS.md was flagged:\n%s", out)
	}

	out = mem(t, teammate, "memory", "set", "architecture", strings.Repeat("Keep handlers thin and put domain rules in the service layer. ", 600))
	if !strings.Contains(out, "⚠️ AGENTS.md is") || !strings.Contains(out, "Stop here: tell the user about the ⚠️ above") {
		t.Fatalf("memory set did not flag AGENTS.md over 32 KiB:\n%s", out)
	}
	out = mem(t, teammate, "memory", "set", "logging", "Use the structured logger everywhere.")
	if !strings.Contains(out, "⚠️ AGENTS.md is") {
		t.Fatalf("memory set stopped flagging AGENTS.md while it is still over 32 KiB:\n%s", out)
	}
	run(t, teammate, "commit", "--quiet", "--all", "-m", "Add memories")
	run(t, teammate, "push", "--quiet")

	out = mem(t, mine, "onboard")
	if !strings.Contains(out, "🔄 UPDATES") || !strings.Contains(out, "⚠️ AGENTS.md is") || !strings.Contains(out, "Tell the user about each ⚠️ item under 🔄 UPDATES") {
		t.Fatalf("onboard did not flag the AGENTS.md that arrived over 32 KiB:\n%s", out)
	}

	mem(t, mine, "memory", "remove", "architecture")
	if out = mem(t, mine, "onboard"); strings.Contains(out, "⚠️ AGENTS.md is") {
		t.Fatalf("onboard still flags AGENTS.md after it was trimmed:\n%s", out)
	}
}
