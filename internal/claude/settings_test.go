package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func settings(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, SettingsPath))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestSyncCompactHookInstallsIntoAFreshProject(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCompactHook(root, true); err != nil || !changed {
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
    ]
  }
}
`
	if got != want {
		t.Fatalf("settings:\n%s", got)
	}
	if changed, err := SyncCompactHook(root, true); err != nil || changed {
		t.Fatalf("a second sync changed the file: %v, %v", changed, err)
	}
}

func TestSyncCompactHookKeepsOtherSettingsInOrderAndRemovesOnlyItsOwn(t *testing.T) {
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
	os.WriteFile(filepath.Join(root, SettingsPath), []byte(existing), 0o644)

	if _, err := SyncCompactHook(root, true); err != nil {
		t.Fatal(err)
	}
	got := settings(t, root)
	order := []string{`"permissions"`, `"hooks"`, `"SessionStart"`, `"echo hi"`, `mem hook compact`, `"PreToolUse"`, `"model"`}
	last := -1
	for _, s := range order {
		i := strings.Index(got, s)
		if i <= last {
			t.Fatalf("%s is missing or out of order:\n%s", s, got)
		}
		last = i
	}

	if changed, err := SyncCompactHook(root, false); err != nil || !changed {
		t.Fatalf("opting out changed = %v, %v", changed, err)
	}
	got = settings(t, root)
	if strings.Contains(got, "mem hook compact") || !strings.Contains(got, `"echo hi"`) || !strings.Contains(got, `"./check.sh"`) {
		t.Fatalf("after opting out:\n%s", got)
	}
}

func TestSyncCompactHookOptOutDoesNotCreateSettings(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCompactHook(root, false); err != nil || changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, SettingsPath)); !os.IsNotExist(err) {
		t.Fatal("opting out created .claude/settings.json")
	}
}
