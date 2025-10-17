package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	templatepkg "github.com/AkatukiSora/twitter-dore/internal/template"
)

func TestNewNonInteractive(t *testing.T) {
	withTerminal(t, false)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "tpl.yaml")

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"new",
		"--out", outPath,
		"--title", "t",
		"--description", "d",
		"--template-inline", `A:{}\nB:{}`,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	doc, err := templatepkg.LoadFile(outPath)
	if err != nil {
		t.Fatalf("load output: %v", err)
	}

	if doc.Title != "t" || doc.Description != "d" {
		t.Fatalf("unexpected metadata: %+v", doc)
	}

	if want := "A:{}\nB:{}"; doc.Template != want {
		t.Fatalf("unexpected template: want %q, got %q", want, doc.Template)
	}
}

func TestNewInteractiveSmoke(t *testing.T) {
	withTerminal(t, true)
	responses := []string{
		"My title",
		"My description",
		"呼び方: {}",
		"好感度: {}",
		templateEndToken,
	}
	withNewPrompter(t, responses)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "tpl.yaml")

	errBuf := &bytes.Buffer{}
	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{
		"new",
		"--out", outPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	doc, err := templatepkg.LoadFile(outPath)
	if err != nil {
		t.Fatalf("load output: %v", err)
	}

	if doc.Template != "呼び方: {}\n好感度: {}" {
		t.Fatalf("unexpected template body: %q", doc.Template)
	}

	if !strings.Contains(errBuf.String(), "Preview:") {
		t.Fatalf("expected preview in stderr, got %q", errBuf.String())
	}

	if !strings.Contains(errBuf.String(), "\x1b[") {
		t.Fatalf("expected ANSI codes in preview output when terminal is TTY")
	}
}

func withNewPrompter(t *testing.T, responses []string) {
	old := newPromptBuilder
	newPromptBuilder = func(*cobra.Command) (prompter, error) {
		return &stubPrompter{responses: append([]string(nil), responses...)}, nil
	}
	t.Cleanup(func() {
		newPromptBuilder = old
	})
}
