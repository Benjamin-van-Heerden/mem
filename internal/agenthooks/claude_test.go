package agenthooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func settings(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ClaudeSettingsPath))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestSyncClaudeInstallsIntoAFreshProject(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncClaude(root, ClaudeHooks{Compact: true, Prompt: true}); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	got := settings(t, root)
	want := `{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "compact",
        "hooks": [
          {
            "type": "command",
            "command": "command -v mem >/dev/null 2>&1 && mem hook compact || true",
            "timeout": 60
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "command -v mem >/dev/null 2>&1 && mem hook prompt || true",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
`
	if got != want {
		t.Fatalf("settings:\n%s", got)
	}
	if changed, err := SyncClaude(root, ClaudeHooks{Compact: true, Prompt: true}); err != nil || changed {
		t.Fatalf("a second sync changed the file: %v, %v", changed, err)
	}
}

func TestSyncClaudeKeepsOtherSettingsInOrderAndRemovesOnlyItsOwn(t *testing.T) {
	root := t.TempDir()
	existing := `{
  "permissions": {"allow": ["Bash(go test:*)"]},
  "hooks": {
    "SessionStart": [{"matcher": "startup", "hooks": [{"type": "command", "command": "echo hi"}]}],
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "./check.sh"}]}]
  },
  "model": "opus"
}`
	os.MkdirAll(filepath.Join(root, ".claude"), 0o755)
	os.WriteFile(filepath.Join(root, ClaudeSettingsPath), []byte(existing), 0o644)

	if _, err := SyncClaude(root, ClaudeHooks{Compact: true, Prompt: true}); err != nil {
		t.Fatal(err)
	}
	got := settings(t, root)
	order := []string{`"permissions"`, `"hooks"`, `"SessionStart"`, `"echo hi"`, `mem hook compact`, `"PreToolUse"`, `"UserPromptSubmit"`, `mem hook prompt`, `"model"`}
	last := -1
	for _, s := range order {
		i := strings.Index(got, s)
		if i <= last {
			t.Fatalf("%s is missing or out of order:\n%s", s, got)
		}
		last = i
	}

	if changed, err := SyncClaude(root, ClaudeHooks{Prompt: true}); err != nil || !changed {
		t.Fatalf("opting out of the compaction hook changed = %v, %v", changed, err)
	}
	if got = settings(t, root); strings.Contains(got, "mem hook compact") || !strings.Contains(got, "mem hook prompt") {
		t.Fatalf("after opting out of the compaction hook:\n%s", got)
	}
	if changed, err := SyncClaude(root, ClaudeHooks{}); err != nil || !changed {
		t.Fatalf("opting out changed = %v, %v", changed, err)
	}
	got = settings(t, root)
	if strings.Contains(got, "mem hook compact") || strings.Contains(got, "UserPromptSubmit") || !strings.Contains(got, `"echo hi"`) || !strings.Contains(got, `"./check.sh"`) {
		t.Fatalf("after opting out:\n%s", got)
	}
}

func TestSyncClaudeOptOutDoesNotCreateSettings(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncClaude(root, ClaudeHooks{}); err != nil || changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, ClaudeSettingsPath)); !os.IsNotExist(err) {
		t.Fatal("opting out created .claude/settings.json")
	}
}
