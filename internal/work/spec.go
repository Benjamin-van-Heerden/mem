package work

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

const (
	SpecDraft     = "draft"
	SpecActive    = "active"
	SpecCompleted = "completed"
	SpecAbandoned = "abandoned"
)

const specTemplate = `## Overview

{Describe the feature or change}

## Goals

- {Goal 1}
- {Goal 2}

## Technical Approach

{How to implement this}

## Success Criteria

- {Criterion 1}
- {Criterion 2}

## Notes

{Additional context}
`

type SpecMeta struct {
	Title      string `yaml:"title"`
	Status     string `yaml:"status"`
	AssignedTo string `yaml:"assigned_to,omitempty"`
	Created    string `yaml:"created_at"`
	Updated    string `yaml:"updated_at"`
	Completed  string `yaml:"completed_at,omitempty"`
	Reason     string `yaml:"abandon_reason,omitempty"`
}

type Spec struct {
	Slug string
	Dir  string
	Meta SpecMeta
	Body string
}

func (s Spec) slug() string  { return s.Slug }
func (s Spec) title() string { return s.Meta.Title }
func (s Spec) Path() string  { return filepath.Join(s.Dir, "spec.md") }
func (s Spec) Archived() bool {
	return s.Meta.Status == SpecCompleted || s.Meta.Status == SpecAbandoned
}

func specsDir(p project.Project) string   { return p.Path("specs") }
func archiveDir(p project.Project) string { return p.Path("specs", "archive") }

// Specs lists open specs, and archived ones too when includeArchived is set.
func Specs(p project.Project, includeArchived bool) ([]Spec, error) {
	specs, err := readSpecs(specsDir(p))
	if err != nil || !includeArchived {
		return specs, err
	}
	archived, err := readSpecs(archiveDir(p))
	return append(specs, archived...), err
}

func readSpecs(dir string) ([]Spec, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var specs []Spec
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "archive" || !pathExists(filepath.Join(dir, e.Name(), "spec.md")) {
			continue
		}
		s := Spec{Slug: e.Name(), Dir: filepath.Join(dir, e.Name())}
		if s.Body, err = ReadMarkdown(s.Path(), &s.Meta); err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Meta.Created < specs[j].Meta.Created })
	return specs, nil
}

func FindSpec(p project.Project, ref string) (Spec, error) {
	specs, err := Specs(p, true)
	if err != nil {
		return Spec{}, err
	}
	return resolve("spec", ref, specs)
}

func NewSpec(p project.Project, title string) (Spec, error) {
	slug, err := uniqueSlug(project.Slugify(title), func(s string) bool {
		return pathExists(filepath.Join(specsDir(p), s)) || pathExists(filepath.Join(archiveDir(p), s))
	})
	if err != nil {
		return Spec{}, err
	}
	s := Spec{Slug: slug, Dir: filepath.Join(specsDir(p), slug), Body: specTemplate}
	s.Meta = SpecMeta{Title: title, Status: SpecDraft, Created: now(), Updated: now()}
	return s, SaveSpec(s)
}

func SaveSpec(s Spec) error {
	s.Meta.Updated = now()
	return WriteMarkdown(s.Path(), s.Meta, s.Body)
}

// ArchiveSpec records a final status and moves the spec, with its tasks, into specs/archive/.
func ArchiveSpec(p project.Project, s Spec, status, reason string) (Spec, error) {
	s.Meta.Status = status
	s.Meta.Completed = now()
	s.Meta.Reason = reason
	if err := SaveSpec(s); err != nil {
		return s, err
	}
	target := filepath.Join(archiveDir(p), s.Slug)
	if err := os.MkdirAll(archiveDir(p), 0o755); err != nil {
		return s, err
	}
	if err := os.Rename(s.Dir, target); err != nil {
		return s, fmt.Errorf("archive spec: %w", err)
	}
	s.Dir = target
	return s, nil
}

// ActiveSpecFor returns the single active spec assigned to user, if there is exactly one.
func ActiveSpecFor(p project.Project, user string) (Spec, bool, error) {
	specs, err := Specs(p, false)
	if err != nil {
		return Spec{}, false, err
	}
	var mine []Spec
	for _, s := range specs {
		if s.Meta.Status == SpecActive && s.Meta.AssignedTo == user {
			mine = append(mine, s)
		}
	}
	if len(mine) != 1 {
		return Spec{}, false, nil
	}
	return mine[0], true, nil
}
