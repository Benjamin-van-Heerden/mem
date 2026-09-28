package templates

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/pelletier/go-toml/v2"
)

const LockPath = ".mem/templates.lock"

// LockItem records the content hash an installed item last had in common with its template.
type LockItem struct {
	Kind     string `toml:"kind"`
	Name     string `toml:"name"`
	Template string `toml:"template"`
	Hash     string `toml:"hash"`
}

func (l LockItem) Key() string { return l.Kind + ":" + l.Name }

type lockFile struct {
	Items []LockItem `toml:"item"`
}

// Result describes what a sync changed. Lines are for the user; lines that start
// with ⚠️ need a decision. Paths are the repository paths that changed.
type Result struct {
	Lines []string
	Paths []string
}

func (r *Result) changed(paths ...string) {
	for _, path := range paths {
		if !slices.Contains(r.Paths, path) {
			r.Paths = append(r.Paths, path)
		}
	}
}

// Sync reconciles the project with the items of the templates it uses: it
// installs missing items, updates items the project has not edited, records
// deletions as opt-outs and reports everything that needs a decision.
func Sync(p *project.Project, lib Library) (Result, error) {
	var res Result
	items, notes, err := lib.Items(p.Config.Templates.Use)
	if err != nil {
		return res, err
	}
	res.Lines = append(res.Lines, notes...)
	s, err := load(p)
	if err != nil {
		return res, err
	}
	s.res = &res

	provided := map[string]bool{}
	for _, item := range items {
		provided[item.Key()] = true
		if err := s.reconcile(item); err != nil {
			return res, err
		}
	}
	for key, rec := range s.lock {
		if provided[key] || s.excluded(key) {
			continue
		}
		delete(s.lock, key)
		s.lockChanged = true
		res.Lines = append(res.Lines, fmt.Sprintf("%s %s is no longer provided by the project's templates. The project copy stays as an ordinary %s.", rec.Kind, rec.Name, rec.Kind))
	}
	return res, s.save()
}

type state struct {
	p             *project.Project
	agents        string
	lock          map[string]LockItem
	res           *Result
	agentsChanged bool
	lockChanged   bool
	configChanged bool
}

func (s *state) excluded(key string) bool {
	return slices.Contains(s.p.Config.Templates.Exclude, key)
}

// Item states, as shown by `mem template list`.
const (
	StateExcluded   = "excluded"
	StateMissing    = "not installed"
	StateDeleted    = "deleted here"
	StateInSync     = "in sync"
	StateUpdated    = "template updated"
	StateLocalEdits = "local edits"
	StateBothEdited = "edited on both sides"
	StateDiffers    = "differs, not tracked"
)

type ItemStatus struct {
	Item  Item
	State string
}

// Status reports the state of each item the project's templates provide, without changing anything.
func Status(p *project.Project, lib Library) ([]ItemStatus, error) {
	items, _, err := lib.Items(p.Config.Templates.Use)
	if err != nil {
		return nil, err
	}
	s, err := load(p)
	if err != nil {
		return nil, err
	}
	var statuses []ItemStatus
	for _, item := range items {
		st, _, err := s.classify(item)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, ItemStatus{Item: item, State: st})
	}
	return statuses, nil
}

func (s *state) classify(item Item) (string, string, error) {
	key := item.Key()
	if s.excluded(key) {
		return StateExcluded, "", nil
	}
	rec, recorded := s.lock[key]
	template, err := templateHash(item)
	if err != nil {
		return "", "", err
	}
	local, exists, err := s.localHash(item.Kind, item.Name)
	if err != nil {
		return "", "", err
	}
	switch {
	case !exists && !recorded:
		return StateMissing, template, nil
	case !exists:
		return StateDeleted, template, nil
	case local == template:
		return StateInSync, template, nil
	case recorded && local == rec.Hash:
		return StateUpdated, template, nil
	case recorded && template == rec.Hash:
		return StateLocalEdits, template, nil
	case recorded:
		return StateBothEdited, template, nil
	default:
		return StateDiffers, template, nil
	}
}

