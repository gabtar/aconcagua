package aconcagua

import (
	"testing"
)

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

func TestKingCanCastleShort(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(WhiteRook, "h1")
	pos.castlingRights = Kk
	kingBB := bitboardFromCoordinate("e1")

	expected := true
	got := canCastleShort(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCanCastleShortTwo(t *testing.T) {
	pos := From("rnb2k1r/pp1Pbppp/2p5/q7/2B5/8/PPPQNnPP/RNB1K2R w KQ - 3 9")
	pos.castlingRights = Kk
	kingBB := bitboardFromCoordinate("e1")

	expected := true
	got := canCastleShort(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleShortIfPathBlocked(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(WhiteRook, "h1")
	pos.AddPiece(WhiteKnight, "f1")
	pos.castlingRights = Kk
	kingBB := bitboardFromCoordinate("e1")

	expected := false
	got := canCastleShort(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleShortIfItsInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(WhiteRook, "h1")
	pos.AddPiece(BlackRook, "f4")
	pos.castlingRights = Kk
	kingBB := bitboardFromCoordinate("e1")

	expected := false
	got := canCastleShort(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCanCastleLong(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(WhiteRook, "a1")
	pos.castlingRights = Qq
	kingBB := bitboardFromCoordinate("e1")

	expected := true
	got := canCastleLong(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackKingCannotCastleLongIfPathBlocked(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e8")
	pos.AddPiece(BlackRook, "a8")
	pos.AddPiece(BlackKnight, "b8")
	pos.castlingRights = Qq
	kingBB := bitboardFromCoordinate("e8")

	expected := false
	got := canCastleLong(&kingBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleLongIfItsInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e8")
	pos.AddPiece(BlackRook, "a8")
	pos.AddPiece(WhiteRook, "c6")
	pos.castlingRights = Qq
	kingBB := bitboardFromCoordinate("e8")

	expected := false
	got := canCastleLong(&kingBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNewKingMoves(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(WhiteRook, "h1")
	pos.AddPiece(WhiteRook, "a1")
	pos.castlingRights = KQ
	kingBB := bitboardFromCoordinate("e1")

	expected := 7
	got := len(newKingMoves(&kingBB, pos, White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
