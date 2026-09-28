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
	cmd.AddCommand(a.templateUse(), a.templateList())
	return cmd
}

// librarySource picks the library URL: the flag, then the project's, then the user default.
func librarySource(flag, configured string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if configured != "" {
		return configured, nil
	}
	source, err := templates.DefaultSource()
	if err != nil || source != "" {
		return source, err
	}
	path, _ := templates.UserConfigPath()
	return "", fmt.Errorf("no template library is configured. Pass --template-source <git url>, or set a default for all projects by adding `template_source = \"<git url>\"` to %s", path)
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
			if p.Config.Templates.Source, err = librarySource(source, p.Config.Templates.Source); err != nil {
				return err
			}
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
			output.Instruction(out, templateInstructions(res, ".mem/config.toml")...)
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "template-source", "", "Git URL of the template library (defaults to the project's, then template_source in the user config)")
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
			src, err := librarySource(source, p.Config.Templates.Source)
			if err != nil {
				return err
			}
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
				fmt.Fprintln(out, "\nThe library has no templates yet. A template is a directory with a template.toml; `mem template promote --to <name>` creates one.")
			} else {
				rows := [][]string{{"TEMPLATE", "USED HERE", "DESCRIPTION"}}
				for _, t := range available {
					used := ""
					if projectErr == nil && slices.Contains(p.Config.Templates.Use, t.Name) {
						used = "yes"
					}
					rows = append(rows, []string{t.Name, used, t.Description})
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
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "template-source", "", "Git URL of the template library (defaults to the project's, then template_source in the user config)")
	return cmd
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
	paths := append(extra, res.Paths...)
	lines := []string{fmt.Sprintf("1. Show the user what changed (%s), then commit it.", strings.Join(slices.Compact(paths), ", "))}
	if slices.ContainsFunc(res.Lines, func(l string) bool { return strings.HasPrefix(l, "⚠️") }) {
		lines = append(lines, "2. Tell the user about each ⚠️ item above and settle it with them.")
	}
	return lines
}

// openTemplates opens the project's library for a sync, or explains why it cannot.
func openTemplates(ctx context.Context, p project.Project, pull bool) (templates.Library, string, error) {
	if p.Config.Templates.Source == "" {
		return templates.Library{}, "", errors.New("the project uses templates but [templates] source is not set in .mem/config.toml; run `mem template use <template> --template-source <git url>`")
	}
	return templates.Open(ctx, p.Config.Templates.Source, pull)
}
