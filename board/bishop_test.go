package board

import "testing"

// Bishop moves tests
func TestBishopAttacksOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_BISHOP, "h1")
	bishop, _ := pos.PieceAt("h1")

	expectedSquares := []string{"g2", "f3", "e4", "d5", "c6", "b7", "a8"}

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopAttacksWithBlockedSquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_BISHOP, "e3")
	pos.AddPiece(BLACK_ROOK, "g5")
	bishop, _ := pos.PieceAt("e3")

	expectedSquares := []string{"f2", "g1", "d4", "c5", "b6", "a7", "f4", "g5", "d2", "c1"}

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_BISHOP, "c4")
	pos.AddPiece(WHITE_ROOK, "f7")   // Can move(capture) white rook on f7
	pos.AddPiece(WHITE_KNIGHT, "d3") // Can move(capture) knight on d3
	bishop, _ := pos.PieceAt("c4")

	expectedSquares := []string{"a2", "b3", "d3", "d5", "e6", "f7", "b5", "a6"}

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithBlockingPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_BISHOP, "g6")
	pos.AddPiece(WHITE_KNIGHT, "e8") // Cannot move, blocked by same color knight
	pos.AddPiece(WHITE_ROOK, "f5")   // Cannot move to f5, because its blocked by Rook
	bishop, _ := pos.PieceAt("g6")

	expectedSquares := []string{"h7", "h5", "f7"}

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMoveWhenCanBlockCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "d1") // White king is in check, only legal move is Bd2
	pos.AddPiece(BLACK_ROOK, "d8") // And also Bxd8 by capturing the Rook which is checking the king
	pos.AddPiece(WHITE_BISHOP, "g5")
	bishop, _ := pos.PieceAt("g5")

	expectedSquares := []string{"d2", "d8"}

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWhenPinnedAndInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "d1")
	pos.AddPiece(BLACK_ROOK, "h1") // Gives check to the white king on d1
	pos.AddPiece(BLACK_ROOK, "d8") // Gives check to the white king on d1 (by xrays) -> pins the bishop
	pos.AddPiece(WHITE_BISHOP, "d4")
	bishop, _ := pos.PieceAt("d4")

	expectedSquares := []string{} // The bishop cannot move at all, because of the double check

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishpMovesWhenTheBishopIsPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "c4")
	pos.AddPiece(BLACK_BISHOP, "g8")
	pos.AddPiece(WHITE_BISHOP, "d5")

	bishop, _ := pos.PieceAt("d5")
	expectedSquares := []string{"e6", "f7", "g8"} // Can only move along the g8 c4 diagonal because of the pin

	expected := sqaureToBitboard(expectedSquares)
	got := bishop.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
