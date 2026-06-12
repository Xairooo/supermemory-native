package memory

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/supermemory-native/supermemory-native/internal/db"
	"github.com/supermemory-native/supermemory-native/internal/embedding"
)

type Engine struct {
	DB       *db.SqliteDB
	Provider embedding.EmbeddingProvider
}

func NewEngine(database *db.SqliteDB, provider embedding.EmbeddingProvider) *Engine {
	return &Engine{
		DB:       database,
		Provider: provider,
	}
}

// GenerateUUID v4 helper for standalone execution
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (e *Engine) AddMemory(content, containerTag string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	embeddings, err := e.Provider.GenerateEmbeddings([]string{content})
	if err != nil {
		return "", err
	}
	if len(embeddings) == 0 {
		return "", fmt.Errorf("failed to generate embeddings for content")
	}

	memID := GenerateUUID()
	mem := db.Memory{
		ID:           memID,
		Content:      content,
		Vector:       embeddings[0],
		ContainerTag: containerTag,
		CreatedAt:    time.Now(),
	}

	if err := e.DB.SaveMemory(mem); err != nil {
		return "", err
	}

	return memID, nil
}

func (e *Engine) QueryMemories(query, containerTag string, threshold float32, limit int) ([]db.SearchResult, error) {
	if query == "" {
		return []db.SearchResult{}, nil
	}

	embeddings, err := e.Provider.GenerateEmbeddings([]string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("failed to generate embeddings for query")
	}

	return e.DB.SearchMemories(containerTag, embeddings[0], threshold, limit)
}
