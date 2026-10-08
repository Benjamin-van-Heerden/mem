package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

func TestCompletingTasksAndSpecsCommitsTheRecordSyncsAndPushes(t *testing.T) {
	mine, teammate := sharedProject(t)
	mem(t, mine, "spec", "new", "Search")
	mem(t, mine, "task", "new", "Index", "Build the index.", "--spec", "search")
	mem(t, mine, "task", "new", "Query", "Answer queries.", "--spec", "search")
	mem(t, mine, "spec", "start", "search")
	run(t, teammate, "pull", "--quiet")
	commit(t, teammate, "theirs.go")
	run(t, teammate, "push", "--quiet")

	writeFile(t, mine, "index.go", "package index\n")
	writeFile(t, mine, "index_test.go", "package index\n")
	out := mem(t, mine, "task", "complete", "index", "Built the index; go test passes.")
	for _, want := range []string{"Committed: Index (every change in the working tree, with the task record)", "✔ Rebased 1 local commit(s) onto origin/dev", "✔ Pushed 1 commit(s) to origin/dev.", "📥 INCOMING", "Test: theirs.go", "Continue with the next task now, without waiting for approval: Query (query)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("task complete output lacks %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "📥 INCOMING") > strings.Index(out, "AGENT INSTRUCTION") {
		t.Fatalf("incoming changes print after the instruction:\n%s", out)
	}
	if run(t, mine, "rev-parse", "HEAD") != run(t, mine, "rev-parse", "origin/dev") {
		t.Fatal("task complete did not push")
	}
	if status := run(t, mine, "status", "--porcelain"); status != "" {
		t.Fatalf("task complete left work uncommitted:\n%s", status)
	}
	if message := run(t, mine, "log", "-1", "--format=%B"); message != "Index\n\nBuilt the index; go test passes.\n\nMem-Task: search/index" {
		t.Fatalf("task commit message = %q", message)
	}
	if files := run(t, mine, "show", "--name-only", "--format=", "HEAD"); !strings.Contains(files, "index.go") || !strings.Contains(files, "index_test.go") || !strings.Contains(files, ".mem/specs/search/tasks/01_index.md") {
		t.Fatalf("the task commit lacks the work or the record:\n%s", files)
	}

	mem(t, mine, "task", "complete", "query", "Answered.")
	out = mem(t, mine, "spec", "complete", "search")
	if !strings.Contains(out, "Committed: Complete spec search") || !strings.Contains(out, "✔ Pushed 1 commit(s) to origin/dev.") {
		t.Fatalf("spec complete output:\n%s", out)
	}
	if status := run(t, mine, "status", "--porcelain"); status != "" {
		t.Fatalf("spec complete left changes uncommitted:\n%s", status)
	}
	run(t, teammate, "pull", "--quiet")
	if files := run(t, teammate, "ls-files", ".mem/specs"); !strings.Contains(files, ".mem/specs/archive/search/spec.md") || strings.Contains(files, ".mem/specs/search/") {
		t.Fatalf("the archived spec did not reach the teammate:\n%s", files)
	}
}

func TestSyncPushesLocalCommits(t *testing.T) {
	mine, _ := sharedProject(t)
	commit(t, mine, "parser.go")
	out := mem(t, mine, "sync")
	if !strings.Contains(out, "✔ Pushed 1 commit(s) to origin/dev.") || strings.Contains(out, "⚠️") {
		t.Fatalf("sync output:\n%s", out)
	}
}

func TestSpecCompletionWaitsUntilConflictsAreResolved(t *testing.T) {
	mine, teammate := sharedProject(t)
	mem(t, mine, "spec", "new", "Search")
	mem(t, mine, "task", "new", "Index", "Build the index.", "--spec", "search")
	mem(t, mine, "spec", "start", "search")
	run(t, teammate, "pull", "--quiet")
	writeFile(t, teammate, "README.md", "theirs\n")
	run(t, teammate, "commit", "--quiet", "-am", "Their readme")
	run(t, teammate, "push", "--quiet")

	writeFile(t, mine, "README.md", "mine\n")
	mem(t, mine, "task", "complete", "index", "Rewrote the readme.")
	err := memErr(t, mine, "spec", "complete", "search")
	if err == nil || !strings.Contains(err.Error(), "spec search cannot be completed until the branch is in step") || !strings.Contains(err.Error(), "behind origin/dev") {
		t.Fatalf("spec complete with an unresolved conflict: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(mine, ".mem", "specs", "search", "spec.md")); statErr != nil {
		t.Fatal("the spec was archived despite the conflict")
	}

	if _, rebaseErr := git.Run(context.Background(), mine, "rebase", "origin/dev"); rebaseErr == nil {
		t.Fatal("the rebase was expected to conflict")
	}
	writeFile(t, mine, "README.md", "both\n")
	run(t, mine, "add", "README.md")
	if _, err := git.RunEnv(context.Background(), mine, []string{"GIT_EDITOR=true"}, "rebase", "--continue"); err != nil {
		t.Fatal(err)
	}
	if out := mem(t, mine, "spec", "complete", "search"); !strings.Contains(out, "Committed: Complete spec search") {
		t.Fatalf("spec complete after resolving:\n%s", out)
	}
}

func TestSpecCompletionNeedsAReachableRemote(t *testing.T) {
	mine, _ := sharedProject(t)
	mem(t, mine, "spec", "new", "Search")
	mem(t, mine, "task", "new", "Index", "Build the index.", "--spec", "search")
	mem(t, mine, "spec", "start", "search")
	mem(t, mine, "task", "complete", "index", "Built.")
	run(t, mine, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	if err := memErr(t, mine, "spec", "complete", "search"); err == nil || !strings.Contains(err.Error(), "could not fetch from origin") {
		t.Fatalf("spec complete without the remote: %v", err)
	}
}
