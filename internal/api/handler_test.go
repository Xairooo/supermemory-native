package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/supermemory-native/supermemory-native/internal/db"
	"github.com/supermemory-native/supermemory-native/internal/embedding"
	"github.com/supermemory-native/supermemory-native/internal/memory"
	"github.com/supermemory-native/supermemory-native/internal/vault"
)

func setupTestHandler(t *testing.T) (*Handler, func()) {
	sdb, err := db.NewSqliteDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	tempVaultDir, err := os.MkdirTemp("", "supermemory_api_vault_test_*")
	if err != nil {
		sdb.Close()
		t.Fatalf("failed to create temp vault dir: %v", err)
	}

	v, err := vault.NewVault(tempVaultDir)
	if err != nil {
		sdb.Close()
		os.RemoveAll(tempVaultDir)
		t.Fatalf("failed to init vault: %v", err)
	}

	prov := embedding.NewMockProvider()
	eng := memory.NewEngine(sdb, prov, v)
	handler := NewHandler(eng)

	return handler, func() {
		sdb.Close()
		os.RemoveAll(tempVaultDir)
	}
}

func TestHandlerOPTIONSAndMethodNotAllowed(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// 1. OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/v3/documents", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}

	// 2. GET request (Method Not Allowed)
	req = httptest.NewRequest("GET", "/v3/documents", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandlerNotFound(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/v3/invalid_path", bytes.NewBuffer([]byte(`{}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandlerAddDocument(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// 1. Successful add
	reqBody := addRequest{
		Content:      "Alice prefers Node.js for quick MCP bridges.",
		ContainerTag: "user_alice",
	}
	payload, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v3/documents", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool   `json:"success"`
		ID      string `json:"id"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.ID == "" {
		t.Error("expected non-empty ID")
	}

	// 2. Bad JSON request
	req = httptest.NewRequest("POST", "/v3/documents", bytes.NewBuffer([]byte(`{broken`)))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandlerProfileAndSearch(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add a dummy document first
	_, err := handler.Engine.AddMemory("Alice prefers Node.js for quick MCP bridges.", "user_alice")
	if err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	// 1. Profile query
	reqBody := queryRequest{
		Query:        "What language does Alice prefer?",
		ContainerTag: "user_alice",
	}
	payload, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v4/profile", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var profileResp struct {
		Profile struct {
			Dynamic []string `json:"dynamic"`
		} `json:"profile"`
		SearchResults struct {
			Total int `json:"total"`
		} `json:"searchResults"`
	}
	json.NewDecoder(rr.Body).Decode(&profileResp)
	if profileResp.SearchResults.Total == 0 {
		t.Error("expected non-zero search results")
	}
	if len(profileResp.Profile.Dynamic) == 0 {
		t.Error("expected non-empty dynamic profile facts")
	}

	// 2. Search query (/v4/search)
	req = httptest.NewRequest("POST", "/v4/search", bytes.NewBuffer(payload))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var searchResp struct {
		Results []map[string]interface{} `json:"results"`
		Total   int                      `json:"total"`
	}
	json.NewDecoder(rr.Body).Decode(&searchResp)
	if searchResp.Total == 0 {
		t.Error("expected non-zero search results")
	}

	// 3. Bad JSON on search
	req = httptest.NewRequest("POST", "/v4/search", bytes.NewBuffer([]byte(`{broken`)))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}

	// 4. Bad JSON on profile
	req = httptest.NewRequest("POST", "/v4/profile", bytes.NewBuffer([]byte(`{broken`)))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandlerSync(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// Hit the sync endpoint
	req := httptest.NewRequest("POST", "/v3/sync", bytes.NewBuffer([]byte(`{}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.Message != "Vault synchronization completed successfully" {
		t.Errorf("unexpected success message: %q", resp.Message)
	}
}
