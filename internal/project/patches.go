package project

import "fmt"

// patches upgrade a project from the schema they are keyed by to the next one.
// Add one whenever Schema increases.
var patches = map[int]func(p Project) error{}

// Upgrade applies pending patches and records the new schema. It returns the
// schemas that were upgraded from.
func Upgrade(p Project) ([]int, error) {
	var applied []int
	for p.Config.Schema < Schema {
		patch, ok := patches[p.Config.Schema]
		if !ok {
			return applied, fmt.Errorf("no upgrade path from project schema %d", p.Config.Schema)
		}
		if err := patch(p); err != nil {
			return applied, fmt.Errorf("upgrade from schema %d: %w", p.Config.Schema, err)
		}
		applied = append(applied, p.Config.Schema)
		p.Config.Schema++
		if err := WriteConfig(p.Root, p.Config); err != nil {
			return applied, err
		}
	}
	return applied, nil
}
