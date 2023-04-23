package board

import "testing"

// King moves tests

func TestKingAttacks(t *testing.T) {
	pos := InitialPosition()
	king, _ := pos.PieceAt("e1")

	expected := Bitboard(0b11100000101000)
	got := king.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesToEmptySquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "e4")
	king, _ := pos.PieceAt("e4")

	expectedSquares := []string{"d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"}

	expected := sqaureToBitboard(expectedSquares)
	got := king.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotMoveToAttackedSquare(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "e4")
	pos.AddPiece(BLACK_KNIGHT, "c6")
	king, _ := pos.PieceAt("e4")

	// Cannot move to d4 or e5 because it's attacked by the black knight
	expectedSquares := []string{"d3", "d5", "e3", "f3", "f4", "f5"}

	expected := sqaureToBitboard(expectedSquares)
	got := king.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesWhenInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_KING, "e1")
	pos.AddPiece(BLACK_ROOK, "h1")
	king, _ := pos.PieceAt("e1")

	// Can only move to the second rank, becuase first rank is attacked by the rook, by x rays
	expectedSquares := []string{"d2", "e2", "f2"}

	expected := sqaureToBitboard(expectedSquares)
	got := king.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

