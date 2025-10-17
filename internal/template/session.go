package template

import (
	"fmt"
	"strings"
)

// Placeholder describes a detected "{}" token in the template body.
type Placeholder struct {
	Index int
	Label string
	Line  string
}

// Session represents a prepared template ready to be filled.
type Session struct {
	sanitized    string
	placeholders []Placeholder
	protector    literalProtector
}

// NewSession prepares the template body for interactive filling.
func NewSession(raw string) (*Session, error) {
	sanitized, protector := protectLiterals(raw)
	placeholders := extractPlaceholders(sanitized, protector)

	return &Session{
		sanitized:    sanitized,
		placeholders: placeholders,
		protector:    protector,
	}, nil
}

// Placeholders returns a copy of the detected placeholders.
func (s *Session) Placeholders() []Placeholder {
	result := make([]Placeholder, len(s.placeholders))
	copy(result, s.placeholders)
	return result
}

// Fill applies the supplied values to the template in order.
func (s *Session) Fill(values []string) (string, error) {
	if len(values) != len(s.placeholders) {
		return "", fmt.Errorf("expected %d values but received %d", len(s.placeholders), len(values))
	}

	result := s.sanitized
	for _, value := range values {
		result = strings.Replace(result, "{}", value, 1)
	}

	return s.protector.Restore(result), nil
}

// HighlightPreview returns the template with placeholders visually highlighted.
func HighlightPreview(raw string, highlight func(string) string) string {
	if highlight == nil {
		return raw
	}

	sanitized, protector := protectLiterals(raw)
	highlighted := strings.ReplaceAll(sanitized, "{}", highlight("{}"))
	return protector.Restore(highlighted)
}

type literalProtector struct {
	tokens []string
}

func protectLiterals(input string) (string, literalProtector) {
	var builder strings.Builder
	var tokens []string

	for i := 0; i < len(input); {
		if strings.HasPrefix(input[i:], "{{}}") {
			token := fmt.Sprintf("__TWITTER_DORE_LITERAL_%d__", len(tokens))
			tokens = append(tokens, token)
			builder.WriteString(token)
			i += len("{{}}")
			continue
		}

		builder.WriteByte(input[i])
		i++
	}

	return builder.String(), literalProtector{tokens: tokens}
}

func (p literalProtector) Restore(value string) string {
	if len(p.tokens) == 0 {
		return value
	}

	result := value
	for _, token := range p.tokens {
		result = strings.ReplaceAll(result, token, "{}")
	}

	return result
}

func extractPlaceholders(input string, protector literalProtector) []Placeholder {
	lines := strings.Split(input, "\n")
	placeholders := make([]Placeholder, 0)

	fieldCounter := 1
	for _, line := range lines {
		displayLine := protector.Restore(line)
		offset := 0
		segmentStart := 0

		for {
			idx := strings.Index(line[offset:], "{}")
			if idx < 0 {
				break
			}

			absolute := offset + idx
			labelSegment := line[segmentStart:absolute]
			labelText := strings.TrimSpace(protector.Restore(labelSegment))
			if labelText == "" {
				labelText = fmt.Sprintf("field%d", fieldCounter)
			}

			placeholders = append(placeholders, Placeholder{
				Index: len(placeholders),
				Label: labelText,
				Line:  displayLine,
			})

			fieldCounter++
			offset = absolute + len("{}")
			segmentStart = offset
		}
	}

	return placeholders
}
