package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAndFormatOKF(t *testing.T) {
	// Test case 1: Raw text without frontmatter
	rawText := "This is a simple raw text memory."
	doc1, err := ParseOKF(rawText)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if doc1.Body != rawText {
		t.Errorf("expected body to match, got: %q", doc1.Body)
	}
	if len(doc1.Frontmatter) != 0 {
		t.Errorf("expected empty frontmatter, got: %v", doc1.Frontmatter)
	}

	formatted1, err := FormatOKF(doc1)
	if err != nil {
		t.Fatalf("failed to format: %v", err)
	}
	if formatted1 != rawText {
		t.Errorf("expected formatted output to match original text, got: %q", formatted1)
	}

	// Test case 2: Content with valid frontmatter
	okfText := `---
type: custom
tags:
  - testing
  - go
---
This is the body of the OKF file.
It can have multiple lines.`

	doc2, err := ParseOKF(okfText)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if doc2.Body != "This is the body of the OKF file.\nIt can have multiple lines." {
		t.Errorf("unexpected body content: %q", doc2.Body)
	}

	if doc2.Frontmatter["type"] != "custom" {
		t.Errorf("expected type frontmatter to be 'custom', got: %v", doc2.Frontmatter["type"])
	}

	tags, ok := doc2.Frontmatter["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Errorf("expected 2 tags, got: %v", doc2.Frontmatter["tags"])
	} else if tags[0] != "testing" || tags[1] != "go" {
		t.Errorf("unexpected tags: %v", tags)
	}

	formatted2, err := FormatOKF(doc2)
	if err != nil {
		t.Fatalf("failed to format: %v", err)
	}
	if !strings.HasPrefix(formatted2, "---") || !strings.Contains(formatted2, "type: custom") || !strings.Contains(formatted2, "This is the body of the OKF file.") {
		t.Errorf("malformed formatted OKF: %q", formatted2)
	}
}

func TestVaultOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "supermemory_vault_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	v, err := NewVault(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize vault: %v", err)
	}

	id := "test-mem-123"
	content := "Obsidian works perfectly with file-first systems."
	tag := "test-tag"

	filePath, err := v.WriteMemory(id, content, tag)
	if err != nil {
		t.Fatalf("failed to write memory: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "test-mem-123.okf")
	if filePath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, filePath)
	}

	// Verify file actually exists and contains frontmatter
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	fileContent := string(bytes)
	if !strings.Contains(fileContent, "type: memory") || !strings.Contains(fileContent, "container_tag: test-tag") || !strings.Contains(fileContent, content) {
		t.Errorf("written file does not look like expected OKF: %q", fileContent)
	}

	// Read via vault API
	readContent, err := v.ReadMemory(id)
	if err != nil {
		t.Fatalf("failed to read memory: %v", err)
	}
	if readContent != fileContent {
		t.Errorf("read content does not match written file")
	}

	// Read via absolute path directly
	readContentAbs, err := v.ReadMemory(filePath)
	if err != nil {
		t.Fatalf("failed to read memory by abs path: %v", err)
	}
	if readContentAbs != fileContent {
		t.Errorf("read content by abs path does not match")
	}

	// Delete memory
	err = v.DeleteMemory(id)
	if err != nil {
		t.Fatalf("failed to delete memory: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted, but it still exists")
	}
}
