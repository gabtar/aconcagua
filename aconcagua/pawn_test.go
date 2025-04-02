package aconcagua

import (
	"testing"
)

// Pawn moves tests

func TestPawnAttacks(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "e2")
	pawnBB := bitboardFromCoordinate("e2")

	expectedSquares := []string{"d3", "f3"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := pawnAttacks(&pawnBB, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnAttacksOnEdgeFiles(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "h2")
	// pawn, _ := pos.PieceAt("h2")
	pawnBB := bitboardFromCoordinate("h2")

	expectedSquares := []string{"g3"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := pawnAttacks(&pawnBB, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "e2")
	pawnBB := bitboardFromCoordinate("e2")

	expectedSquares := []string{"e3", "e4"}

	expected := bitboardFromCoordinates(expectedSquares)
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
	pawnBB := bitboardFromCoordinate("b7")

	// Can capture white bishop on a6 and is blocked by black knight on c6
	// Can also move to b6 and b7
	expectedSquares := []string{"a6", "b6", "b5"}

	expected := bitboardFromCoordinates(expectedSquares)
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
	pawnBB := bitboardFromCoordinate("f2")

	// The only legal move of the pawn is to block the check on f4
	expectedSquares := []string{"f4"}

	expected := bitboardFromCoordinates(expectedSquares)
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
	pawnBB := bitboardFromCoordinate("f2")

	// The only legal move of the pawn is to capture the bishop on e3
	expectedSquares := []string{"e3"}

	expected := bitboardFromCoordinates(expectedSquares)
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
	pawnBB := bitboardFromCoordinate("f2")

	expectedSquares := []string{}

	expected := bitboardFromCoordinates(expectedSquares)
	got := pawnMoves(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnInA4Moves(t *testing.T) {
	pos := From("rnbqkbnr/1ppppppp/8/8/p7/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	pawnBB := bitboardFromCoordinate("a4")

	expected := bitboardFromCoordinate("a3")
	got := pawnMoves(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPinnedIfCapturesThePinnedPiece(t *testing.T) {
	pos := From("r1bqkbnr/7p/2p1p1p1/p1pp1p1Q/P4P2/3PP3/1PPBN1PP/RN3RK1 b kq - 1 9")
	pawnBB := bitboardFromCoordinate("g6")

	expected := bitboardFromCoordinate("h5")
	got := pawnMoves(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNewPawnsMoves(t *testing.T) {
	pos := InitialPosition()
	pawnBB := bitboardFromCoordinate("e2")
	ml := newMoveList()

	expected := 2
	newPawnMoves(&pawnBB, pos, White, ml)
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNewPawnsMovesPromo(t *testing.T) {
	pos := From("8/7P/2k5/8/8/8/8/4K3 w - - 0 1")
	pawnBB := bitboardFromCoordinate("h7")
	ml := newMoveList()

	expected := 4
	newPawnMoves(&pawnBB, pos, White, ml)
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGenEpPawnCaptures(t *testing.T) {
	ml := newMoveList()

	pos := From("4r3/8/8/R7/3Pp2k/8/8/4K3 b - d3 0 1")

	genEpPawnCaptures(pos, Black, ml)

	expected := 1
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
