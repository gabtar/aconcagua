package board

import "testing"

// Pawn moves tests

func TestPawnAttacks(t *testing.T) {
	pos := EmptyPosition()
  pos.AddPiece(WHITE_PAWN, "e2")
	pawn, _ := pos.PieceAt("e2")

	expectedSquares := []string{"d3", "f3"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnAttacksOnEdgeFiles(t *testing.T) {
	pos := EmptyPosition()
  pos.AddPiece(WHITE_PAWN, "h2")
	pawn, _ := pos.PieceAt("h2")

	expectedSquares := []string{"g3"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Attacks(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesOnEmptyBoard(t *testing.T) {
	pos := EmptyPosition()
  pos.AddPiece(WHITE_PAWN, "e2")
	pawn, _ := pos.PieceAt("e2")

	expectedSquares := []string{"e3", "e4"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Moves(pos)

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

  // Can capture white bishop on a6 and is blocked by black knight on c6
  // Can also move to b6 and b7
	expectedSquares := []string{"a6", "b6", "b5"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Moves(pos)

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

  // The only legal move of the pawn is to block the check on f4
	expectedSquares := []string{"f4"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Moves(pos)

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

  // The only legal move of the pawn is to capture the bishop on e3
	expectedSquares := []string{"e3"}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Moves(pos)

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

	expectedSquares := []string{}

	expected := squareToBitboard(expectedSquares)
	got := pawn.Moves(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
