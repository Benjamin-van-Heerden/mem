package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sharedProject returns a mem project on dev and a teammate's clone of it, sharing a bare remote.
func sharedProject(t *testing.T) (mine, teammate string) {
	t.Helper()
	base := t.TempDir()
	t.Setenv("HOME", filepath.Join(base, "home"))
	remote := filepath.Join(base, "remote.git")
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	mine = filepath.Join(base, "mine")
	run(t, base, "clone", "--quiet", remote, mine)
	run(t, mine, "config", "user.name", "Test User")
	run(t, mine, "config", "user.email", "test@example.com")
	commit(t, mine, "README.md")
	run(t, mine, "push", "--quiet", "origin", "HEAD:main")
	mem(t, mine, "init", "--protect=false")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "mem setup")
	run(t, mine, "push", "--quiet")
	teammate = filepath.Join(base, "teammate")
	run(t, base, "clone", "--quiet", "-b", "dev", remote, teammate)
	return mine, teammate
}

// fillLog writes a finished log body in place of the template.
func fillLog(t *testing.T, root string) string {
	t.Helper()
	logs, _ := filepath.Glob(filepath.Join(root, ".mem/logs/*.md"))
	if len(logs) != 1 {
		t.Fatalf("logs = %v", logs)
	}
	data, _ := os.ReadFile(logs[0])
	front := strings.SplitN(string(data), "---\n", 3)
	body := "# Work Log - Tidy the parser\n\n## What Was Accomplished\n\nSplit the parser.\n"
	if err := os.WriteFile(logs[0], []byte("---\n"+front[1]+"---\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
	return logs[0]
}

func TestLogCommitRefusesAnUnfilledLog(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "log", "new")
	cmd := New()
	cmd.SetOut(new(strings.Builder))
	cmd.SetArgs([]string{"log", "commit", "--dir", mine})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "unfilled placeholder") {
		t.Fatalf("an unfilled log was committed: %v", err)
	}
}

func TestLogCommitCommitsRecordsSyncsAndPushes(t *testing.T) {
	mine, teammate := sharedProject(t)
	commit(t, teammate, "theirs.txt")
	run(t, teammate, "push", "--quiet")

	commit(t, mine, "parser.go")
	mem(t, mine, "todo", "new", "Benchmark the parser", "Compare with the old one.")
	mem(t, mine, "log", "new")
	fillLog(t, mine)
	writeFile(t, mine, "scratch.go", "package scratch\n")

	out := mem(t, mine, "log", "commit")
	if !strings.Contains(out, "Committed .mem/ records: Work log: Tidy the parser") || !strings.Contains(out, "Rebased") || !strings.Contains(out, "Pushed") {
		t.Fatalf("log commit output:\n%s", out)
	}
	if local, remote := run(t, mine, "rev-parse", "HEAD"), run(t, mine, "rev-parse", "origin/dev"); local != remote {
		t.Fatalf("dev %s does not match origin/dev %s", local, remote)
	}
	if _, err := os.Stat(filepath.Join(mine, "theirs.txt")); err != nil {
		t.Fatal("the teammate's commit was not brought in")
	}
	if status := run(t, mine, "status", "--porcelain"); status != "?? scratch.go" {
		t.Fatalf("status after log commit = %q, want only the uncommitted scratch.go", status)
	}
	if !strings.Contains(out, "⚠️ The session ends with uncommitted work in 1 file(s)") || strings.Contains(out, "session is closed") {
		t.Fatalf("uncommitted code was not reported:\n%s", out)
	}

	os.Remove(filepath.Join(mine, "scratch.go"))
	if out := mem(t, mine, "log", "commit"); !strings.Contains(out, "already committed") || !strings.Contains(out, "session is closed: the log is committed and dev matches origin/dev") {
		t.Fatalf("a clean close was not confirmed:\n%s", out)
	}
}

func writeLog(t *testing.T, root, name, user string, created time.Time, title string) {
	t.Helper()
	writeFile(t, root, ".mem/logs/"+name+".md", fmt.Sprintf("---\ncreated_at: %q\nuser: %s\n---\n# Work Log - %s\n\nBody of %s.\n", created.Format(time.RFC3339), user, title, title))
}

func TestOnboardShowsTheLatestLogInFullAndListsRecentOnes(t *testing.T) {
	mine, _ := sharedProject(t)
	now := time.Now()
	writeLog(t, mine, "test_user_old", "test_user", now.AddDate(0, 0, -5), "Earlier work")
	writeLog(t, mine, "test_user_new", "test_user", now.AddDate(0, 0, -1), "Latest work")
	writeLog(t, mine, "ana_recent", "ana", now.AddDate(0, 0, -3), "Ana's recent work")
	writeLog(t, mine, "ana_old", "ana", now.AddDate(0, 0, -30), "Ana's old work")
	writeFile(t, mine, ".mem/todos/benchmark.md", fmt.Sprintf("---\ntitle: Benchmark the parser\nstatus: open\ncreated_at: %q\n---\n\nCompare.\n", now.AddDate(0, 0, -12).Format(time.RFC3339)))

	out := mem(t, mine, "onboard", "--offline")
	if !strings.Contains(out, "Body of Latest work.") || strings.Contains(out, "Body of Earlier work.") || strings.Contains(out, "Body of Ana's recent work.") {
		t.Fatalf("only the latest log should be shown in full:\n%s", out)
	}
	if !strings.Contains(out, "Earlier work") || !strings.Contains(out, "Ana's recent work") || strings.Contains(out, "Ana's old work") {
		t.Fatalf("recent logs are not listed as expected:\n%s", out)
	}
	if !strings.Contains(out, "12 days") {
		t.Fatalf("todo age missing:\n%s", out)
	}
	if strings.Contains(out, "what the recent work logs say comes next") || !strings.Contains(out, "Work logs are background") {
		t.Fatalf("the onboard instruction still takes open work from logs:\n%s", out)
	}
}

func TestLogNewAsksToUpdateTheStructureDocAfterAnyCodeChange(t *testing.T) {
	mine, _ := sharedProject(t)
	writeFile(t, mine, ".mem/structure.md", "# Structure\n\nThe app is main.go.\n")
	writeFile(t, mine, "main.go", "package main\n")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Add the structure doc")

	if out := mem(t, mine, "log", "new"); strings.Contains(out, "structure doc") {
		t.Fatalf("log new mentions the structure doc without code changes:\n%s", out)
	}
	writeFile(t, mine, "server.go", "package main\n\nfunc serve() {}\n")
	run(t, mine, "add", "server.go")
	run(t, mine, "commit", "--quiet", "-m", "Add a server")
	if out := mem(t, mine, "log", "new"); !strings.Contains(out, "Code changed since the codebase structure doc was last updated: server.go.") {
		t.Fatalf("log new does not ask about the structure doc:\n%s", out)
	}
}
