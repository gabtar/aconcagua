package board

import (
	"testing"
)

// Pawn moves tests

func TestPawnAttacks(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "e2")
	// pawn, _ := pos.PieceAt("e2")
	pawnBB := squareToBitboard([]string{"e2"})

	expectedSquares := []string{"d3", "f3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnAttacks(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnAttacksOnEdgeFiles(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "h2")
	// pawn, _ := pos.PieceAt("h2")
	pawnBB := squareToBitboard([]string{"h2"})

	expectedSquares := []string{"g3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnAttacks(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "e2")
	pawnBB := squareToBitboard([]string{"e2"})

	expectedSquares := []string{"e3", "e4"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesWithCapturesFrom7thRank(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackPawn, "b7")
	pos.AddPiece(WhiteBishop, "a6")
	pos.AddPiece(BlackKnight, "c6")
	pawnBB := squareToBitboard([]string{"b7"})

	// Can capture white bishop on a6 and is blocked by black knight on c6
	// Can also move to b6 and b7
	expectedSquares := []string{"a6", "b6", "b5"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanBlockACheckOnFirstMove(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "f2")
	pos.AddPiece(BlackRook, "h4")
	pos.AddPiece(WhiteKing, "c4")
	pawnBB := squareToBitboard([]string{"f2"})

	// The only legal move of the pawn is to block the check on f4
	expectedSquares := []string{"f4"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanOnlyMoveInThePinnedDirection(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "f2")
	pos.AddPiece(BlackBishop, "e3")
	pos.AddPiece(WhiteKing, "g1")
	pawnBB := squareToBitboard([]string{"f2"})

	// The only legal move of the pawn is to capture the bishop on e3
	expectedSquares := []string{"e3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPinnedAndInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "f2")
	pos.AddPiece(BlackBishop, "e3")
	pos.AddPiece(BlackRook, "g8")
	pos.AddPiece(WhiteKing, "g1")
	pawnBB := squareToBitboard([]string{"f2"})

	expectedSquares := []string{}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnInA4Moves(t *testing.T) {
	pos := From("rnbqkbnr/1ppppppp/8/8/p7/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	pawnBB := squareToBitboard([]string{"a4"})

	expectedSquares := []string{"a3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPinnedIfCapturesThePinnedPiece(t *testing.T) {
	pos := From("r1bqkbnr/7p/2p1p1p1/p1pp1p1Q/P4P2/3PP3/1PPBN1PP/RN3RK1 b kq - 1 9")
	pawnBB := squareToBitboard([]string{"g6"})

	expectedSquares := []string{"h5"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
