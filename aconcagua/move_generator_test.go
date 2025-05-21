package aconcagua

import "testing"

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

func TestCheckRestrictedSquares(t *testing.T) {
	pos := From("8/8/8/8/1k4PP/1bp1r2K/8/5N2 w - - 0 1")
	checkingSliders := pos.Bitboards[BlackRook] // Black Rook on e3
	checkingNonSliders := Bitboard(0)

	expected := checkRestrictedMoves(White, pos)
	got := checkRestrictedSquares(pos.KingPosition(White), checkingSliders, checkingNonSliders)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinRestrictedSquares(t *testing.T) {
	pos := From("2br2k1/5pp1/5p2/R4BP1/5PKP/8/8/8 w - - 0 1") // bishop on f5 is pinned can only move along the h4-c8 diagonal

	piece := pos.Bitboards[WhiteBishop]
	king := pos.KingPosition(White)
	pinnedPieces := pos.pinnedPieces(White)

	expected := pinRestrictedDirection(&piece, White, pos)
	got := pinRestrictedSquares(piece, king, pinnedPieces)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
