// Package hooks installs mem's Git hooks: pre-push and pre-commit keep staging and production promotion-only (in a
// single-branch project, they keep the base branch free of commits), and post-commit reports checkpoint nudges. It implements the protection checks those hooks call.
package hooks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
)

const marker = "# Installed by mem."

type hook struct {
	name string
	// protection hooks exist only while protect is set; the others are installed in every project.
	protection bool
}

var installed = []hook{{"pre-push", true}, {"pre-commit", true}, {"post-commit", false}}

func script(h hook) string {
	if !h.protection {
		// mem's own commits and rebases carry git.InternalEnv; skipping them here saves starting mem. An older mem
		// without this subcommand would print its help on every commit, so the script checks for it first.
		return "#!/bin/sh\n" + marker + " Prints mem's checkpoint nudges after each commit.\n" +
			"[ -n \"$" + git.InternalEnv + "\" ] && exit 0\n" +
			"mem hook 2>/dev/null | grep -q \" " + h.name + "\" || exit 0\n" +
			"exec mem hook " + h.name + " \"$@\"\n"
	}
	return "#!/bin/sh\n" + marker + " Set protect = false in .mem/config.toml to remove.\nmem hook --help >/dev/null 2>&1 || exit 0\nexec mem hook " + h.name + " \"$@\"\n"
}

// Sync installs or removes mem's hooks according to the protect setting and
// describes anything it changed or could not do.
func Sync(ctx context.Context, p project.Project) ([]string, error) {
	dir, err := git.Run(ctx, p.Root, "rev-parse", "--git-path", "hooks")
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(p.Root, dir)
	}
	var lines []string
	for _, h := range installed {
		name := h.name
		path := filepath.Join(dir, name)
		existing, err := os.ReadFile(path)
		ours := err == nil && strings.Contains(string(existing), marker)
		wanted := !h.protection || p.Config.Git.Protect
		switch {
		case !wanted && ours:
			if err := os.Remove(path); err != nil {
				return lines, err
			}
			lines = append(lines, fmt.Sprintf("Removed the mem %s hook (protect = false).", name))
		case !wanted:
		case err == nil && !ours && h.protection:
			lines = append(lines, fmt.Sprintf("A %s hook that mem did not install already exists at %s. Tell the user; to %s, add this line to it: mem hook %s \"$@\" || exit 1", name, path, purpose(p), name))
		case err == nil && !ours:
			lines = append(lines, fmt.Sprintf("A %s hook that mem did not install already exists at %s. Tell the user; to get mem's checkpoint nudges after each commit, add this line to it: mem hook %s || true", name, path, name))
		case string(existing) != script(h):
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return lines, err
			}
			if err := os.WriteFile(path, []byte(script(h)), 0o755); err != nil {
				return lines, err
			}
			lines = append(lines, fmt.Sprintf("Installed the mem %s hook.", name))
		}
	}
	return lines, nil
}

// purpose says what the protection hooks are for in this project.
func purpose(p project.Project) string {
	if p.Config.Git.SingleBranch {
		return "keep the base branch free of commits"
	}
	return "keep staging and production promotion-only"
}

// PrePush rejects pushes to staging or production that do not come from mem promote, and in a single-branch
// project pushes to the base branch.
func PrePush(p project.Project, remote string, refs io.Reader) error {
	g := p.Config.Git
	if os.Getenv(release.PromoteEnv) == "1" || remote != g.Remote {
		return nil
	}
	scanner := bufio.NewScanner(refs)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 4 || !strings.HasPrefix(fields[2], "refs/heads/") {
			continue
		}
		branch := strings.TrimPrefix(fields[2], "refs/heads/")
		switch {
		case !g.Guarded(branch):
		case g.SingleBranch:
			return fmt.Errorf("%s is the base branch, which only takes in released code, so it is not pushed from here. Push your work to %s", branch, g.Development)
		default:
			return fmt.Errorf("%s only moves by promotion. Push your work to %s, then run `mem promote staging` or `mem promote production`", branch, g.Development)
		}
	}
	return scanner.Err()
}

// PreCommit rejects commits made on staging or production, or on a single-branch project's base branch.
func PreCommit(ctx context.Context, p project.Project) error {
	g := p.Config.Git
	branch := git.CurrentBranch(ctx, p.Root)
	switch {
	case !g.Guarded(branch):
		return nil
	case g.SingleBranch:
		return fmt.Errorf("%s is the base branch, which only takes in released code, so commits are not made on it. Switch to %s (`git switch %s`, which keeps your changes) and commit there", branch, g.Development, g.Development)
	}
	return fmt.Errorf("%s only moves by promotion, so commits are not made on it. Switch to %s (`git switch %s`, which keeps your changes) and commit there", branch, g.Development, g.Development)
}
