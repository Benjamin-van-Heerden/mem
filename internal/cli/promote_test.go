package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionReleaseWithoutNotesTagsAGeneratedSummary(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "promote", "staging")

	out := mem(t, mine, "promote", "production")
	if !strings.Contains(out, "tagged with a summary of its specs and commits") {
		t.Fatalf("release output:\n%s", out)
	}
	if main, test := run(t, mine, "rev-parse", "origin/main"), run(t, mine, "rev-parse", "origin/test"); main != test {
		t.Fatalf("main %s, test %s", main, test)
	}
	tag := run(t, mine, "describe", "--tags", "--exact-match", "origin/main")
	if notes := run(t, mine, "for-each-ref", "--format=%(contents)", "refs/tags/"+tag); !strings.HasPrefix(notes, "Release "+tag) || !strings.Contains(notes, "Commits:") {
		t.Fatalf("tag message = %q", notes)
	}
	if _, err := os.Stat(filepath.Join(mine, ".mem", "local", "release-notes.md")); !os.IsNotExist(err) {
		t.Fatalf("a draft was written: %v", err)
	}
}

func TestProductionReleaseDraftsNotesAndReleasesThemOnConfirm(t *testing.T) {
	mine, _ := sharedProject(t)
	releaseConfig(t, mine, "notes = true")
	draft := filepath.Join(mine, ".mem", "local", "release-notes.md")
	mem(t, mine, "promote", "staging")

	if err := memErr(t, mine, "promote", "production", "--confirm"); err == nil || !strings.Contains(err.Error(), "no release notes to confirm") {
		t.Fatalf("confirm without a draft: %v", err)
	}
	mainBefore := run(t, mine, "ls-remote", "origin", "refs/heads/main")
	out := mem(t, mine, "promote", "production")
	if !strings.Contains(out, "Nothing has been pushed yet.") || !strings.Contains(out, "(drafted)") || !strings.Contains(out, "mem promote production --confirm") {
		t.Fatalf("first run output:\n%s", out)
	}
	if after := run(t, mine, "ls-remote", "origin", "refs/heads/main"); after != mainBefore {
		t.Fatalf("drafting moved main: %q -> %q", mainBefore, after)
	}
	data, err := os.ReadFile(draft)
	if err != nil || !strings.HasPrefix(string(data), "<!-- mem:release ") {
		t.Fatalf("draft = %q, %v", data, err)
	}
	edited := strings.SplitN(string(data), "\n", 2)[0] + "\n# Release notes\n\nThe first release, reviewed by the user.\n"
	if err := os.WriteFile(draft, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := mem(t, mine, "promote", "production"); !strings.Contains(out, "(kept: it was written for this release)") {
		t.Fatalf("re-run did not keep the edited draft:\n%s", out)
	}

	mem(t, mine, "promote", "production", "--confirm")
	tag := run(t, mine, "describe", "--tags", "--abbrev=0", "origin/main")
	if notes := run(t, mine, "for-each-ref", "--format=%(contents)", "refs/tags/"+tag); !strings.Contains(notes, "# Release notes\n\nThe first release, reviewed by the user.") || strings.Contains(notes, "<!--") {
		t.Fatalf("tag notes = %q", notes)
	}
	if _, err := os.Stat(draft); !os.IsNotExist(err) {
		t.Fatalf("draft not removed after the release: %v", err)
	}

	commit(t, mine, "feature.txt")
	run(t, mine, "push", "--quiet")
	mem(t, mine, "promote", "staging")
	mem(t, mine, "promote", "production")
	commit(t, mine, "later.txt")
	run(t, mine, "push", "--quiet")
	mem(t, mine, "promote", "staging")
	if err := memErr(t, mine, "promote", "production", "--confirm"); err == nil || !strings.Contains(err.Error(), "drafted for") {
		t.Fatalf("confirm with a stale draft: %v", err)
	}
}

// releaseConfig adds a [release] section to the project's config and publishes it.
func releaseConfig(t *testing.T, root, settings string) {
	t.Helper()
	config := filepath.Join(root, ".mem", "config.toml")
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, append(data, []byte("\n[release]\n"+settings+"\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, root, "commit", "--quiet", "-am", "Release settings")
	run(t, root, "push", "--quiet")
}

// memErr runs mem and returns its error, for commands expected to fail.
func memErr(t *testing.T, root string, args ...string) error {
	t.Helper()
	cmd := New()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(append(args, "--dir", root))
	return cmd.Execute()
}
