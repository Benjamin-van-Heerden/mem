// Package importer converts a project that uses the Python coding harness
// (.agent_core/) into a mem project. The original files are left in place.
package importer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
	"github.com/pelletier/go-toml/v2"
)

const Source = ".agent_core"

// placeholderDescription is the harness's default project description, which nobody filled in.
const placeholderDescription = "Add your project description here."

// retiredDocs are stock docs of earlier harness versions. A harness update replaced them with the general principles
// in its AGENTS.md block, which the mem block carries on, so importing them would duplicate older guidance.
var retiredDocs = map[string]bool{"coding_general.md": true, "coding_testing.md": true}

var heading = regexp.MustCompile(`^#{1,5} `)

type legacyConfig struct {
	Project struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
	} `toml:"project"`
	Files []struct {
		Path        string `toml:"path"`
		Description string `toml:"description"`
	} `toml:"files"`
	TreeDirs []struct {
		Path        string `toml:"path"`
		Description string `toml:"description"`
	} `toml:"tree_dirs"`
	Runnables []struct {
		Name        string `toml:"name"`
		Command     string `toml:"command"`
		Description string `toml:"description"`
	} `toml:"runnables"`
	Branches struct {
		Dev  string `toml:"dev"`
		Test string `toml:"test"`
		Main string `toml:"main"`
	} `toml:"branches"`
}

type Summary struct {
	Project   project.Project
	Memories  int
	Docs      int
	Structure bool
	Specs     int
	Archived  int
	Todos     int
	Logs      int
	Runnables []string
	Notes     []string
	// NextSteps is the "What Comes Next" section of the newest log, LatestLog. The import removes these sections
	// from every log, so what is still open there has to become todos.
	LatestLog, NextSteps string
}

// Import writes the mem configuration, AGENTS.md and work records converted from .agent_core/.
func Import(root, version string) (Summary, error) {
	var sum Summary
	old := filepath.Join(root, Source)
	if _, err := os.Stat(project.ConfigPath(root)); err == nil {
		return sum, errors.New("this repository already has .mem/config.toml; the import only runs once")
	}
	data, err := os.ReadFile(filepath.Join(old, "config.toml"))
	if err != nil {
		return sum, fmt.Errorf("no Python coding harness found: %w", err)
	}
	var legacy legacyConfig
	if err := toml.Unmarshal(data, &legacy); err != nil {
		return sum, fmt.Errorf("invalid %s/config.toml: %w", Source, err)
	}

	config := project.Config{
		Schema:      project.Schema,
		Name:        orDefault(legacy.Project.Name, filepath.Base(root)),
		Description: strings.TrimPrefix(strings.Join(strings.Fields(legacy.Project.Description), " "), placeholderDescription),
		Git: project.GitConfig{
			Remote:      "origin",
			Development: orDefault(legacy.Branches.Dev, "dev"),
			Staging:     orDefault(legacy.Branches.Test, "test"),
			Production:  orDefault(legacy.Branches.Main, "main"),
			Protect:     true,
		},
	}
	sum.Project = project.Project{Root: root, Config: config}
	p := sum.Project
	users := userMappings(filepath.Join(old, "user_mappings.toml"))

	agents, err := convertAgents(root, version, filepath.Join(old, "memories"), &sum)
	if err != nil {
		return sum, err
	}
	if err := convertDocs(p, filepath.Join(old, "docs"), &sum); err != nil {
		return sum, err
	}
	if err := convertSpecs(p, filepath.Join(old, "specs"), users, &sum); err != nil {
		return sum, err
	}
	if err := convertTodos(p, filepath.Join(old, "todos"), users, &sum); err != nil {
		return sum, err
	}
	if err := convertLogs(p, filepath.Join(old, "logs"), users, &sum); err != nil {
		return sum, err
	}
	if err := convertOnboardConfig(p, legacy, &sum); err != nil {
		return sum, err
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		return sum, err
	}
	return sum, project.WriteConfig(root, config)
}

func convertAgents(root, version, memoriesDir string, sum *Summary) (string, error) {
	path := filepath.Join(root, "AGENTS.md")
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("AGENTS.md is a symlink; make it a regular file first")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text, err := agentsmd.ReplaceLegacy(string(data), version)
	if err != nil {
		return "", err
	}
	if info, err := os.Lstat(filepath.Join(root, "CLAUDE.md")); err == nil && info.Mode()&os.ModeSymlink == 0 {
		sum.Notes = append(sum.Notes, "CLAUDE.md is a separate file. Claude Code reads it instead of AGENTS.md, so the mem instructions would be skipped. Tell the user; move anything still needed from CLAUDE.md into AGENTS.md and delete CLAUDE.md.")
	}
	files, _ := filepath.Glob(filepath.Join(memoriesDir, "*.md"))
	sort.Strings(files)
	for _, file := range files {
		var meta map[string]any
		body, err := work.ReadMarkdown(file, &meta)
		if err != nil {
			return "", err
		}
		name := strings.TrimSuffix(filepath.Base(file), ".md")
		body = regexp.MustCompile(`(?m)^## `).ReplaceAllString(body, "### ")
		if text, err = agentsmd.SetMemory(text, name, body); err != nil {
			return "", fmt.Errorf("memory %s: %w", name, err)
		}
		sum.Memories++
	}
	return text, nil
}

