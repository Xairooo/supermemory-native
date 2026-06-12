package memory

import (
	"testing"

	"github.com/supermemory-native/supermemory-native/internal/db"
	"github.com/supermemory-native/supermemory-native/internal/embedding"
)

func TestEngine(t *testing.T) {
	// Initialize in-memory DB
	sdb, err := db.NewSqliteDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer sdb.Close()

	// Initialize mock provider
	prov := embedding.NewMockProvider()

	// Initialize engine
	eng := NewEngine(sdb, prov)

	// 1. Add memory
	id1, err := eng.AddMemory("Alice prefers Node.js for quick MCP bridges.", "user_alice")
	if err != nil {
		t.Fatalf("failed to add memory: %v", err)
	}
	if id1 == "" {
		t.Error("expected non-empty memory ID")
	}

	id2, err := eng.AddMemory("Rust has zero-cost abstractions.", "user_alice")
	if err != nil {
		t.Fatalf("failed to add memory: %v", err)
	}
	if id2 == "" {
		t.Error("expected non-empty memory ID")
	}

	// Boundary check: empty content
	_, err = eng.AddMemory("", "user_alice")
	if err == nil {
		t.Error("expected error for empty content, got nil")
	}

	// 2. Query memories
	results, err := eng.QueryMemories("What language does Alice prefer?", "user_alice", 0.1, 10)
	if err != nil {
		t.Fatalf("failed to query memories: %v", err)
	}

	// Ensure we got results
	if len(results) == 0 {
		t.Error("expected at least 1 query result, got 0")
	}

	// Ensure sorting and matching works
	found := false
	for _, res := range results {
		if res.Memory.ID == id1 || res.Memory.ID == id2 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find one of the saved memories in results")
	}

	// Boundary check: empty query
	emptyResults, err := eng.QueryMemories("", "user_alice", 0.1, 10)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if len(emptyResults) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(emptyResults))
	}

	// 3. Provider failure modes (for 100% code coverage)
	prov.Fail = true
	_, err = eng.AddMemory("broken", "user_alice")
	if err == nil {
		t.Error("expected error during provider failure on add, got nil")
	}

	_, err = eng.QueryMemories("broken", "user_alice", 0.1, 10)
	if err == nil {
		t.Error("expected error during provider failure on query, got nil")
	}
}
