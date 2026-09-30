package release

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDraftNotesDescribeTheReleaseFromSpecsLogsAndCommits(t *testing.T) {
	ctx := context.Background()
	p := setup(t)
	pl, _ := Prepare(ctx, p, "staging", "")
	if err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	pl, _ = Prepare(ctx, p, "production", "")
	if err := Execute(ctx, p, pl, "First"); err != nil {
		t.Fatal(err)
	}

	write(t, p.Root, ".mem/specs/archive/login/spec.md", "---\ntitle: Password login\nstatus: completed\n---\n\n## Overview\n\nUsers sign in with email and password. Sessions last a week.\n\n## Goals\n\n- x\n")
	write(t, p.Root, ".mem/specs/archive/dropped/spec.md", "---\ntitle: Dropped idea\nstatus: abandoned\n---\n\n## Overview\n\nNo.\n")
	write(t, p.Root, ".mem/logs/tester_20261001_120000.md", "---\nuser: tester\n---\n\n# Work Log - Login and sessions\n\n## Overarching Goals\n\nx\n\n## What Was Accomplished\n\n### Login form\n\ntext\n\n### Session cookies\n\ntext\n\n## Decisions\n\n### Not a section\n")
	run(t, p.Root, "add", ".")
	run(t, p.Root, "commit", "--quiet", "-m", "Add password login")
	run(t, p.Root, "push", "--quiet")
	pl, _ = Prepare(ctx, p, "staging", "")
	if err := Execute(ctx, p, pl, ""); err != nil {
		t.Fatal(err)
	}
	pl, _ = Prepare(ctx, p, "production", "")

	draft := DraftNotes(ctx, p, pl)
	for _, want := range []string{
		"# Release " + pl.Tag,
		"- **Password login**: Users sign in with email and password.",
		"- Login and sessions\n  - Login form\n  - Session cookies\n",
		"- Add password login (Test)",
	} {
		if !strings.Contains(draft, want) {
			t.Fatalf("draft lacks %q:\n%s", want, draft)
		}
	}
	if strings.Contains(draft, "Dropped idea") || strings.Contains(draft, "Not a section") || strings.Contains(draft, "- one (Test)") {
		t.Fatalf("draft includes what is outside the release:\n%s", draft)
	}
	commit, notes := DraftCommit(draft)
	if commit != pl.To || strings.HasPrefix(notes, "<!--") || !strings.HasPrefix(notes, "# Release ") {
		t.Fatalf("DraftCommit = %q, %q", commit, notes[:20])
	}
	if msg := GeneratedMessage(ctx, p, pl); !strings.Contains(msg, "Specs completed: Password login") || !strings.Contains(msg, "- Add password login") {
		t.Fatalf("generated message:\n%s", msg)
	}
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
