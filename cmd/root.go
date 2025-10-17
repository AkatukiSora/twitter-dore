package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/AkatukiSora/twitter-dore/internal/ui"
)

const (
	flagColor = "color"
)

type colorContextKey struct{}

var (
	rootCmd         = newRootCommand()
	isTerminalFunc  = ui.IsTerminalWriter
	defaultColorStr = "auto"
)

// Execute runs the CLI. This is kept small so main can stay trivial.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrln(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}

// NewRootCmd returns a fresh instance of the root command, mainly for tests.
func NewRootCmd() *cobra.Command {
	return newRootCommand()
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twitter-dore",
		Short: "CLI helper for filling Twitter reply templates",
		Long: `twitter-dore loads YAML templates for reply-style questions and guides you
through filling placeholders interactively.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return injectColorSettings(cmd)
		},
	}

	cmd.PersistentFlags().String(flagColor, defaultColorStr, "Color output mode (auto|always|never)")

	cmd.AddCommand(
		newRunCmd(),
		newNewCmd(),
		newVersionCmd(),
		newCompletionCmd(),
	)

	return cmd
}

func injectColorSettings(cmd *cobra.Command) error {
	colorValue, err := cmd.Flags().GetString(flagColor)
	if err != nil {
		return err
	}

	mode, err := ui.ParseMode(colorValue)
	if err != nil {
		return fmt.Errorf("invalid color mode %q: %w", colorValue, err)
	}

	useColor := mode.Enabled(detectTTY(cmd.ErrOrStderr()))
	settings := ui.ColorSettings{
		Mode:    mode,
		Enabled: useColor,
	}

	ctx := context.WithValue(cmd.Context(), colorContextKey{}, settings)
	cmd.SetContext(ctx)

	return nil
}

func getColorSettings(cmd *cobra.Command) ui.ColorSettings {
	value := cmd.Context().Value(colorContextKey{})
	if settings, ok := value.(ui.ColorSettings); ok {
		return settings
	}
	// Should not happen, but fall back to defaults if pre-run didn't execute.
	mode, _ := ui.ParseMode(defaultColorStr)
	return ui.ColorSettings{
		Mode:    mode,
		Enabled: mode.Enabled(detectTTY(cmd.ErrOrStderr())),
	}
}

func detectTTY(writer io.Writer) bool {
	return isTerminalFunc(writer)
}
