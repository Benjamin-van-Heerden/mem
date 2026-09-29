package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/selfupdate"
	"github.com/spf13/cobra"
)

func updateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Replace this mem with the latest release if it is newer",
		Long:  "Onboard does this automatically. Set MEM_NO_UPDATE=1 to turn the automatic update off. Development builds are never replaced.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			output.Section(out, "⬆️ MEM UPDATE")
			if !selfupdate.IsRelease(buildinfo.Version) {
				fmt.Fprintf(out, "This is a development build (%s), which never replaces itself. Install a release to get updates.\n", buildinfo.Version)
				return nil
			}
			latest, err := selfupdate.Latest(cmd.Context())
			if err != nil {
				return fmt.Errorf("could not check for a newer mem: %w", err)
			}
			if !selfupdate.Newer(buildinfo.Version, latest) {
				fmt.Fprintf(out, "mem %s is the latest release.\n", buildinfo.Version)
				return nil
			}
			exe, err := selfupdate.Executable()
			if err != nil {
				return err
			}
			if err := selfupdate.Install(cmd.Context(), latest, exe); err != nil {
				return fmt.Errorf("could not install mem %s: %w", latest, err)
			}
			fmt.Fprintf(out, "Updated mem %s → %s at %s.\n", buildinfo.Version, latest, exe)
			return nil
		},
	}
}

// autoUpdate installs a newer release before onboard runs. It returns a line to
// report when an update is available but could not be installed, and whether
// the new executable must take over this invocation.
func autoUpdate(ctx context.Context) (string, bool) {
	if os.Getenv(selfupdate.SkipEnv) == "1" || !selfupdate.IsRelease(buildinfo.Version) {
		return "", false
	}
	latest, err := selfupdate.Latest(ctx)
	if err != nil || !selfupdate.Newer(buildinfo.Version, latest) {
		return "", false
	}
	exe, err := selfupdate.Executable()
	if err == nil {
		err = selfupdate.Install(ctx, latest, exe)
	}
	if err != nil {
		return fmt.Sprintf("⚠️ mem %s is available but could not be installed automatically (%v). Tell the user to run `mem update`.", latest, err), false
	}
	return fmt.Sprintf("⬆️ Updated mem %s → %s. Continuing with the new version.", buildinfo.Version, latest), true
}

// newerRelease names a newer mem release for mid-session checks, which leave
// replacing the executable to onboard and `mem update`.
func newerRelease(ctx context.Context) string {
	if os.Getenv(selfupdate.SkipEnv) == "1" || !selfupdate.IsRelease(buildinfo.Version) {
		return ""
	}
	latest, err := selfupdate.Latest(ctx)
	if err != nil || !selfupdate.Newer(buildinfo.Version, latest) {
		return ""
	}
	return fmt.Sprintf("mem %s is available (this is %s). Tell the user; `mem update` installs it, and the next onboard does so automatically.", latest, buildinfo.Version)
}

// rerun runs the updated executable with this invocation's arguments and exits with its status.
func rerun(out io.Writer) {
	exe, _ := selfupdate.Executable()
	child := exec.Command(exe, os.Args[1:]...)
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, out, os.Stderr
	child.Env = append(os.Environ(), selfupdate.SkipEnv+"=1")
	err := child.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		os.Exit(exit.ExitCode())
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
