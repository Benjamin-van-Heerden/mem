package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, dir string, args ...string) (string, error) {
	return RunEnv(ctx, dir, nil, args...)
}

// RunEnv runs git with extra environment variables such as "NAME=value".
func RunEnv(ctx context.Context, dir string, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(append(os.Environ(), "GIT_TERMINAL_PROMPT=0"), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return strings.TrimSpace(stdout.String()), fmt.Errorf("git %s: %s", args[0], message)
	}
	return strings.TrimRight(stdout.String(), "\r\n"), nil
}

func Toplevel(ctx context.Context, dir string) (string, error) {
	root, err := Run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a Git repository", dir)
	}
	return root, nil
}

func CurrentBranch(ctx context.Context, root string) string {
	branch, _ := Run(ctx, root, "branch", "--show-current")
	return branch
}

func UserName(ctx context.Context, root string) string {
	name, _ := Run(ctx, root, "config", "user.name")
	return name
}

// CommitPaths commits only the given paths, leaving any other changes in the
// working tree and index untouched, then pushes when the branch has an upstream.
func CommitPaths(ctx context.Context, root, message string, paths ...string) (pushed bool, err error) {
	if _, err := Run(ctx, root, append([]string{"add", "--all", "--"}, paths...)...); err != nil {
		return false, err
	}
	if _, err := Run(ctx, root, append([]string{"commit", "--quiet", "-m", message, "--"}, paths...)...); err != nil {
		return false, err
	}
	if _, err := Run(ctx, root, "rev-parse", "--abbrev-ref", "@{upstream}"); err != nil {
		return false, nil
	}
	if _, err := Run(ctx, root, "push", "--quiet"); err != nil {
		return false, err
	}
	return true, nil
}
