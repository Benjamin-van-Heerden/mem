package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

const PushTimeout = 30 * time.Second

// ErrPush reports that CommitPaths committed but could not push.
var ErrPush = errors.New("push failed")

// Commit commits only the given paths, leaving any other changes in the
// working tree and index untouched.
func Commit(ctx context.Context, root, message string, paths ...string) error {
	if _, err := Run(ctx, root, append([]string{"add", "--all", "--"}, paths...)...); err != nil {
		return err
	}
	_, err := Run(ctx, root, append([]string{"commit", "--quiet", "-m", message, "--"}, paths...)...)
	return err
}

// Push pushes the current branch to its upstream within PushTimeout. Failures wrap ErrPush.
func Push(ctx context.Context, root string) error {
	pushCtx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()
	if _, err := Run(pushCtx, root, "push", "--quiet"); err != nil {
		switch {
		case pushCtx.Err() != nil:
			err = fmt.Errorf("no response from the remote within %s", PushTimeout)
		case strings.Contains(err.Error(), "(fetch first)") || strings.Contains(err.Error(), "(non-fast-forward)"):
			err = errors.New("the remote has commits this checkout does not have yet")
		}
		return pushError{err}
	}
	return nil
}

// pushError matches ErrPush and reads as its cause alone.
type pushError struct{ cause error }

func (e pushError) Error() string        { return e.cause.Error() }
func (e pushError) Is(target error) bool { return target == ErrPush }

// CommitPaths commits only the given paths, then pushes when the branch has an
// upstream. A failed or timed-out push wraps ErrPush; the commit is kept.
func CommitPaths(ctx context.Context, root, message string, paths ...string) (pushed bool, err error) {
	if err := Commit(ctx, root, message, paths...); err != nil {
		return false, err
	}
	if _, err := Run(ctx, root, "rev-parse", "--abbrev-ref", "@{upstream}"); err != nil {
		return false, nil
	}
	if err := Push(ctx, root); err != nil {
		return false, err
	}
	return true, nil
}
