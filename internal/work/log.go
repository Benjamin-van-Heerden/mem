package work

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

const logTemplate = `# Work Log - {short title}

## Overarching Goals

{
Broad goals and what we were trying to achieve with this work in the context of our interaction so far.
}

## What Was Accomplished

{
Description of what was done. Use appropriate subtitles to organize work done and things achieved.

Don't mention anything that is not relevant to actual changes made, e.g. deliberations or context building actions. You can be technical here and use actual code snippets and examples.
}

## Key Files Affected

{
List of files affected and changes made. Be reasonably detailed here.
}

## Errors and Barriers

{
Implementation errors and barriers encountered that have not been resolved yet. Mention approaches which were tried and failed so we can learn from them and avoid repeating mistakes.

(Omit this entire section if there were no errors or barriers)
}

## What Comes Next

{
If there are next steps or logical progressions from where we were, mention/list them here.

Be explicit about work done and expectations for follow up sessions. Remember that future sessions won't have any context about what was discussed.

If we were on an active spec, mention which parts of the spec were completed and which parts need further work.

(This section may be omitted entirely if nothing major needs to happen next)
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
