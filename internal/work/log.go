package work

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// logGuidance is for the agent writing the log; Finish removes it before the log is committed.
const logGuidance = "<!-- A work log records a stretch of work since the previous log, as fact and as things stand. It is not updated later. Anything still to be done, including blockers and decisions waiting on the user, belongs in a todo, not here. -->"

// The placeholders for the key files: one under the files mem listed from Git, one when it found none.
const (
	listedFilesPlaceholder = `{
Say what changed in the files above where the rest of the log does not already make it clear, remove files that belong to someone else's work, such as another session in this checkout, and add files changed outside this repository. Then remove this placeholder.
}`
	filesPlaceholder = `{
Files affected and the changes made to them, including files outside this repository.
}`
)

const logTemplate = "# Work Log - {short title}\n\n" + logGuidance + `

## Goal

{
What this stretch of work was for, in one or two sentences.
}

## What Was Accomplished

{
What was done. Use subheadings to organize it. Leave out deliberation and context building; be technical and include code snippets where they help.
}

## Decisions

{
Decisions made and the reasons for them, so later sessions do not reopen them without cause.

(Omit this section if no notable decisions were made)
}

## Key Files Affected

<files>

## Notes

{
Anything else worth keeping: errors and barriers met, approaches that were tried and dropped and why, so they are not repeated.

(Omit this section if there is nothing to note)
}
`

type LogMeta struct {
	Created string `yaml:"created_at"`
	User    string `yaml:"user"`
	Spec    string `yaml:"spec,omitempty"`
	// Commits are the user's work commits the log covers, oldest first, as they were when it was written.
	Commits []string `yaml:"commits,omitempty,flow"`
}

type Log struct {
	Name string
	Path string
	Meta LogMeta
	Body string
}

func (l Log) slug() string  { return l.Name }
func (l Log) title() string { return l.Name }

var placeholder = regexp.MustCompile(`\{[^{}]*\}`)

// Unfilled returns the template placeholders still present in the log, each
// by its first line of guidance.
func (l Log) Unfilled() []string {
	var left []string
	for _, ph := range placeholder.FindAllString(logTemplate+listedFilesPlaceholder+filesPlaceholder, -1) {
		if !strings.Contains(l.Body, ph) {
			continue
		}
		for _, line := range strings.Split(strings.Trim(ph, "{}"), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				left = append(left, line)
				break
			}
		}
	}
	return left
}

// Finish removes the template's guidance comment from the log file.
func (l *Log) Finish() error {
	body := strings.Replace(l.Body, logGuidance+"\n\n", "", 1)
	if body == l.Body {
		return nil
	}
	l.Body = body
	return WriteMarkdown(l.Path, l.Meta, body)
}

// Heading is the log's title from its first heading, or its file name.
func (l Log) Heading() string {
	for _, line := range strings.Split(l.Body, "\n") {
		if title, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(strings.TrimPrefix(title, "Work Log - "))
		}
	}
	return l.Name
}

// LatestLog returns the user's newest log.
func LatestLog(p project.Project, user string) (Log, bool, error) {
	logs, err := Logs(p)
	if err != nil {
		return Log{}, false, err
	}
	for _, l := range logs {
		if l.Meta.User == user {
			return l, true, nil
		}
	}
	return Log{}, false, nil
}

func logsDir(p project.Project) string { return p.Path("logs") }

// Logs returns all logs, newest first.
func Logs(p project.Project) ([]Log, error) {
	entries, err := os.ReadDir(logsDir(p))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var logs []Log
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || e.IsDir() {
			continue
		}
		l := Log{Name: name, Path: filepath.Join(logsDir(p), e.Name())}
		if l.Body, err = ReadMarkdown(l.Path, &l.Meta); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].Meta.Created > logs[j].Meta.Created })
	return logs, nil
}

func FindLog(p project.Project, ref string) (Log, error) {
	logs, err := Logs(p)
	if err != nil {
		return Log{}, err
	}
	return resolve("log", ref, logs)
}

// LogFile is a file the work since the last log changed, with its changed lines.
type LogFile struct {
	Path  string
	Lines int
}

// maxLogFiles caps the files listed in a new log; the agent names the rest where they matter.
const maxLogFiles = 30

// NewLog creates a log from the template, listing the files and commits the work since the last log covers.
func NewLog(p project.Project, user, spec string, commits []string, files []LogFile) (Log, error) {
	created := time.Now()
	name := user + "_" + created.Format("20060102_150405")
	l := Log{Name: name, Path: filepath.Join(logsDir(p), name+".md"), Body: strings.Replace(logTemplate, "<files>", keyFiles(files), 1)}
	l.Meta = LogMeta{Created: created.Format(time.RFC3339), User: user, Spec: spec, Commits: commits}
	return l, WriteMarkdown(l.Path, l.Meta, l.Body)
}

func keyFiles(files []LogFile) string {
	if len(files) == 0 {
		return filesPlaceholder
	}
	var b strings.Builder
	for i, f := range files {
		if i == maxLogFiles {
			fmt.Fprintf(&b, "- … and %d more\n", len(files)-i)
			break
		}
		fmt.Fprintf(&b, "- `%s` (%d lines)\n", f.Path, f.Lines)
	}
	return b.String() + "\n" + listedFilesPlaceholder
}
