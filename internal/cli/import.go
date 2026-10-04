package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/claude"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/importer"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/spf13/cobra"
)

func (a *app) importCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "import", Short: "Convert a project from another harness"}
	cmd.AddCommand(&cobra.Command{
		Use:   "agent-core",
		Short: "Convert a project that uses the Python coding harness (.agent_core/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Toplevel(cmd.Context(), a.dir)
			if err != nil {
				return err
			}
			sum, err := importer.Import(root, buildinfo.Version)
			if err != nil {
				return err
			}
			if _, err := ensureIgnored(root, localIgnore); err != nil {
				return err
			}
			if _, err := claude.SyncCompactHook(root, true); err != nil {
				return err
			}
			hookLines, err := hooks.Sync(cmd.Context(), sum.Project)
			if err != nil {
				return err
			}
			g := sum.Project.Config.Git
			out := cmd.OutOrStdout()
			output.Heading(out, "📦 IMPORTED FROM .agent_core")
			fmt.Fprintf(out, "Project: %s\nBranches: %s → %s → %s (remote %s)\n", sum.Project.Config.Name, g.Development, g.Staging, g.Production, g.Remote)
			output.Section(out, "📄 CONVERTED")
			fmt.Fprintf(out, "AGENTS.md: old harness block replaced by the mem block, %d memories added, other content kept\n", sum.Memories)
			fmt.Fprintf(out, "Docs: %d in .mem/docs/\n", sum.Docs)
			if sum.Structure {
				fmt.Fprintln(out, "Structure doc: codebase_and_structure.md moved to .mem/structure.md")
			}
			fmt.Fprintf(out, "Specs: %d open, %d archived\nTodos: %d\nWork logs: %d\n", sum.Specs, sum.Archived, sum.Todos, sum.Logs)
			fmt.Fprintln(out, "Claude Code: "+claude.SettingsPath+" runs `mem hook compact` after compaction")
			if len(sum.Runnables) > 0 {
				fmt.Fprintf(out, "Runnables: %s in .mem/runnables/, from the old files, tree_dirs and runnables settings\n", strings.Join(sum.Runnables, ", "))
			}
			if removed, err := removeClaudeLink(root); err != nil {
				return err
			} else if removed {
				sum.Notes = append(sum.Notes, claudeLinkRemoved)
			}
			if settingsIgnored(cmd.Context(), root) {
				sum.Notes = append(sum.Notes, sharedSettingsWarning)
			}
			for _, line := range append(hookLines, sum.Notes...) {
				fmt.Fprintln(out, line)
			}
			output.Instruction(out,
				"Nothing was committed, and .agent_core/ is still in place.",
				"1. Show the user the result: `git status`, AGENTS.md and .mem/.",
				"2. Once the user is happy, remove the old harness: `git rm -r -q .agent_core`, then `rm -rf .agent_core` for the ignored files it leaves behind. Drop its entries from .gitignore.",
				"3. Find what still refers to the old harness (`git grep -n -e agent_core -e harness/main.py -- ':!.mem'`), such as setup docs, scripts and CI, and update it with the user to use mem.",
				fmt.Sprintf("4. Commit everything on %s and push it.", g.Development),
				"5. Run `mem onboard`.",
			)
			return nil
		},
	})
	return cmd
}

// settingsIgnored reports whether .gitignore keeps .claude/settings.json, and with it mem's compaction hook, out of Git.
func settingsIgnored(ctx context.Context, root string) bool {
	_, err := git.Run(ctx, root, "check-ignore", "--quiet", "--no-index", claude.SettingsPath)
	return err == nil
}

const sharedSettingsWarning = "⚠️ .gitignore ignores " + claude.SettingsPath + ", so it is not shared. Each checkout's onboard still installs the compaction hook, but other project settings stay local. Tell the user; Claude Code's convention is to ignore only .claude/settings.local.json (personal settings) and commit " + claude.SettingsPath + "."
