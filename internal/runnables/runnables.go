// Package runnables executes the scripts in .mem/runnables/ whose output augments onboard context.
package runnables

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
)

const (
	timeout   = 15 * time.Second
	maxOutput = 20000
)

type Result struct {
	Name   string
	Output string
	Error  string
}

// Run executes every runnable from the repository root, in name order.
func Run(ctx context.Context, p project.Project) ([]Result, error) {
	dir := p.Path("runnables")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var results []Result
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		results = append(results, run(ctx, p.Root, filepath.Join(dir, e.Name())))
	}
	return results, nil
}

func run(ctx context.Context, root, script string) Result {
	r := Result{Name: filepath.Base(script)}
	info, err := os.Stat(script)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	if info.Mode()&0o111 == 0 {
		r.Error = fmt.Sprintf("not executable; run `chmod +x %s` and give it a shebang line", filepath.ToSlash(filepath.Join(".mem", "runnables", r.Name)))
		return r
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, script)
	cmd.Dir = root
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	r.Output = strings.TrimSpace(stdout.String())
	if len(r.Output) > maxOutput {
		r.Output = r.Output[:maxOutput] + fmt.Sprintf("\n... output truncated at %d characters", maxOutput)
	}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		r.Error = fmt.Sprintf("timed out after %s", timeout)
	case err != nil:
		message, _, _ := strings.Cut(strings.TrimSpace(stderr.String()), "\n")
		r.Error = strings.TrimSpace(err.Error() + " " + message)
	}
	return r
}