func (s *state) reconcile(item Item) error {
	key := item.Key()
	st, template, err := s.classify(item)
	if err != nil {
		return err
	}
	rec, recorded := s.lock[key]
	promote := fmt.Sprintf("`mem template promote %s %s`", item.Kind, item.Name)
	reset := fmt.Sprintf("`mem template reset %s %s`", item.Kind, item.Name)
	switch st {
	case StateExcluded:
		if recorded {
			delete(s.lock, key)
			s.lockChanged = true
		}
	case StateMissing:
		if err := s.install(item); err != nil {
			s.res.Lines = append(s.res.Lines, fmt.Sprintf("⚠️ Could not add %s %s from %s: %v. Tell the user; the template copy needs fixing.", item.Kind, item.Name, item.Template, err))
			return nil
		}
		s.record(item, template)
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("Added %s %s from %s.", item.Kind, item.Name, item.Template))
	case StateDeleted:
		delete(s.lock, key)
		s.lockChanged = true
		s.p.Config.Templates.Exclude = append(s.p.Config.Templates.Exclude, key)
		s.configChanged = true
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("%s %s was deleted in this project, so it is now excluded from %s and will not come back. To restore it, remove %q from [templates] exclude in .mem/config.toml.", capitalize(item.Kind), item.Name, item.Template, key))
	case StateInSync:
		if !recorded || rec.Hash != template || rec.Template != item.Template {
			s.record(item, template)
		}
	case StateUpdated:
		if err := s.install(item); err != nil {
			s.res.Lines = append(s.res.Lines, fmt.Sprintf("⚠️ Could not update %s %s from %s: %v. Tell the user; the template copy needs fixing.", item.Kind, item.Name, item.Template, err))
			return nil
		}
		s.record(item, template)
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("Updated %s %s from %s.", item.Kind, item.Name, item.Template))
	case StateLocalEdits:
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("%s %s has local edits that %s does not have. If they would help similar projects, suggest promoting them to the user: %s.", capitalize(item.Kind), item.Name, item.Template, promote))
	case StateBothEdited:
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("⚠️ %s %s was edited both in this project and in %s, so it was left as it is. Tell the user, and ask which to keep: this project's version (%s, which replaces the template's) or the template's (%s, which replaces the local edits).", capitalize(item.Kind), item.Name, item.Template, promote, reset))
	case StateDiffers:
		s.res.Lines = append(s.res.Lines, fmt.Sprintf("⚠️ This project already has a %s named %s that differs from the one in %s, so it was left as it is. Tell the user, and ask which to keep: this project's version (%s) or the template's (%s).", item.Kind, item.Name, item.Template, promote, reset))
	}
	return nil
}

func (s *state) record(item Item, hash string) {
	s.lock[item.Key()] = LockItem{Kind: item.Kind, Name: item.Name, Template: item.Template, Hash: hash}
	s.lockChanged = true
}

