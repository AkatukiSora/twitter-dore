package template

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document represents the YAML schema for templates.
type Document struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Template    string `yaml:"template"`
}

// ErrTemplateMissing indicates that no template body was provided.
var ErrTemplateMissing = errors.New("template is not defined")

// LoadFile reads the YAML document from disk.
func LoadFile(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}

	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("failed to decode template YAML: %w", err)
	}

	return doc, nil
}

// WriteFile writes the document to disk, creating parent directories when required.
func WriteFile(path string, doc Document) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal template YAML: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create parent directory %q: %w", dir, err)
		}
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate ensures the template body is present.
func (d Document) Validate() error {
	if strings.TrimSpace(d.Template) == "" {
		return ErrTemplateMissing
	}
	return nil
}
