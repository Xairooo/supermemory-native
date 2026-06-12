package embedding

import "errors"

type MockProvider struct {
	Fail bool
}

func NewMockProvider() *MockProvider {
	return &MockProvider{Fail: false}
}

func (m *MockProvider) Dimension() int {
	return 384 // Similar to MiniLM-L6-v2
}

func (m *MockProvider) GenerateEmbeddings(texts []string) ([][]float32, error) {
	if m.Fail {
		return nil, errors.New("mock embedding error")
	}
	out := make([][]float32, len(texts))
	for i, text := range texts {
		emb := make([]float32, m.Dimension())
		// Deterministic dummy vector based on string length and first character
		var base float32
		if len(text) > 0 {
			base = float32(text[0]) / 255.0
		}
		for j := range emb {
			emb[j] = base + float32(j)*0.001
		}
		out[i] = emb
	}
	return out, nil
}
