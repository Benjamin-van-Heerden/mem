package release

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// NotesPath is where `mem promote production` keeps the release notes draft until the release is confirmed.
const NotesPath = ".mem/local/release-notes.md"

const maxNoteCommits = 30

var draftMarker = regexp.MustCompile(`^<!-- mem:release ([0-9a-f]{7,64}) -->\n`)

type specSummary struct {
	Slug, Title, Overview string
}

type logSummary struct {
	Title    string
	Sections []string
}

// DraftNotes gathers what a release delivers from the records mem keeps: specs completed in the range, work
// logs written in it, and its commits. The first line marks the commit the draft was written for.
func DraftNotes(ctx context.Context, p project.Project, pl Plan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- mem:release %s -->\n# Release %s\n", pl.To, pl.Tag)
	if specs := specSummaries(ctx, p.Root, pl.From, pl.To); len(specs) > 0 {
		b.WriteString("\n## Specs completed\n\n")
		for _, s := range specs {
			if s.Overview != "" {
				fmt.Fprintf(&b, "- **%s**: %s\n", s.Title, s.Overview)
			} else {
				fmt.Fprintf(&b, "- **%s**\n", s.Title)
			}
		}
	}
	if logs := logSummaries(ctx, p.Root, pl.From, pl.To); len(logs) > 0 {
		b.WriteString("\n## Work\n\n")
		for _, l := range logs {
			fmt.Fprintf(&b, "- %s\n", l.Title)
			for _, s := range l.Sections {
				fmt.Fprintf(&b, "  - %s\n", s)
			}
		}
	}
	b.WriteString("\n## Commits\n\n")
	for i, c := range pl.Commits {
		if i == maxNoteCommits {
			fmt.Fprintf(&b, "- … and %d more\n", len(pl.Commits)-i)
			break
		}
		fmt.Fprintf(&b, "- %s (%s)\n", c.Subject, c.Author)
	}
	return b.String()
}

// GeneratedMessage is the tag message of a release nobody wrote notes for (`mem deploy`).
func GeneratedMessage(ctx context.Context, p project.Project, pl Plan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Release %s\n", pl.Tag)
	if specs := specSummaries(ctx, p.Root, pl.From, pl.To); len(specs) > 0 {
		titles := make([]string, len(specs))
		for i, s := range specs {
			titles[i] = s.Title
		}
		fmt.Fprintf(&b, "\nSpecs completed: %s\n", strings.Join(titles, "; "))
	}
	b.WriteString("\nCommits:\n")
	for i, c := range pl.Commits {
		if i == maxNoteCommits {
			fmt.Fprintf(&b, "- … and %d more\n", len(pl.Commits)-i)
			break
		}
		fmt.Fprintf(&b, "- %s\n", c.Subject)
	}
	return b.String()
}

// DraftCommit returns the commit a draft was written for, and the notes without the marker.
func DraftCommit(draft string) (commit, notes string) {
	m := draftMarker.FindStringSubmatch(draft)
	if m == nil {
		return "", strings.TrimSpace(draft)
	}
	return m[1], strings.TrimSpace(draft[len(m[0]):])
}

// specSummaries finds specs archived as completed within the range, with their title and first overview sentence (when it stands on its own).
func specSummaries(ctx context.Context, root, from, to string) []specSummary {
	var specs []specSummary
	for _, file := range addedFiles(ctx, root, from, to, ".mem/specs/archive") {
		if path.Base(file) != "spec.md" {
			continue
		}
		content, err := git.Run(ctx, root, "show", to+":"+file)
		if err != nil || !strings.Contains(content, "\nstatus: completed\n") {
			continue
		}
		slug := path.Base(path.Dir(file))
		title := frontmatterValue(content, "title")
		if title == "" {
			title = slug
		}
		specs = append(specs, specSummary{Slug: slug, Title: title, Overview: firstSentence(section(content, "## Overview"))})
	}
	return specs
}

// logSummaries lists the work logs added within the range, with the headings under "What Was Accomplished".
func logSummaries(ctx context.Context, root, from, to string) []logSummary {
	var logs []logSummary
	for _, file := range addedFiles(ctx, root, from, to, ".mem/logs") {
		content, err := git.Run(ctx, root, "show", to+":"+file)
		if err != nil {
			continue
		}
		title := ""
		for _, line := range strings.Split(content, "\n") {
			if rest, ok := strings.CutPrefix(line, "# Work Log - "); ok {
				title = strings.TrimSpace(rest)
				break
			}
		}
		if title == "" {
			continue
		}
		var sections []string
		for _, line := range strings.Split(section(content, "## What Was Accomplished"), "\n") {
			if rest, ok := strings.CutPrefix(line, "### "); ok {
				sections = append(sections, strings.TrimSpace(rest))
			}
		}
		logs = append(logs, logSummary{Title: title, Sections: sections})
	}
	return logs
}

// addedFiles lists files under dir added between from and to (everything under dir at to when from is empty).
func addedFiles(ctx context.Context, root, from, to, dir string) []string {
	args := []string{"ls-tree", "-r", "--name-only", to, "--", dir}
	if from != "" {
		args = []string{"diff", "--name-only", "--diff-filter=A", from, to, "--", dir}
	}
	out, err := git.Run(ctx, root, args...)
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func frontmatterValue(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if rest, ok := strings.CutPrefix(line, key+":"); ok {
			return strings.Trim(strings.TrimSpace(rest), `"'`)
		}
	}
	return ""
}

// section returns the text under a `## ` heading, up to the next `## ` or `# ` heading.
func section(content, heading string) string {
	_, rest, ok := strings.Cut(content, "\n"+heading+"\n")
	if !ok {
		return ""
	}
	for _, stop := range []string{"\n## ", "\n# "} {
		if i := strings.Index(rest, stop); i >= 0 {
			rest = rest[:i]
		}
	}
	return strings.TrimSpace(rest)
}

func firstSentence(text string) string {
	paragraph, _, _ := strings.Cut(strings.TrimSpace(text), "\n\n")
	paragraph = strings.Join(strings.Fields(paragraph), " ")
	if i := strings.Index(paragraph, ". "); i >= 0 {
		paragraph = paragraph[:i+1]
	}
	// A sentence ending in a colon introduces a list and reads as a fragment on its own.
	if strings.HasSuffix(paragraph, ":") {
		return ""
	}
	return paragraph
}