// install writes the template's copy of item into the project.
func (s *state) install(item Item) error {
	switch item.Kind {
	case Memory:
		body, err := os.ReadFile(item.Path)
		if err != nil {
			return err
		}
		updated, err := agentsmd.SetMemory(s.agents, item.Name, string(body))
		if err != nil {
			return err
		}
		s.agents, s.agentsChanged = updated, true
		return nil
	case Skill:
		dest := filepath.Join(s.p.Root, SkillPath(item.Name))
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		if err := copyTree(item.Path, dest); err != nil {
			return err
		}
		s.res.changed(filepath.ToSlash(SkillPath(item.Name)))
		if line, path := linkSkill(s.p.Root, item.Name); line != "" {
			s.res.Lines = append(s.res.Lines, line)
		} else if path != "" {
			s.res.changed(path)
		}
		return nil
	default:
		dest := filepath.Join(s.p.Root, DocPath(item.Name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := copyFile(item.Path, dest); err != nil {
			return err
		}
		s.res.changed(filepath.ToSlash(DocPath(item.Name)))
		return nil
	}
}

func (s *state) localHash(kind, name string) (string, bool, error) {
	if kind == Memory {
		memories, err := agentsmdMemories(s.agents)
		if err != nil {
			return "", false, err
		}
		body, ok := memories[name]
		if !ok {
			return "", false, nil
		}
		return hashBytes([]byte(body)), true, nil
	}
	path := filepath.Join(s.p.Root, DocPath(name))
	if kind == Skill {
		path = filepath.Join(s.p.Root, SkillPath(name))
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	hash, err := hashPath(path)
	return hash, err == nil, err
}

func agentsmdMemories(text string) (map[string]string, error) {
	memories, err := agentsmd.Memories(text)
	if err != nil {
		return nil, err
	}
	bodies := map[string]string{}
	for _, m := range memories {
		bodies[m.Name] = m.Body
	}
	return bodies, nil
}

func (s *state) save() error {
	if s.agentsChanged {
		if err := os.WriteFile(filepath.Join(s.p.Root, "AGENTS.md"), []byte(s.agents), 0o644); err != nil {
			return err
		}
		s.res.changed("AGENTS.md")
	}
	if s.configChanged {
		if err := project.WriteConfig(s.p.Root, s.p.Config); err != nil {
			return err
		}
		s.res.changed(".mem/config.toml")
	}
	if s.lockChanged {
		if err := writeLock(s.p.Root, s.lock); err != nil {
			return err
		}
		s.res.changed(LockPath)
	}
	sort.Strings(s.res.Paths)
	return nil
}

func SkillPath(name string) string { return filepath.Join(".agents", "skills", name) }
func DocPath(name string) string   { return filepath.Join(".mem", "docs", name+".md") }

// linkSkill makes .claude/skills/<name> point at .agents/skills/<name> so Claude
// Code finds the skill. It returns a warning line when the link cannot be made,
// or the link's path when it was created.
func linkSkill(root, name string) (warning, created string) {
	link := filepath.Join(".claude", "skills", name)
	abs := filepath.Join(root, link)
	if _, err := os.Lstat(abs); err == nil {
		return "", ""
	}
	target := filepath.Join("..", "..", SkillPath(name))
	err := os.MkdirAll(filepath.Dir(abs), 0o755)
	if err == nil {
		err = os.Symlink(target, abs)
	}
	if err != nil {
		return fmt.Sprintf("⚠️ Could not link %s to %s (%v). Tell the user; Claude Code only finds the skill once that link exists, or once the skill is copied there.", filepath.ToSlash(link), filepath.ToSlash(target), err), ""
	}
	return "", filepath.ToSlash(link)
}

func templateHash(item Item) (string, error) {
	if item.Kind == Memory {
		body, err := os.ReadFile(item.Path)
		if err != nil {
			return "", err
		}
		return hashBytes([]byte(strings.TrimSpace(string(body)))), nil
	}
	return hashPath(item.Path)
}

// hashBytes ignores carriage returns, so checkouts with CRLF line endings
// (Git's autocrlf on Windows) hash the same as LF ones.
func hashBytes(data []byte) string {
	sum := sha256.Sum256(bytes.ReplaceAll(data, []byte("\r"), nil))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// hashPath hashes a file's content, or a directory's relative paths and file contents.
func hashPath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return hashBytes(data), nil
	}
	h := sha256.New()
	err = filepath.WalkDir(path, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == ".DS_Store" {
			return err
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(path, file)
		fmt.Fprintf(h, "%s\x00%s\x00", filepath.ToSlash(rel), hashBytes(data))
		return nil
	})
	if err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.Name() == ".DS_Store" {
			return err
		}
		rel, _ := filepath.Rel(from, file)
		dest := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		return copyFile(file, dest)
	})
}

func copyFile(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	info, err := os.Stat(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, data, info.Mode().Perm())
}

func readLock(root string) (map[string]LockItem, error) {
	lock := map[string]LockItem{}
	data, err := os.ReadFile(filepath.Join(root, LockPath))
	if errors.Is(err, os.ErrNotExist) {
		return lock, nil
	}
	if err != nil {
		return nil, err
	}
	var file lockFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", LockPath, err)
	}
	for _, item := range file.Items {
		lock[item.Key()] = item
	}
	return lock, nil
}

func writeLock(root string, lock map[string]LockItem) error {
	var file lockFile
	for _, item := range lock {
		file.Items = append(file.Items, item)
	}
	sort.Slice(file.Items, func(i, j int) bool { return file.Items[i].Key() < file.Items[j].Key() })
	data, err := toml.Marshal(file)
	if err != nil {
		return err
	}
	header := "# Written by mem: the content each template item last had in common with its template.\n\n"
	return os.WriteFile(filepath.Join(root, LockPath), append([]byte(header), data...), 0o644)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
