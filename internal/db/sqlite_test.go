package db

import (
	"os"
	"testing"
	"time"
)

func TestSqliteDB(t *testing.T) {
	// Use an in-memory SQLite database for testing
	dbFile := ":memory:"
	sdb, err := NewSqliteDB(dbFile)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer sdb.Close()

	// 1. Save memory
	m1 := Memory{
		ID:           "mem1",
		Content:      "Alice prefers Node.js for quick MCP bridges.",
		Vector:       []float32{1.0, 0.0, 0.0},
		ContainerTag: "user_alice",
		CreatedAt:    time.Now(),
	}
	if err := sdb.SaveMemory(m1); err != nil {
		t.Fatalf("failed to save memory: %v", err)
	}

	m2 := Memory{
		ID:           "mem2",
		Content:      "Rust provides ultra-fast zero-overhead vector calculations.",
		Vector:       []float32{0.0, 1.0, 0.0},
		ContainerTag: "user_alice",
		CreatedAt:    time.Now(),
	}
	if err := sdb.SaveMemory(m2); err != nil {
		t.Fatalf("failed to save memory: %v", err)
	}

	// 2. List memories
	list, err := sdb.ListMemories("user_alice", 10)
	if err != nil {
		t.Fatalf("failed to list memories: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 memories, got %d", len(list))
	}

	// 3. Search memories (Semantic Search)
	// Query is highly similar to m1
	queryVec := []float32{1.0, 0.1, 0.0}
	results, err := sdb.SearchMemories("user_alice", queryVec, 0.5, 10)
	if err != nil {
		t.Fatalf("failed to search memories: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Memory.ID != "mem1" {
		t.Errorf("expected mem1, got %s", results[0].Memory.ID)
	}

	// 4. Delete memory
	if err := sdb.DeleteMemory("mem1"); err != nil {
		t.Fatalf("failed to delete memory: %v", err)
	}
	list, err = sdb.ListMemories("user_alice", 10)
	if err != nil {
		t.Fatalf("failed to list memories: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 memory after deletion, got %d", len(list))
	}
}

func TestNewSqliteDBError(t *testing.T) {
	// Attempt to open a database at an invalid location to trigger error
	_, err := NewSqliteDB("/nonexistent_dir/db.sqlite")
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestMain(m *testing.M) {
	// We make sure the test runner cleans up any generated databases
	code := m.Run()
	os.Remove("test.db")
	os.Exit(code)
}
