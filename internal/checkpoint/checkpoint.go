// Package checkpoint measures work that has not reached a checkpoint yet: the user's commits and changed files since
// their last work log or completed task, commits not pushed, and a stale structure doc. mem's post-commit hook, the
// prompt hook and the compaction digest report it, so records and branches stay current in sessions that never end.
package checkpoint

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
)

// Thresholds for the work-log nudges; whichever measure reaches its threshold first sets the level.
const (
	suggestCommits, suggestFiles, suggestLines = 5, 10, 500
	requireCommits, requireFiles, requireLines = 10, 20, 1200
	pushAt                                     = 3
	// firstWalk and maxWalk bound the history read when looking for the last checkpoint, which is usually recent.
	firstWalk = 100
	maxWalk   = 1000
)

// ProjectFilesCommit is the subject of the commit in which onboard publishes mem's own project files; it is not work.
const ProjectFilesCommit = "Update mem project files"

// TaskTrailer marks the commit `mem task complete` makes. Its completion note records what was done and how it
// was verified, so it resets the count like a work log.
const TaskTrailer = "Mem-Task:"

// completions are the subjects of mem's record-only completion commits: spec completion, and task completion
// before task commits carried the work and TaskTrailer.
var completions = []string{"Complete task ", "Complete spec "}

type change struct {
	file  string
	lines int
}

type commit struct {
	hash, email, subject, task string
	changes              []change
	// setsUp marks the commit that added .mem/config.toml.
	setsUp bool
}

// own reports whether c is the user's own, changing something outside .mem/, rather than a record commit or
// another person's.
func (c commit) own(email string) bool {
	if !strings.EqualFold(c.email, email) || c.subject == ProjectFilesCommit {
		return false
	}
	for _, ch := range c.changes {
		if !strings.HasPrefix(ch.file, ".mem/") {
			return true
		}
	}
	return false
}

// work reports whether c is the user's own and changes a file that counts as work.
func (c commit) work(email string, s project.StructureConfig) bool {
	if !c.own(email) {
		return false
	}
	for _, ch := range c.changes {
		if structure.Relevant(ch.file, s) {
			return true
		}
	}
	return false
}

// checkpoint reports whether c ends the count: a work log of the user's, their completed task or spec, or the
// commit that set up mem in the repository.
func (c commit) checkpoint(user, email string) bool {
	if c.setsUp {
		return true
	}
	if strings.EqualFold(c.email, email) {
		if c.task != "" {
			return true
		}
		for _, prefix := range completions {
			if strings.HasPrefix(c.subject, prefix) {
				return true
			}
		}
	}
	for _, ch := range c.changes {
		if strings.HasPrefix(ch.file, ".mem/logs/"+user+"_") {
			return true
		}
	}
	return false
}

