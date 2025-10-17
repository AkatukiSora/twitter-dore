package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// Mode controls how colorized output should behave.
type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeAlways Mode = "always"
	ModeNever  Mode = "never"
)

// ParseMode parses a string flag into a color Mode.
func ParseMode(value string) (Mode, error) {
	switch strings.ToLower(value) {
	case string(ModeAuto):
		return ModeAuto, nil
	case string(ModeAlways):
		return ModeAlways, nil
	case string(ModeNever):
		return ModeNever, nil
	default:
		return "", fmt.Errorf("unknown color mode")
	}
}

// Enabled reports whether color output should be used under the supplied terminal condition.
func (m Mode) Enabled(isTTY bool) bool {
	switch m {
	case ModeAlways:
		return true
	case ModeNever:
		return false
	case ModeAuto:
		return isTTY
	default:
		// Treat invalid values as auto to stay resilient.
		return isTTY
	}
}

// ColorSettings bundles mode and computed enablement.
type ColorSettings struct {
	Mode    Mode
	Enabled bool
}

const (
	resetCode       = "\x1b[0m"
	boldUnderline   = "\x1b[1;4m"
	placeholderCode = boldUnderline
	lineCode        = boldUnderline
)

// Styler wraps helper methods for color-aware string styling.
type Styler struct {
	Enabled bool
}

// NewStyler creates a Styler from ColorSettings.
func NewStyler(settings ColorSettings) Styler {
	return Styler{Enabled: settings.Enabled}
}

// HighlightLine applies bold+underline to the entire line if colors are enabled.
func (s Styler) HighlightLine(line string) string {
	if !s.Enabled {
		return line
	}
	return lineCode + line + resetCode
}

// HighlightPlaceholder wraps "{}" tokens with bold+underline when colors are enabled.
func (s Styler) HighlightPlaceholder(text string) string {
	if !s.Enabled {
		return text
	}
	return placeholderCode + text + resetCode
}

// IsTerminalWriter reports whether the provided writer ultimately targets a TTY.
func IsTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