func convertDocs(p project.Project, dir string, sum *Summary) error {
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	for _, file := range files {
		if retiredDocs[filepath.Base(file)] {
			sum.Notes = append(sum.Notes, fmt.Sprintf("Left out %s: the general principles in the mem block replace it.", filepath.Base(file)))
			continue
		}
		target := p.Path("docs", filepath.Base(file))
		if filepath.Base(file) == "codebase_and_structure.md" {
			target = p.Path("structure.md")
			sum.Structure = true
		} else {
			sum.Docs++
		}
		if err := copyFile(file, target); err != nil {
			return err
		}
	}
	return nil
}

func convertSpecs(p project.Project, dir string, users map[string]string, sum *Summary) error {
	var specDirs []string
	for _, pattern := range []string{"*/spec.md", "completed/*/spec.md", "abandoned/*/spec.md"} {
		matches, _ := filepath.Glob(filepath.Join(dir, pattern))
		specDirs = append(specDirs, matches...)
	}
	for _, specFile := range specDirs {
		var meta map[string]any
		body, err := work.ReadMarkdown(specFile, &meta)
		if err != nil {
			return err
		}
		slug := filepath.Base(filepath.Dir(specFile))
		assigned := user(str(meta["assigned_to"]), users)
		spec := work.SpecMeta{
			Title:      str(meta["title"]),
			Status:     specStatus(str(meta["status"]), assigned),
			AssignedTo: assigned,
			Created:    timestamp(meta["created_at"]),
			Updated:    timestamp(meta["updated_at"]),
			Completed:  timestamp(meta["completed_at"]),
		}
		target := p.Path("specs", slug)
		if spec.Status == work.SpecCompleted || spec.Status == work.SpecAbandoned {
			target = p.Path("specs", "archive", slug)
			sum.Archived++
		} else {
			sum.Specs++
		}
		if err := copyDir(filepath.Dir(specFile), target); err != nil {
			return err
		}
		if err := work.WriteMarkdown(filepath.Join(target, "spec.md"), spec, withIssue(body, meta)); err != nil {
			return err
		}
		tasks, _ := filepath.Glob(filepath.Join(target, "tasks", "*.md"))
		for _, taskFile := range tasks {
			var meta map[string]any
			body, err := work.ReadMarkdown(taskFile, &meta)
			if err != nil {
				return err
			}
			task := work.TaskMeta{Title: str(meta["title"]), Status: work.TaskTodo, Created: timestamp(meta["created_at"]), Updated: timestamp(meta["updated_at"]), Completed: timestamp(meta["completed_at"])}
			if str(meta["status"]) == "completed" {
				task.Status = work.TaskCompleted
			}
			if err := work.WriteMarkdown(taskFile, task, body); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertTodos(p project.Project, dir string, users map[string]string, sum *Summary) error {
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	claimed, _ := filepath.Glob(filepath.Join(dir, "claimed", "*.md"))
	for _, file := range append(files, claimed...) {
		var meta map[string]any
		body, err := work.ReadMarkdown(file, &meta)
		if err != nil {
			return err
		}
		todo := work.TodoMeta{Title: str(meta["title"]), Status: work.TodoOpen, Created: timestamp(meta["created_at"])}
		if str(meta["status"]) == "claimed" {
			todo.Status = work.TodoClaimed
			todo.ClaimedBy = user(str(meta["claimed_by"]), users)
			todo.ClaimedAt = timestamp(meta["claimed_at"])
		}
		if err := work.WriteMarkdown(p.Path("todos", filepath.Base(file)), todo, withIssue(body, meta)); err != nil {
			return err
		}
		sum.Todos++
	}
	return nil
}

func convertLogs(p project.Project, dir string, users map[string]string, sum *Summary) error {
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	var latest time.Time
	for _, file := range files {
		var meta map[string]any
		body, err := work.ReadMarkdown(file, &meta)
		if err != nil {
			return err
		}
		username := str(meta["username"])
		log := work.LogMeta{Created: timestamp(meta["created_at"]), User: user(username, users), Spec: str(meta["spec_slug"])}
		name := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(file), ".md"), "_session") + ".md"
		if rest, ok := strings.CutPrefix(name, username+"_"); ok && username != "" {
			name = log.User + "_" + rest
		}
		body, next := withoutNextSteps(titledLog(body, log.Created))
		if err := work.WriteMarkdown(p.Path("logs", name), log, body); err != nil {
			return err
		}
		if created, err := time.Parse(time.RFC3339, log.Created); err == nil && created.After(latest) {
			latest, sum.LatestLog, sum.NextSteps = created, strings.TrimSuffix(name, ".md"), next
		}
		sum.Logs++
	}
	return nil
}

// withoutNextSteps takes the harness's "What Comes Next" section out of a log and returns it separately. mem logs
// hold facts only: a list of next steps is never updated, so it reads as pending work long after it is done.
func withoutNextSteps(body string) (rest, next string) {
	lines := strings.Split(body, "\n")
	start, end, fenced := -1, len(lines), false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
		}
		if fenced {
			continue
		}
		if start < 0 && strings.HasPrefix(line, "## What Comes Next") {
			start = i
		} else if start >= 0 && (strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ")) {
			end = i
			break
		}
	}
	if start < 0 {
		return body, ""
	}
	next = strings.TrimSpace(strings.Join(lines[start+1:end], "\n"))
	return strings.TrimSpace(strings.Join(append(lines[:start:start], lines[end:]...), "\n")), next
}

