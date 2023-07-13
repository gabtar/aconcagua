package board

import "testing"

func TestKnightAttacks(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKnight, "e4")
	knightBB := squareToBitboard([]string{"e4"})

	expectedSquares := []string{"d6", "f6", "d2", "f2", "g5", "g3", "c5", "c3"}

	expected := squareToBitboard(expectedSquares)
	got := knightAttacks(&knightBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenBlockedBySameColorPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKnight, "e4")
	pos.AddPiece(BlackRook, "d6")
	pos.AddPiece(BlackRook, "f6")
	pos.AddPiece(BlackKing, "d2")
	pos.AddPiece(BlackBishop, "f2")
	knightBB := squareToBitboard([]string{"e4"})

	expectedSquares := []string{"g5", "g3", "c5", "c3"}

	expected := squareToBitboard(expectedSquares)
	got := knightMoves(&knightBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKnight, "b1")
	pos.AddPiece(BlackBishop, "c3")
	pos.AddPiece(WhiteRook, "a3")
	pos.AddPiece(WhiteRook, "d2") // Blocks Knight move
	knightBB := squareToBitboard([]string{"b1"})

	expectedSquares := []string{"c3"} // The Knight can only capture the bishop. "a3" and "d2" are blocked by the rook, so it cannot move there

	expected := squareToBitboard(expectedSquares)
	got := knightMoves(&knightBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKnight, "e4")
	pos.AddPiece(BlackRook, "e8")
	pos.AddPiece(WhiteKing, "e1")
	knightBB := squareToBitboard([]string{"e4"})

	expectedSquares := []string{} // The Knight is pinned, it cannot move at all

	expected := squareToBitboard(expectedSquares)
	got := knightMoves(&knightBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
