// Package runnables executes the commands in [runnables] of .mem/config.toml whose output augments onboard context.
package runnables

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/shell"
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

// Run executes every command in [runnables] from the repository root, in name order.
func Run(ctx context.Context, p project.Project) []Result {
	names := make([]string, 0, len(p.Config.Runnables))
	for name := range p.Config.Runnables {
		names = append(names, name)
	}
	sort.Strings(names)
	results := make([]Result, 0, len(names))
	for _, name := range names {
		results = append(results, run(ctx, p.Root, name, p.Config.Runnables[name]))
	}
	return results
}

func run(ctx context.Context, root, name, command string) Result {
	r := Result{Name: name}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := shell.Command(ctx, command)
	cmd.Dir = root
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
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
