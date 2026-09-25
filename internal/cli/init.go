package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

const localIgnore = "/.mem/local/"

func (a *app) initCommand() *cobra.Command {
	config := project.Config{Schema: project.Schema}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up mem in the current Git repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Toplevel(cmd.Context(), a.dir)
			if err != nil {
				return err
			}
			if _, err := os.Stat(project.ConfigPath(root)); err == nil {
				return errors.New("this repository already has .mem/config.toml")
			}
			if config.Name == "" {
				config.Name = filepath.Base(root)
			}
			agentsPath := filepath.Join(root, "AGENTS.md")
			existing, err := os.ReadFile(agentsPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			agents, err := agentsmd.Install(string(existing), buildinfo.Version)
			if err != nil {
				return err
			}
			if err := project.WriteConfig(root, config); err != nil {
				return err
			}
			if err := os.WriteFile(agentsPath, []byte(agents), 0o644); err != nil {
				return err
			}
			if err := ensureIgnored(root, localIgnore); err != nil {
				return err
			}
			p := project.Project{Root: root, Config: config}
			hookLines, err := hooks.Sync(cmd.Context(), p)
			if err != nil {
				return err
			}
			branchLine := ensureDevelopmentBranch(cmd.Context(), p)

			out := cmd.OutOrStdout()
			output.Heading(out, "📦 MEM INITIALIZED")
			fmt.Fprintf(out, "Project: %s\n", config.Name)
			fmt.Fprintf(out, "Branches: %s → %s → %s (remote %s)\n", config.Git.Development, config.Git.Staging, config.Git.Production, config.Git.Remote)
			output.Section(out, "📄 FILES")
			fmt.Fprintln(out, ".mem/config.toml   project configuration")
			fmt.Fprintln(out, "AGENTS.md           mem instructions and project memories added; existing content kept")
			fmt.Fprintln(out, ".gitignore          ignores .mem/local/")
			for _, line := range hookLines {
				fmt.Fprintln(out, line)
			}
			if branchLine != "" {
				fmt.Fprintln(out, branchLine)
			}
			dev := config.Git.Development
			output.Instruction(out,
				"1. Read AGENTS.md now: it contains the working instructions for this project.",
				fmt.Sprintf("2. Show the user these files. Commit them on %s and push it (`git push -u %s %s`) so every clone shares the setup.", dev, config.Git.Remote, dev),
				fmt.Sprintf("3. Run `mem onboard` to build the project context. %s and %s are created on %s by their first `mem promote`.", config.Git.Staging, config.Git.Production, config.Git.Remote),
			)
			return nil
		},
	}
	cmd.Flags().StringVar(&config.Name, "name", "", "Project name (defaults to the repository directory name)")
	cmd.Flags().StringVar(&config.Description, "description", "", "One-line project description")
	cmd.Flags().StringVar(&config.Git.Remote, "remote", "origin", "Shared Git remote")
	cmd.Flags().StringVar(&config.Git.Development, "development", "dev", "Development branch, where day-to-day work happens")
	cmd.Flags().StringVar(&config.Git.Staging, "staging", "test", "Staging branch, deployed as preview releases")
	cmd.Flags().StringVar(&config.Git.Production, "production", "main", "Production branch")
	cmd.Flags().BoolVar(&config.Git.Protect, "protect", true, "Install Git hooks that keep staging and production promotion-only (use --protect=false for solo projects)")
	return cmd
}

// ensureDevelopmentBranch creates the development branch at HEAD when it does not exist yet.
func ensureDevelopmentBranch(ctx context.Context, p project.Project) string {
	dev := p.Config.Git.Development
	if git.CurrentBranch(ctx, p.Root) == dev {
		return ""
	}
	if _, err := git.Run(ctx, p.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+dev); err == nil {
		return fmt.Sprintf("You are not on the development branch %s. Switch with `git switch %s`; uncommitted changes come along.", dev, dev)
	}
	if _, err := git.Run(ctx, p.Root, "branch", dev); err != nil {
		return fmt.Sprintf("The development branch %s does not exist yet. Create it once the repository has a commit: `git switch -c %s`.", dev, dev)
	}
	return fmt.Sprintf("Created the development branch %s at HEAD. Switch to it with `git switch %s`; uncommitted changes come along.", dev, dev)
}

func ensureIgnored(root, entry string) error {
	path := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	text := string(data)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return os.WriteFile(path, []byte(text+entry+"\n"), 0o644)
}
