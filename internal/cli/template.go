package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Benjamin-van-Heerden/mem/internal/output"
	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/templates"
	"github.com/spf13/cobra"
)

func (a *app) templateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Draw memories, skills and docs from a template library, and promote them back",
		Long:  "A template library is a Git repository with one directory per template (for example nextjs-web/), each holding template.toml, memories/<name>.md, skills/<name>/ and docs/<name>.md. Projects list the templates they use in .mem/config.toml; onboard keeps them in sync.",
	}
	cmd.AddCommand(a.templateUse(), a.templateList(), a.templatePromote(), a.templateReset())
	return cmd
}

// librarySource picks the library URL: the flag, then the project's, then mem's default library.
func librarySource(flag, configured string) string {
	if flag != "" {
		return flag
	}
	if configured != "" {
		return configured
	}
	return templates.DefaultSource
}

func (a *app) templateUse() *cobra.Command {
	var source string
	cmd := &cobra.Command{
		Use:   "use <template>",
		Short: "Start using a template in this project and draw in its items",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			p.Config.Templates.Source = librarySource(source, p.Config.Templates.Source)
			lib, warning, err := templates.Open(cmd.Context(), p.Config.Templates.Source, true)
			if err != nil {
				return err
			}
			name := args[0]
			if slices.Contains(p.Config.Templates.Use, name) {
				return fmt.Errorf("this project already uses %s; onboard keeps it in sync", name)
			}
			p.Config.Templates.Use = append(p.Config.Templates.Use, name)
			if _, _, err := lib.Items(p.Config.Templates.Use); err != nil {
				return err
			}
			if err := project.WriteConfig(p.Root, p.Config); err != nil {
				return err
			}
			res, err := templates.Sync(&p, lib)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "🧩 TEMPLATE ADDED: "+name)
			fmt.Fprintf(out, "Library: %s\nTemplates in use: %s\n", p.Config.Templates.Source, strings.Join(p.Config.Templates.Use, ", "))
			renderTemplateSync(out, warning, res)
			available, err := lib.Templates()
			if err != nil {
				return err
			}
			if slices.ContainsFunc(available, func(t templates.Template) bool { return t.Name == name && t.HasSetup }) {
				fmt.Fprintf(out, "%s has a setup for new projects (%s/setup.md in the library); this project did not run it. Its steps can guide adding what is missing.\n", name, name)
			}
			output.Instruction(out, templateInstructions(res, ".mem/config.toml")...)
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "template-source", "", "Git URL of the template library (defaults to the project's, then "+templates.DefaultSource+")")
	return cmd
}

func (a *app) templateList() *cobra.Command {
	var source string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the library's templates and the state of this project's template items",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, projectErr := a.project(cmd)
			src := librarySource(source, p.Config.Templates.Source)
			lib, warning, err := templates.Open(cmd.Context(), src, true)
			if err != nil {
				return err
			}
			available, err := lib.Templates()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, "🧩 TEMPLATES")
			fmt.Fprintf(out, "Library: %s\n", src)
			if warning != "" {
				fmt.Fprintln(out, "⚠️ "+warning)
			}
			if len(available) == 0 {
				fmt.Fprintln(out, "\nThe library has no templates yet. A template is a directory with a template.toml; `mem template promote <memory|skill|doc> <name> --to <template>` creates one from a project item.")
			} else {
				rows := [][]string{{"TEMPLATE", "USED HERE", "SETUP", "DESCRIPTION"}}
				for _, t := range available {
					used, setup := "", ""
					if projectErr == nil && slices.Contains(p.Config.Templates.Use, t.Name) {
						used = "yes"
					}
					if t.HasSetup {
						setup = "yes"
					}
					rows = append(rows, []string{t.Name, used, setup, t.Description})
				}
				fmt.Fprintln(out)
				table(out, rows)
			}
			if projectErr != nil || len(p.Config.Templates.Use) == 0 {
				return nil
			}
			statuses, err := templates.Status(&p, lib)
			if err != nil {
				return err
			}
			output.Section(out, "📦 THIS PROJECT'S TEMPLATE ITEMS")
			if len(statuses) == 0 {
				fmt.Fprintln(out, "The templates in use provide no items yet.")
				return nil
			}
			rows := [][]string{{"KIND", "NAME", "TEMPLATE", "STATE"}}
			for _, st := range statuses {
				rows = append(rows, []string{st.Item.Kind, st.Item.Name, st.Item.Template, st.State})
			}
			table(out, rows)
			hints := map[string]string{
				templates.StateLocalEdits: "Local edits can be shared with similar projects: `mem template promote <kind> <name>`.",
				templates.StateBothEdited: "Items edited on both sides keep their local version until the user decides: `mem template promote <kind> <name>` keeps it, `mem template reset <kind> <name>` takes the template's.",
				templates.StateDiffers:    "Items that differ but are not tracked: `mem template promote <kind> <name>` keeps the project's version, `mem template reset <kind> <name>` takes the template's.",
				templates.StateUpdated:    "Template updates arrive at the next `mem onboard`.",
				templates.StateMissing:    "Missing items arrive at the next `mem onboard`.",
				templates.StateExcluded:   "Excluded items are listed under [templates] exclude in .mem/config.toml; `mem template reset <kind> <name>` brings one back.",
			}
			var notes []string
			for _, st := range statuses {
				if hint, ok := hints[st.State]; ok && !slices.Contains(notes, hint) {
					notes = append(notes, hint)
				}
			}
			if len(notes) > 0 {
				fmt.Fprintln(out)
				for _, note := range notes {
					fmt.Fprintln(out, note)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "template-source", "", "Git URL of the template library (defaults to the project's, then "+templates.DefaultSource+")")
	return cmd
}

