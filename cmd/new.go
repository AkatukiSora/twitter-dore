package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	templatepkg "github.com/AkatukiSora/twitter-dore/internal/template"
	"github.com/AkatukiSora/twitter-dore/internal/ui"
)

const templateEndToken = "EOF"

var newPromptBuilder = defaultPromptFactory

func newNewCmd() *cobra.Command {
	var (
		outPath         string
		force           bool
		titleFlag       string
		descriptionFlag string
		inlineTemplate  string
		templateFile    string
	)

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new template YAML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outPath == "" {
				return errors.New("--out is required")
			}

			if err := ensureWritable(outPath, force); err != nil {
				return err
			}

			modeSettings := getColorSettings(cmd)
			styler := ui.NewStyler(modeSettings)

			switch {
			case inlineTemplate != "" && templateFile != "":
				return errors.New("only one of --template-inline or --template-file may be set")
			case inlineTemplate != "":
				return writeTemplateFile(cmd, outPath, titleFlag, descriptionFlag, decodeInline(inlineTemplate), styler, force)
			case templateFile != "":
				body, err := os.ReadFile(templateFile)
				if err != nil {
					return fmt.Errorf("failed to read template file: %w", err)
				}

				return writeTemplateFile(cmd, outPath, titleFlag, descriptionFlag, string(body), styler, force)
			default:
				return runInteractiveNew(cmd, interactiveInputs{
					outPath:     outPath,
					force:       force,
					title:       titleFlag,
					description: descriptionFlag,
				})
			}
		},
	}

	cmd.Flags().StringVar(&outPath, "out", "", "Path for the generated YAML template")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite the output file if it exists")
	cmd.Flags().StringVar(&titleFlag, "title", "", "Template title")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "Template description")
	cmd.Flags().StringVar(&inlineTemplate, "template-inline", "", "Template body provided inline (supports \\n escape sequences)")
	cmd.Flags().StringVar(&templateFile, "template-file", "", "Read template body from file")

	_ = cmd.MarkFlagRequired("out")

	return cmd
}

type interactiveInputs struct {
	outPath     string
	force       bool
	title       string
	description string
}

func runInteractiveNew(cmd *cobra.Command, inputs interactiveInputs) error {
	prompter, err := newPromptBuilder(cmd)
	if err != nil {
		return err
	}

	title := inputs.title
	if title == "" {
		title, err = prompter.Ask("title", true)
		if err != nil {
			return err
		}
	}

	description := inputs.description
	if description == "" {
		description, err = prompter.Ask("description", true)
		if err != nil {
			return err
		}
	}

	lines := make([]string, 0)
	for {
		label := fmt.Sprintf("template line %d (enter %s to finish)", len(lines)+1, templateEndToken)
		line, err := prompter.Ask(label, true)
		if err != nil {
			return err
		}

		if line == templateEndToken {
			if len(lines) == 0 {
				if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "テンプレートは1行以上必要です。"); err != nil {
					return err
				}
				continue
			}
			break
		}

		lines = append(lines, line)
	}

	body := strings.Join(lines, "\n")
	if strings.TrimSpace(body) == "" {
		return errors.New("template must contain at least one non-empty line")
	}

	styler := ui.NewStyler(getColorSettings(cmd))
	preview := templatepkg.HighlightPreview(body, styler.HighlightPlaceholder)
	if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "Preview:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(cmd.ErrOrStderr(), preview); err != nil {
		return err
	}

	if err := writeTemplateFile(cmd, inputs.outPath, title, description, body, styler, inputs.force); err != nil {
		return err
	}

	return nil
}

func writeTemplateFile(cmd *cobra.Command, outPath, title, description, body string, styler ui.Styler, force bool) error {
	if strings.TrimSpace(body) == "" {
		return errors.New("template body is empty")
	}

	if err := ensureWritable(outPath, force); err != nil {
		return err
	}

	doc := templatepkg.Document{
		Title:       title,
		Description: description,
		Template:    body,
	}

	if err := templatepkg.WriteFile(outPath, doc); err != nil {
		return err
	}

	if styler.Enabled {
		message := fmt.Sprintf("Template saved to %s", outPath)
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), styler.HighlightLine(message)); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Template saved to %s\n", outPath); err != nil {
			return err
		}
	}

	return nil
}

func ensureWritable(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite)", path)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

func decodeInline(value string) string {
	var builder strings.Builder

	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch != '\\' || i == len(value)-1 {
			builder.WriteByte(ch)
			continue
		}

		next := value[i+1]
		switch next {
		case 'n':
			builder.WriteByte('\n')
			i++
		case 't':
			builder.WriteByte('\t')
			i++
		case 'r':
			builder.WriteByte('\r')
			i++
		case '\\':
			builder.WriteByte('\\')
			i++
		default:
			builder.WriteByte(ch)
			builder.WriteByte(next)
			i++
		}
	}

	return builder.String()
}
