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
	"github.com/Benjamin-van-Heerden/mem/internal/release"
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
			branchLines, err := ensureBranches(cmd.Context(), root, config.Git)
			if err != nil {
				return err
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
			output.Section(out, "🌿 BRANCHES")
			for _, line := range branchLines {
				fmt.Fprintln(out, line)
			}
			dev := config.Git.Development
			output.Instruction(out,
				"1. Read AGENTS.md now: it contains the working instructions for this project.",
				fmt.Sprintf("2. Show the user these files. Commit them on %s and push it (`git push -u %s %s`) so every clone shares the setup.", dev, config.Git.Remote, dev),
				"3. Run `mem onboard` to build the project context.",
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

// ensureBranches creates missing branches in promotion order (staging from production, development from staging),
// tracking the remote's copy where one exists, publishes those the remote lacks, and switches to development.
func ensureBranches(ctx context.Context, root string, g project.GitConfig) ([]string, error) {
	if g.Development == g.Staging || g.Staging == g.Production || g.Development == g.Production {
		return nil, errors.New("the development, staging and production branches must have different names")
	}
	if !refExists(ctx, root, "HEAD") {
		return nil, errors.New("this repository has no commits yet; make a first commit, then run `mem init`")
	}
	_, err := git.Run(ctx, root, "remote", "get-url", g.Remote)
	hasRemote := err == nil
	if hasRemote {
		if _, err := git.Run(ctx, root, "fetch", "--quiet", g.Remote); err != nil {
			return nil, fmt.Errorf("could not fetch %s: %w", g.Remote, err)
		}
	}
	order := []string{g.Production, g.Staging, g.Development}
	var lines []string
	for i, branch := range order {
		switch {
		case refExists(ctx, root, "refs/heads/"+branch):
			lines = append(lines, fmt.Sprintf("%s exists", branch))
		case hasRemote && refExists(ctx, root, "refs/remotes/"+g.Remote+"/"+branch):
			if _, err := git.Run(ctx, root, "branch", "--quiet", "--track", branch, g.Remote+"/"+branch); err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprintf("Created %s tracking %s/%s", branch, g.Remote, branch))
		case i == 0:
			return nil, fmt.Errorf("the production branch %s does not exist locally or on %s; pass --production with the branch that holds your releases (you are on %s)", branch, g.Remote, git.CurrentBranch(ctx, root))
		default:
			if _, err := git.Run(ctx, root, "branch", branch, order[i-1]); err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprintf("Created %s from %s", branch, order[i-1]))
		}
	}
	if hasRemote {
		for _, branch := range order {
			if refExists(ctx, root, "refs/remotes/"+g.Remote+"/"+branch) {
				continue
			}
			if _, err := git.RunEnv(ctx, root, []string{release.PromoteEnv + "=1"}, "push", "--quiet", "-u", g.Remote, branch+":"+branch); err != nil {
				return nil, fmt.Errorf("could not publish %s to %s: %w", branch, g.Remote, err)
			}
			lines = append(lines, fmt.Sprintf("Published %s to %s", branch, g.Remote))
		}
	} else {
		lines = append(lines, fmt.Sprintf("No remote %s: add it and push all three branches so every clone shares them.", g.Remote))
	}
	if git.CurrentBranch(ctx, root) != g.Development {
		if _, err := git.Run(ctx, root, "switch", "--quiet", g.Development); err != nil {
			lines = append(lines, fmt.Sprintf("⚠️ Could not switch to %s (%v). Commit or stash the conflicting changes, then run `git switch %s`.", g.Development, err, g.Development))
		} else {
			lines = append(lines, fmt.Sprintf("Switched to %s", g.Development))
		}
	}
	return lines, nil
}

func refExists(ctx context.Context, root, ref string) bool {
	_, err := git.Run(ctx, root, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
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
