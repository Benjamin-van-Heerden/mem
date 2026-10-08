package cli

import (
	"strings"
	"testing"
)

func TestCompactHookSyncsAndPrintsAShortDigest(t *testing.T) {
	mine, teammate := sharedProject(t)
	mem(t, mine, "spec", "new", "Parser")
	mem(t, mine, "task", "new", "Tokenize", "Split input into tokens.", "--spec", "parser")
	mem(t, mine, "task", "new", "Parse", "Build the tree.", "--spec", "parser")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Draft the parser spec")
	mem(t, mine, "spec", "start", "parser")
	mem(t, mine, "todo", "new", "Parser docs", "Document the grammar.")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Record the docs todo")
	mem(t, mine, "todo", "claim", "parser_docs")
	run(t, mine, "push", "--quiet")

	run(t, teammate, "pull", "--quiet")
	mem(t, teammate, "memory", "set", "logging", "Use the structured logger.")
	mem(t, teammate, "todo", "new", "Flaky CI", "The cache step fails sometimes.")
	commit(t, teammate, "lexer.go")
	run(t, teammate, "add", "--all")
	run(t, teammate, "commit", "--quiet", "-m", "Add the lexer")
	run(t, teammate, "push", "--quiet")

	out := mem(t, mine, "hook", "compact")
	for _, want := range []string{
		"🔄 MEM AFTER COMPACTION",
		"Branch: dev (tracking origin/dev: 0 ahead, 0 behind)",
		"✔ Fast-forwarded dev",
		"Active spec: Parser (parser), 0 of 2 task(s) done; next: Tokenize (tokenize)",
		"Your claimed todos: parser_docs",
		"Test: Add the lexer",
		"Todo opened: Flaky CI",
		"New: logging\nUse the structured logger.",
		"Follow the changed memories",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("digest lacks %q:\n%s", want, out)
		}
	}
	if len(out) > 3000 {
		t.Fatalf("digest is %d bytes:\n%s", len(out), out)
	}
	if strings.Contains(out, "CODEBASE AND STRUCTURE") || strings.Contains(out, "WORK LOGS") {
		t.Fatalf("digest repeats onboard context:\n%s", out)
	}
}

func TestCompactHookNeverFailsOutsideAMemProject(t *testing.T) {
	out := mem(t, t.TempDir(), "hook", "compact")
	if !strings.Contains(out, "mem could not sync this checkout after compaction") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestCompactHookAsksForAWorkLogOnceWorkHasGathered(t *testing.T) {
	mine, _ := sharedProject(t)
	for _, file := range []string{"a.go", "b.go"} {
		commit(t, mine, file)
	}
	out := mem(t, mine, "hook", "compact")
	if !strings.Contains(out, "Work log: 2 commits since your last work log or completed task.") || strings.Contains(out, "Write a work log") {
		t.Fatalf("digest after 2 commits:\n%s", out)
	}
	if !strings.Contains(out, "✔ Pushed 2 commit(s) to origin/dev.") || run(t, mine, "rev-parse", "HEAD") != run(t, mine, "rev-parse", "origin/dev") {
		t.Fatalf("the compaction catch-up did not push the committed work:\n%s", out)
	}
	for _, file := range []string{"c.go", "d.go", "e.go"} {
		commit(t, mine, file)
	}
	out = mem(t, mine, "hook", "compact")
	if !strings.Contains(out, "Work log: 5 commits since your last work log or completed task; you should stop and write one now") || !strings.Contains(out, "Write a work log for the work since your last one now") {
		t.Fatalf("digest after 5 commits:\n%s", out)
	}
}
