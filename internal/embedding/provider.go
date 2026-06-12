package embedding

// EmbeddingProvider defines the interface for creating vector embeddings from plain text
type EmbeddingProvider interface {
	GenerateEmbeddings(texts []string) ([][]float32, error)
	Dimension() int
}
