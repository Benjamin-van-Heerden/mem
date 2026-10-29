package templates

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

// LockChange records a change note the project has seen, so it becomes a todo only once.
type LockChange struct {
	Template string `toml:"template"`
	Name     string `toml:"name"`
}

func (l LockChange) Key() string { return l.Template + "/" + l.Name }

type lockFile struct {
	Items   []LockItem   `toml:"item"`
	Changes []LockChange `toml:"change,omitempty"`
}

func readLock(root string) (map[string]LockItem, map[string]LockChange, error) {
	lock, seen := map[string]LockItem{}, map[string]LockChange{}
	data, err := os.ReadFile(filepath.Join(root, LockPath))
	if errors.Is(err, os.ErrNotExist) {
		return lock, seen, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var file lockFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return nil, nil, fmt.Errorf("invalid %s: %w", LockPath, err)
	}
	for _, item := range file.Items {
		lock[item.Key()] = item
	}
	for _, change := range file.Changes {
		seen[change.Key()] = change
	}
	return lock, seen, nil
}

func writeLock(root string, lock map[string]LockItem, seen map[string]LockChange) error {
	var file lockFile
	for _, item := range lock {
		file.Items = append(file.Items, item)
	}
	sort.Slice(file.Items, func(i, j int) bool { return file.Items[i].Key() < file.Items[j].Key() })
	for _, change := range seen {
		file.Changes = append(file.Changes, change)
	}
	sort.Slice(file.Changes, func(i, j int) bool { return file.Changes[i].Key() < file.Changes[j].Key() })
	data, err := toml.Marshal(file)
	if err != nil {
		return err
	}
	header := "# Written by mem: the content each template item last had in common with its template, and the change notes the project has seen.\n\n"
	return os.WriteFile(filepath.Join(root, LockPath), append([]byte(header), data...), 0o644)
}
