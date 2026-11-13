package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/pelletier/go-toml/v2"
)

const Schema = 4

// ModeFolder marks a project that is a plain folder, synced by cloud storage if at all, with no Git.
const ModeFolder = "folder"

// ErrFolder refuses a command that only means something with Git.
var ErrFolder = errors.New("this is a folder project without Git, so this command does not apply here")

type Config struct {
	Schema      int    `toml:"schema"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	// Mode is ModeFolder for a folder project and empty for a Git repository.
	Mode      string          `toml:"mode,omitempty"`
	Git       GitConfig       `toml:"git,omitempty"`
	Structure StructureConfig `toml:"structure,omitempty"`
	Templates TemplatesConfig `toml:"templates,omitempty"`
	Claude    ClaudeConfig    `toml:"claude,omitempty"`
	Cursor    CursorConfig    `toml:"cursor,omitempty"`
	Codex     CodexConfig     `toml:"codex,omitempty"`
	Release   ReleaseConfig   `toml:"release,omitempty"`
	Checks    ChecksConfig    `toml:"checks,omitempty"`
	// Runnables name the commands whose output onboard includes, run from the repository root in name order.
	Runnables map[string]string `toml:"runnables,omitempty"`
	// Precommit is schema 2's name for Checks. ParseConfig moves it there, so a project that has not been upgraded yet
	// still runs its checks, and the schema 3 patch writes the config without it.
	Precommit *ChecksConfig `toml:"precommit,omitempty"`
}

// ChecksConfig lists the project's checks, which `mem check` runs and promotion to staging requires.
type ChecksConfig struct {
	Commands []string `toml:"commands,omitempty"`
}

// ReleaseConfig sets how production releases are made.
type ReleaseConfig struct {
	// ProductionPR makes production releases go through a pull request, which mem completes by fast-forward.
	ProductionPR bool `toml:"production_pr,omitempty"`
	// Notes makes production releases wait for release notes the agent drafts and the user confirms. Without
	// them, a release is tagged with a generated summary of its specs and commits.
	Notes bool `toml:"notes,omitempty"`
	// VersionTags names each production release after the version given to promote, such as v1.4.0, for a project
	// whose CI publishes tags. Without it, a release gets a date tag.
	VersionTags bool `toml:"version_tags,omitempty"`
}

// ClaudeConfig controls what mem installs for Claude Code.
type ClaudeConfig struct {
	// CompactHook turns the post-compaction catch-up hook off when false; unset means on.
	CompactHook *bool `toml:"compact_hook,omitempty"`
	// PromptHook turns the work-log nudge at each user message off when false; unset means on.
	PromptHook *bool `toml:"prompt_hook,omitempty"`
}

func (c ClaudeConfig) CompactHookEnabled() bool { return c.CompactHook == nil || *c.CompactHook }

func (c ClaudeConfig) PromptHookEnabled() bool { return c.PromptHook == nil || *c.PromptHook }

// CursorConfig controls what mem installs for Cursor.
type CursorConfig struct {
	// StopHook turns the work-log follow-up at the end of a turn off when false; unset means on.
	StopHook *bool `toml:"stop_hook,omitempty"`
}

func (c CursorConfig) StopHookEnabled() bool { return c.StopHook == nil || *c.StopHook }

// CodexConfig controls what mem installs for Codex.
type CodexConfig struct {
	// CompactHook turns the post-compaction catch-up hook off when false; unset means on.
	CompactHook *bool `toml:"compact_hook,omitempty"`
	// PromptHook turns the work-log nudge at each user message off when false; unset means on.
	PromptHook *bool `toml:"prompt_hook,omitempty"`
}

func (c CodexConfig) CompactHookEnabled() bool { return c.CompactHook == nil || *c.CompactHook }

func (c CodexConfig) PromptHookEnabled() bool { return c.PromptHook == nil || *c.PromptHook }

type GitConfig struct {
	Remote      string `toml:"remote"`
	Development string `toml:"development"`
	// SingleBranch marks a project with no staging or production branches: mem makes no releases in it.
	SingleBranch bool   `toml:"single_branch,omitempty"`
	Staging      string `toml:"staging,omitempty"`
	Production   string `toml:"production,omitempty"`
	Protect      bool   `toml:"protect"`
	// Base names the branch a long-lived development branch, such as a partner's own main, is built on: it takes in
	// that branch's released code by merge commits only.
	Base string `toml:"base,omitempty"`
}

// Guarded reports whether branch is one mem never commits to or pushes: staging and production, or in a
// single-branch project the base branch.
func (g GitConfig) Guarded(branch string) bool {
	switch {
	case branch == "":
		return false
	case g.SingleBranch:
		return branch == g.Base
	}
	return branch == g.Staging || branch == g.Production
}

// Stage returns the branch for a promotion stage and the branch it is promoted from. A single-branch project has
// no stages.
func (g GitConfig) Stage(stage string) (branch, source string, ok bool) {
	if g.SingleBranch {
		return "", "", false
	}
	switch stage {
	case "staging":
		return g.Staging, g.Development, true
	case "production":
		return g.Production, g.Staging, true
	}
	return "", "", false
}

// TemplatesConfig names the template library and the templates this project draws items from.
type TemplatesConfig struct {
	Source  string   `toml:"source,omitempty"`
	Use     []string `toml:"use,omitempty"`
	Exclude []string `toml:"exclude,omitempty"`
}

// StructureConfig sets which files count as work, for the structure doc's drift and the work-log nudges.
type StructureConfig struct {
	Ignore  []string `toml:"ignore,omitempty"`
	Include []string `toml:"include,omitempty"`
}

type Project struct {
	Root   string
	Config Config
}

func (p Project) Path(parts ...string) string {
	return filepath.Join(append([]string{p.Root, ".mem"}, parts...)...)
}

func (p Project) Rel(path string) string {
	rel, err := filepath.Rel(p.Root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func ConfigPath(root string) string {
	return filepath.Join(root, ".mem", "config.toml")
}

// Load finds the project around dir. A folder project is found by its config alone, so that Git is never run for one;
// anything else is the Git repository around dir.
func Load(ctx context.Context, dir string) (Project, error) {
	if root, ok := nearestConfig(dir); ok {
		config, err := readConfig(root)
		if err != nil {
			return Project{}, err
		}
		if config.Folder() {
			return Project{Root: root, Config: config}, nil
		}
	}
	root, err := git.Toplevel(ctx, dir)
	if err != nil {
		return Project{}, err
	}
	config, err := readConfig(root)
	if errors.Is(err, os.ErrNotExist) {
		return Project{}, fmt.Errorf("%s is not a mem project; run `mem init` to set it up", root)
	}
	if err != nil {
		return Project{}, err
	}
	return Project{Root: root, Config: config}, nil
}

// nearestConfig walks up from dir to the closest directory with a .mem/config.toml.
func nearestConfig(dir string) (string, bool) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(ConfigPath(root)); err == nil {
			return root, true
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", false
		}
		root = parent
	}
}

func readConfig(root string) (Config, error) {
	data, err := os.ReadFile(ConfigPath(root))
	if err != nil {
		return Config{}, err
	}
	return ParseConfig(data)
}

// ParseConfig reads a .mem/config.toml, such as one taken from another branch, and rejects the same configs Load does.
func ParseConfig(data []byte) (Config, error) {
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("invalid .mem/config.toml: %w", err)
	}
	if config.Precommit != nil {
		if len(config.Checks.Commands) == 0 {
			config.Checks = *config.Precommit
		}
		config.Precommit = nil
	}
	if config.Schema > Schema {
		return Config{}, fmt.Errorf("this project uses mem schema %d but the installed mem supports schema %d; update mem", config.Schema, Schema)
	}
	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("invalid .mem/config.toml: %w", err)
	}
	return config, nil
}

// Folder reports whether this is a folder project.
func (c Config) Folder() bool { return c.Mode == ModeFolder }

// Folder reports whether this is a folder project.
func (p Project) Folder() bool { return p.Config.Folder() }

// validate rejects a config that mixes modes, such as a folder project with Git settings or a single-branch project
// that still names release settings, so a mode is never half applied.
func (c Config) validate() error {
	if c.Mode != "" && !c.Folder() {
		return fmt.Errorf("unknown mode %q; the only mode is %q, for a folder project", c.Mode, ModeFolder)
	}
	if c.Folder() {
		if c.Git != (GitConfig{}) || c.Release != (ReleaseConfig{}) || c.Templates.Source != "" || len(c.Templates.Use) > 0 || len(c.Templates.Exclude) > 0 {
			return errors.New("mode = \"folder\" has no Git, releases or templates, so [git], [release] and [templates] must not be set; remove them")
		}
		return nil
	}
	if c.Release.VersionTags && c.Release.ProductionPR {
		return errors.New("[release] version_tags and production_pr cannot be combined; remove one")
	}
	if !c.Git.SingleBranch {
		return nil
	}
	if c.Git.Staging != "" || c.Git.Production != "" {
		return errors.New("[git] single_branch = true, so staging and production must not be set; remove them")
	}
	if c.Release != (ReleaseConfig{}) {
		return errors.New("[git] single_branch = true, so mem makes no releases and [release] must not be set; remove it")
	}
	return nil
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

// User identifies the current developer from Git's user.name. A folder project has one person and no identity to
// tell apart, so its user is empty.
func User(ctx context.Context, p Project) (string, error) {
	if p.Folder() {
		return "", nil
	}
	name := Slugify(git.UserName(ctx, p.Root))
	if name == "" {
		return "", errors.New("Git user.name is not set; run `git config --global user.name \"Your Name\"` so mem can attribute work")
	}
	return name, nil
}

var (
	separators = regexp.MustCompile(`[\s\-/\\.:]+`)
	invalid    = regexp.MustCompile(`[^a-z0-9_]`)
	repeats    = regexp.MustCompile(`_+`)
)

func Slugify(text string) string {
	slug := separators.ReplaceAllString(strings.ToLower(text), "_")
	slug = invalid.ReplaceAllString(slug, "")
	return strings.Trim(repeats.ReplaceAllString(slug, "_"), "_")
}
