package vault

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type OKFDocument struct {
	Frontmatter map[string]interface{}
	Body        string
}

// ParseOKF parses a markdown string with optional YAML frontmatter.
func ParseOKF(content string) (*OKFDocument, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return &OKFDocument{
			Frontmatter: make(map[string]interface{}),
			Body:        content,
		}, nil
	}

	// Split by "---" to isolate YAML and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return &OKFDocument{
			Frontmatter: make(map[string]interface{}),
			Body:        content,
		}, nil
	}

	yamlStr := strings.TrimSpace(parts[1])
	bodyStr := strings.TrimSpace(parts[2])

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	if fm == nil {
		fm = make(map[string]interface{})
	}

	return &OKFDocument{
		Frontmatter: fm,
		Body:        bodyStr,
	}, nil
}

// FormatOKF formats an OKFDocument back into a string with frontmatter.
func FormatOKF(doc *OKFDocument) (string, error) {
	if len(doc.Frontmatter) == 0 {
		return doc.Body, nil
	}

	yamlBytes, err := yaml.Marshal(doc.Frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML frontmatter: %w", err)
	}

	// Format with standard frontmatter layout
	return fmt.Sprintf("---\n%s---\n%s", string(yamlBytes), doc.Body), nil
}

// SetDefault sets a default value for a frontmatter key if it doesn't exist.
func (d *OKFDocument) SetDefault(key string, val interface{}) {
	if d.Frontmatter == nil {
		d.Frontmatter = make(map[string]interface{})
	}
	if _, exists := d.Frontmatter[key]; !exists {
		d.Frontmatter[key] = val
	}
}
