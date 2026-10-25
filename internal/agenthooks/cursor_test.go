package agenthooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncCursorInstallsIntoAFreshProject(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCursor(root, true); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	data, _ := os.ReadFile(filepath.Join(root, CursorHooksPath))
	want := `{
  "version": 1,
  "hooks": {
    "stop": [
      {
        "command": "mem hook stop",
        "timeout": 30
      }
    ]
  }
}
`
	if string(data) != want {
		t.Fatalf("hooks file:\n%s", data)
	}
	if changed, err := SyncCursor(root, true); err != nil || changed {
		t.Fatalf("a second sync changed the file: %v, %v", changed, err)
	}
}

func TestSyncCursorKeepsOtherHooksAndRemovesOnlyItsOwn(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, CursorHooksPath)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(`{"version": 1, "hooks": {"afterFileEdit": [{"command": "./hooks/format.sh"}], "stop": [{"command": "./hooks/audit.sh", "loop_limit": null}, {"command": "mem hook stop", "timeout": 5}]}, "note": "team hooks"}`), 0o644)

	if changed, err := SyncCursor(root, true); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	data, _ := os.ReadFile(path)
	want := `{
  "version": 1,
  "hooks": {
    "afterFileEdit": [
      {
        "command": "./hooks/format.sh"
      }
    ],
    "stop": [
      {
        "command": "./hooks/audit.sh",
        "loop_limit": null
      },
      {
        "command": "mem hook stop",
        "timeout": 30
      }
    ]
  },
  "note": "team hooks"
}
`
	if string(data) != want {
		t.Fatalf("after updating mem's entry:\n%s", data)
	}

	if changed, err := SyncCursor(root, false); err != nil || !changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	data, _ = os.ReadFile(path)
	want = `{
  "version": 1,
  "hooks": {
    "afterFileEdit": [
      {
        "command": "./hooks/format.sh"
      }
    ],
    "stop": [
      {
        "command": "./hooks/audit.sh",
        "loop_limit": null
      }
    ]
  },
  "note": "team hooks"
}
`
	if string(data) != want {
		t.Fatalf("after turning the hook off:\n%s", data)
	}
}

func TestSyncCursorOptOutDoesNotCreateTheFile(t *testing.T) {
	root := t.TempDir()
	if changed, err := SyncCursor(root, false); err != nil || changed {
		t.Fatalf("changed = %v, %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, CursorHooksPath)); !os.IsNotExist(err) {
		t.Fatalf("the hooks file was created: %v", err)
	}
}
