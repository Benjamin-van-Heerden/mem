package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// knowledge is what an agent picks up at session start: project memories from
// AGENTS.md and skill files. Onboard compares it before and after syncing, so
// that what arrives mid-session is shown instead of waiting for the next session.
type knowledge struct {
	memories map[string]string
	skills   map[string]skill
}

type skill struct {
	path        string
	description string
	content     string
}

var skillDirs = []string{filepath.Join(".agents", "skills"), filepath.Join(".claude", "skills")}

func readKnowledge(p project.Project) knowledge {
	k := knowledge{memories: map[string]string{}, skills: map[string]skill{}}
	if text, err := readAgents(p); err == nil {
		memories, _ := agentsmd.Memories(text)
		for _, m := range memories {
			k.memories[m.Name] = m.Body
		}
	}
	for _, dir := range skillDirs {
		paths, _ := filepath.Glob(filepath.Join(p.Root, dir, "*", "SKILL.md"))
		for _, path := range paths {
			name := filepath.Base(filepath.Dir(path))
			if _, seen := k.skills[name]; seen {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var meta struct {
				Description string `yaml:"description"`
			}
			work.ReadMarkdown(path, &meta)
			k.skills[name] = skill{path: filepath.ToSlash(filepath.Join(dir, name, "SKILL.md")), description: meta.Description, content: string(data)}
		}
	}
	return k
}

// renderKnowledgeChanges prints the memories and skills that changed during
// onboard and reports whether there were any.
func renderKnowledgeChanges(out io.Writer, before, after knowledge) bool {
	memories := changedNames(before.memories, after.memories, func(a, b string) bool { return a == b })
	skills := changedNames(before.skills, after.skills, func(a, b skill) bool { return a.content == b.content })
	if len(memories) > 0 {
		output.Section(out, "🧠 CHANGED MEMORIES")
		for i, name := range memories {
			if i > 0 {
				fmt.Fprintln(out)
			}
			body, ok := after.memories[name]
			switch {
			case !ok:
				fmt.Fprintf(out, "Removed: %s\n", name)
			case !hasKey(before.memories, name):
				fmt.Fprintf(out, "New: %s\n%s\n", name, strings.TrimSpace(body))
			default:
				fmt.Fprintf(out, "Updated: %s\n%s\n", name, strings.TrimSpace(body))
			}
		}
	}
	if len(skills) > 0 {
		output.Section(out, "🛠️ CHANGED SKILLS")
		for _, name := range skills {
			s, ok := after.skills[name]
			label := "Updated"
			switch {
			case !ok:
				fmt.Fprintf(out, "Removed: %s\n", name)
				continue
			case !hasKey(before.skills, name):
				label = "New"
			}
			fmt.Fprintf(out, "%s: %s (%s)\n", label, name, s.path)
			if s.description != "" {
				fmt.Fprintf(out, "  %s\n", s.description)
			}
		}
	}
	return len(memories)+len(skills) > 0
}

func changedNames[T any](before, after map[string]T, equal func(a, b T) bool) []string {
	var names []string
	for name, a := range after {
		if b, ok := before[name]; !ok || !equal(a, b) {
			names = append(names, name)
		}
	}
	for name := range before {
		if !hasKey(after, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func hasKey[T any](m map[string]T, key string) bool {
	_, ok := m[key]
	return ok
}
