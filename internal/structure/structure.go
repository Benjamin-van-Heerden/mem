// Package structure maintains .memr/structure.md, the living map of the
// codebase, and measures how far the code has moved since it was last updated.
package structure

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
)

const (
	RelPath        = ".memr/structure.md"
	staleFiles     = 5
	staleLines     = 1000
	maxTreeEntries = 300
)

const Template = `# Codebase and Structure

## Overview
{One concise factual summary of the repository as found. Describe the main artifact or application type only when confirmed by repo files. Do not infer goals or intent.}

## Tech Stack
{Bullet list: languages, frameworks, key libraries, databases, and infrastructure that matter for day-to-day development.}

## Directory Layout
{Annotated tree of top-level and important nested directories. Describe what lives in each important directory and which directories are source, tests, generated output, configuration, or durable project state. Do not list every file.}

## Key Modules
{For each significant module or component, describe what it does, key files, and what it depends on or what depends on it.}

## Data Flow
{Describe the primary data paths through the system. Cover the main flows, not every edge case.}

## Entry Points
{Describe how the application, service, package, or tool is started. Include development and production entry points when they differ.}

## Commands and Workflows
{Document repo-defined commands, scripts, task runners and developer workflows that are visible in repo files.}

## External Interfaces
{Describe APIs exposed, external services consumed, databases, message queues, filesystem dependencies, or other codebase boundaries. Omit this section if none apply.}

## Tests and Verification
{Describe the test layout, meaningful test categories, and the focused commands used to verify changes.}

## Conventions and Patterns
{Describe patterns that would help an agent or developer make fitting changes in this codebase. Only include patterns confirmed from the code.}
`

var (
	excludedDirs = map[string]bool{
		".git": true, ".memr": true, ".agent_core": true, ".next": true, ".nuxt": true, ".pytest_cache": true, ".ruff_cache": true,
		".tox": true, ".venv": true, "__pycache__": true, "build": true, "coverage": true, "deps": true, "dist": true,
		"node_modules": true, "target": true, "vendor": true, "venv": true,
	}
	excludedSuffixes = []string{
		".md", ".lock", ".sum", ".7z", ".avif", ".bin", ".bmp", ".class", ".dll", ".dmg", ".eot", ".exe", ".gif", ".gz", ".ico",
		".jpeg", ".jpg", ".map", ".mp3", ".mp4", ".o", ".pdf", ".png", ".pyc", ".so", ".svg", ".tar", ".ttf", ".webp", ".woff", ".woff2", ".zip",
	}
	excludedNames = map[string]bool{
		".DS_Store": true, "package-lock.json": true, "pnpm-lock.yaml": true, "yarn.lock": true, "npm-shrinkwrap.json": true, "composer.lock": true,
	}
)

type Change struct {
	Path  string
	Lines int
}

type Drift struct {
	Missing  bool
	Editing  bool
	Baseline string
	Changes  []Change
	Lines    int
}

func (d Drift) Stale() bool {
	return !d.Missing && !d.Editing && (len(d.Changes) > staleFiles || d.Lines >= staleLines)
}

func Path(p project.Project) string { return filepath.Join(p.Root, RelPath) }

// Measure compares the code with the last commit that touched the structure doc.
func Measure(ctx context.Context, p project.Project) (Drift, error) {
	if _, err := os.Stat(Path(p)); os.IsNotExist(err) {
		return Drift{Missing: true}, nil
	}
	if status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", RelPath); status != "" {
		return Drift{Editing: true}, nil
	}
	baseline, err := git.Run(ctx, p.Root, "log", "-1", "--format=%H", "--", RelPath)
	if err != nil || baseline == "" {
		return Drift{Editing: true}, err
	}
	d, err := ChangesSince(ctx, p, baseline)
	d.Baseline = baseline
	return d, err
}

// ChangesSince counts code changes between base and the working tree, including new untracked files.
func ChangesSince(ctx context.Context, p project.Project, base string) (Drift, error) {
	var d Drift
	numstat, err := git.Run(ctx, p.Root, "diff", "--numstat", base)
	if err != nil {
		return d, err
	}
	for _, line := range strings.Split(numstat, "\n") {
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) != 3 || fields[0] == "-" {
			continue
		}
		added, _ := strconv.Atoi(fields[0])
		deleted, _ := strconv.Atoi(fields[1])
		d.add(p, renamedTo(fields[2]), added+deleted)
	}
	untracked, err := git.Run(ctx, p.Root, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return d, err
	}
	for _, file := range strings.Split(untracked, "\n") {
		if data, err := os.ReadFile(filepath.Join(p.Root, file)); err == nil && file != "" {
			d.add(p, file, strings.Count(string(data), "\n"))
		}
	}
	sort.Slice(d.Changes, func(i, j int) bool { return d.Changes[i].Lines > d.Changes[j].Lines })
	return d, nil
}

func (d *Drift) add(p project.Project, file string, lines int) {
	if lines == 0 || !Relevant(file, p.Config.Structure.Ignore) {
		return
	}
	d.Changes = append(d.Changes, Change{Path: file, Lines: lines})
	d.Lines += lines
}

// renamedTo resolves numstat rename notation such as "src/{old => new}/f.go".
func renamedTo(file string) string {
	if !strings.Contains(file, " => ") {
		return file
	}
	if open, close := strings.Index(file, "{"), strings.Index(file, "}"); open >= 0 && close > open {
		_, target, _ := strings.Cut(file[open+1:close], " => ")
		return path.Clean(file[:open] + target + file[close+1:])
	}
	_, target, _ := strings.Cut(file, " => ")
	return target
}

// Relevant reports whether a repository path counts as code for the structure doc.
func Relevant(file string, ignore []string) bool {
	if file == "AGENTS.md" || file == "CLAUDE.md" || excludedNames[path.Base(file)] || inExcludedDir(file) {
		return false
	}
	for _, suffix := range excludedSuffixes {
		if strings.HasSuffix(strings.ToLower(file), suffix) {
			return false
		}
	}
	for _, pattern := range ignore {
		if prefix, ok := strings.CutSuffix(pattern, "/**"); ok && strings.HasPrefix(file, prefix+"/") {
			return false
		}
		if matched, _ := path.Match(pattern, file); matched {
			return false
		}
		if matched, _ := path.Match(pattern, path.Base(file)); matched {
			return false
		}
	}
	return true
}

func inExcludedDir(file string) bool {
	for _, dir := range strings.Split(path.Dir(file), "/") {
		if excludedDirs[dir] {
			return true
		}
	}
	return false
}

// Tree renders the repository's tracked and unignored files as an indented tree.
func Tree(ctx context.Context, p project.Project) (string, error) {
	out, err := git.Run(ctx, p.Root, "ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return "", err
	}
	files := strings.Split(out, "\n")
	sort.Strings(files)
	var lines []string
	seen := map[string]bool{}
	for _, file := range files {
		if file == "" || inExcludedDir(file) || !strings.HasSuffix(file, ".md") && !Relevant(file, nil) {
			continue
		}
		parts := strings.Split(file, "/")
		for depth := range parts {
			key := strings.Join(parts[:depth+1], "/")
			if seen[key] {
				continue
			}
			seen[key] = true
			name := parts[depth]
			if depth < len(parts)-1 {
				name += "/"
			}
			lines = append(lines, strings.Repeat("  ", depth)+name)
		}
	}
	if len(lines) > maxTreeEntries {
		omitted := len(lines) - maxTreeEntries
		lines = append(lines[:maxTreeEntries], "... "+strconv.Itoa(omitted)+" more entries")
	}
	return strings.Join(lines, "\n"), nil
}
