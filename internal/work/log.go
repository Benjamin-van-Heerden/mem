package work

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

const logTemplate = `# Work Log - {short title}

<!-- A work log records what happened in this session, as fact. It is not updated later. Anything still to be done, including blockers and decisions waiting on the user, belongs in a todo, not here. -->

## Overarching Goals

{
What we set out to achieve in this session, in the context of the interaction so far.
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
