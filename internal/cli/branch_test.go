package cli

import (
	"strings"
	"testing"
)

func TestRecordChangesOnAFeatureBranchStayThereAndSaySo(t *testing.T) {
	mine, _ := sharedProject(t)
	draft := func(title, slug string) {
		mem(t, mine, "spec", "new", title)
		mem(t, mine, "task", "new", "First", "Do it.", "--spec", slug)
		run(t, mine, "add", "--all")
		run(t, mine, "commit", "--quiet", "-m", "Draft "+slug)
	}
	draft("Parser", "parser")
	if out := mem(t, mine, "spec", "start", "parser"); strings.Contains(out, "teammates see this") {
		t.Fatalf("spec start on dev mentioned a merge:\n%s", out)
	}

	run(t, mine, "switch", "--quiet", "-c", "refunds")
	draft("Refunds", "refunds")
	out := mem(t, mine, "spec", "start", "refunds")
	if !strings.Contains(out, "You are on refunds, not dev: teammates see this record change once refunds merges into dev.") {
		t.Fatalf("spec start on a feature branch:\n%s", out)
	}
	if subject := run(t, mine, "log", "-1", "--format=%s", "refunds"); subject != "Start spec refunds" {
		t.Fatalf("the spec start was not committed on the feature branch: %q", subject)
	}
	if subject := run(t, mine, "log", "-1", "--format=%s", "dev"); subject == "Start spec refunds" {
		t.Fatal("the spec start was committed on dev")
	}
}

func TestSpecCompletedOnAFeatureBranchAsksToMergeItBack(t *testing.T) {
	mine, _ := sharedProject(t)
	run(t, mine, "switch", "--quiet", "-c", "refunds")
	mem(t, mine, "spec", "new", "Refunds")
	mem(t, mine, "task", "new", "Refund flow", "Build it.", "--spec", "refunds")
	run(t, mine, "add", "--all")
	run(t, mine, "commit", "--quiet", "-m", "Draft the refunds spec")
	mem(t, mine, "spec", "start", "refunds")
	mem(t, mine, "task", "complete", "refund_flow", "Built and tested.")
	out := mem(t, mine, "spec", "complete", "refunds")
	if !strings.Contains(out, "The spec was built on refunds. Merge it into dev now") {
		t.Fatalf("spec completion on a feature branch:\n%s", out)
	}
}
