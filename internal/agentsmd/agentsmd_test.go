package agentsmd

import (
	"strings"
	"testing"
)

func TestInstallAndRefreshPreserveUserContentAndMemories(t *testing.T) {
	text, err := Install("# My project\n\nUser notes.\n", "1.2.0")
	if err != nil {
		t.Fatal(err)
	}
	text, err = SetMemory(text, "logging", "Use the project logger.\n\n### Detail\nNever print.")
	if err != nil {
		t.Fatal(err)
	}
	stale := strings.Replace(text, "# Working With mem", "# Edited by hand", 1)
	refreshed, newer, err := Refresh(stale, "1.2.0")
	if err != nil || newer != "" {
		t.Fatalf("refresh newer=%q err=%v", newer, err)
	}
	if refreshed != text {
		t.Fatal("refresh did not restore the managed block exactly")
	}
	if !strings.HasPrefix(refreshed, "# My project\n\nUser notes.\n") {
		t.Fatal("user content outside the block was not preserved")
	}
	memories, err := Memories(refreshed)
	if err != nil || len(memories) != 1 || memories[0].Body != "Use the project logger.\n\n### Detail\nNever print." {
		t.Fatalf("memories after refresh = %#v, %v", memories, err)
	}
	if again, _, _ := Refresh(refreshed, "1.2.0"); again != refreshed {
		t.Fatal("refreshing current instructions changed the file")
	}
}

func TestRefreshLeavesBlockFromNewerVersion(t *testing.T) {
	text, _ := Install("", "1.10.0")
	updated, newer, err := Refresh(text, "1.9.3")
	if err != nil || newer != "1.10.0" || updated != text {
		t.Fatalf("newer=%q err=%v changed=%v", newer, err, updated != text)
	}
	if updated, newer, _ := Refresh(text, "dev"); newer != "" || !strings.Contains(updated, "Managed by mem dev.") {
		t.Fatal("a development build did not refresh the block")
	}
}

func TestSetMemoryReplacesByNameAndRemoveDeletesOnlyThatMemory(t *testing.T) {
	text, _ := Install("", "1.0.0")
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
