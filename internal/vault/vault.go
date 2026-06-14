package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Vault struct {
	Dir string
}

// NewVault initializes a new Vault and ensures its directory exists.
func NewVault(dir string) (*Vault, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create vault directory %s: %w", absDir, err)
	}
	return &Vault{Dir: absDir}, nil
}

// WriteMemory writes a memory to a physical OKF file in the vault.
func (v *Vault) WriteMemory(id, content, containerTag string) (string, error) {
	return v.WriteMemoryWithTime(id, content, containerTag, time.Now())
}

// WriteMemoryWithTime writes a memory to a physical OKF file in the vault with a specific timestamp.
func (v *Vault) WriteMemoryWithTime(id, content, containerTag string, t time.Time) (string, error) {
	doc, err := ParseOKF(content)
	if err != nil {
		return "", err
	}

	// Ensure standard OKF frontmatter properties are populated
	doc.SetDefault("type", "memory")
	doc.SetDefault("title", fmt.Sprintf("Memory %s", id))
	doc.SetDefault("container_tag", containerTag)
	doc.SetDefault("timestamp", t.Format(time.RFC3339))

	formatted, err := FormatOKF(doc)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.okf", id)
	filePath := filepath.Join(v.Dir, filename)

	if err := os.WriteFile(filePath, []byte(formatted), 0644); err != nil {
		return "", fmt.Errorf("failed to write OKF file: %w", err)
	}

	return filePath, nil
}

// ReadMemory reads and parses an OKF file by its physical path or ID.
func (v *Vault) ReadMemory(id string) (string, error) {
	var filePath string
	if filepath.IsAbs(id) || strings.Contains(id, string(filepath.Separator)) {
		filePath = id
	} else {
		filePath = filepath.Join(v.Dir, fmt.Sprintf("%s.okf", id))
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(bytes), nil
}

// DeleteMemory deletes the physical file from the vault.
func (v *Vault) DeleteMemory(id string) error {
	filePath := filepath.Join(v.Dir, fmt.Sprintf("%s.okf", id))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete OKF file: %w", err)
	}
	return nil
}

// ListPhysicalFiles returns a list of all memory IDs present in the vault.
func (v *Vault) ListPhysicalFiles() ([]string, error) {
	entries, err := os.ReadDir(v.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".okf") {
			id := strings.TrimSuffix(entry.Name(), ".okf")
			ids = append(ids, id)
		}
	}
	return ids, nil
}
