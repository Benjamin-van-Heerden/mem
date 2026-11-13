package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// patches upgrade a project from the schema they are keyed by to the next one and return the paths they changed,
// relative to the repository root. Add one whenever Schema increases.
var patches = map[int]func(p *Project) ([]string, error){
	1: dropTodoClaims,
	2: moveRunnablesToConfig,
	3: prioritiseRunnablesReview,
}

// Upgrade applies pending patches and records the new schema, updating p's config to match. It returns the schemas
// that were upgraded from and the paths the patches changed.
func Upgrade(p *Project) ([]int, []string, error) {
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

// moveRunnablesToConfig upgrades to schema 3, where only [runnables] entries run at onboard. Every executable file in
// .mem/runnables/ becomes an entry that runs it as before, and a todo asks for the entries to be reviewed with the
// user, since a command can now name its interpreter or the project's own script. Schema 3 also renames [precommit]
// to [checks], which ParseConfig does as it reads and Upgrade's write of the config completes.
func moveRunnablesToConfig(p *Project) ([]string, error) {
	entries, err := os.ReadDir(p.Path("runnables"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var listed, skipped []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		// Windows keeps no executable bit, and the upgrade is shared, so a Windows checkout lists every file.
		if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			skipped = append(skipped, e.Name())
			continue
		}
		if _, ok := p.Config.Runnables[e.Name()]; ok {
			continue
		}
		if p.Config.Runnables == nil {
			p.Config.Runnables = map[string]string{}
		}
		p.Config.Runnables[e.Name()] = ".mem/runnables/" + e.Name()
		listed = append(listed, e.Name())
	}
	if len(listed) == 0 && len(skipped) == 0 {
		return nil, nil
	}
	todo := p.Path("todos", runnablesTodoFile)
	if _, err := os.Stat(todo); err == nil {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(todo), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(todo, []byte(runnablesTodo(listed, skipped, time.Now())), 0o644); err != nil {
		return nil, err
	}
	return []string{p.Rel(todo)}, nil
}

func runnablesTodo(listed, skipped []string, now time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: Review the runnables moved into the config\npriority: critical\ncreated_at: %q\n---\n\n", now.Format(time.RFC3339))
	b.WriteString("mem now runs only the commands listed under `[runnables]` in `.mem/config.toml`: each entry is a name and a command, run through the shell from the repository root at onboard. A runnable can name its interpreter or reuse a project script, such as `python3 .mem/runnables/models.py` or `bun run db:summary`, and then needs no executable bit or shebang. Files in `.mem/runnables/` no longer run on their own. `.mem/README.md` describes runnables in full.\n\n")
	if len(listed) > 0 {
		fmt.Fprintf(&b, "The upgrade listed each executable file in `.mem/runnables/` as an entry that runs it as before: %s.\n\n", codeList(listed))
	}
	if len(skipped) > 0 {
		fmt.Fprintf(&b, "These files were not executable, so they never ran and were not listed: %s. Add an entry for any that should run, or delete them.\n\n", codeList(skipped))
	}
	b.WriteString("Go through the entries with the user: name the interpreter in the command where that is clearer than a shebang, point an entry at the project's own script where one does the same job, and remove entries (and their files) that are no longer needed. Delete this todo when done.\n")
	return b.String()
}

const runnablesTodoFile = "review_runnables_config.md"

// prioritiseRunnablesReview upgrades to schema 4, where todos have priorities: the runnables review that schema 3's
// upgrade recorded is critical, since onboard's context depends on it.
func prioritiseRunnablesReview(p *Project) ([]string, error) {
	path := p.Path("todos", runnablesTodoFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") || strings.Contains(text, "\npriority:") {
		return nil, nil
	}
	title := strings.Index(text, "\ntitle:")
	if title < 0 {
		return nil, nil
	}
	end := title + 1 + strings.Index(text[title+1:], "\n")
	updated := text[:end] + "\npriority: critical" + text[end:]
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return nil, err
	}
	return []string{p.Rel(path)}, nil
}

func codeList(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = "`" + n + "`"
	}
	return strings.Join(quoted, ", ")
}

// dropTodoClaims removes the status, claimed_by and claimed_at fields from every todo: schema 2 has no claims, a
// todo is open until it is deleted. Only the frontmatter is touched.
func dropTodoClaims(p *Project) ([]string, error) {
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
