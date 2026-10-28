package agentsmd

import (
	"strings"
	"testing"
)

func TestInstallAndRefreshPreserveUserContentAndMemories(t *testing.T) {
	text, err := Install("# My project\n\nUser notes.\n", "1.2.0", false)
	if err != nil {
		t.Fatal(err)
	}
	text, err = SetMemory(text, "logging", "Use the project logger.\n\n### Detail\nNever print.")
	if err != nil {
		t.Fatal(err)
	}
	stale := strings.Replace(text, "# Working With mem", "# Edited by hand", 1)
	refreshed, newer, err := Refresh(stale, "1.2.0", false)
	if err != nil || newer != "" {
		t.Fatalf("refresh newer=%q err=%v", newer, err)
	}
	if refreshed != text {
		t.Fatal("refresh did not restore the managed block exactly")
	}
	if !strings.HasPrefix(refreshed, "<mem>\n") || !strings.HasSuffix(refreshed, "\n\n# My project\n\nUser notes.\n") {
		t.Fatal("the managed block does not lead, or user content after it was not preserved")
	}
	memories, err := Memories(refreshed)
	if err != nil || len(memories) != 1 || memories[0].Body != "Use the project logger.\n\n### Detail\nNever print." {
		t.Fatalf("memories after refresh = %#v, %v", memories, err)
	}
	if again, _, _ := Refresh(refreshed, "1.2.0", false); again != refreshed {
		t.Fatal("refreshing current instructions changed the file")
	}
}

func TestRefreshLeavesBlockFromNewerVersion(t *testing.T) {
	text, _ := Install("", "1.10.0", false)
	updated, newer, err := Refresh(text, "1.9.3", false)
	if err != nil || newer != "1.10.0" || updated != text {
		t.Fatalf("newer=%q err=%v changed=%v", newer, err, updated != text)
	}
	if updated, newer, _ := Refresh(text, "dev", false); newer != "" || !strings.Contains(updated, "Managed by mem dev.") {
		t.Fatal("a development build did not refresh the block")
	}
}

func TestReplaceLegacyDropsEveryHarnessBlock(t *testing.T) {
	text := "<AGENT_CORE>\nnewer harness\n</AGENT_CORE>\n\n<core_instructions>\nolder harness\n</core_instructions>\n\n## UV\n\nKeep this.\n"
	updated, err := ReplaceLegacy(text, "1.2.0", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(updated, "harness") || !strings.HasPrefix(updated, "<mem>\n") || !strings.Contains(updated, "</mem>\n\n## UV\n\nKeep this.\n") {
		t.Fatalf("legacy blocks left behind or project content lost:\n%s", updated)
	}
}

func TestSetMemoryReplacesByNameAndRemoveDeletesOnlyThatMemory(t *testing.T) {
	text, _ := Install("", "1.0.0", false)
	text, _ = SetMemory(text, "a", "first")
	text, _ = SetMemory(text, "b", "second")
	text, _ = SetMemory(text, "a", "replaced")
	text, err := RemoveMemory(text, "b")
	if err != nil {
		t.Fatal(err)
	}
	memories, _ := Memories(text)
	if len(memories) != 1 || memories[0].Name != "a" || memories[0].Body != "replaced" {
		t.Fatalf("memories = %#v", memories)
	}
	if _, err := RemoveMemory(text, "missing"); err == nil {
		t.Fatal("removing a missing memory succeeded")
	}
	if _, err := SetMemory(text, "c", "## heading would split the memory"); err == nil {
		t.Fatal("a level-2 heading inside a memory was accepted")
	}
}

func TestTagMentionsInUserContentDoNotBreakTheBlocks(t *testing.T) {
	text, err := Install("# Notes\n\nThe `<mem>` block and the <memories> section are managed by mem.\n", "1.0.0", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := Refresh(text, "1.0.0", false); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := SetMemory(text, "a", "first"); err != nil {
		t.Fatalf("set memory: %v", err)
	}
}

func TestSingleBranchInstructionsLeaveOutReleasesAndRenumberTheRules(t *testing.T) {
	staged, single := render(false), render(true)
	for _, text := range []string{staged, single} {
		if strings.Contains(text, "{{") {
			t.Fatalf("template syntax left in the instructions:\n%s", text)
		}
	}
	for _, want := range []string{"3. Never commit to or push the staging or production branch", "4. Release only when the user asks", "6. Use the package manager", "`mem promote production`", "before a release,"} {
		if !strings.Contains(staged, want) {
			t.Errorf("staged instructions lack %q", want)
		}
	}
	for _, want := range []string{"3. Never commit to or push the base branch", "4. Never edit generated files", "5. Use the package manager", "This project has a single branch: mem makes no releases here"} {
		if !strings.Contains(single, want) {
			t.Errorf("single-branch instructions lack %q", want)
		}
	}
	for _, unwanted := range []string{"6. ", "mem promote", "mem deploy", "staging", "release status", "before a release"} {
		if strings.Contains(single, unwanted) {
			t.Errorf("single-branch instructions contain %q", unwanted)
		}
	}
}

func TestRefreshSwitchesTheInstructionsWhenTheModeChanges(t *testing.T) {
	text, _ := Install("", "1.0.0", false)
	single, _, err := Refresh(text, "1.0.0", true)
	if err != nil || !strings.Contains(single, "This project has a single branch") || strings.Contains(single, "mem promote") {
		t.Fatalf("refreshed for single branch (%v):\n%s", err, single)
	}
	if back, _, _ := Refresh(single, "1.0.0", false); back != text {
		t.Fatal("refreshing back to staged did not restore the staged instructions")
	}
}
