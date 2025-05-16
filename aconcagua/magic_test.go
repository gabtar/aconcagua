package aconcagua

import "testing"

func TestRookRelevantMask(t *testing.T) {
	expectedSquares := []string{"e2", "e3", "e5", "e6", "e7",
		"b4", "c4", "d4", "f4", "g4"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMask(28)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopRelevantMask(t *testing.T) {
	expectedSquares := []string{"c2", "d3", "f5", "g6", "b7", "c6", "d5", "f3", "g2"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMask(28)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRook(t *testing.T) {
	// pos := EmptyPosition()

	// bishopBB := bitboardFromCoordinate("c4")
	expectedSquares := []string{"a2", "b3", "d3", "d5", "e6", "f7", "b5", "a6"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rook(32, Bitboard(1<<35))

	got.Print()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockers(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteRook, "c4")
	pos.AddPiece(WhiteBishop, "b4")

	expectedSquares := []string{"c1", "c2", "c3", "c5", "c6", "c7", "c8",
		"b4", "d4", "e4", "f4", "g4", "h4"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rooksAttacksWithBlockers(26, pos.Pieces(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
