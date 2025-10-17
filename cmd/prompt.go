package cmd

import (
	"errors"
	"io"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type prompter interface {
	Ask(label string, allowEmpty bool) (string, error)
}

type promptFactory func(*cobra.Command) (prompter, error)

var defaultPromptFactory promptFactory = newPromptUIPrompter

type promptUIPrompter struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func newPromptUIPrompter(cmd *cobra.Command) (prompter, error) {
	reader := toReadCloser(cmd.InOrStdin())
	writer := toWriteCloser(cmd.ErrOrStderr())

	return &promptUIPrompter{
		reader: reader,
		writer: writer,
	}, nil
}

func (p *promptUIPrompter) Ask(label string, allowEmpty bool) (string, error) {
	validate := func(input string) error {
		if allowEmpty {
			return nil
		}
		if strings.TrimSpace(input) == "" {
			return errors.New("入力が必要です")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     label,
		AllowEdit: true,
		Validate:  validate,
		Stdin:     p.reader,
		Stdout:    p.writer,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

type nopWriteCloser struct {
	io.Writer
}

func (n nopWriteCloser) Close() error {
	return nil
}

func toWriteCloser(w io.Writer) io.WriteCloser {
	if wc, ok := w.(io.WriteCloser); ok {
		return wc
	}
	return nopWriteCloser{Writer: w}
}

func toReadCloser(r io.Reader) io.ReadCloser {
	if rc, ok := r.(io.ReadCloser); ok {
		return rc
	}
	return io.NopCloser(r)
}
