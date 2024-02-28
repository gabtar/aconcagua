package board

import "testing"

// King moves tests

func TestKingAttacks(t *testing.T) {
	kingBB := bitboardFromCoordinate("e1")

	expected := Bitboard(0b11100000101000)
	got := kingAttacks(&kingBB) // The king defends... all pieces around him

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesToEmptySquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e4")
	kingBB := bitboardFromCoordinate("e4")

	expectedSquares := []string{"d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotMoveToAttackedSquare(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e4")
	pos.AddPiece(BlackKnight, "c6")
	kingBB := bitboardFromCoordinate("e4")

	// Cannot move to d4 or e5 because it's attacked by the black knight
	expectedSquares := []string{"d3", "d5", "e3", "f3", "f4", "f5"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesWhenInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(BlackRook, "h1")
	kingBB := bitboardFromCoordinate("e1")

	// Can only move to the second rank, becuase first rank is attacked by the rook, by x rays
	expectedSquares := []string{"d2", "e2", "f2"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingValidMoves(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(BlackRook, "h1")
	kingBB := bitboardFromCoordinate("e1")

	expected := 3
	got := kingMoves(&kingBB, pos, White).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
