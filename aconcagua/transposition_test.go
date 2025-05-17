package aconcagua

import "testing"

func TestNewTranspositionTable(t *testing.T) {
	tt := NewTranspositionTable(64)

	expected := uint64(64 * 1024 * 1024 / 20)
	got := tt.size

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestProbeFlagExact(t *testing.T) {
	tt := NewTranspositionTable(64)

	tt.store(1, 1, FlagExact, 1)

	got, _ := tt.probe(1, 1, 0, 0)

	expected := 1

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestProbeFlagAlpha(t *testing.T) {
	tt := NewTranspositionTable(64)
	alpha := 0

	tt.store(1, 1, FlagAlpha, 1)

	got, _ := tt.probe(1, 1, alpha, 0)
	expected := alpha

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestProbeFlagBeta(t *testing.T) {
	tt := NewTranspositionTable(64)
	beta := 0

	tt.store(1, 1, FlagBeta, 1)

	got, _ := tt.probe(1, 1, 0, beta)
	expected := beta

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
