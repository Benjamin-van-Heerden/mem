package work

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// logGuidance is for the agent writing the log; Finish removes it before the log is committed.
const logGuidance = "<!-- A work log records a stretch of work since the previous log, as fact. It is not updated later. Anything still to be done, including blockers and decisions waiting on the user, belongs in a todo, not here. -->"

const logTemplate = "# Work Log - {short title}\n\n" + logGuidance + `

## Overarching Goals

{
What we set out to achieve in this stretch of work, in the context of the interaction so far.
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

{
Files affected and the changes made to them. Be reasonably detailed.
}

## Errors and Barriers

{
Errors and barriers met, approaches that were tried and failed, and why, so they are not repeated.

(Omit this section if there were none)
}
`

type LogMeta struct {
	Created string `yaml:"created_at"`
	User    string `yaml:"user"`
	Spec    string `yaml:"spec,omitempty"`
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
	for _, ph := range placeholder.FindAllString(logTemplate, -1) {
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

func NewLog(p project.Project, user, spec string) (Log, error) {
	created := time.Now()
	name := user + "_" + created.Format("20060102_150405")
	l := Log{Name: name, Path: filepath.Join(logsDir(p), name+".md"), Body: logTemplate}
	l.Meta = LogMeta{Created: created.Format(time.RFC3339), User: user, Spec: spec}
	return l, WriteMarkdown(l.Path, l.Meta, l.Body)
}
