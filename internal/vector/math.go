package vector

import (
	"errors"
	"math"
)

// DotProduct calculates the dot product of two vectors
func DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vector dimensions must be equal")
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot, nil
}

// L2Norm calculates the L2 norm (magnitude) of a vector
func L2Norm(a []float32) float32 {
	var sum float32
	for _, val := range a {
		sum += val * val
	}
	return float32(math.Sqrt(float64(sum)))
}

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vector dimensions must be equal")
	}
	dot, err := DotProduct(a, b)
	if err != nil {
		return 0, err
	}
	normA := L2Norm(a)
	normB := L2Norm(b)
	if normA == 0 || normB == 0 {
		return 0, nil // Avoid division by zero
	}
	return dot / (normA * normB), nil
}
