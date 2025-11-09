package engine

import "testing"

func TestCountPiecesInBitboard(t *testing.T) {
	bb := Bitboard(0)

	expected := 0
	got := bb.count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCountPiecesInBitboard2(t *testing.T) {
	bb := Bitboard(AllSquares)

	expected := 64
	got := bb.count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
