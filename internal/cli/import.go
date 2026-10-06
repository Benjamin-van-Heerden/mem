package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
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
			if _, err := agentsmd.WriteGuide(root); err != nil {
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
			fmt.Fprintln(out, "Install guide: "+agentsmd.GuidePath+", for teammates and agents without mem")
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
			if sum.NextSteps != "" {
				output.Section(out, "🔜 WHAT COMES NEXT, FROM THE LATEST LOG")
				fmt.Fprintf(out, "The import removed the \"What Comes Next\" sections from the logs; mem logs record facts and open work lives in todos. This is the one from %s:\n\n%s\n", sum.LatestLog, sum.NextSteps)
			}
			var lines []string
			step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines), text)) }
			lines = append(lines, "Nothing was committed, and .agent_core/ is still in place.")
			step("Show the user the result: `git status`, AGENTS.md and .mem/.")
			if sum.NextSteps != "" {
				step("Go through the items under 🔜 WHAT COMES NEXT with the user. Much of it is usually done by now; record what is still open as todos (`mem todo new \"<title>\" \"<description>\"`).")
			}
			if sum.Todos > 0 {
				step("Check the imported todos against the code with the user and delete the ones that are done (`mem todo delete <todo>`).")
			}
			step("Once the user is happy, remove the old harness: `git rm -r -q .agent_core`, then `rm -rf .agent_core` for the ignored files it leaves behind. Drop its entries from .gitignore.")
			step("Find what still refers to the old harness (`git grep -n -e agent_core -e harness/main.py -- ':!.mem'`), such as setup docs, scripts and CI, and update it with the user to use mem.")
			step(fmt.Sprintf("Commit everything on %s and push it.", g.Development))
			step("Run `mem onboard`.")
			if sum.Specs > 0 {
				step("Check the open specs against the code with the user. Complete the ones whose work is done with `mem spec complete <spec>`; it commits and pushes, so do this only now.")
			}
			output.Instruction(out, lines...)
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
