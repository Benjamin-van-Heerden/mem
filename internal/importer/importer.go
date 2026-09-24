// Package importer converts a project that uses the Python coding harness
// (.agent_core/) into a memr project. The original files are left in place.
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

	"github.com/Benjamin-van-Heerden/memr/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
	"github.com/Benjamin-van-Heerden/memr/internal/work"
	"github.com/pelletier/go-toml/v2"
)

const Source = ".agent_core"

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
}

// Import writes the memr configuration, AGENTS.md and work records converted from .agent_core/.
func Import(root, version string) (Summary, error) {
	var sum Summary
	old := filepath.Join(root, Source)
	if _, err := os.Stat(project.ConfigPath(root)); err == nil {
		return sum, errors.New("this repository already has .memr/config.toml; the import only runs once")
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
		Description: strings.Join(strings.Fields(legacy.Project.Description), " "),
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
	if err := convertLogs(p, filepath.Join(old, "logs"), &sum); err != nil {
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
		sum.Notes = append(sum.Notes, "CLAUDE.md is a separate file. Claude Code reads it instead of AGENTS.md, so the memr instructions would be skipped. Tell the user; move anything still needed from CLAUDE.md into AGENTS.md and delete CLAUDE.md.")
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
	for _, pattern := range []string{"*/spec.md", "completed/*/spec.md"} {
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
		if err := work.WriteMarkdown(filepath.Join(target, "spec.md"), spec, body); err != nil {
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
	for _, file := range files {
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
		if err := work.WriteMarkdown(p.Path("todos", filepath.Base(file)), todo, body); err != nil {
			return err
		}
		sum.Todos++
	}
	return nil
}

func convertLogs(p project.Project, dir string, sum *Summary) error {
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	for _, file := range files {
		var meta map[string]any
		body, err := work.ReadMarkdown(file, &meta)
		if err != nil {
			return err
		}
		log := work.LogMeta{Created: timestamp(meta["created_at"]), User: str(meta["username"]), Spec: str(meta["spec_slug"])}
		name := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(file), ".md"), "_session") + ".md"
		if err := work.WriteMarkdown(p.Path("logs", name), log, body); err != nil {
			return err
		}
		sum.Logs++
	}
	return nil
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

// userMappings maps GitHub usernames to memr identities (slugified Git names).
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
			users[github] = project.Slugify(m.Name)
		}
	}
	return users
}

func user(name string, users map[string]string) string {
	if mapped, ok := users[name]; ok {
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
