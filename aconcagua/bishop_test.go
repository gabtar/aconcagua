package aconcagua

import (
	"testing"
)

// Bishop moves tests
func TestBishopAttacksOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteBishop, "h1")
	bishopBB := bitboardFromCoordinate("h1")

	expectedSquares := []string{"g2", "f3", "e4", "d5", "c6", "b7", "a8"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopAttacks(&bishopBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopAttacksWithBlockedSquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteBishop, "e3")
	pos.AddPiece(BlackRook, "g5")
	bishopBB := bitboardFromCoordinate("e3")

	expectedSquares := []string{"f2", "g1", "d4", "c5", "b6", "a7", "f4", "g5", "d2", "c1"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopAttacks(&bishopBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackBishop, "c4")
	pos.AddPiece(WhiteRook, "f7")   // Can move(capture) white rook on f7
	pos.AddPiece(WhiteKnight, "d3") // Can move(capture) knight on d3
	bishopBB := bitboardFromCoordinate("c4")

	expectedSquares := []string{"a2", "b3", "d3", "d5", "e6", "f7", "b5", "a6"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMoves(&bishopBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithBlockingPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteBishop, "g6")
	pos.AddPiece(WhiteKnight, "e8") // Cannot move, blocked by same color knight
	pos.AddPiece(WhiteRook, "f5")   // Cannot move to f5, because its blocked by Rook
	bishopBB := bitboardFromCoordinate("g6")

	expectedSquares := []string{"h7", "h5", "f7"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMoves(&bishopBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMoveWhenCanBlockCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "d1") // White king is in check, only legal move is Bd2
	pos.AddPiece(BlackRook, "d8") // And also Bxd8 by capturing the Rook which is checking the king
	pos.AddPiece(WhiteBishop, "g5")
	bishopBB := bitboardFromCoordinate("g5")

	expectedSquares := []string{"d2", "d8"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMoves(&bishopBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWhenPinnedAndInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "d1")
	pos.AddPiece(BlackRook, "h1") // Gives check to the white king on d1
	pos.AddPiece(BlackRook, "d8") // Gives check to the white king on d1 (by xrays) -> pins the bishop
	pos.AddPiece(WhiteBishop, "d4")
	bishopBB := bitboardFromCoordinate("d4")

	expectedSquares := []string{} // The bishop cannot move at all, because of the double check

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMoves(&bishopBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishpMovesWhenTheBishopIsPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "c4")
	pos.AddPiece(BlackBishop, "g8")
	pos.AddPiece(WhiteBishop, "d5")

	bishopBB := bitboardFromCoordinate("d5")

	expectedSquares := []string{"e6", "f7", "g8"} // Can only move along the g8 c4 diagonal because of the pin

	expected := bitboardFromCoordinates(expectedSquares)
	got := bishopMoves(&bishopBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNewBishopMoves(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteBishop, "g6")
	pos.AddPiece(WhiteKnight, "e8") // Cannot move, blocked by same color knight
	pos.AddPiece(WhiteRook, "f5")   // Cannot move to f5, because its blocked by Rook
	bishopBB := bitboardFromCoordinate("g6")

	expectedSquares := []string{"h7", "h5", "f7"}

	expected := len(expectedSquares)
	got := len(newBishopMoves(&bishopBB, pos, White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
