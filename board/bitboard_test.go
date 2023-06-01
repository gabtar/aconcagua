package board

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
	bb := Bitboard(ALL_SQUARES)

	expected := 64
	got := bb.count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
