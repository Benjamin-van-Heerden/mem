package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// patches upgrade a project from the schema they are keyed by to the next one and return the paths they changed,
// relative to the repository root. Add one whenever Schema increases.
var patches = map[int]func(p Project) ([]string, error){
	1: dropTodoClaims,
}

// Upgrade applies pending patches and records the new schema. It returns the
// schemas that were upgraded from and the paths the patches changed.
func Upgrade(p Project) ([]int, []string, error) {
	var applied []int
	var paths []string
	for p.Config.Schema < Schema {
		patch, ok := patches[p.Config.Schema]
		if !ok {
			return applied, paths, fmt.Errorf("no upgrade path from project schema %d", p.Config.Schema)
		}
		changed, err := patch(p)
		if err != nil {
			return applied, paths, fmt.Errorf("upgrade from schema %d: %w", p.Config.Schema, err)
		}
		paths = append(paths, changed...)
		applied = append(applied, p.Config.Schema)
		p.Config.Schema++
		if err := WriteConfig(p.Root, p.Config); err != nil {
			return applied, paths, err
		}
	}
	return applied, paths, nil
}

// dropTodoClaims removes the status, claimed_by and claimed_at fields from every todo: schema 2 has no claims, a
// todo is open until it is deleted. Only the frontmatter is touched.
func dropTodoClaims(p Project) ([]string, error) {
	files, err := filepath.Glob(p.Path("todos", "*.md"))
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return changed, err
		}
		text := strings.ReplaceAll(string(data), "\r\n", "\n")
		if !strings.HasPrefix(text, "---\n") {
			continue
		}
		end := strings.Index(text[4:], "\n---\n")
		if end < 0 {
			continue
		}
		var kept []string
		for _, line := range strings.Split(text[4:4+end], "\n") {
			if strings.HasPrefix(line, "status:") || strings.HasPrefix(line, "claimed_by:") || strings.HasPrefix(line, "claimed_at:") {
				continue
			}
			kept = append(kept, line)
		}
		updated := "---\n" + strings.Join(kept, "\n") + text[4+end:]
		if updated == text {
			continue
		}
		if err := os.WriteFile(file, []byte(updated), 0o644); err != nil {
			return changed, err
		}
		changed = append(changed, p.Rel(file))
	}
	return changed, nil
}
