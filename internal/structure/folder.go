package structure

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

// folderFiles visits the files of a folder project that count as work. It only ever stats them: reading a file could
// make cloud storage download it. Everything counts except dot-files, dot-directories, the agent instructions and
// [structure] ignore patterns, since PDFs and Markdown are the work here and not by-products of it.
func folderFiles(p project.Project, visit func(rel string, info fs.FileInfo)) error {
	return filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// A folder that cannot be read is no reason to lose the rest.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == p.Root {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel := p.Rel(path)
		if d.IsDir() || rel == "AGENTS.md" || rel == "CLAUDE.md" || matchesAny(p.Config.Structure.Ignore, rel) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		visit(rel, info)
		return nil
	})
}

// ChangedSince lists the files of a folder project modified after since, most recent first. Without Git there is no
// measure of lines, so they count none.
func ChangedSince(p project.Project, since time.Time) ([]Change, error) {
	modified := map[string]time.Time{}
	err := folderFiles(p, func(rel string, info fs.FileInfo) {
		if info.ModTime().After(since) {
			modified[rel] = info.ModTime()
		}
	})
	if err != nil {
		return nil, err
	}
	changes := make([]Change, 0, len(modified))
	for file := range modified {
		changes = append(changes, Change{Path: file})
	}
	sort.Slice(changes, func(i, j int) bool {
		a, b := changes[i].Path, changes[j].Path
		if !modified[a].Equal(modified[b]) {
			return modified[a].After(modified[b])
		}
		return a < b
	})
	return changes, nil
}
