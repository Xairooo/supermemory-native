package db

import (
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	_ "modernc.org/sqlite"

	"github.com/supermemory-native/supermemory-native/internal/vector"
)

type Memory struct {
	ID           string    `json:"id"`
	Content      string    `json:"content"`
	Vector       []float32 `json:"vector"`
	ContainerTag string    `json:"container_tag"`
	CreatedAt    time.Time `json:"created_at"`
}

type SearchResult struct {
	Memory     Memory  `json:"memory"`
	Similarity float32 `json:"similarity"`
}

type SqliteDB struct {
	db *sql.DB
}

func NewSqliteDB(dbPath string) (*SqliteDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		vector TEXT NOT NULL,
		container_tag TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_memories_container ON memories(container_tag);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return &SqliteDB{db: db}, nil
}

func (s *SqliteDB) Close() error {
	return s.db.Close()
}

func (s *SqliteDB) SaveMemory(m Memory) error {
	vectorJSON, err := json.Marshal(m.Vector)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO memories (id, content, vector, container_tag, created_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		content = excluded.content,
		vector = excluded.vector,
		container_tag = excluded.container_tag,
		created_at = excluded.created_at;
	`
	_, err = s.db.Exec(query, m.ID, m.Content, string(vectorJSON), m.ContainerTag, m.CreatedAt)
	return err
}

func (s *SqliteDB) DeleteMemory(id string) error {
	_, err := s.db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

func (s *SqliteDB) ListMemories(containerTag string, limit int) ([]Memory, error) {
	rows, err := s.db.Query("SELECT id, content, vector, container_tag, created_at FROM memories WHERE container_tag = ? ORDER BY created_at DESC LIMIT ?", containerTag, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		var vectorStr string
		if err := rows.Scan(&m.ID, &m.Content, &vectorStr, &m.ContainerTag, &m.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(vectorStr), &m.Vector); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, nil
}

func (s *SqliteDB) SearchMemories(containerTag string, queryVec []float32, threshold float32, limit int) ([]SearchResult, error) {
	rows, err := s.db.Query("SELECT id, content, vector, container_tag, created_at FROM memories WHERE container_tag = ?", containerTag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var m Memory
		var vectorStr string
		if err := rows.Scan(&m.ID, &m.Content, &vectorStr, &m.ContainerTag, &m.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(vectorStr), &m.Vector); err != nil {
			return nil, err
		}

		// Calculate similarity using our ultra-fast pure Go vector package
		sim, err := vector.CosineSimilarity(queryVec, m.Vector)
		if err != nil {
			return nil, err
		}

		if sim >= threshold {
			results = append(results, SearchResult{
				Memory:     m,
				Similarity: sim,
			})
		}
	}

	// Sort results in descending order of similarity
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Slice by limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
