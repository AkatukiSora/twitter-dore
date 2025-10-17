package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	templatepkg "github.com/AkatukiSora/twitter-dore/internal/template"
	"github.com/AkatukiSora/twitter-dore/internal/ui"
)

var runPromptBuilder = defaultPromptFactory

func newRunCmd() *cobra.Command {
	var (
		inputPath string
		output    string
		noEmpty   bool
		quiet     bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Fill a YAML template by replacing placeholders",
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return errors.New("--in is required")
			}

			doc, err := templatepkg.LoadFile(inputPath)
			if err != nil {
				return err
			}

			if err := doc.Validate(); err != nil {
				if errors.Is(err, templatepkg.ErrTemplateMissing) {
					return fmt.Errorf("%s: %w", inputPath, err)
				}
				return err
			}

			session, err := templatepkg.NewSession(doc.Template)
			if err != nil {
				return err
			}

			placeholders := session.Placeholders()
			values := make([]string, len(placeholders))
			prompter, err := runPromptBuilder(cmd)
			if err != nil {
				return err
			}

			styler := ui.NewStyler(getColorSettings(cmd))
			allowEmpty := !noEmpty

			for idx, placeholder := range placeholders {
				highlighted := styler.HighlightLine(placeholder.Line)
				if _, err := fmt.Fprintln(cmd.ErrOrStderr(), highlighted); err != nil {
					return err
				}

				value, err := prompter.Ask(placeholder.Label, allowEmpty)
				if err != nil {
					return err
				}

				values[idx] = value
			}

			result, err := session.Fill(values)
			if err != nil {
				return err
			}

			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0o644); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
			}

			if !quiet {
				if _, err := fmt.Fprint(cmd.OutOrStdout(), result); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputPath, "in", "", "Path to template YAML")
	cmd.Flags().StringVar(&output, "out", "", "Path to write filled template")
	cmd.Flags().BoolVar(&noEmpty, "no-empty", false, "Require non-empty answers for placeholders")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress completed output")

	_ = cmd.MarkFlagRequired("in")

	return cmd
}
