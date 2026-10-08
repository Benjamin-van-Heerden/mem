// Package checkpoint measures work that has not reached a checkpoint yet: the user's commits since their last work
// log or completed task, commits not pushed, and a stale structure doc. mem's post-commit hook and the compaction
// digest report it, so records and branches stay current in sessions that never end.
package checkpoint

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
)

const (
	suggestLogAt = 3
	requireLogAt = 5
	pushAt       = 3
	// maxWalk bounds the history read when looking for the last checkpoint.
	maxWalk = 1000
)

// ProjectFilesCommit is the subject of the commit in which onboard publishes mem's own project files; it is not work.
const ProjectFilesCommit = "Update mem project files"

// completions are the subjects of mem's task and spec completion commits. A completion note records what was done
// and how it was verified, so it resets the count like a work log.
var completions = []string{"Complete task ", "Complete spec "}

type commit struct {
	email, subject string
	files          []string
	// setsUp marks the commit that added .mem/config.toml.
	setsUp bool
}

// work reports whether c is the user's own work rather than a record commit or another person's.
func (c commit) work(email string) bool {
	if !strings.EqualFold(c.email, email) || c.subject == ProjectFilesCommit {
		return false
	}
	for _, f := range c.files {
		if !strings.HasPrefix(f, ".mem/") {
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
		for _, prefix := range completions {
			if strings.HasPrefix(c.subject, prefix) {
				return true
			}
		}
	}
	for _, f := range c.files {
		if strings.HasPrefix(f, ".mem/logs/"+user+"_") {
			return true
		}
	}
	return false
}

// history reads up to maxWalk non-merge commits on HEAD, newest first, with the files each changed.
func history(ctx context.Context, root string) ([]commit, error) {
	out, err := git.Run(ctx, root, "log", "--no-merges", "-n", strconv.Itoa(maxWalk), "--format=%x00%ae%x09%s", "--name-status", "HEAD")
	if err != nil {
		return nil, err
	}
	var commits []commit
	for _, chunk := range strings.Split(out, "\x00")[1:] {
		lines := strings.Split(strings.TrimSpace(chunk), "\n")
		email, subject, _ := strings.Cut(lines[0], "\t")
		c := commit{email: email, subject: subject}
		for _, line := range lines[1:] {
			fields := strings.Split(line, "\t")
			if len(fields) < 2 {
				continue
			}
			file := fields[len(fields)-1]
			if fields[0] == "A" && file == ".mem/config.toml" {
				c.setsUp = true
			}
			c.files = append(c.files, file)
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// WorkSinceLog counts the user's work commits on HEAD since their last work log or completed task. It also reports
// whether the newest commit is one of them.
func WorkSinceLog(ctx context.Context, root, user, email string) (count int, latestIsWork bool, err error) {
	commits, err := history(ctx, root)
	if err != nil {
		return 0, false, err
	}
	for i, c := range commits {
		if c.checkpoint(user, email) {
			break
		}
		if c.work(email) {
			count++
			if i == 0 {
				latestIsWork = true
			}
		}
	}
	return count, latestIsWork, nil
}

// LogLine describes n work commits since the last work log, escalating from a count to an instruction to stop.
func LogLine(n int) string {
	noun := "commits"
	if n == 1 {
		noun = "commit"
	}
	line := fmt.Sprintf("%d %s since your last work log or completed task", n, noun)
	switch {
	case n >= requireLogAt:
		return line + "; you should write one now, before continuing (`mem log new`)."
	case n >= suggestLogAt:
		return line + "; consider writing one now (`mem log new`)."
	}
	return line + "."
}

// LogRequired reports whether n work commits since the last log call for stopping to write one.
func LogRequired(n int) bool { return n >= requireLogAt }

// CommitLines are the nudges after a commit, from local state only: the work-log count after each of the user's
// work commits, unpushed commits past a threshold, and a stale structure doc.
func CommitLines(ctx context.Context, p project.Project, user, email string) []string {
	var lines []string
	if n, latest, err := WorkSinceLog(ctx, p.Root, user, email); err == nil && latest {
		lines = append(lines, LogLine(n))
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