// withIssue keeps the link to the GitHub issue the harness mirrored a record to; mem does not mirror records.
func withIssue(body string, meta map[string]any) string {
	url := str(meta["issue_url"])
	if url == "" {
		return body
	}
	return strings.TrimSpace(body) + "\n\nGitHub issue: " + url
}

// titledLog gives an imported log the `# Work Log - <title>` heading mem reads its title from. Later harness
// versions wrote the title as a plain line; early ones had none and used top-level headings for sections, which
// move down a level under a title naming the session's date.
func titledLog(body, created string) string {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	switch {
	case strings.HasPrefix(lines[0], "# Work Log - "):
		return body
	case strings.HasPrefix(lines[0], "Work Log - "):
		lines[0] = "# " + lines[0]
		return strings.Join(lines, "\n")
	}
	var headings []int
	topLevel, fenced := false, false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
		}
		if !fenced && heading.MatchString(line) {
			headings = append(headings, i)
			topLevel = topLevel || strings.HasPrefix(line, "# ")
		}
	}
	if topLevel {
		for _, i := range headings {
			lines[i] = "#" + lines[i]
		}
	}
	date, _, _ := strings.Cut(created, "T")
	return "# Work Log - Session of " + date + "\n\n" + strings.Join(lines, "\n")
}

// convertOnboardConfig turns the old [[files]], [[tree_dirs]] and [[runnables]] into runnables.
func convertOnboardConfig(p project.Project, legacy legacyConfig, sum *Summary) error {
	var files, trees []string
	for _, f := range legacy.Files {
		if strings.HasPrefix(f.Path, Source+"/docs/") {
			continue
		}
		files = append(files, fmt.Sprintf("echo %s\ncat %s\necho", quote("## "+f.Path+": "+f.Description), quote(f.Path)))
	}
	for _, d := range legacy.TreeDirs {
		trees = append(trees, fmt.Sprintf("echo %s\ngit ls-files --cached --others --exclude-standard -- %s | sed 's|/[^/]*$||' | sort | uniq -c | awk '{print $2 \" (\" $1 \" files)\"}'\necho", quote("## "+d.Path+": "+d.Description), quote(d.Path)))
	}
	scripts := map[string][]string{"10_important_files": files, "20_directory_trees": trees}
	for i, r := range legacy.Runnables {
		scripts[fmt.Sprintf("%d_%s", 30+i, project.Slugify(r.Name))] = []string{fmt.Sprintf("# %s\n%s", r.Description, r.Command)}
	}
	for name, parts := range scripts {
		if len(parts) == 0 {
			continue
		}
		script := "#!/bin/sh\n# Imported from .agent_core/config.toml.\n" + strings.Join(parts, "\n") + "\n"
		path := p.Path("runnables", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			return err
		}
		sum.Runnables = append(sum.Runnables, name)
	}
	sort.Strings(sum.Runnables)
	return nil
}

func specStatus(old, assigned string) string {
	switch old {
	case "completed":
		return work.SpecCompleted
	case "abandoned":
		return work.SpecAbandoned
	case "merge_ready":
		return work.SpecActive
	}
	if assigned != "" {
		return work.SpecActive
	}
	return work.SpecDraft
}

// userMappings maps GitHub usernames to mem identities (slugified Git names). GitHub usernames are case-insensitive,
// and the harness wrote them lowercased into log usernames, so keys are lowercased.
func userMappings(path string) map[string]string {
	users := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return users
	}
	var mappings map[string]struct {
		Name string `toml:"name"`
	}
	if toml.Unmarshal(data, &mappings) == nil {
		for github, m := range mappings {
			users[strings.ToLower(github)] = project.Slugify(m.Name)
		}
	}
	return users
}

func user(name string, users map[string]string) string {
	if mapped, ok := users[strings.ToLower(name)]; ok {
		return mapped
	}
	return project.Slugify(name)
}

// timestamp converts the harness's naive local timestamps to RFC 3339.
func timestamp(v any) string {
	if t, ok := v.(time.Time); ok {
		return t.Format(time.RFC3339)
	}
	s := str(v)
	if s == "" {
		return ""
	}
	for _, layout := range []string{"2006-01-02T15:04:05.999999", "2006-01-02T15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return s
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func copyFile(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	return os.WriteFile(to, data, 0o644)
}

func copyDir(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		return copyFile(path, filepath.Join(to, rel))
	})
}
