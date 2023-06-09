package board

import "testing"

// Rook moves tests
func TestRookAttacksOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "e4")
	rook, _ := pos.PieceAt("e4")

	expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"a4", "b4", "c4", "d4", "f4", "g4", "h4"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockedSquares(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "e4")
	pos.AddPiece(WHITE_KNIGHT, "c4") // Knight blocking on c4
	rook, _ := pos.PieceAt("e4")

	expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"c4", "d4", "f4", "g4", "h4"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithAllSquaresBlocked(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "b3")
	pos.AddPiece(WHITE_KNIGHT, "b4") // Knight blocking on b4
	pos.AddPiece(WHITE_KNIGHT, "b2") // Knight blocking on b5
	pos.AddPiece(WHITE_KNIGHT, "a3") // Knight blocking on a3
	pos.AddPiece(WHITE_KNIGHT, "c3") // Knight blocking on c3
	rook, _ := pos.PieceAt("b3")

	expectedSquares := []string{"b4", "b2", "a3", "c3"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithCaptures(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "a8")
	pos.AddPiece(WHITE_KNIGHT, "a4") // Can move(capture) white knight on a4
	pos.AddPiece(WHITE_KNIGHT, "c8") // Can move(capture) knight on c8
	rook, _ := pos.PieceAt("a8")

	expectedSquares := []string{"a7", "a6", "a5", "a4", "b8", "c8"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithBlockingPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WHITE_ROOK, "g2")
	pos.AddPiece(BLACK_KNIGHT, "g4") // Can move(capture) white knight on g4
	pos.AddPiece(WHITE_KNIGHT, "f2") // Cannot move to f2, because its blocked by Knight
	rook, _ := pos.PieceAt("g2")

	expectedSquares := []string{"g1", "h2", "g3", "g4"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMoveWhenCanBlockCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "c6") // Black king is in check, so only legal move is Re6 blocking the check
	pos.AddPiece(WHITE_ROOK, "e1")
	pos.AddPiece(BLACK_KING, "e8")
	rook, _ := pos.PieceAt("c6")

	expectedSquares := []string{"e6"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenKingInDoubleCheck(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "c6")
	pos.AddPiece(WHITE_ROOK, "e1") // Gives check to the black king on e8
	pos.AddPiece(WHITE_ROOK, "a8") // Gives check to the black king on e8
	pos.AddPiece(BLACK_KING, "e8")
	rook, _ := pos.PieceAt("c6")

	expectedSquares := []string{} // The rook cannot move at all because of the double check in own king

	expected := squareToBitboard(expectedSquares)
	got := rook.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenTheRookIsPinned(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KING, "e8")
	pos.AddPiece(BLACK_ROOK, "e4")
	pos.AddPiece(WHITE_ROOK, "e2")
	pos.AddPiece(WHITE_ROOK, "e3")

	rook, _ := pos.PieceAt("e4")
	// The rook can only move along the e file, becuase it's pinned if moves
	// along the 4 rank, the king will be in check!
	expectedSquares := []string{"e3", "e5", "e6", "e7"}

	expected := squareToBitboard(expectedSquares)
	got := rook.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
