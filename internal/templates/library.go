// Package templates draws memories, skills and docs from a Git template library
// into a project, and promotes project items back into the library.
package templates

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/pelletier/go-toml/v2"
)

const networkTimeout = 20 * time.Second

const (
	Memory = "memory"
	Skill  = "skill"
	Doc    = "doc"
)

var Kinds = []string{Memory, Skill, Doc}

// Library is a local clone of a template library repository.
type Library struct {
	Source string
	Dir    string
}

type Template struct {
	Name        string
	Description string
}

// Item is one memory, skill or doc provided by a template.
type Item struct {
	Kind     string
	Name     string
	Template string
	Path     string
}

func (i Item) Key() string { return i.Kind + ":" + i.Name }

var unsafe = regexp.MustCompile(`[^A-Za-z0-9]+`)

// Open returns the cached clone of source, cloning it on first use. With pull,
// it fast-forwards the clone; a failed pull is returned as a warning so callers
// can continue with the cached copy.
func Open(ctx context.Context, source string, pull bool) (Library, string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return Library{}, "", err
	}
	lib := Library{Source: source, Dir: filepath.Join(cache, "mem", "templates", strings.Trim(unsafe.ReplaceAllString(source, "_"), "_"))}
	netCtx, cancel := context.WithTimeout(ctx, networkTimeout)
	defer cancel()
	if _, err := os.Stat(filepath.Join(lib.Dir, ".git")); err != nil {
		if err := os.MkdirAll(filepath.Dir(lib.Dir), 0o755); err != nil {
			return Library{}, "", err
		}
		if _, err := git.Run(netCtx, filepath.Dir(lib.Dir), "clone", "--quiet", source, lib.Dir); err != nil {
			os.RemoveAll(lib.Dir)
			return Library{}, "", fmt.Errorf("could not clone the template library %s: %w", source, err)
		}
		return lib, "", nil
	}
	if pull {
		if _, err := git.Run(netCtx, lib.Dir, "pull", "--ff-only", "--quiet"); err != nil {
			first, _, _ := strings.Cut(err.Error(), "\n")
			return lib, fmt.Sprintf("Could not update the template library from %s (%s); using the cached copy.", source, first), nil
		}
	}
	return lib, "", nil
}

// Templates lists the directories that contain a template.toml.
func (l Library) Templates() ([]Template, error) {
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil, err
	}
	var templates []Template
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(l.Dir, e.Name(), "template.toml"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var meta struct {
			Description string `toml:"description"`
		}
		if err := toml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("invalid %s/template.toml: %w", e.Name(), err)
		}
		templates = append(templates, Template{Name: e.Name(), Description: meta.Description})
	}
	return templates, nil
}

// Items lists the items of the named templates. When templates provide the same
// item, the later one wins; each such case is described in the returned notes.
func (l Library) Items(names []string) ([]Item, []string, error) {
	available, err := l.Templates()
	if err != nil {
		return nil, nil, err
	}
	known := map[string]bool{}
	for _, t := range available {
		known[t.Name] = true
	}
	byKey := map[string]Item{}
	var notes []string
	for _, name := range names {
		if !known[name] {
			return nil, nil, fmt.Errorf("the template library has no template %q; see `mem template list`", name)
		}
		items, err := l.templateItems(name)
		if err != nil {
			return nil, nil, err
		}
		for _, item := range items {
			if prev, ok := byKey[item.Key()]; ok {
				notes = append(notes, fmt.Sprintf("%s %s comes from %s, overriding %s.", item.Kind, item.Name, item.Template, prev.Template))
			}
			byKey[item.Key()] = item
		}
	}
	items := make([]Item, 0, len(byKey))
	for _, item := range byKey {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key() < items[j].Key() })
	return items, notes, nil
}

func (l Library) templateItems(name string) ([]Item, error) {
	var items []Item
	for _, kind := range Kinds {
		dir := filepath.Join(l.Dir, name, kindDir(kind))
		entries, err := os.ReadDir(dir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			itemName, ok := itemName(kind, e)
			if !ok {
				continue
			}
			items = append(items, Item{Kind: kind, Name: itemName, Template: name, Path: filepath.Join(dir, e.Name())})
		}
	}
	return items, nil
}

// ItemPath is where an item of a template lives in the library.
func (l Library) ItemPath(template, kind, name string) string {
	if kind == Skill {
		return filepath.Join(l.Dir, template, kindDir(kind), name)
	}
	return filepath.Join(l.Dir, template, kindDir(kind), name+".md")
}

func kindDir(kind string) string {
	switch kind {
	case Memory:
		return "memories"
	case Skill:
		return "skills"
	default:
		return "docs"
	}
}

func itemName(kind string, e os.DirEntry) (string, bool) {
	if strings.HasPrefix(e.Name(), ".") {
		return "", false
	}
	if kind == Skill {
		return e.Name(), e.IsDir()
	}
	name, ok := strings.CutSuffix(e.Name(), ".md")
	return name, ok && !e.IsDir()
}

// DefaultSource reads template_source from the user's mem config
// ($XDG_CONFIG_HOME/mem/config.toml, or ~/.config/mem/config.toml).
func DefaultSource() (string, error) {
	path, err := UserConfigPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var config struct {
		TemplateSource string `toml:"template_source"`
	}
	if err := toml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("invalid %s: %w", path, err)
	}
	return config.TemplateSource, nil
}

func UserConfigPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "mem", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "mem", "config.toml"), nil
}
