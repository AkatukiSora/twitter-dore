package cmd

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	templatepkg "github.com/AkatukiSora/twitter-dore/internal/template"
)

type stubPrompter struct {
	responses []string
	index     int
}

func (s *stubPrompter) Ask(_ string, allowEmpty bool) (string, error) {
	for s.index < len(s.responses) {
		value := s.responses[s.index]
		s.index++
		if !allowEmpty && value == "" {
			continue
		}
		return value, nil
	}
	return "", errors.New("no more stub responses")
}

func TestRunBasic(t *testing.T) {
	withTerminal(t, false)
	responses := []string{"Alice", "100"}
	withRunPrompter(t, responses)

	dir := t.TempDir()
	templatePath := filepath.Join(dir, "tpl.yaml")
	doc := templatepkg.Document{
		Title:       "sample",
		Description: "desc",
		Template:    "呼び方: {}\n好感度: {}",
	}
	if err := templatepkg.WriteFile(templatePath, doc); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cmd := NewRootCmd()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"run", "--in", templatePath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "呼び方: Alice\n好感度: 100"
	if outBuf.String() != expected {
		t.Fatalf("unexpected output:\nwant: %q\nhave: %q", expected, outBuf.String())
	}
}

func TestRunLiteralPlaceholders(t *testing.T) {
	withTerminal(t, false)
	responses := []string{"OK"}
	withRunPrompter(t, responses)

	dir := t.TempDir()
	path := filepath.Join(dir, "tpl.yaml")
	doc := templatepkg.Document{
		Template: "literal: {{}}\nvalue: {}",
	}
	if err := templatepkg.WriteFile(path, doc); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cmd := NewRootCmd()
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"run", "--in", path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	expected := "literal: {}\nvalue: OK"
	if outBuf.String() != expected {
		t.Fatalf("unexpected output: want %q, got %q", expected, outBuf.String())
	}
}

func TestRunNoEmpty(t *testing.T) {
	withTerminal(t, false)
	responses := []string{"", "retry"}
	withRunPrompter(t, responses)

	dir := t.TempDir()
	path := filepath.Join(dir, "tpl.yaml")
	doc := templatepkg.Document{
		Template: "value: {}",
	}
	if err := templatepkg.WriteFile(path, doc); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cmd := NewRootCmd()
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"run", "--in", path, "--no-empty"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got, want := strings.TrimSpace(outBuf.String()), "value: retry"; got != want {
		t.Fatalf("unexpected result: want %q, got %q", want, got)
	}
}

func TestRunColorModes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tpl.yaml")
	if err := templatepkg.WriteFile(path, templatepkg.Document{Template: "value: {}"}); err != nil {
		t.Fatalf("write template: %v", err)
	}

	responses := []string{"answer"}
	withRunPrompter(t, responses)
	withTerminal(t, true)

	cmdNever := NewRootCmd()
	stdoutNever := &bytes.Buffer{}
	stderrNever := &bytes.Buffer{}
	cmdNever.SetOut(stdoutNever)
	cmdNever.SetErr(stderrNever)
	cmdNever.SetArgs([]string{"run", "--in", path, "--color=never"})
	if err := cmdNever.Execute(); err != nil {
		t.Fatalf("execute never: %v", err)
	}
	if strings.Contains(stderrNever.String(), "\x1b[") {
		t.Fatalf("expected no ANSI codes when color=never, got %q", stderrNever.String())
	}

	// auto mode with simulated TTY should include ANSI.
	cmdAuto := NewRootCmd()
	stdoutAuto := &bytes.Buffer{}
	stderrAuto := &bytes.Buffer{}
	cmdAuto.SetOut(stdoutAuto)
	cmdAuto.SetErr(stderrAuto)
	cmdAuto.SetArgs([]string{"run", "--in", path, "--color=auto"})
	if err := cmdAuto.Execute(); err != nil {
		t.Fatalf("execute auto: %v", err)
	}
	if !strings.Contains(stderrAuto.String(), "\x1b[") {
		t.Fatalf("expected ANSI codes when color=auto with TTY, got %q", stderrAuto.String())
	}
}

func withRunPrompter(t *testing.T, responses []string) {
	old := runPromptBuilder
	runPromptBuilder = func(*cobra.Command) (prompter, error) {
		return &stubPrompter{responses: append([]string(nil), responses...)}, nil
	}
	t.Cleanup(func() {
		runPromptBuilder = old
	})
}

func withTerminal(t *testing.T, value bool) {
	prev := isTerminalFunc
	isTerminalFunc = func(io.Writer) bool { return value }
	t.Cleanup(func() { isTerminalFunc = prev })
}
