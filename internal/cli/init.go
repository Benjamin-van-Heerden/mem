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
	"github.com/Benjamin-van-Heerden/mem/internal/claude"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/release"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/spf13/cobra"
)

const localIgnore = "/.mem/local/"

func (a *app) initCommand() *cobra.Command {
	config := project.Config{Schema: project.Schema}
	var templateNames []string
	var templateSource string
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
			var lib templates.Library
			var libWarning string
			if len(templateNames) > 0 {
				config.Templates.Source = librarySource(templateSource, "")
				if lib, libWarning, err = templates.Open(cmd.Context(), config.Templates.Source, true); err != nil {
					return err
				}
				config.Templates.Use = templateNames
				if _, _, err := lib.Items(templateNames); err != nil {
					return err
				}
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
			if _, err := ensureIgnored(root, localIgnore); err != nil {
				return err
			}
			if _, err := claude.SyncCompactHook(root, true); err != nil {
				return err
			}
			removed, err := removeClaudeLink(root)
			if err != nil {
				return err
			}
			p := project.Project{Root: root, Config: config}
			hookLines, err := hooks.Sync(cmd.Context(), p)
			if err != nil {
				return err
			}
			var synced templates.Result
			var setup string
			if len(templateNames) > 0 {
				if synced, err = templates.Sync(&p, lib); err != nil {
					return err
				}
				if setup, err = lib.Setup(templateNames); err != nil {
					return err
				}
			}
			if setup != "" {
				if err := os.WriteFile(filepath.Join(root, templates.SetupPath), []byte(setup), 0o644); err != nil {
					return err
				}
				synced.Paths = append(synced.Paths, templates.SetupPath)
			}

			out := cmd.OutOrStdout()
			output.Heading(out, "📦 MEM INITIALIZED")
			fmt.Fprintf(out, "Project: %s\n", config.Name)
			fmt.Fprintf(out, "Branches: %s → %s → %s (remote %s)\n", config.Git.Development, config.Git.Staging, config.Git.Production, config.Git.Remote)
			output.Section(out, "📄 FILES")
			fmt.Fprintln(out, ".mem/config.toml   project configuration")
			fmt.Fprintln(out, "AGENTS.md           mem instructions and project memories added; existing content kept")
			fmt.Fprintln(out, ".gitignore          ignores .mem/local/")
			fmt.Fprintln(out, claude.SettingsPath+"  Claude Code hook that runs `mem hook compact` after compaction")
			for _, line := range hookLines {
				fmt.Fprintln(out, line)
			}
			if removed {
				fmt.Fprintln(out, claudeLinkRemoved)
			}
			if settingsIgnored(cmd.Context(), root) {
				fmt.Fprintln(out, sharedSettingsWarning)
			}
			if setup != "" {
				fmt.Fprintln(out, templates.SetupPath+"      one-time setup from the template; onboard walks through it until the file is deleted")
			}
			if len(templateNames) > 0 {
				output.Section(out, "🧩 TEMPLATES: "+strings.Join(templateNames, ", "))
				fmt.Fprintf(out, "Library: %s\n", config.Templates.Source)
				renderTemplateSync(out, libWarning, synced)
			}
			output.Section(out, "🌿 BRANCHES")
			for _, line := range branchLines {
				fmt.Fprintln(out, line)
			}
			dev := config.Git.Development
			onboard := "3. Run `mem onboard` to build the project context."
			if setup != "" {
				onboard = "3. Run `mem onboard`: it builds the project context and presents the template's setup, which comes before any other work."
			}
			output.Instruction(out,
				"1. Read AGENTS.md now: it contains the working instructions for this project.",
				fmt.Sprintf("2. Show the user these files%s. Commit them on %s and push it (`git push -u %s %s`) so every clone shares the setup.", templateNote(synced), dev, config.Git.Remote, dev),
				onboard,
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
	cmd.Flags().StringArrayVar(&templateNames, "template", nil, "Template to draw memories, skills and docs from; repeat for several (later ones win on name clashes)")
	cmd.Flags().StringVar(&templateSource, "template-source", "", "Git URL of the template library (defaults to "+templates.DefaultSource+")")
	cmd.Flags().BoolVar(&config.Release.ProductionPR, "production-pr", false, "Release production through a pull request, which mem completes by fast-forward")
	cmd.Flags().BoolVar(&config.Git.Protect, "protect", true, "Install Git hooks that keep staging and production promotion-only (use --protect=false for solo projects)")
	return cmd
}

// ensureBranches creates missing branches in promotion order (staging from production, development from staging),
// tracking the remote's copy where one exists, publishes those the remote lacks, and switches to development.
func ensureBranches(ctx context.Context, root string, g project.GitConfig) ([]string, error) {
	if g.Development == g.Staging || g.Staging == g.Production || g.Development == g.Production {
		return nil, errors.New("the development, staging and production branches must have different names")
	}
	_, err := git.Run(ctx, root, "remote", "get-url", g.Remote)
	hasRemote := err == nil
	if hasRemote {
		if _, err := git.Run(ctx, root, "fetch", "--quiet", g.Remote); err != nil {
			return nil, fmt.Errorf("could not fetch %s: %w", g.Remote, err)
		}
	}
	var lines []string
	if !refExists(ctx, root, "HEAD") && !(hasRemote && refExists(ctx, root, "refs/remotes/"+g.Remote+"/"+g.Production)) {
		if err := firstCommit(ctx, root, g.Production); err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("Created an empty first commit on %s", g.Production))
	}
	order := []string{g.Production, g.Staging, g.Development}
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

// firstCommit gives a repository without commits, such as a fresh clone of an empty
// GitHub repository, an empty commit on the production branch. It is built from the
// empty tree so files the user already staged stay staged and out of the commit.
func firstCommit(ctx context.Context, root, production string) error {
	emptyIndex := filepath.Join(os.TempDir(), fmt.Sprintf("mem-empty-index-%d", os.Getpid()))
	defer os.Remove(emptyIndex)
	tree, err := git.RunEnv(ctx, root, []string{"GIT_INDEX_FILE=" + emptyIndex}, "write-tree")
	if err != nil {
		return err
	}
	sha, err := git.Run(ctx, root, "commit-tree", strings.TrimSpace(tree), "-m", "Initial commit")
	if err != nil {
		return fmt.Errorf("could not create the first commit (is git's user.name and user.email set?): %w", err)
	}
	if _, err := git.Run(ctx, root, "update-ref", "refs/heads/"+production, strings.TrimSpace(sha)); err != nil {
		return err
	}
	_, err = git.Run(ctx, root, "symbolic-ref", "HEAD", "refs/heads/"+production)
	return err
}

func refExists(ctx context.Context, root, ref string) bool {
	_, err := git.Run(ctx, root, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

func templateNote(res templates.Result) string {
	if len(res.Paths) == 0 {
		return ""
	}
	return ", including the template items (" + strings.Join(describePaths(res.Paths), ", ") + ")"
}

// ensureIgnored adds entry to .gitignore unless a line already matches it, and reports whether it did.
func ensureIgnored(root, entry string) (bool, error) {
	path := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return false, nil
		}
	}
	text := string(data)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text+entry+"\n"), 0o644)
}
