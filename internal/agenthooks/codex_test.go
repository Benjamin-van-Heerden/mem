package agenthooks

import (
	"os"
	"path/filepath"
	"testing"
)

// Codex skips a project hook whose definition changed until every user approves it again, so this pins the entries
// byte for byte: a change here must be deliberate.
func TestSyncCodexInstallsStableEntriesIntoAFreshProject(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCodex(root, CodexHooks{Compact: true, Prompt: true}); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	data, _ := os.ReadFile(filepath.Join(root, CodexHooksPath))
	want := `{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "compact",
        "hooks": [
          {
            "type": "command",
            "command": "command -v mem >/dev/null 2>&1 && mem hook compact || true",
            "commandWindows": "mem hook compact",
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
            "commandWindows": "mem hook prompt",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
`
	if string(data) != want {
		t.Fatalf("hooks file:\n%s", data)
	}
	if changed, err := SyncCodex(root, CodexHooks{Compact: true, Prompt: true}); err != nil || changed {
		t.Fatalf("a second sync changed the file: %v, %v", changed, err)
	}
}

func TestSyncCodexKeepsOtherHooksAndRemovesOnlyItsOwn(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, CodexHooksPath)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(`{"description": "team hooks", "hooks": {"SessionStart": [{"matcher": "startup", "hooks": [{"type": "command", "command": "./hooks/env.sh"}]}, {"matcher": "compact", "hooks": [{"type": "command", "command": "mem hook compact", "timeout": 5}]}]}}`), 0o644)

	if changed, err := SyncCodex(root, CodexHooks{}); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	data, _ := os.ReadFile(path)
	want := `{
  "description": "team hooks",
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/env.sh"
          }
        ]
      }
    ]
  }
}
`
	if string(data) != want {
		t.Fatalf("after turning mem's hooks off:\n%s", data)
	}
}

func TestSyncCodexOptOutDoesNotCreateTheFile(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCodex(root, CodexHooks{}); err != nil || changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, CodexHooksPath)); !os.IsNotExist(err) {
		t.Fatalf("the hooks file was created: %v", err)
	}
}