// history reads up to n non-merge commits on HEAD, newest first, with the files each changed and their line counts.
func history(ctx context.Context, root string, n int) ([]commit, error) {
	out, err := git.Run(ctx, root, "log", "--no-merges", "-n", strconv.Itoa(n), "--format=%x00%h%x09%ae%x09%s%x09%(trailers:key="+strings.TrimSuffix(TaskTrailer, ":")+",valueonly,separator=%x2C)", "--numstat", "--summary", "HEAD")
	if err != nil {
		return nil, err
	}
	var commits []commit
	for _, chunk := range strings.Split(out, "\x00")[1:] {
		lines := strings.Split(strings.TrimSpace(chunk), "\n")
		fields := strings.SplitN(lines[0], "\t", 4)
		c := commit{hash: fields[0]}
		if len(fields) > 1 {
			c.email = fields[1]
		}
		if len(fields) > 2 {
			c.subject = fields[2]
		}
		if len(fields) > 3 {
			c.task = strings.TrimSpace(fields[3])
		}
		for _, line := range lines[1:] {
			if summary := strings.TrimSpace(line); strings.HasPrefix(summary, "create mode ") && strings.HasSuffix(summary, " .mem/config.toml") {
				c.setsUp = true
				continue
			}
			fields := strings.SplitN(line, "\t", 3)
			if len(fields) != 3 {
				continue
			}
			added, _ := strconv.Atoi(fields[0])
			deleted, _ := strconv.Atoi(fields[1])
			c.changes = append(c.changes, change{file: structure.RenamedTo(fields[2]), lines: added + deleted})
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// Work is the user's work since their last work log or completed task: their work commits, and the files and lines
// that count as work (structure.Relevant) changed by those commits and by uncommitted changes.
type Work struct {
	Commits, Files, Lines int
	// LatestIsWork reports whether the newest commit is one of the user's work commits.
	LatestIsWork bool
}

// Level is how strongly the work since the last checkpoint calls for a work log.
type Level int

const (
	Quiet Level = iota
	Suggested
	Due
)

func (w Work) Level() Level {
	switch {
	case w.Commits >= requireCommits || w.Files >= requireFiles || w.Lines >= requireLines:
		return Due
	case w.Commits >= suggestCommits || w.Files >= suggestFiles || w.Lines >= suggestLines:
		return Suggested
	}
	return Quiet
}

// sinceCheckpoint reads the commits on HEAD back to the user's last checkpoint, newest first, without it.
func sinceCheckpoint(ctx context.Context, root, user, email string) ([]commit, error) {
	commits, err := history(ctx, root, firstWalk)
	if err == nil && len(commits) == firstWalk && !anyCheckpoint(commits, user, email) {
		commits, err = history(ctx, root, maxWalk)
	}
	for i, c := range commits {
		if c.checkpoint(user, email) {
			return commits[:i], err
		}
	}
	return commits, err
}

// Since measures the user's work since their last checkpoint. Teammates' commits do not count, and neither do
// mem's records; uncommitted changes in the checkout do, whoever made them.
func Since(ctx context.Context, p project.Project, user, email string) (Work, error) {
	commits, err := sinceCheckpoint(ctx, p.Root, user, email)
	if err != nil {
		return Work{}, err
	}
	var w Work
	files := map[string]bool{}
	count := func(file string, lines int) {
		if structure.Relevant(file, p.Config.Structure) {
			files[file] = true
			w.Lines += lines
		}
	}
	for i, c := range commits {
		if !c.work(email, p.Config.Structure) {
			continue
		}
		w.Commits++
		w.LatestIsWork = w.LatestIsWork || i == 0
		for _, ch := range c.changes {
			count(ch.file, ch.lines)
		}
	}
	if d, err := structure.ChangesSince(ctx, p, "HEAD"); err == nil {
		for _, ch := range d.Changes {
			count(ch.Path, ch.Lines)
		}
	}
	w.Files = len(files)
	return w, nil
}

// Covered is what a work log written now covers: the user's own commits since their last checkpoint, oldest first,
// and every file outside .mem/ they and the uncommitted changes touched, most changed first.
type Covered struct {
	Commits []string
	Files   []structure.Change
}

func Cover(ctx context.Context, root, user, email string) (Covered, error) {
	commits, err := sinceCheckpoint(ctx, root, user, email)
	if err != nil {
		return Covered{}, err
	}
	var cov Covered
	lines := map[string]int{}
	var order []string
	add := func(file string, n int) {
		if strings.HasPrefix(file, ".mem/") || path.Base(file) == ".DS_Store" {
			return
		}
		if _, seen := lines[file]; !seen {
			order = append(order, file)
		}
		lines[file] += n
	}
	for i := len(commits) - 1; i >= 0; i-- {
		if c := commits[i]; c.own(email) {
			cov.Commits = append(cov.Commits, c.hash)
			for _, ch := range c.changes {
				add(ch.file, ch.lines)
			}
		}
	}
	if changes, err := structure.Diff(ctx, root, "HEAD"); err == nil {
		for _, ch := range changes {
			add(ch.Path, ch.Lines)
		}
	}
	for _, file := range order {
		cov.Files = append(cov.Files, structure.Change{Path: file, Lines: lines[file]})
	}
	sort.SliceStable(cov.Files, func(i, j int) bool { return cov.Files[i].Lines > cov.Files[j].Lines })
	return cov, nil
}

func anyCheckpoint(commits []commit, user, email string) bool {
	for _, c := range commits {
		if c.checkpoint(user, email) {
			return true
		}
	}
	return false
}

// Summary describes the work in one clause, such as "3 commits, 12 files and 640 lines of work since your last work
// log or completed task".
func (w Work) Summary() string {
	return fmt.Sprintf("%s, %s and %s of work since your last work log or completed task", plural(w.Commits, "commit"), plural(w.Files, "file"), plural(w.Lines, "line"))
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

const carryOn = "A log does not end your turn: carry on with the work afterwards."

// MidTurnLine is the nudge for an agent that may be in the middle of something, after a commit or a compaction: it
// never asks the agent to drop the work at hand.
func MidTurnLine(w Work) string {
	switch w.Level() {
	case Due:
		return w.Summary() + "; a work log is due: write one at the next natural stopping point, such as when the current change is done (`mem log new`). " + carryOn
	case Suggested:
		return w.Summary() + "; write a work log at the next natural stopping point (`mem log new`). " + carryOn
	}
	return w.Summary() + "."
}

// TurnStartLine is the nudge when the user sends a message, usually at the start of a turn but possibly while the
// agent is still working; it is empty while no log is called for.
func TurnStartLine(w Work) string {
	switch w.Level() {
	case Due:
		return w.Summary() + ". A work log is due: write one before taking up substantial new work (`mem log new`, then `mem log commit`); if you are in the middle of a change, finish it first. Answer a quick question first, and if the user asked you to stop, stop. Writing a log does not end your turn: continue with the user's request afterwards."
	case Suggested:
		return w.Summary() + ". Write a work log at the next natural stopping point in this turn (`mem log new`). " + carryOn
	}
	return ""
}

// TurnEndLine is Cursor's follow-up when an agent ends its turn with a work log due. Cursor sends it as the next user
// message, so it says where it comes from and leaves room for a turn that ended waiting on the user; it is empty
// while no log is due.
func TurnEndLine(w Work) string {
	if w.Level() != Due {
		return ""
	}
	return "This is an automatic follow-up from mem, not a message from the user. " + w.Summary() + ", and a work log is due. If your last turn finished a piece of work, write the log now (`mem log new`, then `mem log commit`) and then end your turn. If it ended with a question for the user or in the middle of a change, do not write it yet: say in one line that a work log is due and end your turn. Do not start other work."
}

// CommitLines are the nudges after a commit, from local state only: the work-log count after each of the user's
// work commits, unpushed commits past a threshold, and a stale structure doc.
func CommitLines(ctx context.Context, p project.Project, user, email string) []string {
	var lines []string
	if w, err := Since(ctx, p, user, email); err == nil && w.LatestIsWork {
		lines = append(lines, MidTurnLine(w))
	}
	if out, err := git.Run(ctx, p.Root, "rev-list", "--count", "@{upstream}..HEAD"); err == nil {
		if n, _ := strconv.Atoi(out); n >= pushAt {
			branch := git.CurrentBranch(ctx, p.Root)
			lines = append(lines, fmt.Sprintf("%d commits on %s are not pushed: run `mem sync` to share them.", n, branch))
		}
	}
	if d, err := structure.Measure(ctx, p); err == nil && d.Stale() {
		lines = append(lines, fmt.Sprintf("The structure doc is out of date: %d code file(s) and %d line(s) changed since it was last updated. Refresh it with `mem structure`.", len(d.Changes), d.Lines))
	}
	return lines
}
