package embedding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMockProvider(t *testing.T) {
	p := NewMockProvider()
	if p.Dimension() != 384 {
		t.Errorf("expected 384, got %d", p.Dimension())
	}

	// Basic generation
	texts := []string{"hello", "world"}
	embs, err := p.GenerateEmbeddings(texts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embs) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embs))
	}
	if len(embs[0]) != 384 {
		t.Errorf("expected embedding dimension 384, got %d", len(embs[0]))
	}

	// Fail mode
	p.Fail = true
	_, err = p.GenerateEmbeddings(texts)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGeminiProviderMockServer(t *testing.T) {
	// Mock HTTP server to simulate Gemini API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1beta/models/text-embedding-004:batchEmbedContents" {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}

		// Decode body
		var req geminiBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if len(req.Requests) != 2 {
			t.Errorf("expected 2 requests, got %d", len(req.Requests))
		}

		// Mock response
		resp := geminiEmbedResponse{
			Embeddings: []struct {
				Values []float32 "json:\"values\""
			}{
				{Values: make([]float32, 768)},
				{Values: make([]float32, 768)},
			},
		}
		for i := range resp.Embeddings {
			resp.Embeddings[i].Values = make([]float32, 768)
			resp.Embeddings[i].Values[0] = float32(i) + 1.0
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Initialize provider pointing to our mock server
	p := &GeminiProvider{
		APIKey:  "fake_key",
		Model:   "models/text-embedding-004",
		BaseURL: server.URL,
		client:  server.Client(),
	}

	embs, err := p.GenerateEmbeddings([]string{"hello", "world"})
	if err != nil {
		t.Fatalf("failed to generate embeddings: %v", err)
	}

	if len(embs) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embs))
	}
	if embs[0][0] != 1.0 || embs[1][0] != 2.0 {
		t.Errorf("unexpected embedding values: got %v and %v", embs[0][0], embs[1][0])
	}
}

func TestOllamaProvider(t *testing.T) {
	// Mock HTTP server to simulate Ollama API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}

		// Decode body
		var req ollamaEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		// Mock response
		resp := ollamaEmbedResponse{
			Embedding: make([]float32, 768),
		}
		resp.Embedding[0] = 42.0

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OllamaProvider{
		Model:   "nomic-embed-text",
		BaseURL: server.URL,
		client:  server.Client(),
	}

	if p.Dimension() != 768 {
		t.Errorf("expected 768, got %d", p.Dimension())
	}

	pLarge := &OllamaProvider{
		Model:   "mxbai-embed-large",
		BaseURL: server.URL,
		client:  server.Client(),
	}
	if pLarge.Dimension() != 1024 {
		t.Errorf("expected 1024, got %d", pLarge.Dimension())
	}

	embs, err := p.GenerateEmbeddings([]string{"hello", "world"})
	if err != nil {
		t.Fatalf("failed to generate embeddings: %v", err)
	}

	if len(embs) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embs))
	}
	if embs[0][0] != 42.0 || embs[1][0] != 42.0 {
		t.Errorf("unexpected embedding values: got %f", embs[0][0])
	}
}