func (a *app) templatePromote() *cobra.Command {
	var to string
	cmd := &cobra.Command{
		Use:   "promote <memory|skill|doc> <name>",
		Short: "Send a project memory, skill or doc to its template so similar projects receive it",
		Long:  "Copies the item into the template library, commits and pushes it. Every project that uses the template receives it at its next onboard. Items that did not come from a template go to the project's only template, or to the one named with --to, which is created if it does not exist.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			p.Config.Templates.Source = librarySource("", p.Config.Templates.Source)
			lib, _, err := templates.Open(cmd.Context(), p.Config.Templates.Source, false)
			if err != nil {
				return err
			}
			res, err := templates.Promote(cmd.Context(), &p, lib, args[0], args[1], to)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, fmt.Sprintf("⬆️ PROMOTED: %s %s", args[0], args[1]))
			for _, line := range res.Lines {
				fmt.Fprintln(out, line)
			}
			output.Instruction(out,
				fmt.Sprintf("1. Commit %s in this project.", strings.Join(res.Paths, ", ")),
				"2. Tell the user that every project using the template receives the item at its next `mem onboard`.",
			)
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "Template to promote to (needed when the item is not from a template and the project uses several)")
	return cmd
}

func (a *app) templateReset() *cobra.Command {
	return &cobra.Command{
		Use:   "reset <memory|skill|doc> <name>",
		Short: "Replace the project's copy of a template item with the template's",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.project(cmd)
			if err != nil {
				return err
			}
			lib, warning, err := openTemplates(cmd.Context(), p, true)
			if err != nil {
				return err
			}
			res, err := templates.Reset(&p, lib, args[0], args[1])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			output.Section(out, fmt.Sprintf("↩️ RESET: %s %s", args[0], args[1]))
			renderTemplateSync(out, warning, res)
			output.Instruction(out, templateInstructions(res)...)
			return nil
		},
	}
}

func renderTemplateSync(out io.Writer, warning string, res templates.Result) {
	if warning != "" {
		fmt.Fprintln(out, "⚠️ "+warning)
	}
	if len(res.Lines) == 0 {
		fmt.Fprintln(out, "Everything from the templates is already in place.")
	}
	for _, line := range res.Lines {
		fmt.Fprintln(out, line)
	}
}

// templateInstructions tells the agent what to do with a sync's changes.
func templateInstructions(res templates.Result, extra ...string) []string {
	paths := append(extra, describePaths(res.Paths)...)
	lines := []string{fmt.Sprintf("1. Show the user what changed (%s), then commit it.", strings.Join(slices.Compact(paths), ", "))}
	if slices.ContainsFunc(res.Lines, func(l string) bool { return strings.HasPrefix(l, "⚠️") }) {
		lines = append(lines, "2. Tell the user about each ⚠️ item above and settle it with them.")
	}
	return lines
}

// describePaths keeps instructions readable when a template brings many skills: skill
// directories and their .claude links are named by directory with a count.
func describePaths(paths []string) []string {
	groups := []struct{ dir, noun string }{{".agents/skills/", "skill"}, {".claude/skills/", "link"}}
	counts := make([]int, len(groups))
	var out []string
	for _, path := range paths {
		grouped := false
		for i, g := range groups {
			if strings.HasPrefix(path, g.dir) {
				counts[i]++
				grouped = true
			}
		}
		if !grouped {
			out = append(out, path)
		}
	}
	for i, g := range groups {
		switch {
		case counts[i] == 1:
			out = append(out, fmt.Sprintf("%s (1 %s)", g.dir, g.noun))
		case counts[i] > 1:
			out = append(out, fmt.Sprintf("%s (%d %ss)", g.dir, counts[i], g.noun))
		}
	}
	return out
}

// openTemplates opens the project's library for a sync, or explains why it cannot.
func openTemplates(ctx context.Context, p project.Project, pull bool) (templates.Library, string, error) {
	if p.Config.Templates.Source == "" {
		return templates.Library{}, "", errors.New("the project uses templates but [templates] source is not set in .mem/config.toml; run `mem template use <template> --template-source <git url>`")
	}
	return templates.Open(ctx, p.Config.Templates.Source, pull)
}
