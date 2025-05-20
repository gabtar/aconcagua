package aconcagua

import "testing"

func TestDirectionsFromA1ToA8(t *testing.T) {
	expected := NORTH
	got := directions[0][56] // a1 to a8

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDirectionsFromA1ToH8(t *testing.T) {
	expected := NORTHEAST
	got := directions[0][63] // a1 to h8

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDirectionsFromE4ToD5(t *testing.T) {
	expected := NORTHWEST
	got := directions[28][35] // e4 to d5

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDirectionsFromE4ToC5(t *testing.T) {
	expected := INVALID
	got := directions[28][34] // e4 to c5

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestStraightPinnedPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhiteKing, "e1")
	pos.AddPiece(BlackRook, "e8")
	pos.AddPiece(WhitePawn, "e4")

	expected := bitboardFromCoordinate("e4")
	got := pos.pinnedPieces(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDiagonalPinnedPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e8")
	pos.AddPiece(BlackPawn, "d7")
	pos.AddPiece(BlackKnight, "f7")
	pos.AddPiece(WhiteQueen, "h5")
	pos.AddPiece(WhiteBishop, "a4")

	expected := bitboardFromCoordinate("f7") | bitboardFromCoordinate("d7")
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDiagonalNoPinnedPieces(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e8")
	pos.AddPiece(BlackPawn, "d7")
	pos.AddPiece(BlackPawn, "c6")
	pos.AddPiece(WhiteBishop, "a4")

	expected := Bitboard(0)
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPiecesOnBlack(t *testing.T) {
	pos := From("7k/6pn/6P1/3B4/7Q/7p/PPP4R/1K6 b - - 0 1")

	expected := Bitboard(1 << 55) // knight on h7
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
