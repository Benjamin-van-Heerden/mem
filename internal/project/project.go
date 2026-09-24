package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/pelletier/go-toml/v2"
)

const Schema = 1

type Config struct {
	Schema      int             `toml:"schema"`
	Name        string          `toml:"name"`
	Description string          `toml:"description"`
	Git         GitConfig       `toml:"git"`
	Structure   StructureConfig `toml:"structure,omitempty"`
}

type GitConfig struct {
	Remote      string `toml:"remote"`
	Development string `toml:"development"`
	Staging     string `toml:"staging"`
	Production  string `toml:"production"`
	Protect     bool   `toml:"protect"`
}

// Stage returns the branch for a promotion stage and the branch it is promoted from.
func (g GitConfig) Stage(stage string) (branch, source string, ok bool) {
	switch stage {
	case "staging":
		return g.Staging, g.Development, true
	case "production":
		return g.Production, g.Staging, true
	}
	return "", "", false
}

type StructureConfig struct {
	Ignore []string `toml:"ignore,omitempty"`
}

type Project struct {
	Root   string
	Config Config
}

func (p Project) Path(parts ...string) string {
	return filepath.Join(append([]string{p.Root, ".memr"}, parts...)...)
}

func (p Project) Rel(path string) string {
	rel, err := filepath.Rel(p.Root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func ConfigPath(root string) string {
	return filepath.Join(root, ".memr", "config.toml")
}

func Load(ctx context.Context, dir string) (Project, error) {
	root, err := git.Toplevel(ctx, dir)
	if err != nil {
		return Project{}, err
	}
	data, err := os.ReadFile(ConfigPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return Project{}, fmt.Errorf("%s is not a memr project; run `memr init` to set it up", root)
	}
	if err != nil {
		return Project{}, err
	}
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return Project{}, fmt.Errorf("invalid .memr/config.toml: %w", err)
	}
	if config.Schema > Schema {
		return Project{}, fmt.Errorf("this project uses memr schema %d but the installed memr supports schema %d; update memr", config.Schema, Schema)
	}
	return Project{Root: root, Config: config}, nil
}

func WriteConfig(root string, config Config) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ConfigPath(root)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(root), data, 0o644)
}

// User identifies the current developer from Git's user.name.
func User(ctx context.Context, root string) (string, error) {
	name := Slugify(git.UserName(ctx, root))
	if name == "" {
		return "", errors.New("Git user.name is not set; run `git config --global user.name \"Your Name\"` so memr can attribute work")
	}
	return name, nil
}

var (
	separators = regexp.MustCompile(`[\s\-]+`)
	invalid    = regexp.MustCompile(`[^a-z0-9_]`)
	repeats    = regexp.MustCompile(`_+`)
)

func Slugify(text string) string {
	slug := separators.ReplaceAllString(strings.ToLower(text), "_")
	slug = invalid.ReplaceAllString(slug, "")
	return strings.Trim(repeats.ReplaceAllString(slug, "_"), "_")
}
