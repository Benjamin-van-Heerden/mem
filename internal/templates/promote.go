package templates

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// Promote copies the project's item into a template of the library, commits
// and pushes the library, and records the item as in sync. template may be
// empty when the lock or the project's single template decides it. Promoting to
// a template the project does not use yet creates it if needed and adds it to
// the project.
func Promote(ctx context.Context, p *project.Project, lib Library, kind, name, template string) (Result, error) {
	var res Result
	if !slices.Contains(Kinds, kind) {
		return res, fmt.Errorf("unknown kind %q; use memory, skill or doc", kind)
	}
	s, err := load(p)
	if err != nil {
		return res, err
	}
	s.res = &res
	key := kind + ":" + name
	if template == "" {
		switch rec, ok := s.lock[key]; {
		case ok:
			template = rec.Template
		case len(p.Config.Templates.Use) == 1:
			template = p.Config.Templates.Use[0]
		default:
			return res, fmt.Errorf("%s %s does not come from a template and this project uses %d templates; choose one with --to <template>", kind, name, len(p.Config.Templates.Use))
		}
	}
	if _, exists, err := s.localHash(kind, name); err != nil {
		return res, err
	} else if !exists {
		return res, fmt.Errorf("this project has no %s named %s", kind, name)
	}

	netCtx, cancel := context.WithTimeout(ctx, networkTimeout)
	defer cancel()
	if _, err := git.Run(netCtx, lib.Dir, "pull", "--ff-only", "--quiet"); err != nil {
		return res, fmt.Errorf("could not update the template library before promoting: %w", err)
	}
	meta := filepath.Join(lib.Dir, template, "template.toml")
	if _, err := os.Stat(meta); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(meta), 0o755); err != nil {
			return res, err
		}
		if err := os.WriteFile(meta, []byte("description = \"\"\n"), 0o644); err != nil {
			return res, err
		}
		res.Lines = append(res.Lines, fmt.Sprintf("Created the template %s in the library. Add a one-line description to %s/template.toml in the library.", template, template))
	}
	dest := lib.ItemPath(template, kind, name)
	if err := s.export(kind, name, dest); err != nil {
		return res, err
	}
	if status, _ := git.Run(ctx, lib.Dir, "status", "--porcelain"); status == "" {
		res.Lines = append(res.Lines, fmt.Sprintf("%s already has this version of %s %s; nothing to push.", template, kind, name))
	} else {
		message := fmt.Sprintf("Promote %s %s from %s", kind, name, p.Config.Name)
		if _, err := git.Run(ctx, lib.Dir, "add", "--all"); err != nil {
			return res, err
		}
		name, _ := git.Run(ctx, p.Root, "config", "user.name")
		email, _ := git.Run(ctx, p.Root, "config", "user.email")
		if _, err := git.Run(ctx, lib.Dir, "-c", "user.name="+name, "-c", "user.email="+email, "commit", "--quiet", "-m", message); err != nil {
			return res, err
		}
		pushCtx, cancel := context.WithTimeout(ctx, git.PushTimeout)
		defer cancel()
		if _, err := git.Run(pushCtx, lib.Dir, "push", "--quiet"); err != nil {
			git.Run(ctx, lib.Dir, "reset", "--quiet", "--hard", "@{upstream}")
			return res, fmt.Errorf("could not push the template library, so nothing was promoted (%v); try again once %s is reachable", err, lib.Source)
		}
		res.Lines = append(res.Lines, fmt.Sprintf("Pushed %q to the template library.", message))
	}

	if !slices.Contains(p.Config.Templates.Use, template) {
		p.Config.Templates.Use = append(p.Config.Templates.Use, template)
		s.configChanged = true
		res.Lines = append(res.Lines, fmt.Sprintf("This project now uses %s.", template))
	}
	hash, err := hashItem(kind, dest)
	if err != nil {
		return res, err
	}
	s.record(Item{Kind: kind, Name: name, Template: template, Path: dest}, hash)
	return res, s.save()
}

// Reset replaces the project's copy of an item with its template's copy and
// records it as in sync. A reset item that was excluded is included again.
func Reset(p *project.Project, lib Library, kind, name string) (Result, error) {
	var res Result
	items, _, err := lib.Items(p.Config.Templates.Use)
	if err != nil {
		return res, err
	}
	key := kind + ":" + name
	i := slices.IndexFunc(items, func(item Item) bool { return item.Key() == key })
	if i < 0 {
		return res, fmt.Errorf("none of this project's templates (%v) provides %s %s", p.Config.Templates.Use, kind, name)
	}
	s, err := load(p)
	if err != nil {
		return res, err
	}
	s.res = &res
	if j := slices.Index(p.Config.Templates.Exclude, key); j >= 0 {
		p.Config.Templates.Exclude = slices.Delete(p.Config.Templates.Exclude, j, j+1)
		s.configChanged = true
	}
	if err := s.install(items[i]); err != nil {
		return res, err
	}
	hash, err := templateHash(items[i])
	if err != nil {
		return res, err
	}
	s.record(items[i], hash)
	res.Lines = append(res.Lines, fmt.Sprintf("Replaced %s %s with the copy from %s.", kind, name, items[i].Template))
	return res, s.save()
}

func load(p *project.Project) (*state, error) {
	lock, err := readLock(p.Root)
	if err != nil {
		return nil, err
	}
	agents, err := os.ReadFile(filepath.Join(p.Root, "AGENTS.md"))
	if err != nil {
		return nil, err
	}
	return &state{p: p, agents: string(agents), lock: lock, res: &Result{}}, nil
}

// export writes the project's copy of an item to dest in the library.
func (s *state) export(kind, name, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	switch kind {
	case Memory:
		memories, err := agentsmdMemories(s.agents)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, []byte(memories[name]+"\n"), 0o644)
	case Skill:
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		return copyTree(filepath.Join(s.p.Root, SkillPath(name)), dest)
	default:
		return copyFile(filepath.Join(s.p.Root, DocPath(name)), dest)
	}
}

func hashItem(kind, path string) (string, error) {
	return templateHash(Item{Kind: kind, Path: path})
}
