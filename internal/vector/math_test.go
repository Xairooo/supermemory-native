package vector

import (
	"math"
	"testing"
)

func TestDotProduct(t *testing.T) {
	// 1. Equal dimensions
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{4.0, 5.0, 6.0}
	dot, err := DotProduct(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := float32(32.0) // 1*4 + 2*5 + 3*6 = 4+10+18 = 32
	if dot != expected {
		t.Errorf("expected %f, got %f", expected, dot)
	}

	// 2. Unequal dimensions
	_, err = DotProduct(a, []float32{1.0})
	if err == nil {
		t.Error("expected error for unequal dimensions, got nil")
	}
}

func TestL2Norm(t *testing.T) {
	a := []float32{3.0, 4.0}
	norm := L2Norm(a)
	expected := float32(5.0) // sqrt(9 + 16) = 5
	if math.Abs(float64(norm-expected)) > 1e-6 {
		t.Errorf("expected %f, got %f", expected, norm)
	}
}

func TestCosineSimilarity(t *testing.T) {
	// 1. Perfect match (1.0)
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{2.0, 0.0, 0.0}
	sim, err := CosineSimilarity(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(float64(sim-1.0)) > 1e-6 {
		t.Errorf("expected 1.0, got %f", sim)
	}

	// 2. Orthogonal (0.0)
	c := []float32{0.0, 1.0, 0.0}
	sim, err = CosineSimilarity(a, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sim != 0.0 {
		t.Errorf("expected 0.0, got %f", sim)
	}

	// 3. Opposite (-1.0)
	d := []float32{-1.0, 0.0, 0.0}
	sim, err = CosineSimilarity(a, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(float64(sim+1.0)) > 1e-6 {
		t.Errorf("expected -1.0, got %f", sim)
	}

	// 4. Unequal dimensions
	_, err = CosineSimilarity(a, []float32{1.0})
	if err == nil {
		t.Error("expected error for unequal dimensions, got nil")
	}

	// 5. Zero vector boundary check
	zero := []float32{0.0, 0.0, 0.0}
	sim, err = CosineSimilarity(a, zero)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sim != 0.0 {
		t.Errorf("expected 0.0 for zero vector, got %f", sim)
	}
}
