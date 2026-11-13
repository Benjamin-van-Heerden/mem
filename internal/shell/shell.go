// Package shell runs the command lines a project configures, such as its checks and runnables, through the
// platform's shell.
package shell

import (
	"context"
	"os/exec"
	"runtime"
)

// Command runs command through sh, or cmd on Windows, so a configured line can name any interpreter or script.
func Command(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
