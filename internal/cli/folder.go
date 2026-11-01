package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	osuser "os/user"
	"path/filepath"

	"github.com/Benjamin-van-Heerden/mem/internal/agenthooks"
	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/buildinfo"
	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/spf13/cobra"
)

// gitOnlyInitFlags are the init flags that have no meaning in a folder project.
var gitOnlyInitFlags = []string{"remote", "development", "staging", "production", "template", "no-template", "template-source", "production-pr", "release-notes", "single-branch", "protect"}

// initFolder sets up the directory as a folder project: instructions, config and agent hooks, and nothing Git.
func (a *app) initFolder(cmd *cobra.Command, name, description, user string) error {
	for _, flag := range gitOnlyInitFlags {
		if cmd.Flags().Changed(flag) {
			return fmt.Errorf("--folder sets up a plain folder without Git or templates, so it cannot be combined with --%s", flag)
		}
	}
	root, err := filepath.Abs(a.dir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(project.ConfigPath(root)); err == nil {
		return errors.New("this folder already has .mem/config.toml")
	}
	if user == "" {
		user = accountName()
	}
	if project.Slugify(user) == "" {
		return errors.New("mem could not tell who you are from this account; pass --user \"Your Name\" to attribute work")
	}
	if name == "" {
		name = filepath.Base(root)
	}
	config := project.Config{Schema: project.Schema, Name: name, Description: description, Mode: project.ModeFolder, User: user}

	agentsPath := filepath.Join(root, "AGENTS.md")
	existing, err := os.ReadFile(agentsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	agents, err := agentsmd.Install(string(existing), buildinfo.Version, agentsmd.Folder)
	if err != nil {
		return err
	}
	if err := project.WriteConfig(root, config); err != nil {
		return err
	}
	if err := os.WriteFile(agentsPath, []byte(agents), 0o644); err != nil {
		return err
	}
	if _, err := agentsmd.WriteGuide(root); err != nil {
		return err
	}
	if _, err := agenthooks.SyncClaude(root, agenthooks.ClaudeHooks{Compact: true, Prompt: true}); err != nil {
		return err
	}
	if _, err := agenthooks.SyncCursor(root, true); err != nil {
		return err
	}
	removed, err := removeClaudeLink(root)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	output.Heading(out, "📦 MEM INITIALIZED")
	fmt.Fprintf(out, "Project: %s (folder, no Git)\nUser: %s\n", config.Name, project.Slugify(user))
	output.Section(out, "📄 FILES")
	fmt.Fprintln(out, ".mem/config.toml   project configuration")
	fmt.Fprintln(out, agentsmd.GuidePath+"    how to install mem, for agents without it")
	fmt.Fprintln(out, "AGENTS.md           mem instructions and project memories added; existing content kept")
	fmt.Fprintln(out, agenthooks.ClaudeSettingsPath+"  Claude Code hooks for context after compaction and work-log nudges")
	fmt.Fprintln(out, agenthooks.CursorHooksPath+"     Cursor hook that has the agent write a due work log when a turn ends")
	if removed {
		fmt.Fprintln(out, claudeLinkRemoved)
	}
	output.Instruction(out,
		"1. Read AGENTS.md now: it contains the working instructions for this folder.",
		"2. Run `mem onboard` to build the project context.",
	)
	return nil
}

// accountName is who the OS says is logged in. Release builds have no cgo, and without it the full name is unknown.
func accountName() string {
	account, err := osuser.Current()
	if err != nil {
		return ""
	}
	if account.Name != "" {
		return account.Name
	}
	return account.Username
}

// onboardFolder applies updates and prints the context of a folder project. There is nothing to sync: cloud storage
// has already brought the files here.
func (a *app) onboardFolder(cmd *cobra.Command, p project.Project, user, updateLine string) error {
	before := readKnowledge(p)
	updates, err := applyFolderUpdates(p)
	if err != nil {
		return err
	}
	if updateLine != "" {
		updates = append([]string{updateLine}, updates...)
	}
	agents, err := readAgents(p)
	if err != nil {
		return err
	}
	if warning := agentsSizeWarning(agents); warning != "" {
		updates = append(updates, warning)
	}

	var buf bytes.Buffer
	knowledgeChanged := renderKnowledgeChanges(&buf, before, readKnowledge(p))
	state, err := writeContext(cmd.Context(), &buf, p, user)
	if err != nil {
		return err
	}
	state.knowledgeChanged = knowledgeChanged
	state.updateWarnings = hasWarning(updates)

	out := cmd.OutOrStdout()
	output.Heading(out, "📄 MEM ONBOARD: "+p.Config.Name)
	if p.Config.Description != "" {
		fmt.Fprintln(out, p.Config.Description)
	}
	fmt.Fprintf(out, "User: %s\n", user)
	if len(updates) > 0 {
		output.Section(out, "🔄 UPDATES")
		for _, line := range updates {
			fmt.Fprintln(out, line)
		}
	}
	if err := emitContext(out, &buf, p); err != nil {
		return err
	}
	renderFolderInstruction(out, state)
	return nil
}

// applyFolderUpdates brings a folder project's managed files up to date with this executable.
func applyFolderUpdates(p project.Project) ([]string, error) {
	var lines []string
	applied, _, err := project.Upgrade(p)
	if err != nil {
		return nil, err
	}
	if len(applied) > 0 {
		lines = append(lines, fmt.Sprintf("Upgraded the project format to schema %d.", project.Schema))
	}
	text, err := readAgents(p)
	if err != nil {
		return nil, err
	}
	refreshed, newer, err := agentsmd.Refresh(text, buildinfo.Version, agentsmd.Folder)
	if err != nil {
		return nil, err
	}
	switch {
	case newer != "":
		lines = append(lines, fmt.Sprintf("AGENTS.md was written by mem %s, which is newer than this mem (%s). Tell the user to update mem.", newer, buildinfo.Version))
	case refreshed != text:
		if err := writeAgents(p, refreshed); err != nil {
			return nil, err
		}
		lines = append(lines, "Refreshed the mem instructions in AGENTS.md.")
	}
	switch changed, err := agentsmd.WriteGuide(p.Root); {
	case err != nil:
		return nil, err
	case changed:
		lines = append(lines, "Updated "+agentsmd.GuidePath+", the install guide AGENTS.md points to when mem is missing.")
	}
	claudeWant := agenthooks.ClaudeHooks{Compact: p.Config.Claude.CompactHookEnabled(), Prompt: p.Config.Claude.PromptHookEnabled()}
	cursorStop := p.Config.Cursor.StopHookEnabled()
	for _, h := range []struct {
		agent, line string
		sync        func() (bool, error)
	}{
		{"Claude Code", claudeHooksLine(claudeWant), func() (bool, error) { return agenthooks.SyncClaude(p.Root, claudeWant) }},
		{"Cursor", cursorHookLine(cursorStop), func() (bool, error) { return agenthooks.SyncCursor(p.Root, cursorStop) }},
	} {
		switch changed, err := h.sync(); {
		case err != nil:
			lines = append(lines, fmt.Sprintf("⚠️ Could not update mem's %s hooks: %v. Tell the user; fix the file, and the next onboard installs the hooks.", h.agent, err))
		case changed:
			lines = append(lines, h.line)
		}
	}
	return lines, nil
}

func renderFolderInstruction(out io.Writer, state contextState) {
	lines := []string{"Your next response must:"}
	step := func(text string) { lines = append(lines, fmt.Sprintf("%d. %s", len(lines), text)) }
	if state.knowledgeChanged {
		step("Follow the memories under 🧠 CHANGED MEMORIES for the rest of this session, and use the skills under 🛠️ CHANGED SKILLS where they apply: they changed after this session started.")
	}
	if state.updateWarnings {
		step("Tell the user about each ⚠️ item under 🔄 UPDATES and settle it with them.")
	}
	if state.drift.Missing {
		step("Mention that there is no structure doc yet, and offer to create one with `mem structure`.")
	}
	step("Summarize where things stand from the open todos and any runnable output. Use tables where they help. Work logs are background: do not present what an old log planned as open work unless a todo still holds it.")
	step("Ask the user how they would like to proceed.")
	lines = append(lines, "", "If the user has already said what to work on, confirm it briefly and continue with that instead of asking.", nameSession())
	output.Instruction(out, lines...)
}
