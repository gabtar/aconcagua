package aconcagua

import "testing"

// Rook moves tests
func TestRookAttacksOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "e4")
	rookBB := bitboardFromCoordinate("e4")

	expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"a4", "b4", "c4", "d4", "f4", "g4", "h4"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookAttacks(&rookBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockedSquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "e4")
	pos.AddPiece(WhiteKnight, "c4") // Knight blocking on c4
	rookBB := bitboardFromCoordinate("e4")

	expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"c4", "d4", "f4", "g4", "h4"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookAttacks(&rookBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithAllSquaresBlocked(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "b3")
	pos.AddPiece(WhiteKnight, "b4") // Knight blocking on b4
	pos.AddPiece(WhiteKnight, "b2") // Knight blocking on b5
	pos.AddPiece(WhiteKnight, "a3") // Knight blocking on a3
	pos.AddPiece(WhiteKnight, "c3") // Knight blocking on c3
	rookBB := bitboardFromCoordinate("b3")

	expectedSquares := []string{"b4", "b2", "a3", "c3"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookAttacks(&rookBB, pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "a8")
	pos.AddPiece(WhiteKnight, "a4") // Can move(capture) white knight on a4
	pos.AddPiece(WhiteKnight, "c8") // Can move(capture) knight on c8
	rookBB := bitboardFromCoordinate("a8")

	expectedSquares := []string{"a7", "a6", "a5", "a4", "b8", "c8"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMoves(&rookBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithBlockingPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteRook, "g2")
	pos.AddPiece(BlackKnight, "g4") // Can move(capture) white knight on g4
	pos.AddPiece(WhiteKnight, "f2") // Cannot move to f2, because its blocked by Knight
	rookBB := bitboardFromCoordinate("g2")

	expectedSquares := []string{"g1", "h2", "g3", "g4"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMoves(&rookBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMoveWhenCanBlockCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "c6") // Black king is in check, so only legal move is Re6 blocking the check
	pos.AddPiece(WhiteRook, "e1")
	pos.AddPiece(BlackKing, "e8")
	rookBB := bitboardFromCoordinate("c6")

	expectedSquares := []string{"e6"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMoves(&rookBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenKingInDoubleCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "c6")
	pos.AddPiece(WhiteRook, "e1") // Gives check to the black king on e8
	pos.AddPiece(WhiteRook, "a8") // Gives check to the black king on e8
	pos.AddPiece(BlackKing, "e8")
	rookBB := bitboardFromCoordinate("c6")

	expectedSquares := []string{} // The rook cannot move at all because of the double check in own king

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMoves(&rookBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenTheRookIsPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e8")
	pos.AddPiece(BlackRook, "e4")
	pos.AddPiece(WhiteRook, "e2")
	pos.AddPiece(WhiteRook, "e3")

	// The rook can only move along the e file, becuase it's pinned if moves
	// along the 4 rank, the king will be in check!
	rookBB := bitboardFromCoordinate("e4")

	expectedSquares := []string{"e3", "e5", "e6", "e7"}

	expected := bitboardFromCoordinates(expectedSquares)
	got := rookMoves(&rookBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestNewRookMoves(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "a8")
	pos.AddPiece(WhiteKnight, "a4") // Can move(capture) white knight on a4
	pos.AddPiece(WhiteKnight, "c8") // Can move(capture) knight on c8
	rookBB := bitboardFromCoordinate("a8")
	ml := newMoveList()

	newRookMoves(&rookBB, pos, Black, ml)
	expectedSquares := []string{"a7", "a6", "a5", "a4", "b8", "c8"}

	expected := len(expectedSquares)
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expectedSquares, got)
	}

}
