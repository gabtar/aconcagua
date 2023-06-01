package board

import "testing"

func TestKnightAttacks(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KNIGHT, "e4")
	knight, _ := pos.PieceAt("e4")

	expectedSquares := []string{"d6", "f6", "d2", "f2", "g5", "g3", "c5", "c3"}

	expected := squareToBitboard(expectedSquares)
	got := knight.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenBlockedBySameColorPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KNIGHT, "e4")
	pos.AddPiece(BLACK_ROOK, "d6")
	pos.AddPiece(BLACK_ROOK, "f6")
	pos.AddPiece(BLACK_KING, "d2")
	pos.AddPiece(BLACK_BISHOP, "f2")
	knight, _ := pos.PieceAt("e4")

	expectedSquares := []string{"g5", "g3", "c5", "c3"}

	expected := squareToBitboard(expectedSquares)
	got := knight.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KNIGHT, "b1")
	pos.AddPiece(BLACK_BISHOP, "c3")
	pos.AddPiece(WHITE_ROOK, "a3")
	pos.AddPiece(WHITE_ROOK, "d2") // Blocks Knight move
	knight, _ := pos.PieceAt("b1")

	expectedSquares := []string{"c3"} // The Knight can only capture the bishop. "a3" and "d2" are blocked by the rook, so it cannot move there

	expected := squareToBitboard(expectedSquares)
	got := knight.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KNIGHT, "e4")
	pos.AddPiece(BLACK_ROOK, "e8")
	pos.AddPiece(WHITE_KING, "e1")
	knight, _ := pos.PieceAt("e4")

	expectedSquares := []string{} // The Knight is pinned, it cannot move at all

	expected := squareToBitboard(expectedSquares)
	got := knight.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
