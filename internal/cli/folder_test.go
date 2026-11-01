package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// folderProject is a folder project in a directory that is no Git repository, with a git on PATH that fails the test
// if anything runs it.
func folderProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, ".mem/config.toml", "schema = 2\nname = 'work'\nmode = 'folder'\nuser = 'Test Person'\n")
	blockGit(t)
	return root
}

func blockGit(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	called := filepath.Join(bin, "called")
	script := "#!/bin/sh\ntouch '" + called + "'\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Cleanup(func() {
		if _, err := os.Stat(called); err == nil {
			t.Error("a folder project ran git")
		}
	})
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		files[path] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestFolderProjectRefusesCommandsThatNeedGit(t *testing.T) {
	root := folderProject(t)
	before := snapshot(t, root)
	for _, args := range [][]string{
		{"sync"},
		{"promote", "staging"},
		{"deploy"},
		{"import", "agent-core"},
		{"template", "list"},
		{"spec", "list"},
		{"task", "list"},
		{"log", "commit"},
		{"hook", "pre-push", "origin", "url"},
		{"hook", "pre-commit"},
		{"hook", "post-commit"},
	} {
		if err := memErr(t, root, args...); !errors.Is(err, project.ErrFolder) {
			t.Errorf("mem %v: error = %v, want the folder refusal", args, err)
		}
	}
	after := snapshot(t, root)
	if len(after) != len(before) {
		t.Fatalf("a refused command changed the files: %v", after)
	}
	for path, content := range before {
		if after[path] != content {
			t.Errorf("a refused command changed %s", path)
		}
	}
}

func TestFolderProjectAttributesWorkToItsConfiguredUser(t *testing.T) {
	root := folderProject(t)
	mem(t, root, "todo", "new", "Chase the invoice", "From last month")
	if out := mem(t, root, "todo", "list"); !strings.Contains(out, "Chase the invoice") {
		t.Fatalf("todo list:\n%s", out)
	}
	p, err := project.Load(t.Context(), filepath.Join(root, ".mem"))
	if err != nil {
		t.Fatal(err)
	}
	if user, err := project.User(t.Context(), p); err != nil || user != "test_person" {
		t.Fatalf("User = %q, %v", user, err)
	}
}

func TestFolderWorkLogNudgeFollowsChangedFiles(t *testing.T) {
	root := folderProject(t)
	past := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(filepath.Join(root, ".mem", "config.toml"), past, past); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "_generated/invoice.pdf", "built")
	writeFile(t, root, ".claude/settings.json", "{}")
	writeFile(t, root, "AGENTS.md", "instructions")
	touched := past.Add(time.Hour)
	for i := range 20 {
		name := "clients/acme/doc" + strconv.Itoa(i) + ".md"
		if i%2 == 0 {
			name = "clients/acme/doc" + strconv.Itoa(i) + ".pdf"
		}
		writeFile(t, root, name, "content")
		if err := os.Chtimes(filepath.Join(root, name), touched, touched); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, root, ".mem/config.toml", "schema = 2\nname = 'work'\nmode = 'folder'\nuser = 'Test Person'\n[structure]\nignore = ['_generated/**']\n")
	if err := os.Chtimes(filepath.Join(root, ".mem", "config.toml"), past, past); err != nil {
		t.Fatal(err)
	}

	out := mem(t, root, "hook", "prompt")
	if !strings.Contains(out, "20 files changed since your last work log. A work log is due") || !strings.Contains(out, "(`mem log new`)") {
		t.Fatalf("hook prompt:\n%s", out)
	}
	if strings.Contains(out, "mem log commit") {
		t.Fatalf("a folder nudge should not ask for a commit:\n%s", out)
	}

	created := mem(t, root, "log", "new")
	logs, err := filepath.Glob(filepath.Join(root, ".mem", "logs", "test_person_*.md"))
	if err != nil || len(logs) != 1 {
		t.Fatalf("logs = %v, %v\n%s", logs, err, created)
	}
	log, err := os.ReadFile(logs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "- `clients/acme/doc0.pdf`\n") || strings.Contains(string(log), "lines)") || strings.Contains(string(log), "_generated") {
		t.Fatalf("log files:\n%s", log)
	}
	if strings.Contains(created, "mem log commit") {
		t.Fatalf("log new should not ask for a commit:\n%s", created)
	}
	if out := mem(t, root, "hook", "prompt"); out != "" {
		t.Fatalf("a new log should clear the nudge:\n%s", out)
	}
}

func TestFolderCompactionDigestStaysOffGit(t *testing.T) {
	root := folderProject(t)
	out := mem(t, root, "hook", "compact")
	if !strings.Contains(out, "MEM AFTER COMPACTION") || strings.Contains(out, "Branch") || strings.Contains(out, "sync") {
		t.Fatalf("hook compact:\n%s", out)
	}
}
