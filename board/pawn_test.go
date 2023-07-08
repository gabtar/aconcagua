package board

import (
	"testing"
)

// Pawn moves tests

func TestPawnAttacks(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "e2")
	pawn, _ := pos.PieceAt("e2")
	pawnBB := pawn.Square()

	expectedSquares := []string{"d3", "f3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnAttacks(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnAttacksOnEdgeFiles(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "h2")
	pawn, _ := pos.PieceAt("h2")
	pawnBB := pawn.Square()

	expectedSquares := []string{"g3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnAttacks(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "e2")
	pawn, _ := pos.PieceAt("e2")
	pawnBB := pawn.Square()

	expectedSquares := []string{"e3", "e4"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesWithCapturesFrom7thRank(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_PAWN, "b7")
	pos.AddPiece(WHITE_BISHOP, "a6")
	pos.AddPiece(BLACK_KNIGHT, "c6")
	pawn, _ := pos.PieceAt("b7")
	pawnBB := pawn.Square()

	// Can capture white bishop on a6 and is blocked by black knight on c6
	// Can also move to b6 and b7
	expectedSquares := []string{"a6", "b6", "b5"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, BLACK)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanBlockACheckOnFirstMove(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "f2")
	pos.AddPiece(BLACK_ROOK, "h4")
	pos.AddPiece(WHITE_KING, "c4")
	pawn, _ := pos.PieceAt("f2")
	pawnBB := pawn.Square()

	// The only legal move of the pawn is to block the check on f4
	expectedSquares := []string{"f4"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanOnlyMoveInThePinnedDirection(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "f2")
	pos.AddPiece(BLACK_BISHOP, "e3")
	pos.AddPiece(WHITE_KING, "g1")
	pawn, _ := pos.PieceAt("f2")
	pawnBB := pawn.Square()

	// The only legal move of the pawn is to capture the bishop on e3
	expectedSquares := []string{"e3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPinnedAndInCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_PAWN, "f2")
	pos.AddPiece(BLACK_BISHOP, "e3")
	pos.AddPiece(BLACK_ROOK, "g8")
	pos.AddPiece(WHITE_KING, "g1")
	pawn, _ := pos.PieceAt("f2")
	pawnBB := pawn.Square()

	expectedSquares := []string{}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, WHITE)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnInA4Moves(t *testing.T) {
	pos := From("rnbqkbnr/1ppppppp/8/8/p7/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	pawn, _ := pos.PieceAt("a4")
	pawnBB := pawn.Square()

	expectedSquares := []string{"a3"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, BLACK)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPinnedIfCapturesThePinnedPiece(t *testing.T) {
	pos := From("r1bqkbnr/7p/2p1p1p1/p1pp1p1Q/P4P2/3PP3/1PPBN1PP/RN3RK1 b kq - 1 9")
	pawn, _ := pos.PieceAt("g6")
	pawnBB := pawn.Square()

	expectedSquares := []string{"h5"}

	expected := squareToBitboard(expectedSquares)
	got := pawnMoves(&pawnBB, pos, BLACK)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
