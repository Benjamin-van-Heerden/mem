package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/checkpoint"
	"github.com/Benjamin-van-Heerden/mem/internal/claude"
	"github.com/Benjamin-van-Heerden/mem/internal/converge"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
	"github.com/Benjamin-van-Heerden/mem/internal/hooks"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
)

// applyUpdates brings the project and its managed instructions up to date with this executable.
func applyUpdates(ctx context.Context, p project.Project) ([]string, error) {
	var lines, publishPaths []string
	applied, changed, err := project.Upgrade(p)
	if err != nil {
		return nil, err
	}
	if len(applied) > 0 {
		lines = append(lines, fmt.Sprintf("Upgraded the project format to schema %d.", project.Schema))
		publishPaths = append(append(publishPaths, p.Rel(project.ConfigPath(p.Root))), changed...)
	}

	text, err := readAgents(p)
	if err != nil {
		return nil, err
	}
	refreshed, newer, err := agentsmd.Refresh(text, buildinfo.Version)
	if err != nil {
		return nil, err
	}
	switch {
	case newer != "":
		lines = append(lines, fmt.Sprintf("AGENTS.md was written by mem %s, which is newer than this mem (%s). Tell the user to update mem.", newer, buildinfo.Version))
	case refreshed != text:
		status, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", "AGENTS.md")
		if err := writeAgents(p, refreshed); err != nil {
			return nil, err
		}
		if status != "" {
			lines = append(lines, "Refreshed the mem instructions in AGENTS.md. It already had uncommitted edits, so commit it together with them.")
		} else {
			lines = append(lines, "Refreshed the mem instructions in AGENTS.md.")
			publishPaths = append(publishPaths, "AGENTS.md")
		}
	}
	guideStatus, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", agentsmd.GuidePath)
	switch changed, err := agentsmd.WriteGuide(p.Root); {
	case err != nil:
		return nil, err
	case changed && guideStatus != "":
		lines = append(lines, "Updated "+agentsmd.GuidePath+". It already had uncommitted edits, so commit it together with them.")
	case changed:
		lines = append(lines, "Updated "+agentsmd.GuidePath+", the install guide AGENTS.md points to when mem is missing.")
		publishPaths = append(publishPaths, agentsmd.GuidePath)
	}
	ignoreStatus, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", ".gitignore")
	added, err := ensureIgnored(p.Root, localIgnore)
	if err != nil {
		return nil, err
	}
	if added {
		if ignoreStatus != "" {
			lines = append(lines, "Added .mem/local/ to .gitignore. It already had uncommitted edits, so commit it together with them.")
		} else {
			lines = append(lines, "Added .mem/local/ to .gitignore.")
			publishPaths = append(publishPaths, ".gitignore")
		}
	}
	if tracked, _ := git.Run(ctx, p.Root, "ls-files", "--", ".mem/local"); tracked != "" {
		lines = append(lines, "⚠️ Files under .mem/local/ are committed, but they are local to each checkout. Stop tracking them with `git rm -r --cached .mem/local` and commit that.")
	}
	settingsStatus, _ := git.Run(ctx, p.Root, "status", "--porcelain", "--", claude.SettingsPath)
	enabled := p.Config.Claude.CompactHookEnabled()
	switch changed, err := claude.SyncCompactHook(p.Root, enabled); {
	case err != nil:
		lines = append(lines, fmt.Sprintf("⚠️ Could not update the Claude Code compaction hook: %v. Tell the user; fix the file, and the next onboard installs the hook.", err))
	case changed:
		line := "Installed the Claude Code compaction hook in " + claude.SettingsPath + "."
		if !enabled {
			line = "Removed the Claude Code compaction hook from " + claude.SettingsPath + " ([claude] compact_hook = false)."
		}
		switch {
		case settingsIgnored(ctx, p.Root):
			lines = append(lines, line+" .gitignore keeps it out of Git, so it stays on this machine; onboard installs it in every checkout.")
		case settingsStatus != "":
			lines = append(lines, line+" It already had uncommitted edits, so commit it together with them.")
		default:
			lines = append(lines, line)
			publishPaths = append(publishPaths, claude.SettingsPath)
		}
	}
	if len(publishPaths) > 0 {
		lines = append(lines, publish(ctx, p, checkpoint.ProjectFilesCommit, publishPaths...))
	}
	hookLines, err := hooks.Sync(ctx, p)
	return append(lines, hookLines...), err
}

var templatePaths = []string{"AGENTS.md", ".agents/skills", ".claude/skills", ".mem/docs", templates.LockPath, ".mem/config.toml"}

// syncTemplates brings the project's template items up to date and publishes
// the result, unless those paths already had uncommitted edits.
func syncTemplates(ctx context.Context, p *project.Project, pull bool) ([]string, error) {
	if len(p.Config.Templates.Use) == 0 {
		return nil, nil
	}
	lib, warning, err := openTemplates(ctx, *p, pull)
	if err != nil {
		return []string{fmt.Sprintf("⚠️ Could not open the template library: %v. Template items were not synced.", err)}, nil
	}
	var lines []string
	if warning != "" {
		lines = append(lines, "⚠️ "+warning)
	}
	dirty, _ := git.Run(ctx, p.Root, append([]string{"status", "--porcelain", "--"}, templatePaths...)...)
	res, err := templates.Sync(p, lib)
	if err != nil {
		return nil, err
	}
	lines = append(lines, res.Lines...)
	switch {
	case len(res.Paths) == 0:
	case dirty != "":
		lines = append(lines, fmt.Sprintf("These paths already had uncommitted edits, so nothing was committed: commit the template changes (%s) together with them.", strings.Join(res.Paths, ", ")))
	default:
		lines = append(lines, publish(ctx, *p, "Sync template items", res.Paths...))
	}
	return lines, nil
}

func hasWarning(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, "⚠️") {
			return true
		}
	}
	return false
}

// afterUpdates refreshes the ahead/behind counts and the unpushed nudge, since
// publishing updates may have pushed commits that were local during the sync.
func afterUpdates(ctx context.Context, p project.Project, r converge.Report) converge.Report {
	stale := converge.Unpushed(r)
	fresh := converge.Local(ctx, p)
	r.Ahead, r.Behind = fresh.Ahead, fresh.Behind
	nudges := r.Nudges[:0:0]
	for _, line := range r.Nudges {
		if line != stale {
			nudges = append(nudges, line)
		}
	}
	if line := converge.Unpushed(r); line != "" {
		nudges = append(nudges, line)
	}
	r.Nudges = nudges
	return r
}
